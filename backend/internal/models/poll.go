package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PollStatusOpen   = "open"
	PollStatusClosed = "closed"
)

type PollOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatorID primitive.ObjectID `bson:"creator_id" json:"creator_id"`
	Title     string             `bson:"title" json:"title"`
	Options   []PollOption       `bson:"options" json:"options"`
	ShareSlug string             `bson:"share_slug" json:"share_slug"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}