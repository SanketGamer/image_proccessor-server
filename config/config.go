package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisAddr string

	JWTSecret      string
	JWTExpiryHours int

	AWSRegion string
	AWSKeyID  string
	AWSSecret string
	AWSBucket string

	SentryDSN     string
	SentryEnv     string
	SentryRelease string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
	hours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))

	return &Config{
		Port:           os.Getenv("PORT"),
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBSSLMode:      os.Getenv("DB_SSL"),
		RedisAddr:      os.Getenv("REDIS_ADDR"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiryHours: hours,
		AWSRegion:      os.Getenv("AWS_REGION"),
		AWSKeyID:       os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecret:      os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSBucket:      os.Getenv("AWS_BUCKET_NAME"),
		SentryDSN:      os.Getenv("SENTRY_DSN"),
		SentryEnv:      os.Getenv("SENTRY_ENV"),
		SentryRelease:  os.Getenv("SENTRY_RELEASE"),
	}
}
