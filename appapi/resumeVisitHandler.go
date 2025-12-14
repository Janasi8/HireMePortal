package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	localmodel "achievesomethingbro/appmodel"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type VisitResponse struct {
	localmodel.ResumeContactMade        // Embed the original Visit struct
	CompanyName                  string `json:"company_name"`
}

func HandleResumeContactsRequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("hereResume: %+v", r.Body)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	switch r.Method {
	case "POST":
		markResumeVisit(w, r)
	case "GET":
		parts := strings.Split(r.URL.Path, "/")
		// Example URL: /job-posting/company/123
		log.Print(len(parts))
		log.Print("Parts: ", parts)
		if len(parts) > 0 && len(parts) > 3 && len(parts) < 5 && parts[2] == "user" && parts[3] != "" {
			log.Print("Parts: ", parts[2])
			getResumeVisitsByUser(w, r)
		} else if len(parts) > 0 && len(parts) > 3 && len(parts) < 5 && parts[2] == "company" && parts[3] != "" {
			log.Print("Parts: ", parts[2])
			getResumeVisitsByCompany(w, r)
		} else if len(parts) > 0 && len(parts) > 3 && len(parts) < 5 && parts[2] == "resume" && parts[3] != "" {
			log.Print("Parts: ", parts[2])
			getResumeVisitsByResume(w, r)
		} else if len(parts) > 0 && len(parts) > 2 && len(parts) < 4 && parts[2] != "" {
			log.Print("Parts: ", parts[2])
			getResumeVisitsById(w, r)
		} else {
			getAllResumeVisits(w, r)
		}
	case "PUT":
		updateResumeStatusVisit(w, r)
	case "DELETE":
		deleteResumeStatusVisit(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func markResumeVisit(w http.ResponseWriter, r *http.Request) {

	var req localmodel.ResumeContactMade
	log.Printf("hereMarkResumeVisit: %+v", req)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	log.Printf("hereMarkResumeVisit: %+v", req)
	// Basic Validation
	if req.UserID <= 1 || req.ResumeID <= 1 || req.CompanyID <= 0 {
		http.Error(w, `{"message": "UserID, ResumeID and CompanyName are required"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Resume Visit Request: %+v", req)

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("DB Connection Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "INSERT INTO resume_contacts_made (user_id, resume_id, company_id, contact_date, status) VALUES (?, ?, ?, CURDATE(), ?)"
	_, err = db.Exec(query, req.UserID, req.ResumeID, req.CompanyID, "RESUME_CHECKED")
	if err != nil {
		log.Printf("DB Insert Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Resume visit recorded successfully",
	})
}

func updateResumeStatusVisit(w http.ResponseWriter, r *http.Request) {
	log.Printf("hereUpdateResumeStatusVisit: %+v", r.Body)
	var req localmodel.ResumeContactMade
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Basic Validation
	if req.UserID <= 0 || req.ResumeID <= 0 || req.CompanyID <= 0 || req.Status == "" {
		http.Error(w, `{"message": "UserID, ResumeID, CompanyName and Status are required"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Update Resume Status Visit Request: %+v", req)

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("DB Connection Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE resume_contacts_made SET status = ? WHERE user_id = ? AND resume_id = ? AND company_id = ?"
	_, err = db.Exec(query, req.Status, req.UserID, req.ResumeID, req.CompanyID)
	if err != nil {
		log.Printf("DB Update Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Resume visit status updated successfully",
	})
}

func deleteResumeStatusVisit(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /resume/{userID}/{CompanyID}
	log.Printf("Delete Resume Request Path: %s", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	log.Printf("Delete Resume Request Parts: %+v", parts)
	for i, part := range parts {
		log.Printf("Part %d: %s", i, part)
	}
	if len(parts) < 4 || parts[2] == "" || parts[3] == "" {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		log.Printf("Parts: %+v", parts)
		http.Error(w, `{"error": "Missing user ID or resume ID in URL path"}`, http.StatusBadRequest)
		return
	}
	userID := parts[2]
	companyId := parts[3]
	log.Printf("hereDeleteResumeStatusVisit for userID: %s, resumeID: %s", userID,
		companyId)

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "DELETE FROM resume_contacts_made WHERE user_id = $1 AND company_id = $2"
	_, err = db.Exec(query, userID, companyId)
	if err != nil {
		log.Printf("Database error deleting resume visit: %v", err)
		http.Error(w, `{"error": "Database error deleting resume visit"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Resume visit deleted successfully",
	})
}

func getResumeVisitsByCompany(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /resume/{userID}
	log.Printf("Get Resume Request Path: %s", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	log.Printf("Get Resume Request Parts: %+v", parts)
	for i, part := range parts {
		log.Printf("Part %d: %s", i, part)
	}
	if len(parts) < 4 || parts[3] == "" {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		log.Printf("Parts: %+v", parts)
		http.Error(w, `{"error": "Missing user ID in URL path"}`, http.StatusBadRequest)
		return
	}
	// 1. FIX: Convert the userID from string to an integer.
	companyIdStr := parts[3]
	companyId, err := strconv.Atoi(companyIdStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	log.Printf("hereGetResumeVisits for userID: %d", companyId)

	db, err := dbpg.ConnectDB() // Assuming this connects to your MySQL DB
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 2. FIX: Use MySQL's '?' placeholder syntax.
	query := `
        SELECT user_id, resume_id, company_id, contact_date, status
        FROM resume_contacts_made WHERE company_id = ? ORDER BY contact_date DESC`

	result, err := db.Query(query, companyId)
	if err != nil {
		// This is where your original error was being triggered.
		log.Printf("Database error retrieving resume visits: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	// ... rest of your code to process the result
	defer result.Close()

	var visits []localmodel.ResumeContactMade
	for result.Next() {
		var visit localmodel.ResumeContactMade
		if err := result.Scan(&visit.UserID, &visit.ResumeID, &visit.CompanyID, &visit.ContactDate, &visit.Status); err != nil {
			log.Printf("Database error scanning resume visit: %v", err)
			http.Error(w, `{"error": "Database error scanning resume visit"}`, http.StatusInternalServerError)
			return
		}
		visits = append(visits, visit)
	}
	log.Print("Visits: ", visits)
	if err = result.Err(); err != nil {
		log.Printf("Database result error: %v", err)
		http.Error(w, `{"error": "Database result error"}`, http.StatusInternalServerError)
		return
	}
	var companyName string
	var visitResponses []VisitResponse
	visitResponses = make([]VisitResponse, len(visits))
	for i, visit := range visits {
		err := db.QueryRow("SELECT company_name FROM companies WHERE id = ?", visit.CompanyID).Scan(&companyName)
		if err != nil {
			log.Printf("Database error retrieving company name: %v", err)
			companyName = "Unknown"
		}
		visitResponses[i].ResumeContactMade = visit
		visitResponses[i].CompanyName = companyName
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visitResponses)
}

func getResumeVisitsByUser(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /resume/{userID}
	log.Printf("Get Resume Request Path: %s", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	log.Printf("Get Resume Request Parts: %+v", parts)
	for i, part := range parts {
		log.Printf("Part %d: %s", i, part)
	}
	if len(parts) < 4 || parts[3] == "" {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		log.Printf("Parts: %+v", parts)
		http.Error(w, `{"error": "Missing user ID in URL path"}`, http.StatusBadRequest)
		return
	}
	// 1. FIX: Convert the userID from string to an integer.
	userIDStr := parts[3]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	log.Printf("hereGetResumeVisits for userID: %d", userID)

	db, err := dbpg.ConnectDB() // Assuming this connects to your MySQL DB
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	statuses := []string{"REJECTED", "SELECTED"}

	// 2. FIX: Use MySQL's '?' placeholder syntax.
	query := `
        SELECT user_id, resume_id, company_id, contact_date, status
        FROM resume_contacts_made
        WHERE user_id = ? AND status NOT IN (%s)
        ORDER BY contact_date DESC`

	// Generate the correct number of '?' placeholders for the IN clause.
	placeholders := strings.Repeat("?,", len(statuses)-1) + "?"
	finalQuery := fmt.Sprintf(query, placeholders)

	// Prepare all arguments for the query.
	args := make([]interface{}, len(statuses)+1)
	args[0] = userID
	for i, status := range statuses {
		args[i+1] = status
	}

	// Execute the corrected query.
	result, err := db.Query(finalQuery, args...)
	if err != nil {
		// This is where your original error was being triggered.
		log.Printf("Database error retrieving resume visits: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	// ... rest of your code to process the result
	defer result.Close()

	var visits []localmodel.ResumeContactMade
	for result.Next() {
		var visit localmodel.ResumeContactMade
		if err := result.Scan(&visit.UserID, &visit.ResumeID, &visit.CompanyID, &visit.ContactDate, &visit.Status); err != nil {
			log.Printf("Database error scanning resume visit: %v", err)
			http.Error(w, `{"error": "Database error scanning resume visit"}`, http.StatusInternalServerError)
			return
		}
		visits = append(visits, visit)
	}
	log.Print("Visits: ", visits)
	if err = result.Err(); err != nil {
		log.Printf("Database result error: %v", err)
		http.Error(w, `{"error": "Database result error"}`, http.StatusInternalServerError)
		return
	}
	var companyName string
	var visitResponses []VisitResponse
	visitResponses = make([]VisitResponse, len(visits))
	for i, visit := range visits {
		err := db.QueryRow("SELECT company_name FROM companies WHERE id = ?", visit.CompanyID).Scan(&companyName)
		if err != nil {
			log.Printf("Database error retrieving company name: %v", err)
			companyName = "Unknown"
		}
		visitResponses[i].ResumeContactMade = visit
		visitResponses[i].CompanyName = companyName
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visitResponses)
}

func getResumeVisitsByResume(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /resume-visit/resume/resumeid
	log.Printf("Get Resume Request Path: %s", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	log.Printf("Get Resume Request Parts: %+v", parts)
	for i, part := range parts {
		log.Printf("Part %d: %s", i, part)
	}
	if len(parts) < 4 || parts[3] == "" {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		log.Printf("Parts: %+v", parts)
		http.Error(w, `{"error": "Missing resume ID in URL path"}`, http.StatusBadRequest)
		return
	}
	// 1. FIX: Convert the userID from string to an integer.
	resumeIDStr := parts[3]
	resumeID, err := strconv.Atoi(resumeIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid resume ID"}`, http.StatusBadRequest)
		return
	}

	log.Printf("hereGetResumeVisits for userID: %d", resumeID)

	db, err := dbpg.ConnectDB() // Assuming this connects to your MySQL DB
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 2. FIX: Use MySQL's '?' placeholder syntax.
	query := `
        SELECT user_id, resume_id, company_id, contact_date, status
        FROM resume_contacts_made
        WHERE resume_id = ? ORDER BY contact_date DESC`

	// Execute the corrected query.
	result, err := db.Query(query, resumeID)
	if err != nil {
		// This is where your original error was being triggered.
		log.Printf("Database error retrieving resume visits: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	// ... rest of your code to process the result
	defer result.Close()

	var visits []localmodel.ResumeContactMade
	for result.Next() {
		var visit localmodel.ResumeContactMade
		if err := result.Scan(&visit.UserID, &visit.ResumeID, &visit.CompanyID, &visit.ContactDate, &visit.Status); err != nil {
			log.Printf("Database error scanning resume visit: %v", err)
			http.Error(w, `{"error": "Database error scanning resume visit"}`, http.StatusInternalServerError)
			return
		}
		visits = append(visits, visit)
	}
	log.Print("Visits: ", visits)
	if err = result.Err(); err != nil {
		log.Printf("Database result error: %v", err)
		http.Error(w, `{"error": "Database result error"}`, http.StatusInternalServerError)
		return
	}
	var companyName string
	var visitResponses []VisitResponse
	visitResponses = make([]VisitResponse, len(visits))
	for i, visit := range visits {
		err := db.QueryRow("SELECT company_name FROM companies WHERE id = ?", visit.CompanyID).Scan(&companyName)
		if err != nil {
			log.Printf("Database error retrieving company name: %v", err)
			companyName = "Unknown"
		}
		visitResponses[i].ResumeContactMade = visit
		visitResponses[i].CompanyName = companyName
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visitResponses)
}

func getResumeVisitsById(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /resume/{userID}
	log.Printf("Get Resume Request Path: %s", r.URL.Path)
	parts := strings.Split(r.URL.Path, "/")
	log.Printf("Get Resume Request Parts: %+v", parts)
	for i, part := range parts {
		log.Printf("Part %d: %s", i, part)
	}
	if len(parts) < 3 || parts[2] == "" {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		log.Printf("Parts: %+v", parts)
		http.Error(w, `{"error": "Missing user ID in URL path"}`, http.StatusBadRequest)
		return
	}
	// 1. FIX: Convert the userID from string to an integer.
	idStr := parts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	log.Printf("hereGetResumeVisits for userID: %d", id)

	db, err := dbpg.ConnectDB() // Assuming this connects to your MySQL DB
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 2. FIX: Use MySQL's '?' placeholder syntax.
	query := `
        SELECT user_id, resume_id, company_id, contact_date, status
        FROM resume_contacts_made
        WHERE id = ?
        ORDER BY contact_date DESC`

	// Execute the corrected query.
	result, err := db.Query(query, id)
	if err != nil {
		// This is where your original error was being triggered.
		log.Printf("Database error retrieving resume visits: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	// ... rest of your code to process the result
	defer result.Close()

	var visits []localmodel.ResumeContactMade
	for result.Next() {
		var visit localmodel.ResumeContactMade
		if err := result.Scan(&visit.UserID, &visit.ResumeID, &visit.CompanyID, &visit.ContactDate, &visit.Status); err != nil {
			log.Printf("Database error scanning resume visit: %v", err)
			http.Error(w, `{"error": "Database error scanning resume visit"}`, http.StatusInternalServerError)
			return
		}
		visits = append(visits, visit)
	}
	log.Print("Visits: ", visits)
	if err = result.Err(); err != nil {
		log.Printf("Database result error: %v", err)
		http.Error(w, `{"error": "Database result error"}`, http.StatusInternalServerError)
		return
	}
	var companyName string
	var visitResponses []VisitResponse
	visitResponses = make([]VisitResponse, len(visits))
	for i, visit := range visits {
		err := db.QueryRow("SELECT company_name FROM companies WHERE id = ?", visit.CompanyID).Scan(&companyName)
		if err != nil {
			log.Printf("Database error retrieving company name: %v", err)
			companyName = "Unknown"
		}
		visitResponses[i].ResumeContactMade = visit
		visitResponses[i].CompanyName = companyName
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visitResponses)
}

func getAllResumeVisits(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /resume/{userID}
	log.Printf("Get Resume Request Path: %s", r.URL.Path)

	db, err := dbpg.ConnectDB() // Assuming this connects to your MySQL DB
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 2. FIX: Use MySQL's '?' placeholder syntax.
	query := `
        SELECT user_id, resume_id, company_id, contact_date, status
        FROM resume_contacts_made ORDER BY contact_date DESC`

	// Execute the corrected query.
	result, err := db.Query(query)
	if err != nil {
		// This is where your original error was being triggered.
		log.Printf("Database error retrieving resume visits: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	// ... rest of your code to process the result
	defer result.Close()

	var visits []localmodel.ResumeContactMade
	for result.Next() {
		var visit localmodel.ResumeContactMade
		if err := result.Scan(&visit.UserID, &visit.ResumeID, &visit.CompanyID, &visit.ContactDate, &visit.Status); err != nil {
			log.Printf("Database error scanning resume visit: %v", err)
			http.Error(w, `{"error": "Database error scanning resume visit"}`, http.StatusInternalServerError)
			return
		}
		visits = append(visits, visit)
	}
	log.Print("Visits: ", visits)
	if err = result.Err(); err != nil {
		log.Printf("Database result error: %v", err)
		http.Error(w, `{"error": "Database result error"}`, http.StatusInternalServerError)
		return
	}
	var companyName string
	var visitResponses []VisitResponse
	visitResponses = make([]VisitResponse, len(visits))
	for i, visit := range visits {
		err := db.QueryRow("SELECT company_name FROM companies WHERE id = ?", visit.CompanyID).Scan(&companyName)
		if err != nil {
			log.Printf("Database error retrieving company name: %v", err)
			companyName = "Unknown"
		}
		visitResponses[i].ResumeContactMade = visit
		visitResponses[i].CompanyName = companyName
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visitResponses)
}
