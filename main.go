package main

import (
	signupapiv1 "achievesomethingbro/appapi"
	dbpg "achievesomethingbro/appdb"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

// ---------------- API HANDLERS ----------------

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

	// Load env (safe)
	_ = godotenv.Load()

	// Init DB & ES
	dbpg.IntializeDB()
	dbpg.InitElasticsearch()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://janasi8.github.io",
			"http://localhost:8080",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ---------------- 🔥 ADD CORS HERE ----------------

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://janasi8.github.io", // frontend domain
			"http://localhost:3000",     // optional local testing
			"http://localhost:8080",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ---------------- LOAD HTML TEMPLATES ----------------

	router.LoadHTMLGlob("templates/*")

	// Static files
	router.Static("/static", "./static")

	// ---------------- UI ROUTES ----------------

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", nil)
	})

	// ---------------- API ROUTES ----------------

	api := router.Group("/api")
	{
		api.GET("/todos", getTodos)
		api.GET("/todos/:id", getTodoById)
		api.POST("/todos", appendTodos)
	}

	// ---------------- AUTH / OTHER ROUTES ----------------

	signupapiv1.InitializeAPI(router)

	// ---------------- FALLBACK ----------------

	router.NoRoute(func(c *gin.Context) {

		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API route not found",
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", nil)
	})

	// ---------------- SERVER ----------------

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🚀 Server running on port " + port)
	router.Run(":" + port)
}
