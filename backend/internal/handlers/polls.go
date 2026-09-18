package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/chitvanjain/polling-tool/internal/models"
	"github.com/chitvanjain/polling-tool/internal/utils"
)

type PollHandler struct {
	Polls *mongo.Collection
}

type createPollRequest struct {
	Title   string   `json:"title" binding:"required,min=3,max=200"`
	Options []string `json:"options" binding:"required,min=2,max=10,dive,required,min=1,max=100"`
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

	poll := models.Poll{
		CreatorID: creatorID,
		Title:     req.Title,
		Options:   options,
		ShareSlug: slug,
		Status:    models.PollStatusOpen,
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

	c.JSON(http.StatusOK, poll)
}