package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI  string
	RedisAddr string
	Port      string
	JWTSecret string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		MongoURI:  os.Getenv("MONGO_URI"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
		Port:      port,
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
}