package handlers

import (
	"fmt"
	"net/http"
	"time"
"encoding/json"
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

	if initial, err := h.resultsPayload(ctx, poll); err == nil {
		if data, err := jsonMarshal(initial); err == nil {
			writeEvent(data)
		}
	}

	channel := "poll:" + poll.ID.Hex() + ":updates"
	sub := h.Redis.Subscribe(ctx, channel)
	defer sub.Close()

	msgs := sub.Channel()
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			writeEvent([]byte(msg.Payload))
		case <-heartbeat.C:
			fmt.Fprint(c.Writer, ": keepalive\n\n")
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}