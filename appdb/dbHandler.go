package dbpg

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	_ "github.com/go-sql-driver/mysql" // MySQL driver import
)

var Db *sql.DB // Global variable to hold the database connection
// connectDB opens a database connection, pings it, and returns the DB object.
// This function centralizes the database connection logic.
func ConnectDB() (*sql.DB, error) {
	Db, err := sql.Open("mysql", "root:Qwerty@123@tcp(127.0.0.1:3306)/vendor_management?parseTime=true")
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Ping the database to ensure the connection is live.
	if err := Db.Ping(); err != nil {
		Db.Close() // Close the connection on ping failure
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to the database!")
	return Db, nil
}
func IntializeDB() {

	//spring.application.name=vendor-management-application
	//spring.datasource.url=jdbc:mysql://localhost:3306/vendor_management
	//spring.datasource.username=root
	//spring.datasource.password=Qwerty@123
	//spring.jpa.hibernate.ddl-auto=update
	var err error
	Db, err = ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer Db.Close()
}

// Global variable to hold the Elasticsearch client instance
var EsClient *elasticsearch.Client

// InitElasticsearch connects to the ES cluster and sets up the EsClient global variable.
func InitElasticsearch() {
	// 1. Define Configuration
	// It's best practice to load the URL from environment variables.
	// Example: "http://localhost:9200" or "https://elastic-host:9243"
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		esURL = "http://localhost:9200" // Default for local development
	}

	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
		// --- NEW: AUTHENTICATION ADDED HERE ---
		// Username: "elastic",  // <--- REPLACE WITH YOUR USERNAME (e.g., 'elastic' is common)
		// Password: "changeme", // <--- REPLACE WITH YOUR PASSWORD
		// Transport: &http.Transport{
		// 	TLSClientConfig: &tls.Config{
		// 		// WARNING: InsecureSkipVerify should ONLY be used in development/testing.
		// 		// This flag disables certificate chain and host name verification.
		// 		InsecureSkipVerify: true,
		// 	},
		// },
		// Optional: Add authentication if needed (API keys, Basic Auth, etc.)
		// Username: os.Getenv("ES_USERNAME"),
		// Password: os.Getenv("ES_PASSWORD"),
	}

	// 2. Initialize the Client
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the Elasticsearch client: %s", err)
	}

	// Ping the cluster to verify connectivity
	res, err := client.Info()
	if err != nil {
		log.Fatalf("Error pinging Elasticsearch cluster: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error response from Elasticsearch cluster: %s", res.String())
	}

	log.Println("INFO: Successfully connected to Elasticsearch cluster.")
	EsClient = client // Assign the initialized client globally
}
