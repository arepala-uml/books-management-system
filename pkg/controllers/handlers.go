package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/arepala-uml/books-management-system/pkg/cache"
	"github.com/arepala-uml/books-management-system/pkg/config"
	"github.com/arepala-uml/books-management-system/pkg/kafka"
	"github.com/arepala-uml/books-management-system/pkg/models"
	"github.com/arepala-uml/books-management-system/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"

	"net/http"

	"github.com/gin-gonic/gin"
)

type BookListResponse struct {
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
	Books  []models.Book `json:"books"`
}

// ErrorResponse represents the structure of an error response
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Book    models.Book `json:"book,omitempty"`
}

// @Summary Get all books with optional pagination
// @Description Fetches all books, with pagination support using limit and offset query parameters
// @Param limit query int false "Limit the number of books per page" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} BookListResponse "List of books"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 500 {object} ErrorResponse "Error fetching books"
// @Router /books [get]
func GetBooks(c *gin.Context) {
	var books []models.Book
	limit := 10
	offset := 0
	log.Info("Got the request to fetch all the books")

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

// @Summary Get details of a single book by ID
// @Description Fetches the book data for a specific ID, first checking the cache, then the database
// @Param id path int true "Book ID"
// @Success 200 {object} models.Book "Book details"
// @Failure 404 {object} ErrorResponse "Book not found"
// @Failure 500 {object} ErrorResponse "Error fetching book"
// @Router /books/{id} [get]

func GetBook(c *gin.Context) {
	id := c.Param("id")
	log.Infof("Got the request to fectch details of book with id: %d", id)
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
// @Summary Create a new book
// @Description Adds a new book to the system
// @Accept json
// @Produce json
// @Param book body models.Book true "Book details"
// @Success 201 {object} SuccessResponse "Book created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 500 {object} ErrorResponse "Error creating book"
// @Router /books [post]
func CreateBook(c *gin.Context) {
	var book models.Book
	log.Info("Got the request to create a new book")
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

	// Publish the event to Kafka (book created)
	event := fmt.Sprintf("Book created: %s by %s", book.Title, book.Author)
	if err := kafka.PublishEvent("book_events", []byte(event)); err != nil {
		log.Errorf("Failed to publish event to Kafka: %v", err)
	}
	log.Infof("Successfully published an event to the kafka topic book_events about creating book with id:%d", book.ID)

	//Save to cache
	cache.StoreBookInCache(book)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Book created successfully",
		"book":    book,
	})
}

// @Summary Update an existing book
// @Description Updates the details of an existing book by ID
// @Param id path int true "Book ID"
// @Param book body models.Book true "Updated book details"
// @Success 200 {object} SuccessResponse "Book updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Book not found"
// @Failure 500 {object} ErrorResponse "Error updating book"
// @Router /books/{id} [put]

func UpdateBook(c *gin.Context) {
	id := c.Param("id")
	log.Infof("Got the request to update book with id: %d", id)
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

	// Publish the event to Kafka (book updated)
	event := fmt.Sprintf("Book updated: %s by %s", book.Title, book.Author)
	if err := kafka.PublishEvent("book_events", []byte(event)); err != nil {
		log.Errorf("Failed to publish event to Kafka: %v", err)
	}
	log.Infof("Successfully published an event to the kafka topic book_events about updaing book with id :%d", id)

	// Cache the updated book
	cache.StoreBookInCache(book)

	c.JSON(http.StatusOK, gin.H{
		"message": "Book updated successfully",
		"book":    book,
	})
}

// @Summary Delete a book by ID
// @Description Deletes a specific book from the system by its ID
// @Param id path int true "Book ID"
// @Success 200 {object} SuccessResponse "Book deleted successfully"
// @Failure 404 {object} ErrorResponse "Book not found"
// @Failure 500 {object} ErrorResponse "Error deleting book"
// @Router /books/{id} [delete]
func DeleteBook(c *gin.Context) {
	id := c.Param("id")
	log.Infof("Got the request to delete book with id: %d", id)

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

	// Publish the event to Kafka (book deleted)
	event := fmt.Sprintf("Book deleted with id: %s", id)
	if err := kafka.PublishEvent("book_events", []byte(event)); err != nil {
		log.Errorf("Failed to publish event to Kafka: %v", err)
	}
	log.Infof("Successfully published an event to the kafka topic book_events about deleting book with id :%d", id)

	// Remove from cache
	cache.DeleteBookFromCache(id)

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}
