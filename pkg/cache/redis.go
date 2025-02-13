package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arepala-uml/books-management-system/pkg/config"
	"github.com/arepala-uml/books-management-system/pkg/models"
	"github.com/arepala-uml/books-management-system/pkg/utils"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func GetBookFromCache(id string) (*models.Book, error) {
	// Construct the Redis key in the format "BOOKS_ID:<ID_NUMBER>"
	redisKey := fmt.Sprintf("BOOKS_ID:%s", id)
	bookData, err := utils.ReJSONGet(redisKey, ".")
	if err != nil {
		log.Errorf("Error getting book from Redis: %v", err)
	}

	var book models.Book
	if dataMap, ok := bookData.(map[string]interface{}); ok {

		bookBytes, err := json.Marshal(dataMap)
		if err != nil {
			log.Printf("Error marshaling book data to JSON: %v", err)
			return nil, err
		}

		err = json.Unmarshal(bookBytes, &book)
		if err != nil {
			log.Printf("Error unmarshaling book data: %v", err)
			return nil, err
		}
	} else {
		log.Error("Retrieved data is not a valid book format")
		return nil, err
	}
	return &book, nil
}

func GetBooksFromCache() ([]models.Book, error) {
	books := make([]models.Book, 0)

	iter := config.RedisClient.Scan(ctx, 0, "BOOKS_ID:*", 0).Iterator()
	for iter.Next(ctx) {
		redisKey := iter.Val()
		bookData, err := utils.ReJSONGet(redisKey, ".")
		if err == redis.Nil {
			continue
		} else if err != nil {
			log.Errorf("Error fetching book data from Redis for key %s: %v", redisKey, err)
			continue
		}

		if dataMap, ok := bookData.(map[string]interface{}); ok {
			bookBytes, err := json.Marshal(dataMap)
			if err != nil {
				log.Errorf("Error marshaling book data to JSON: %v", err)
				continue
			}
			var book models.Book
			err = json.Unmarshal(bookBytes, &book)
			if err != nil {
				log.Errorf("Error unmarshaling book data: %v", err)
				continue
			}
			books = append(books, book)
		} else {
			log.Printf("Retrieved data for key %s is not a valid book format", redisKey)
		}
	}

	if err := iter.Err(); err != nil {
		log.Printf("Error iterating over Redis keys: %v", err)
		return nil, err
	}

	return books, nil
}

func StoreBookInCache(book models.Book) error {
	redisKey := fmt.Sprintf("BOOKS_ID:%d", book.ID)
	err := utils.ReJSONSet(redisKey, ".", book, viper.GetInt("REDIS_EXPIRY_BOOKS"))
	if err != nil {
		log.Errorf("Failed to set data for the key - %s, %v", redisKey, err)
		return err
	}
	log.Printf("Book cached successfully with ID: %d", book.ID)
	return nil
}

func StoreBooksInCache(books []models.Book) error {
	for _, book := range books {
		err := StoreBookInCache(book)
		if err != nil {
			log.Printf("Error storing book %d in cache: %v", book.ID, err)
			return err
		}
	}
	log.Info("All books cached successfully")
	return nil
}

func DeleteBookFromCache(id string) error {
	redisKey := fmt.Sprintf("BOOKS_ID:%s", id)
	if utils.RedisKeyExists(redisKey) {
		log.Infof("Deleting Book with redis key from redis %v", redisKey)
		err := utils.ReJSONDel(redisKey, ".")
		if err != nil {
			log.Printf("Error deleting book from cache: %v", err)
			return err
		}
	}
	log.Infof("Book removed from cache with ID: %s", id)
	return nil
}
