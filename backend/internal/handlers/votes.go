package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/chitvanjain/polling-tool/internal/models"
)

type VoteHandler struct {
	Polls *mongo.Collection
	Votes *mongo.Collection
	Redis *redis.Client
}

type voteRequest struct {
	OptionID string `json:"option_id" binding:"required"`
}

type optionResult struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

func voterIDCookie(c *gin.Context) string {
	if id, err := c.Cookie("voter_id"); err == nil && id != "" {
		return id
	}
	newID := uuid.NewString()
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("voter_id", newID, 60*60*24*365, "/", "", false, true)
	return newID
}

func (h *VoteHandler) resultsPayload(ctx context.Context, poll models.Poll) (gin.H, error) {
	countsKey := "poll:" + poll.ID.Hex() + ":counts"
	rawCounts, err := h.Redis.HGetAll(ctx, countsKey).Result()
	if err != nil {
		return nil, err
	}

	results := make([]optionResult, 0, len(poll.Options))
	for _, opt := range poll.Options {
		count := 0
		if v, ok := rawCounts[opt.ID]; ok {
			fmt.Sscanf(v, "%d", &count)
		}
		results = append(results, optionResult{ID: opt.ID, Text: opt.Text, Count: count})
	}

	return gin.H{
		"title":   poll.Title,
		"status":  poll.Status,
		"results": results,
	}, nil
}

func (h *VoteHandler) SubmitVote(c *gin.Context) {
	slug := c.Param("slug")

	var req voteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var poll models.Poll
	if err := h.Polls.FindOne(ctx, bson.M{"share_slug": slug}).Decode(&poll); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	if poll.Status != models.PollStatusOpen {
		c.JSON(http.StatusConflict, gin.H{"error": "this poll is closed"})
		return
	}

	validOption := false
	for _, opt := range poll.Options {
		if opt.ID == req.OptionID {
			validOption = true
			break
		}
	}
	if !validOption {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option for this poll"})
		return
	}

	voterID := voterIDCookie(c)
	votersKey := "poll:" + poll.ID.Hex() + ":voters"

	added, err := h.Redis.SAdd(ctx, votersKey, voterID).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not record vote"})
		return
	}
	if added == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "you have already voted on this poll"})
		return
	}

	countsKey := "poll:" + poll.ID.Hex() + ":counts"
	newCount, err := h.Redis.HIncrBy(ctx, countsKey, req.OptionID, 1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not record vote"})
		return
	}

	vote := models.Vote{
		PollID:    poll.ID,
		OptionID:  req.OptionID,
		VoterID:   voterID,
		CreatedAt: time.Now(),
	}
	if _, err := h.Votes.InsertOne(ctx, vote); err != nil {
		log.Printf("warning: vote counted in redis but failed to persist to mongo: %v", err)
	}

	if payload, err := h.resultsPayload(ctx, poll); err == nil {
		if data, err := json.Marshal(payload); err == nil {
			channel := "poll:" + poll.ID.Hex() + ":updates"
			h.Redis.Publish(ctx, channel, data)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"option_id": req.OptionID,
		"count":     newCount,
	})
}

func (h *VoteHandler) GetResults(c *gin.Context) {
	slug := c.Param("slug")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var poll models.Poll
	if err := h.Polls.FindOne(ctx, bson.M{"share_slug": slug}).Decode(&poll); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	payload, err := h.resultsPayload(ctx, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load results"})
		return
	}

	c.JSON(http.StatusOK, payload)
}