package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/chitvanjain/polling-tool/internal/models"
	"github.com/chitvanjain/polling-tool/internal/utils"
)

type PollHandler struct {
	Polls *mongo.Collection
	Votes *mongo.Collection
	Redis *redis.Client
}

type createPollRequest struct {
	Title           string   `json:"title" binding:"required,min=3,max=200"`
	Options         []string `json:"options" binding:"required,min=2,max=10,dive,required,min=1,max=100"`
	DurationMinutes int      `json:"duration_minutes" binding:"omitempty,min=5,max=129600"`
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	var req createPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, _ := c.Get("userID")
	creatorID, err := primitive.ObjectIDFromHex(userIDValue.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	options := make([]models.PollOption, len(req.Options))
	for i, text := range req.Options {
		options[i] = models.PollOption{ID: fmt.Sprintf("opt_%d", i+1), Text: text}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var slug string
	for attempt := 0; attempt < 5; attempt++ {
		candidate, err := utils.GenerateSlug(8)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate share link"})
			return
		}
		count, err := h.Polls.CountDocuments(ctx, bson.M{"share_slug": candidate})
		if err == nil && count == 0 {
			slug = candidate
			break
		}
	}
	if slug == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate a unique share link, please retry"})
		return
	}

	var expiresAt *time.Time
if req.DurationMinutes > 0 {
	t := time.Now().Add(time.Duration(req.DurationMinutes) * time.Minute)
	expiresAt = &t
}

poll := models.Poll{
	CreatorID: creatorID,
	Title:     req.Title,
	Options:   options,
	ShareSlug: slug,
	Status:    models.PollStatusOpen,
	ExpiresAt: expiresAt,
	CreatedAt: time.Now(),
}

	res, err := h.Polls.InsertOne(ctx, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create poll"})
		return
	}
	poll.ID = res.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, poll)
}

func (h *PollHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var poll models.Poll
	if err := h.Polls.FindOne(ctx, bson.M{"share_slug": slug}).Decode(&poll); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	applyExpiry(ctx, h.Polls, &poll)

	hasVoted := false
	if voterID, err := c.Cookie("voter_id"); err == nil && voterID != "" {
		votersKey := "poll:" + poll.ID.Hex() + ":voters"
		if isMember, err := h.Redis.SIsMember(ctx, votersKey, voterID).Result(); err == nil {
			hasVoted = isMember
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         poll.ID,
		"title":      poll.Title,
		"options":    poll.Options,
		"status":     poll.Status,
		"share_slug": poll.ShareSlug,
		"expires_at": poll.ExpiresAt,
		"has_voted":  hasVoted,
	})
}

func (h *PollHandler) ListMine(c *gin.Context) {
	userIDValue, _ := c.Get("userID")
	creatorID, err := primitive.ObjectIDFromHex(userIDValue.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cursor, err := h.Polls.Find(
		ctx,
		bson.M{"creator_id": creatorID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load polls"})
		return
	}
	defer cursor.Close(ctx)

	type pollSummary struct {
		ID         primitive.ObjectID `json:"id"`
		Title      string             `json:"title"`
		ShareSlug  string             `json:"share_slug"`
		Status     string             `json:"status"`
		ExpiresAt  *time.Time         `json:"expires_at,omitempty"`
		CreatedAt  time.Time          `json:"created_at"`
		TotalVotes int                `json:"total_votes"`
	}

	summaries := []pollSummary{}
	for cursor.Next(ctx) {
		var poll models.Poll
		if err := cursor.Decode(&poll); err != nil {
			continue
		}

		applyExpiry(ctx, h.Polls, &poll)

		total := 0
		countsKey := "poll:" + poll.ID.Hex() + ":counts"
		if rawCounts, err := h.Redis.HGetAll(ctx, countsKey).Result(); err == nil {
			for _, v := range rawCounts {
				n := 0
				fmt.Sscanf(v, "%d", &n)
				total += n
			}
		}

		summaries = append(summaries, pollSummary{
			ID:         poll.ID,
			Title:      poll.Title,
			ShareSlug:  poll.ShareSlug,
			Status:     poll.Status,
			ExpiresAt:  poll.ExpiresAt,
			CreatedAt:  poll.CreatedAt,
			TotalVotes: total,
		})
	}

	c.JSON(http.StatusOK, summaries)
}

func (h *PollHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	pollID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	userIDValue, _ := c.Get("userID")
	creatorID, err := primitive.ObjectIDFromHex(userIDValue.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	res, err := h.Polls.DeleteOne(ctx, bson.M{"_id": pollID, "creator_id": creatorID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete poll"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	h.Redis.Del(ctx, "poll:"+pollID.Hex()+":counts", "poll:"+pollID.Hex()+":voters")
	if h.Votes != nil {
		_, _ = h.Votes.DeleteMany(ctx, bson.M{"poll_id": pollID})
	}

	c.Status(http.StatusNoContent)
}