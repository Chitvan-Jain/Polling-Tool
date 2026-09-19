package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/chitvanjain/polling-tool/internal/auth"
	"github.com/chitvanjain/polling-tool/internal/config"
	"github.com/chitvanjain/polling-tool/internal/db"
	"github.com/chitvanjain/polling-tool/internal/handlers"
)

func main() {
	cfg := config.Load()

	mongoClient := db.ConnectMongo(cfg.MongoURI)
	redisClient := db.ConnectRedis(cfg.RedisAddr)

	database := mongoClient.Database("pollingtool")
	usersCollection := database.Collection("users")
	pollsCollection := database.Collection("polls")
	votesCollection := database.Collection("votes")

	db.EnsurePollIndexes(pollsCollection)
	db.EnsureVoteIndexes(votesCollection)

	authHandler := &handlers.AuthHandler{Users: usersCollection, JWTSecret: cfg.JWTSecret}
	pollHandler := &handlers.PollHandler{Polls: pollsCollection}
	voteHandler := &handlers.VoteHandler{Polls: pollsCollection, Votes: votesCollection, Redis: redisClient}

	router := gin.Default()

router.Use(cors.New(cors.Config{
	AllowOrigins:     []string{"http://localhost:5173"},
	AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
	AllowHeaders:     []string{"Content-Type", "Authorization"},
	AllowCredentials: true,
	MaxAge:           12 * time.Hour,
}))

	router.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		status := gin.H{"status": "ok"}
		code := http.StatusOK

		if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
			status["mongo"] = "down"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		} else {
			status["mongo"] = "up"
		}

		if _, err := redisClient.Ping(ctx).Result(); err != nil {
			status["redis"] = "down"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		} else {
			status["redis"] = "up"
		}

		c.JSON(code, status)
	})

	api := router.Group("/api")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/signup", authHandler.Signup)
			authGroup.POST("/login", authHandler.Login)
		}

		pollsGroup := api.Group("/polls")
		pollsGroup.Use(auth.RequireAuth(cfg.JWTSecret))
		{
			pollsGroup.POST("", pollHandler.CreatePoll)
		}

		publicGroup := api.Group("/p")
		{
			publicGroup.GET("/:slug", pollHandler.GetBySlug)
			publicGroup.POST("/:slug/vote", voteHandler.SubmitVote)
			publicGroup.GET("/:slug/results", voteHandler.GetResults)
			publicGroup.GET("/:slug/live", voteHandler.LiveResults)
		}
	}

	log.Printf("server starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}