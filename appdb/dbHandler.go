package dbpg

import (
	"database/sql"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

var Db *sql.DB

// ==================== DATABASE (MOCK / DEMO MODE) ====================

// ConnectDB is a MOCK function for demo purposes
func ConnectDB() (*sql.DB, error) {
	log.Println("Mock DB connection used (demo mode)")
	return nil, nil
}

// IntializeDB does nothing in demo mode
func IntializeDB() {
	log.Println("Database initialization skipped (demo mode)")
	return
}

// ==================== ELASTICSEARCH (OPTIONAL) ====================

var EsClient *elasticsearch.Client

func InitElasticsearch() {
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		log.Println("Elasticsearch disabled (demo mode)")
		return
	}

	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Println("Elasticsearch client error:", err)
		return
	}

	res, err := client.Info()
	if err != nil {
		log.Println("Elasticsearch not reachable:", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Println("Elasticsearch error response")
		return
	}

	log.Println("Elasticsearch connected (optional)")
	EsClient = client
}
