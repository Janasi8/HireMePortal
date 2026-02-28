package dbpg

import (
	"database/sql"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

// 🔴 HARD-CODED DSN (to avoid env issues)
var dsn = "root:Jaiswal123@tcp(127.0.0.1:3306)/hiremeportal?parseTime=true"

// ==================== DATABASE ====================

// ConnectDB creates or reuses a single DB connection
func ConnectDB() (*sql.DB, error) {

	// Reuse connection if already exists
	if Db != nil {
		return Db, nil
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ MySQL connected successfully")
	Db = db
	return Db, nil
}

// IntializeDB initializes DB at app start
func IntializeDB() {
	db, err := ConnectDB()
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}
	Db = db
}

// ==================== ELASTICSEARCH (OPTIONAL) ====================

var EsClient *elasticsearch.Client

func InitElasticsearch() {
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		log.Println("ℹ️ Elasticsearch disabled")
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

	log.Println("✅ Elasticsearch connected")
	EsClient = client
}
