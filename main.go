package main

import (
	signupapiv1 "achievesomethingbro/appapi"
	dbpg "achievesomethingbro/appdb"
	model "achievesomethingbro/appmodel"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	isRender := os.Getenv("RENDER") == "true"

	// ✅ Initialize DB only if NOT running on Render
	if !isRender {
		log.Println("🟢 Running locally — Initializing DB & Elasticsearch")
		dbpg.IntializeDB()
		dbpg.InitElasticsearch()
		initializeTables()
	} else {
		log.Println("🟡 Running on Render — Skipping local DB initialization")
	}

	// ✅ Create Gin router
	router := gin.Default()

	// ✅ Initialize all APIs with router
	signupapiv1.InitializeAPI(router)

	// ✅ Dynamic port for Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Println("🚀 Server running on port " + port)
	router.Run(":" + port)
}

// ---------------- TABLE INITIALIZATION ----------------

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
