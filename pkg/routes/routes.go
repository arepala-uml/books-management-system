package routes

import (
	"github.com/arepala-uml/books-management-system/pkg/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterBookStoreRoutes registers the API routes for the book management store
func RegisterBookStoreRoutes(r *gin.Engine) {
	r.GET("/books", controllers.GetBooks)
	r.GET("/books/:id", controllers.GetBook)
	r.POST("/books", controllers.CreateBook)
	r.PUT("/books/:id", controllers.UpdateBook)
	r.DELETE("/books/:id", controllers.DeleteBook)
}
