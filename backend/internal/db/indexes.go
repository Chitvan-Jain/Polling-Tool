package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsurePollIndexes(polls *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := polls.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "share_slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatalf("could not create poll indexes: %v", err)
	}
}
func EnsureVoteIndexes(votes *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := votes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "poll_id", Value: 1}, {Key: "voter_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatalf("could not create vote indexes: %v", err)
	}
}