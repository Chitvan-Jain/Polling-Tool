package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/chitvanjain/polling-tool/internal/models"
)

func (h *VoteHandler) LiveResults(c *gin.Context) {
	slug := c.Param("slug")
	ctx := c.Request.Context()

	var poll models.Poll
	if err := h.Polls.FindOne(ctx, bson.M{"share_slug": slug}).Decode(&poll); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	applyExpiry(ctx, h.Polls, &poll)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	writeEvent := func(data []byte) {
		fmt.Fprintf(c.Writer, "event: results\ndata: %s\n\n", data)
		flusher.Flush()
	}

	sendCurrentState := func() {
		var fresh models.Poll
		if err := h.Polls.FindOne(ctx, bson.M{"_id": poll.ID}).Decode(&fresh); err == nil {
			applyExpiry(ctx, h.Polls, &fresh)
			poll = fresh
		}

		hasVoted := h.hasVoterVoted(ctx, c, poll)
		visible := canSeeResults(poll, hasVoted)

		if payload, err := h.resultsPayload(ctx, poll, visible); err == nil {
			if data, err := json.Marshal(payload); err == nil {
				writeEvent(data)
			}
		}
	}

	sendCurrentState()

	channel := "poll:" + poll.ID.Hex() + ":updates"
	sub := h.Redis.Subscribe(ctx, channel)
	defer sub.Close()

	msgs := sub.Channel()
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case _, ok := <-msgs:
			if !ok {
				return
			}
			sendCurrentState()
		case <-heartbeat.C:
			fmt.Fprint(c.Writer, ": keepalive\n\n")
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}