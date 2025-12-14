package main

import (
	signupapiv1 "achievesomethingbro/appapi"
	dbpg "achievesomethingbro/appdb"
	model "achievesomethingbro/appmodel"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	host = "localhost"
	port = 5000
)

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

func getTodos(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, todos)
}
func getTodoById(c *gin.Context) {
	id := c.Param("id")
	for _, a := range todos {
		if a.Id == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "todo not found"})
}

func appendTodos(c *gin.Context) {

	var newTodo todo
	if err := c.BindJSON(&newTodo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}
	todos = append(todos, newTodo)
	c.IndentedJSON(http.StatusCreated, newTodo)
}
func main() {

	// IMPORTANT: You must replace the DSN with your actual MySQL credentials.
	// DSN format: "user:password@tcp(127.0.0.1:3306)/database_name"
	dbpg.IntializeDB()       // Ensure the database connection is closed when main exits
	dbpg.InitElasticsearch() // Initialize Elasticsearch connection
	// Initialize the API functions by calling a separate function.
	initializeTables()
	initializeAPI()

}

func initializeTables() {
	model.CreateAllTables()
	model.CreateOrderTable()
	model.CreateItemTable()
	model.CreateCartTable()
	model.CreateCartItemsTable()
	model.CreateCheckoutTable()
	model.CreateUserAddressTable()
	model.CreateUserResumeTable()
	model.CreateAiResumeSummaryTable()
	model.CreatePlanTables()
}

// initializeAPI handles the registration of all API endpoints.
func initializeAPI() {
	// Register the signup handler from the handlers package.
	// We pass the database connection to the handler's constructor.
	signupapiv1.InitializeAPI()
}
