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
	RedisAddr := viper.GetString("REDIS_HOST") + ":" + viper.GetString("REDIS_PORT")
	RedisPassword := viper.GetString("REDIS_PASSWORD")
	RedisDB := viper.GetInt("REDIS_DB")

	log.Infof("Redis is running at %s", RedisAddr)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         RedisAddr,
		Password:     RedisPassword,
		DB:           RedisDB,
		WriteTimeout: time.Duration(120) * time.Second,
		DialTimeout:  time.Duration(10) * time.Second,
		ReadTimeout:  time.Duration(60) * time.Second,
	})
	ReJSONHandler = rejson.NewReJSONHandler()
	ReJSONHandler.SetGoRedisClient(RedisClient)

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Info("Successfully connected to Redis with RedisJSON!")
}

func initPostgres() {
	databaseName := viper.GetString("POSTGRES_DB")
	databaseUser := viper.GetString("POSTGRES_USER")
	databasePassword := viper.GetString("POSTGRES_PASSWORD")
	databaseHost := viper.GetString("POSTGRES_HOST")
	databasePort := viper.GetString("POSTGRES_PORT")

	//PostgreSQL connection string
	connectionLink := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		databaseUser, databasePassword, databaseHost, databasePort, databaseName)

	log.Info("PostgreSQL Connection URL: ", connectionLink)
	d, err := gorm.Open(postgres.Open(connectionLink), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	DB = d

	log.Info("Successfully connected to PostgreSQL (Local)!")
}

func GetDB() *gorm.DB {
	return DB
}

func GetRedisClient() *redis.Client {
	return RedisClient
}
