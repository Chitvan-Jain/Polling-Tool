package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	mongoURI := os.Getenv("MONGO_URI")
	redisAddr := os.Getenv("REDIS_ADDR")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("mongo ping failed: %v", err)
	}
	log.Println("connected to MongoDB")

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}
	log.Println("connected to Redis")

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		healthCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		status := gin.H{"status": "ok"}
		code := http.StatusOK

		if err := mongoClient.Ping(healthCtx, readpref.Primary()); err != nil {
			status["mongo"] = "down"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		} else {
			status["mongo"] = "up"
		}

		if _, err := redisClient.Ping(healthCtx).Result(); err != nil {
			status["redis"] = "down"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		} else {
			status["redis"] = "up"
		}

		c.JSON(code, status)
	})

	log.Printf("server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}