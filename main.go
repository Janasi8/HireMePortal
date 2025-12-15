package main

import (
	signupapiv1 "achievesomethingbro/appapi"
	dbpg "achievesomethingbro/appdb"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ---------------- SAMPLE DEMO DATA ----------------

type todo struct {
	Id        string `json:"id"`
	BookName  string `json:"bookname"`
	Completed bool   `json:"completed"`
}

var todos = []todo{
	{Id: "1", BookName: "Punam padhai karle", Completed: false},
	{Id: "2", BookName: "Punam ko abcd sikhao", Completed: true},
	{Id: "3", BookName: "Punam ko time se uthao", Completed: false},
}

// ---------------- DEMO API HANDLERS ----------------

func getTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}

func getTodoById(c *gin.Context) {
	id := c.Param("id")
	for _, a := range todos {
		if a.Id == id {
			c.JSON(http.StatusOK, a)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "todo not found"})
}

func appendTodos(c *gin.Context) {
	var newTodo todo
	if err := c.BindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}
	todos = append(todos, newTodo)
	c.JSON(http.StatusCreated, newTodo)
}

// ---------------- MAIN ----------------

func main() {

	// Initialize DB & Elasticsearch (keep as-is)
	dbpg.IntializeDB()
	dbpg.InitElasticsearch()

	router := gin.Default()

	// SAFELY load templates (prevents Render crash)
	if _, err := os.Stat("templates"); err == nil {
		router.LoadHTMLGlob("templates/*")
	}

	// Home route
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "HireMePortal API is running",
		})
	})

	// Demo APIs
	router.GET("/todos", getTodos)
	router.GET("/todos/:id", getTodoById)
	router.POST("/todos", appendTodos)

	// Register existing APIs (login/signup/etc.)
	signupapiv1.InitializeAPI()

	// Render-compatible dynamic PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}
