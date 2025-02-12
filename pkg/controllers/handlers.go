package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

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

	if queryLimit := c.DefaultQuery("limit", "10"); queryLimit != "" {
		if parsedLimit, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsedLimit
		}
	}

	if queryOffset := c.DefaultQuery("offset", "0"); queryOffset != "" {
		if parsedOffset, err := strconv.Atoi(queryOffset); err == nil {
			offset = parsedOffset
		}
	}

	// Get books from Redis cache
	log.Info("Checking in cache for the books data")
	booksFromCache, err := cache.GetBooksFromCache()
	if err == nil && booksFromCache != nil && len(booksFromCache) > 0 {
		log.Info("Successfully fetched books data from the cache ")
		c.JSON(http.StatusOK, booksFromCache)
		return
	}

	log.Info("Books data is missing in the cache and fetching from postgres")
	// Otherwise, fetch from Postgres
	err = config.DB.Limit(limit).Offset(offset).Find(&books).Error
	if err != nil {
		log.Infof("Books not found in postgres")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching books"})
		return
	}
	log.Info("Successfully fetched the books data from postgres")
	// Store the books in cache
	cache.StoreBooksInCache(books)
	c.JSON(http.StatusOK, gin.H{
		"limit":  limit,
		"offset": offset,
		"books":  books,
	})
}

// GetBook handles the GET /books/{id} request
func GetBook(c *gin.Context) {
	id := c.Param("id")
	var book models.Book

	// Get book from Redis cache
	log.Infof("Checking in cache for the book data with id:%d", id)
	cachedBook, err := cache.GetBookFromCache(id)
	if err == nil && cachedBook != nil {
		log.Infof("Successfully fetched book data with id: %d from the cache ", id)
		c.JSON(http.StatusOK, cachedBook)
		return
	}

	// Otherwise, fetch from Postgres
	log.Infof("Book data with id: %d is missing in the cache and fetching from postgres", id)
	err = config.DB.First(&book, id).Error
	if err != nil {
		log.Infof("Book not found in postgres with id: %d", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	log.Infof("Successfully fetched the book data with id: %d from postgres", id)
	// Cache the book
	cache.StoreBookInCache(book)
	c.JSON(http.StatusOK, book)
}

// CreateBook handles the POST /books request
func CreateBook(c *gin.Context) {
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
			// Handle type mismatch
			errors := make(map[string]interface{})
			errorMessage := fmt.Sprintf("Invalid type for field '%s', expected %s", jsonErr.Field, jsonErr.Type)
			detailsMessage := fmt.Sprintf("Field '%s' should be of type '%s', but received '%s'", jsonErr.Field, jsonErr.Type, jsonErr.Value)
			errors["error"] = errorMessage
			errors["details"] = detailsMessage
			log.Errorf("Error in the request body: %v", errors)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   errorMessage,
				"details": detailsMessage,
			})
			return
		}

		// Handle validation errors
		var validationErrors []string
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, validationErr := range validationErrs {
				validationErrors = append(validationErrors, utils.FormatErrorMessage(validationErr))
			}
		}
		log.Errorf("Errors in validating the request body for creating book: %v and invalid input: %v", validationErrors, book)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": validationErrors,
		})
		return
	}

	// Save to Postgres
	err := config.DB.Create(&book).Error
	if err != nil {
		log.Errorf("Error in creating the book: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating book"})
		return
	}
	log.Infof("Successfully added the book with id: %d details to postgres", book.ID)

	//Save to cache
	cache.StoreBookInCache(book)
	c.JSON(http.StatusCreated, book)
}

// UpdateBook handles the PUT /books/{id} request
func UpdateBook(c *gin.Context) {
	id := c.Param("id")
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
			errors := make(map[string]interface{})
			errorMessage := fmt.Sprintf("Invalid type for field '%s', expected %s", jsonErr.Field, jsonErr.Type)
			detailsMessage := fmt.Sprintf("Field '%s' should be of type '%s', but received '%s'", jsonErr.Field, jsonErr.Type, jsonErr.Value)
			errors["error"] = errorMessage
			errors["details"] = detailsMessage
			log.Errorf("Error in updating the book due to request body: %v", id, errors)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   errorMessage,
				"details": detailsMessage,
			})
			return
		}

		// Handle validation errors
		var validationErrors []string
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, validationErr := range validationErrs {
				validationErrors = append(validationErrors, utils.FormatErrorMessage(validationErr))
			}
		}
		log.Errorf("Errors in validating the request body for updating book: %v and invalid input: %v", validationErrors, book)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": validationErrors,
		})
		return
	}

	var existingBook models.Book
	if err := config.DB.First(&existingBook, id).Error; err != nil {
		log.Errorf("Failed to find book with id: %d", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Update in Postgres
	book.ID = existingBook.ID
	err := config.DB.Model(&book).Where("id = ?", id).Updates(book).Error
	if err != nil {
		log.Errorf("Failed to updated the book with id: %d", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating book"})
		return
	}
	log.Infof("Successfully updated the book with id: %d in postgres", id)

	// Cache the updated book
	cache.StoreBookInCache(book)

	c.JSON(http.StatusOK, gin.H{
		"message": "Book updated successfully",
		"book":    book,
	})
}

// DeleteBook handles the DELETE /books/{id} request
func DeleteBook(c *gin.Context) {
	id := c.Param("id")

	// Delete from Postgres
	var existingBook models.Book
	if err := config.DB.First(&existingBook, id).Error; err != nil {
		// If the book is not found, return a 404 error
		log.Errorf("Book with id %d not found", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Delete the book from the database
	err := config.DB.Delete(&models.Book{}, id).Error
	if err != nil {
		log.Errorf("Error deleting the book with id: %d", id)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting book"})
		return
	}

	log.Infof("Successfully deleted the book with id: %d from postgres", id)
	// Remove from cache
	cache.DeleteBookFromCache(id)

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}
