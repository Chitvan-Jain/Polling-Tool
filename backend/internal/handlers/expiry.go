package handlers

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/chitvanjain/polling-tool/internal/models"
)

func applyExpiry(ctx context.Context, polls *mongo.Collection, poll *models.Poll) {
	if poll.Status == models.PollStatusOpen && poll.ExpiresAt != nil && time.Now().After(*poll.ExpiresAt) {
		_, _ = polls.UpdateOne(ctx, bson.M{"_id": poll.ID}, bson.M{"$set": bson.M{"status": models.PollStatusClosed}})
		poll.Status = models.PollStatusClosed
	}
}