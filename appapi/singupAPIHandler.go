package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	userModel "achievesomethingbro/appmodel"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // MySQL driver import
)

// handleSignup is the handler function for the /signup endpoint.
// It processes the incoming request, validates the data, hashes the password,
// and saves the new user to our in-memory map.
func handleSignup(w http.ResponseWriter, r *http.Request) {
	// Set the content type to JSON for the response.
	w.Header().Set("Content-Type", "application/json")

	// Decode the request body into a new User struct.
	var companyPayload userModel.CompanyPayload
	err := json.NewDecoder(r.Body).Decode(&companyPayload)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		// If decoding fails, return a bad request error.
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Simple validation: check if username or password are empty.
	if companyPayload.UserName == "" || companyPayload.FirstName == "" || companyPayload.LastName == "" || companyPayload.MobileNumber == "" || companyPayload.Email == "" {
		http.Error(w, `{"error": "All fields are required"}`, http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	var userID int
	var count int
	query := "SELECT count(*) FROM users WHERE user_type = ? AND (user_name = ? OR email = ? OR mobile_number = ?)"
	err = db.QueryRow(query, companyPayload.UserType, companyPayload.UserName, companyPayload.Email, companyPayload.MobileNumber).Scan(&count)
	if err != nil {
		log.Printf("Database error checking for existing user: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, `{"error": "User with the same username, email, or mobile number already exists"}`, http.StatusConflict)
		return
	}
	query = "SELECT id FROM users WHERE user_name = ? OR email = ? OR mobile_number = ?"
	err = db.QueryRow(query, companyPayload.UserName, companyPayload.Email, companyPayload.MobileNumber).Scan(&userID)
	if err != sql.ErrNoRows && err != nil {
		log.Printf("Database error retrieving user ID: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	// Prepare the SQL INSERT statement.
	sqlStmt := `
	INSERT INTO users (user_name, first_name, last_name, mobile_number, email, gender, date_of_birth, password,user_type)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	log.Printf("New user signed up: %s", companyPayload.UserName)
	// Execute the INSERT statement.
	result, err := db.Exec(sqlStmt,
		companyPayload.UserName,
		companyPayload.FirstName,
		companyPayload.LastName,
		companyPayload.MobileNumber,
		companyPayload.Email,
		companyPayload.Gender,
		companyPayload.DateOfBirth,
		companyPayload.Password,
		companyPayload.UserType,
	)
	if err != nil {
		log.Printf("Failed to insert new user: %v user data %v", err, companyPayload)
		http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve last insert ID: %v", err)
		http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"message":   "User created successfully!",
		"user_id":   insertedId,
		"user_name": companyPayload.UserName,
	}
	if companyPayload.UserType == "RECRUITER" || companyPayload.UserType == "ADVERTISER" {
		companyStmt := `
	INSERT INTO companies (user_id, company_name, gst_number)
	VALUES (?, ?, ?)`
		comRes, err := db.Exec(companyStmt,
			insertedId,
			companyPayload.CompanyName,
			companyPayload.GSTNumber,
		)
		if err != nil {
			log.Print(companyPayload)
			log.Printf("Failed to insert company details: %v", err)
			http.Error(w, `{"error": "Failed to create company details"}`, http.StatusInternalServerError)
			return
		}
		fmt.Println("Company details inserted successfully")
		compId, err := comRes.LastInsertId()
		if err != nil {
			log.Printf("Failed to retrieve last insert ID: %v", err)
			http.Error(w, `{"error": "Failed to create company"}`, http.StatusInternalServerError)
			return
		}
		response["company_id"] = compId
	}

	if err != nil {
		log.Printf("Failed to insert new user: %v user data %v", err, companyPayload)
		http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Respond with a success message.
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Printf("New user signed up: %s (ID: %d)", companyPayload.UserName, userID)
	db.Close()
}

// InitializeAPI handles the registration of all API endpoints.
func InitializeAPI() {
	// Register the signup handler from the handlers package.
	// We pass the database connection to the handler's constructor.
	http.HandleFunc("/admin", AdminHandler)
	http.HandleFunc("/signup", handleSignup)
	http.HandleFunc("/user", handleGetUser)
	http.HandleFunc("/user/", handleDeleteUser)
	// --- API Route Setup with CORS middleware ---
	http.HandleFunc("/plans/jobseeker", JobseekerPlansHandler)
	http.HandleFunc("/plans/company", CompanyPlansHandler)
	http.HandleFunc("/plans/advertiser", AdvertiserPlansHandler)
	// http.HandleFunc("/loginorsignup.html", templateLoginOrSignupHandler)
	// http.HandleFunc("/dashboard.html", templateDashboardHandler)

	// Define API routes and link them to handler functions.
	http.HandleFunc("/api/items", handleItemCollection)
	http.HandleFunc("/api/items/", handleItemResource)
	http.HandleFunc("/checkout", handleCheckoutResource)
	http.HandleFunc("/jobseeker/profile/", HandleJobSeekerProfileRequests)
	// --- Address API Wiring ---
	// Register the new address handler for CRUD operations.
	// We register /address and /address/ to allow the handler to process
	// both base requests (POST) and requests with trailing IDs (GET, PUT, DELETE).
	http.HandleFunc("/address", handleAddressRequests)
	http.HandleFunc("/address/", handleAddressRequests)
	// Separate handlers for orders
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			log.Println("Creating a new order")
			handleCreateOrder(w, r)
		case http.MethodGet:
			log.Println("Retrieving all orders")
			handleGetOrders(w, r)
		default:
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetSingleOrder(w, r)
		case http.MethodPut:
			handleUpdateOrder(w, r)
		case http.MethodDelete:
			handleDeleteOrder(w, r)
		default:
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	// Resume upload/download handlers
	MakeDir() // Ensure the resume storage directory exists
	// Handle /resume without trailing slash
	http.HandleFunc("/resume/", handleResumeRequests)

	// 2. Resume File Download Handler (streams the file content)
	// NOTE: Requires a path with a user ID suffix, e.g., /resume/download/123
	http.HandleFunc("/resume/download/", handleDownloadResumeRequests)

	// 3. AI Analysis Handler (handles POST /resume/analyze)
	http.HandleFunc("/resume/analyze", handleAnalyzeResumeRequests)

	// Cart handlers
	// In the main() function:
	http.HandleFunc("/api/cart", handleCartRequests)
	http.HandleFunc("/api/cart/", handleCartRequests)

	// Register the handler for the "/login/admin" endpoint
	// We wrap our handler with the CORS middleware
	http.HandleFunc("/admin/login", adminLoginHandler)
	http.HandleFunc("/login", handleLoginRequests)
	http.HandleFunc("/analyze-resume", ResumeSummaryHandler)
	http.HandleFunc("/analyze-resume/", ResumeSummaryHandler)
	http.HandleFunc("/resume-visit", HandleResumeContactsRequests)
	http.HandleFunc("/resume-visit/", HandleResumeContactsRequests)
	http.HandleFunc("/job-posting", HandleJobDescriptionRequests)
	http.HandleFunc("/job-posting/", HandleJobDescriptionRequests)
	http.HandleFunc("/job-visits", HandleJobVisitsRequests)
	http.HandleFunc("/job-visits/", HandleJobVisitsRequests)
	http.HandleFunc("/index.html", templateHandler)
	http.HandleFunc("/", templateHandler)
	http.HandleFunc("/testing", TestingHandler)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
