package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/arepala-uml/books-management-system/pkg/cache"
	"github.com/arepala-uml/books-management-system/pkg/config"
	"github.com/arepala-uml/books-management-system/pkg/models"
	"github.com/arepala-uml/books-management-system/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"

	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBooks handles the GET /books request
func GetBooks(c *gin.Context) {
	var books []models.Book
	limit := 10
	offset := 0

	// Get books from Redis cache
	log.Info("Checking in cache for the books data")
	booksFromCache, err := cache.GetBooksFromCache()
	if err == nil && booksFromCache != nil {
		c.JSON(http.StatusOK, booksFromCache)
		return
	}

	log.Info("Books data is missing in the cache and fetching from postgres")
	// Otherwise, fetch from DB
	err = models.DB.Limit(limit).Offset(offset).Find(&books).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching books"})
		return
	}

	// Store the books in cache for future requests
	cache.StoreBooksInCache(books)

	c.JSON(http.StatusOK, books)
}

// GetBook handles the GET /books/{id} request
func GetBook(c *gin.Context) {
	id := c.Param("id")
	var book models.Book

	// Check Redis first
	log.Infof("Checking in cache for the book data with id:%d", id)
	cachedBook, err := cache.GetBookFromCache(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching from cache"})
		return
	}

	if cachedBook != nil {
		c.JSON(http.StatusOK, cachedBook)
		return
	}

	// If not in cache, fetch from DB
	log.Infof("Book data with id: %d is missing in the cache and fetching from postgres", id)
	err = config.DB.First(&book, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Cache the book
	//storeBookInCache(book)
	cache.StoreBookInCache(book)

	c.JSON(http.StatusOK, book)
}

// CreateBook handles the POST /books request
func CreateBook(c *gin.Context) {
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		// Check if the error is from JSON unmarshaling (type mismatch)
		if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
			// Handle type mismatch for specific fields
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Invalid type for field '%s', expected %s", jsonErr.Field, jsonErr.Type),
				"details": fmt.Sprintf("Field '%s' should be of type '%s', but received '%s'", jsonErr.Field, jsonErr.Type, jsonErr.Value),
			})
			return
		}

		// Handle validation errors (e.g., missing required fields)
		var validationErrors []string
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, validationErr := range validationErrs {
				// Build detailed error messages for required fields and other constraints
				validationErrors = append(validationErrors, utils.FormatErrorMessage(validationErr))
			}
		}
		// Return validation errors to the user
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": validationErrors,
		})
		return
	}

	// Save to DB
	err := config.DB.Create(&book).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating book"})
		return
	}
	cache.StoreBookInCache(book)

	c.JSON(http.StatusCreated, book)
}

// UpdateBook handles the PUT /books/{id} request
func UpdateBook(c *gin.Context) {
	id := c.Param("id")
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update in DB
	err := config.DB.Model(&book).Where("id = ?", id).Updates(book).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating book"})
		return
	}

	// Cache the updated book
	cache.StoreBookInCache(book)

	c.JSON(http.StatusOK, book)
}

// DeleteBook handles the DELETE /books/{id} request
func DeleteBook(c *gin.Context) {
	id := c.Param("id")

	// Delete from DB
	err := config.DB.Delete(&models.Book{}, id).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting book"})
		return
	}

	// Remove from cache
	cache.DeleteBookFromCache(id)

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}
