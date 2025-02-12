package cache

import (
	"context"
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

	// If the book is found in cache, deserialize it into a book object
	book := bookData.(models.Book)
	return &book, nil
}

func GetBooksFromCache() ([]models.Book, error) {
	books := make([]models.Book, 0)

	// Use Redis to scan all keys matching the pattern "BOOKS_ID:*"
	iter := config.RedisClient.Scan(ctx, 0, "BOOKS_ID:*", 0).Iterator()
	for iter.Next(ctx) {
		// Construct the Redis key in the format "BOOKS_ID:<ID_NUMBER>"
		redisKey := iter.Val()

		// Get the book data from Redis using JSON.GET (RedisJSON)
		bookData, err := utils.ReJSONGet(redisKey, ".")
		if err == redis.Nil {
			// If the key does not exist in Redis, continue to the next iteration
			continue
		} else if err != nil {
			log.Printf("Error fetching book data from Redis for key %s: %v", redisKey, err)
			continue
		}

		book := bookData.(models.Book)

		// Append the book to the books slice
		books = append(books, book)
	}

	// Check for errors during the iteration
	if err := iter.Err(); err != nil {
		log.Printf("Error iterating over Redis keys: %v", err)
		return nil, err
	}

	return books, nil
}

// storeBookInCache stores a book object in Redis cache by its ID
func StoreBookInCache(book models.Book) error {
	// Serialize the book object to JSON before storing it in Redis
	// bookData, err := json.Marshal(book)
	// if err != nil {
	// 	log.Printf("Error marshaling book to JSON: %v", err)
	// 	return err
	// }

	// Set the book in Redis cache as JSON using RedisJSON (JSON.SET)
	redisKey := fmt.Sprintf("BOOKS_ID:%d", book.ID)
	err := utils.ReJSONSet(redisKey, ".", book, viper.GetInt("REDIS_EXPIRY_BOOKS"))
	if err != nil {
		log.Errorf("Failed to set data for the key - %s, %v", redisKey, err)
		return err
	}
	log.Printf("Book cached successfully with ID: %d", book.ID)
	return nil
}

// StoreBooksInCache stores multiple books in Redis cache
func StoreBooksInCache(books []models.Book) error {
	for _, book := range books {
		// Cache each book individually using StoreBookInCache
		err := StoreBookInCache(book)
		if err != nil {
			log.Printf("Error storing book %d in cache: %v", book.ID, err)
			return err
		}
	}
	log.Info("All books cached successfully")
	return nil
}

// DeleteBookFromCache removes a book from the Redis cache by its ID
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
