package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	localmodel "achievesomethingbro/appmodel"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type JobResponses struct {
	localmodel.JobPosting        // Embed the original Visit struct
	UserName              string `json:"user_name"`
	Company_id            int    `json:"user_id"`
}

func HandleJobDescriptionRequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("hereJobDescription: %+v", r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	switch r.Method {
	case "POST":
		log.Print("herePostJobDescription")
		createJobDescription(w, r)
	case "GET":
		log.Print("herePostJobDescription1")
		parts := strings.Split(r.URL.Path, "/")
		// Example URL: /job-posting/company/123
		log.Print(len(parts))
		log.Print("Parts: ", parts)
		if len(parts) > 0 && len(parts) > 2 && len(parts) < 3 && parts[2] != "" {
			log.Print("Parts: ", parts)
			getJobDescriptionByID(w, r)
		} else if len(parts) > 0 && len(parts) > 4 && len(parts) < 6 && parts[2] == "company" && parts[3] != "" && parts[4] == "count" {
			getJobDescriptionCountByCompanyID(w, r)
		} else if len(parts) > 0 && len(parts) > 3 && len(parts) < 5 && parts[2] == "company" && parts[3] != "" {
			getJobDescriptionByCompanyID(w, r)
		} else {
			GetJobDescriptions(w, r)
		}
	case "PUT":
		log.Print("herePostJobDescription2")
		updateJobPosting(w, r)
	case "DELETE":
		log.Print("herePostJobDescription3")
		deleteJobPosting(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	log.Printf("out: %+v", r.Body)
	// Implementation for handling job description requests
}
func createJobDescription(w http.ResponseWriter, r *http.Request) {
	var jobDesc localmodel.JobPosting
	if err := json.NewDecoder(r.Body).Decode(&jobDesc); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	log.Print("Job Description Request: ", jobDesc)
	// Basic Validation
	if jobDesc.JobTitle == "" || jobDesc.JobDescription == "" || jobDesc.CompanyID <= 0 {
		http.Error(w, `{"message": "Title, Description and CompanyID are required"}`, http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	query := `
        INSERT INTO job_postings (
            job_title, job_description, company_id, number_of_openings, job_location, 
            job_type, salary_range, required_experience, required_qualifications, 
            skills_required, posting_date, application_deadline, job_posting_status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	jobDesc.PostingDate = time.Now().UTC()
	res, err := db.Exec(query,
		jobDesc.JobTitle,
		jobDesc.JobDescription,
		jobDesc.CompanyID,
		jobDesc.NumberOfOpenings,
		jobDesc.JobLocation,
		jobDesc.JobType,
		jobDesc.SalaryRange,
		jobDesc.RequiredExperience,
		jobDesc.RequiredQualifications,
		jobDesc.SkillsRequired,
		jobDesc.PostingDate,
		jobDesc.ApplicationDeadline,
		jobDesc.JobPostingStatus,
	)
	if err != nil {
		log.Printf("DB Insert Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	idInserted, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Convert the id (which is int64) to your specific custom type.
	jobDesc.ID = localmodel.StringInt(idInserted)

	response := map[string]interface{}{
		"message": "Job description created successfully",
		"body":    jobDesc,
	}
	log.Print(response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func GetJobDescriptions(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, job_title, job_description, company_id, number_of_openings, job_location, job_type, salary_range, required_experience, required_qualifications, skills_required, posting_date, application_deadline, job_posting_status FROM job_postings")
	if err != nil {
		log.Printf("Database query error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobPostings []localmodel.JobPosting
	for rows.Next() {
		var visit localmodel.JobPosting
		err := rows.Scan(
			&visit.ID,
			&visit.JobTitle,
			&visit.JobDescription,
			&visit.CompanyID,
			&visit.NumberOfOpenings,
			&visit.JobLocation,
			&visit.JobType,
			&visit.SalaryRange,
			&visit.RequiredExperience,
			&visit.RequiredQualifications,
			&visit.SkillsRequired,
			&visit.PostingDate,
			&visit.ApplicationDeadline,
			&visit.JobPostingStatus,
		)
		if err != nil {
			log.Printf("Database error scanning row: %v", err)
			http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
		jobPostings = append(jobPostings, visit)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Database result error: %v", err)
		http.Error(w, `{"error": "Database result error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobPostings)
}

func updateJobPosting(w http.ResponseWriter, r *http.Request) {
	// Implementation for updating a job posting
	var jobDesc localmodel.JobPosting
	if err := json.NewDecoder(r.Body).Decode(&jobDesc); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	log.Print(jobDesc)

	// Basic Validation
	if jobDesc.ID <= 0 || jobDesc.JobTitle == "" || jobDesc.JobDescription == "" || jobDesc.CompanyID <= 0 {
		http.Error(w, `{"message": "ID, Title, Description and CompanyID are required"}`, http.StatusBadRequest)
		return
	}
	log.Print("i am here")
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()
	log.Print("i am here")

	// Get the current time, convert it to UTC, and then format it.
	jobDesc.PostingDate = time.Now().UTC()
	query := `
		UPDATE job_postings SET
			job_title = ?, job_description = ?, company_id = ?, number_of_openings = ?, job_location = ?, 
			job_type = ?, salary_range = ?, required_experience = ?, required_qualifications = ?, 
			skills_required = ?, posting_date = ?, application_deadline = ?, job_posting_status = ?
		WHERE id = ?`
	log.Print("i am here")
	_, err = db.Exec(query,
		jobDesc.JobTitle,
		jobDesc.JobDescription,
		jobDesc.CompanyID,
		jobDesc.NumberOfOpenings,
		jobDesc.JobLocation,
		jobDesc.JobType,
		jobDesc.SalaryRange,
		jobDesc.RequiredExperience,
		jobDesc.RequiredQualifications,
		jobDesc.SkillsRequired,
		jobDesc.PostingDate,
		jobDesc.ApplicationDeadline,
		jobDesc.JobPostingStatus,
		jobDesc.ID,
	)
	log.Print("i am here")
	if err != nil {
		log.Printf("DB Update Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Job posting updated successfully",
		"body":    jobDesc,
	}
	log.Print(response)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func deleteJobPosting(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, `{"error": "Missing user ID in URL path"}`, http.StatusBadRequest)
		return
	}
	JobID := parts[2]
	if JobID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "DELETE FROM job_postings WHERE id = ?"
	result, err := db.Exec(query, JobID)
	if err != nil {
		log.Printf("DB Delete Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("DB RowsAffected Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, `{"message": "No job posting found with the given ID"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message": "Job posting deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getJobDescriptionByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, `{"error": "Missing job ID in URL path"}`, http.StatusBadRequest)
		return
	}
	jobID := parts[2]
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var jobDesc localmodel.JobPosting
	query := "SELECT id, job_title, job_description, company_id, number_of_openings, job_location, job_type, salary_range, required_experience, required_qualifications, skills_required, posting_date, application_deadline, job_posting_status FROM job_postings WHERE id = ?"
	err = db.QueryRow(query, jobID).Scan(
		&jobDesc.ID,
		&jobDesc.JobTitle,
		&jobDesc.JobDescription,
		&jobDesc.CompanyID,
		&jobDesc.NumberOfOpenings,
		&jobDesc.JobLocation,
		&jobDesc.JobType,
		&jobDesc.SalaryRange,
		&jobDesc.RequiredExperience,
		&jobDesc.RequiredQualifications,
		&jobDesc.SkillsRequired,
		&jobDesc.PostingDate,
		&jobDesc.ApplicationDeadline,
		&jobDesc.JobPostingStatus,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"message": "No job posting found with the given ID"}`, http.StatusNotFound)
			return
		}
		log.Printf("Database query error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	var userName string
	err = db.QueryRow("SELECT user_name FROM users WHERE id = ?", jobDesc.CompanyID).Scan(&userName)
	if err != nil {
		log.Printf("Database error retrieving company name: %v", err)
		userName = "Unknown"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobDesc)
}

func getJobDescriptionByCompanyID(w http.ResponseWriter, r *http.Request) {
	// --- Extract company ID from URL ---
	cleanPath := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(cleanPath, "/")
	// parts should be ["job-posting", "company", "123"]
	if len(parts) < 3 {
		http.Error(w, `{"error": "Missing company ID in URL path"}`, http.StatusBadRequest)
		return
	}
	companyID := parts[2]

	// --- Database Logic ---
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query to select job postings for a specific company
	query := `
        SELECT id, job_title,job_description, company_id, number_of_openings, job_location, 
               job_type, salary_range, required_experience, required_qualifications, 
               skills_required, posting_date, application_deadline, job_posting_status 
        FROM job_postings WHERE company_id = ?`

	rows, err := db.Query(query, companyID)
	if err != nil {
		log.Printf("Database query error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobPostings []localmodel.JobPosting
	for rows.Next() {
		var posting localmodel.JobPosting
		// Scan logic is the same as in GetJobDescriptions
		err := rows.Scan(
			&posting.ID, &posting.JobTitle, &posting.JobDescription, &posting.CompanyID,
			&posting.NumberOfOpenings, &posting.JobLocation, &posting.JobType,
			&posting.SalaryRange, &posting.RequiredExperience, &posting.RequiredQualifications,
			&posting.SkillsRequired, &posting.PostingDate, &posting.ApplicationDeadline,
			&posting.JobPostingStatus,
		)
		if err != nil {
			log.Printf("Database error scanning row: %v", err)
			continue // Continue to next row on scan error
		}
		jobPostings = append(jobPostings, posting)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Database iteration error: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobPostings)
}

func getJobDescriptionCountByCompanyID(w http.ResponseWriter, r *http.Request) {
	// --- Extract company ID from URL ---
	cleanPath := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(cleanPath, "/")
	// Assuming path is /job-posting/company/{id}, parts = ["job-posting", "company", "123"]
	if len(parts) < 4 {
		http.Error(w, `{"error": "Invalid URL path, expected /job-posting/company/{id}/count"}`, http.StatusBadRequest)
		return
	}
	companyID := parts[2] // ID is the 3rd part of the path (index 2)
	log.Print("here")
	// --- Database Logic ---
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()
	log.Print("here")
	// 1. Use a more efficient query to get only the count.
	// Use $1 for PostgreSQL placeholders.
	query := `SELECT COUNT(*) FROM job_postings WHERE company_id = ?`

	var count int
	log.Print("here")
	// 2. Use QueryRow since we expect only one row back.
	err = db.QueryRow(query, companyID).Scan(&count)
	if err != nil {
		log.Printf("Database query error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// 3. Create a map for a clean JSON response.
	response := map[string]int{"count": count}
	log.Print(response)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// func getJobDescriptionByCategory(w http.ResponseWriter, r *http.Request) {
//     // --- Extract company ID from URL ---
//     cleanPath := strings.Trim(r.URL.Path, "/")
//     parts := strings.Split(cleanPath, "/")
//     // parts should be ["job-posting", "company", "123"]
//     if len(parts) < 4 {
//         http.Error(w, `{"error": "Missing company ID in URL path"}`, http.StatusBadRequest)
//         return
//     }
//     companyID := parts[3]

//     // --- Database Logic ---
//     db, err := dbpg.ConnectDB()
//     if err != nil {
//         http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
//         return
//     }
//     defer db.Close()

//     // Query to select job postings for a specific company
//     query := `
//         SELECT id, job_title, job_description, company_id, number_of_openings, job_location,
//                job_type, salary_range, required_experience, required_qualifications,
//                skills_required, posting_date, application_deadline, job_posting_status
//         FROM job_postings WHERE job_type = ?`

//     rows, err := db.Query(query, )
//     if err != nil {
//         log.Printf("Database query error: %v", err)
//         http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
//         return
//     }
//     defer rows.Close()

//     var jobPostings []localmodel.JobPosting
//     for rows.Next() {
//         var posting localmodel.JobPosting
//         // Scan logic is the same as in GetJobDescriptions
//         err := rows.Scan(
//             &posting.ID, &posting.JobTitle, &posting.JobDescription, &posting.CompanyID,
//             &posting.NumberOfOpenings, &posting.JobLocation, &posting.JobType,
//             &posting.SalaryRange, &posting.RequiredExperience, &posting.RequiredQualifications,
//             &posting.SkillsRequired, &posting.PostingDate, &posting.ApplicationDeadline,
//             &posting.JobPostingStatus,
//         )
//         if err != nil {
//             log.Printf("Database error scanning row: %v", err)
//             continue // Continue to next row on scan error
//         }
//         jobPostings = append(jobPostings, posting)
//     }

//     if err = rows.Err(); err != nil {
//         log.Printf("Database iteration error: %v", err)
//     }

//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(jobPostings)
// }
