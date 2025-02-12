package config

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/gommon/log"
	"github.com/nitishm/go-rejson/v4"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB            *gorm.DB
	RedisClient   *redis.Client
	ReJSONHandler *rejson.Handler
	ctx           = context.Background()
	RedisAddr     string
)

func Connect() {
	// Initialize Postgres and Redis connection
	initPostgres()
	initRedis()
}

func initRedis() {
	// Get Redis connection details from environment variables
	RedisAddr := viper.GetString("REDIS_HOST") + ":" + viper.GetString("REDIS_PORT")
	RedisPassword := viper.GetString("REDIS_PASSWORD") // If no password, use an empty string
	RedisDB := viper.GetInt("REDIS_DB")                // Default Redis DB is 0

	// Logging Redis connection details
	log.Infof("Redis is running at %s", RedisAddr)

	// Initialize the Redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         RedisAddr,
		Password:     RedisPassword, // Provide password if applicable
		DB:           RedisDB,       // Select the Redis DB (default is 0)
		WriteTimeout: time.Duration(120) * time.Second,
		DialTimeout:  time.Duration(10) * time.Second,
		ReadTimeout:  time.Duration(60) * time.Second,
	})

	// Initialize RedisJSON handler
	ReJSONHandler = rejson.NewReJSONHandler()

	// Set the Redis client for RedisJSON
	//ReJSONHandler.SetGoRedisClientWithContext(ctx, RedisClient)
	ReJSONHandler.SetGoRedisClient(RedisClient)

	// Test the connection to Redis
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Successfully connected to Redis and RedisJSON
	log.Info("Successfully connected to Redis with RedisJSON!")
}

func initPostgres() {
	// Fetch PostgreSQL environment variables
	databaseName := viper.GetString("POSTGRES_DB")
	databaseUser := viper.GetString("POSTGRES_USER")
	databasePassword := viper.GetString("POSTGRES_PASSWORD")
	databaseHost := viper.GetString("POSTGRES_HOST")
	databasePort := viper.GetString("POSTGRES_PORT")

	// Construct the PostgreSQL connection string
	connectionLink := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		databaseUser, databasePassword, databaseHost, databasePort, databaseName)

	log.Info("PostgreSQL Connection URL: ", connectionLink)

	// Initialize the PostgreSQL connection
	d, err := gorm.Open(postgres.Open(connectionLink), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	DB = d

	log.Info("Successfully connected to PostgreSQL (Local)!")
}

// func initPostgres() {
// 	databaseName := viper.GetString("NEON_DB_NAME")
// 	databaseUser := viper.GetString("NEON_DB_USER")
// 	databasePassword := viper.GetString("NEON_DB_PASSWORD")
// 	databaseHost := viper.GetString("NEON_DB_HOST")
// 	databasePort := viper.GetString("NEON_DB_PORT")

// 	connectionLink := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require", databaseUser, databasePassword, databaseHost, databasePort, databaseName)

// 	log.Info(connectionLink)
// 	d, err := gorm.Open(postgres.Open(connectionLink), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("Failed to connect to the database:", err)
// 	}
// 	DB = d
// 	log.Info("Succesfully established the connection to mysql database")
// }

func GetDB() *gorm.DB {
	return DB
}

func GetRedisClient() *redis.Client {
	return RedisClient
}
