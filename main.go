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

	// DB & Elasticsearch intentionally kept in DEMO MODE
	dbpg.IntializeDB()
	dbpg.InitElasticsearch()

	router := gin.Default()

	// Serve HTML templates
	router.LoadHTMLGlob("templates/*")

	// Home page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Demo APIs
	router.GET("/todos", getTodos)
	router.GET("/todos/:id", getTodoById)
	router.POST("/todos", appendTodos)

	// Register your existing APIs (login/signup/etc)
	signupapiv1.InitializeAPI()

	// Dynamic PORT (required for deployment)
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	router.Run(":" + port)
}
