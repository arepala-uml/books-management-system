package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/arepala-uml/books-management-system/docs"

	"github.com/arepala-uml/books-management-system/pkg/config"
	"github.com/arepala-uml/books-management-system/pkg/kafka"
	"github.com/arepala-uml/books-management-system/pkg/models"
	"github.com/arepala-uml/books-management-system/pkg/routes"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

func InitConfig() {
	// Get the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Build the full path to app.env dynamically
	configPath := filepath.Join(pwd, "app.env")
	log.Info(configPath)

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)
	})

	viper.Set("PWD", pwd)
}

func init() {
	InitConfig()
	config.Connect()
	models.DB = config.GetDB()
	// Auto-migrate the Book model to keep the database schema updated
	models.DB.AutoMigrate(&models.Book{})
}

func main() {
	fmt.Println("Hi")
	r := gin.Default()

	go func() {
		brokerList := []string{fmt.Sprintf("%s:%s", viper.GetString("KAFKA_HOST"), viper.GetString("KAFKA_PORT"))}
		topic := viper.GetString("KAFKA_TOPIC")
		err := kafka.StartConsumer(brokerList, topic)
		if err != nil {
			log.Errorf("Error in consumer: %v", err)
		}
		log.Info("Kafka consumer started")
	}()

	// Register the routes for the Book Store API
	routes.RegisterBookStoreRoutes(r)

	// Swagger UI endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	hostname := viper.GetString("SERVER_HOST") + ":" + viper.GetString("SERVER_PORT")
	log.Info("Server running on ", hostname)

	log.Fatal(http.ListenAndServe(hostname, r))
}
