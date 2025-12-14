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

// JobVisitResponse is a composite struct that combines the contact record
// with the full details of the job posting for a comprehensive API response.
type JobVisitResponse struct {
	localmodel.JobContactMade
	localmodel.JobPosting
}

// HandleJobVisitsRequests is the main router for all incoming requests
// related to job visits. It delegates tasks to specific functions based on the HTTP method and URL path.
func HandleJobVisitsRequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received Job Visit Request: Method %s, URL %s", r.Method, r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle pre-flight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 || parts[0] != "job-visits" {
		http.Error(w, `{"message": "Invalid endpoint, must start with /job-visits"}`, http.StatusNotFound)
		return
	}

	switch r.Method {
	case "POST":
		// Handles POST /job-visits
		markJobVisit(w, r)
	case "PUT":
		// Handles PUT /job-visits
		updateJobStatusVisit(w, r)
	case "DELETE":
		// Handles DELETE /job-visits/{userID}/{jobID}
		if len(parts) != 3 {
			http.Error(w, `{"error": "DELETE URL must be in the format /job-visits/{userID}/{jobID}"}`, http.StatusBadRequest)
			return
		}
		deleteJobStatusVisit(w, r)
	case "GET":
		log.Printf("GET request with URL Parts: %v", parts)
		switch len(parts) {
		case 1:
			// Handles GET /job-visits
			getAllJobVisits(w, r)
		case 2:
			// Handles GET /job-visits/{id}
			getJobVisitByID(w, r)
		case 3:
			// Handles GET /job-visits/user/{id}, /job-visits/company/{id}, /job-visits/job/{id}
			entity := parts[1]
			switch entity {
			case "user":
				getJobVisitsByUser(w, r)
			case "company":
				getJobVisitsByCompany(w, r)
			case "job":
				getJobVisitsByJob(w, r)
			default:
				http.Error(w, `{"message": "Invalid GET path. After /job-visits/, expected 'user', 'company', 'job', or a contact ID."}`, http.StatusBadRequest)
			}
		default:
			http.Error(w, `{"message": "Malformed URL for GET request"}`, http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// markJobVisit handles POST requests to record that a user has contacted/applied for a job.
func markJobVisit(w http.ResponseWriter, r *http.Request) {
	var req localmodel.JobContactMade
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.UserID <= 0 || req.JobID <= 0 || req.CompanyID <= 0 {
		http.Error(w, `{"message": "UserID, JobID, and CompanyID are required"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Job Visit Request: %+v", req)

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("DB Connection Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert a new record with the default status of 'CONTACTED'
	query := "INSERT INTO job_contacts_made (user_id, job_id, company_id, contact_date, contacted_status) VALUES (?, ?, ?, CURDATE(), ?)"
	_, err = db.Exec(query, req.UserID, req.JobID, req.CompanyID, "CONTACTED")
	if err != nil {
		log.Printf("DB Insert Error: %v", err)
		http.Error(w, `{"message": "Could not record job visit. You may have already applied."}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Job visit recorded successfully",
	})
}

// updateJobStatusVisit handles PUT requests to change the status of a job contact.
func updateJobStatusVisit(w http.ResponseWriter, r *http.Request) {
	var req localmodel.JobContactMade
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.UserID <= 0 || req.JobID <= 0 || req.CompanyID <= 0 || req.ContactedStatus == "" {
		http.Error(w, `{"message": "UserID, JobID, CompanyID and ContactedStatus are required"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Update Job Status Visit Request: %+v", req)

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("DB Connection Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE job_contacts_made SET contacted_status = ? WHERE user_id = ? AND job_id = ? AND company_id = ?"
	res, err := db.Exec(query, req.ContactedStatus, req.UserID, req.JobID, req.CompanyID)
	if err != nil {
		log.Printf("DB Update Error: %v", err)
		http.Error(w, `{"message": "Internal server error during update"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"message": "No matching record found to update"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Job visit status updated successfully",
	})
}

// deleteJobStatusVisit handles DELETE requests to remove a job contact record.
func deleteJobStatusVisit(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /job-visits/{userID}/{jobID}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, `{"error": "URL must be in the format /job-visits/{userID}/{jobID}"}`, http.StatusBadRequest)
		return
	}
	userID := parts[1]
	jobID := parts[2]

	log.Printf("Attempting to delete job contact for userID: %s, jobID: %s", userID, jobID)

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "DELETE FROM job_contacts_made WHERE user_id = ? AND job_id = ?"
	res, err := db.Exec(query, userID, jobID)
	if err != nil {
		log.Printf("Database error deleting job visit: %v", err)
		http.Error(w, `{"error": "Failed to delete job visit"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"message": "No matching record found to delete"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Job visit deleted successfully",
	})
}

// getJobVisits retrieves job contacts based on a specified column (e.g., user_id, company_id) and value.
// It performs a JOIN with job_postings to return comprehensive data.
func getJobVisits(w http.ResponseWriter, r *http.Request, whereColumn string, id int) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("DB Connection Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := fmt.Sprintf(`
        SELECT
            jcm.id, jcm.job_id, jcm.user_id, jcm.company_id, jcm.contact_date, jcm.contacted_status,
            jp.id, jp.company_id, jp.job_title, jp.job_description, jp.number_of_openings,
            jp.job_location, jp.job_type, jp.salary_range, jp.required_experience,
            jp.required_qualifications, jp.skills_required, jp.posting_date,
            jp.application_deadline, jp.job_posting_status
        FROM
            job_contacts_made jcm
        JOIN
            job_postings jp ON jcm.job_id = jp.id
        WHERE %s = ?
        ORDER BY
            jcm.contact_date DESC`, whereColumn)

	rows, err := db.Query(query, id)
	if err != nil {
		log.Printf("Database query error: %v", err)
		http.Error(w, `{"message": "Failed to retrieve job visits"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var visits []JobVisitResponse
	for rows.Next() {
		var visit JobVisitResponse
		if err := rows.Scan(
			&visit.JobContactMade.ID, &visit.JobContactMade.JobID, &visit.JobContactMade.UserID, &visit.JobContactMade.CompanyID, &visit.JobContactMade.ContactDate, &visit.JobContactMade.ContactedStatus,
			&visit.JobPosting.ID, &visit.JobPosting.CompanyID, &visit.JobPosting.JobTitle, &visit.JobPosting.JobDescription, &visit.JobPosting.NumberOfOpenings,
			&visit.JobPosting.JobLocation, &visit.JobPosting.JobType, &visit.JobPosting.SalaryRange, &visit.JobPosting.RequiredExperience,
			&visit.JobPosting.RequiredQualifications, &visit.JobPosting.SkillsRequired, &visit.JobPosting.PostingDate,
			&visit.JobPosting.ApplicationDeadline, &visit.JobPosting.JobPostingStatus,
		); err != nil {
			log.Printf("Database scan error: %v", err)
			http.Error(w, `{"message": "Failed to process job visit data"}`, http.StatusInternalServerError)
			return
		}
		visits = append(visits, visit)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		http.Error(w, `{"message": "Error reading job visit results"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visits)
}

// getJobVisitsByUser is a specific handler for GET /job-visits/user/{id}.
func getJobVisitsByUser(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, `{"error": "URL must be in the format /job-visits/user/{id}"}`, http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Fetching job visits for User ID: %d", userID)
	getJobVisits(w, r, "jcm.user_id", userID)
}

// getJobVisitsByCompany is a specific handler for GET /job-visits/company/{id}.
func getJobVisitsByCompany(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, `{"error": "URL must be in the format /job-visits/company/{id}"}`, http.StatusBadRequest)
		return
	}
	companyID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, `{"error": "Invalid company ID"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Fetching job visits for Company ID: %d", companyID)
	getJobVisits(w, r, "jcm.company_id", companyID)
}

// getJobVisitsByJob is a specific handler for GET /job-visits/job/{id}.
func getJobVisitsByJob(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, `{"error": "URL must be in the format /job-visits/job/{id}"}`, http.StatusBadRequest)
		return
	}
	jobID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, `{"error": "Invalid job ID"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Fetching job visits for Job ID: %d", jobID)
	getJobVisits(w, r, "jcm.job_id", jobID)
}

// getJobVisitByID is a specific handler for GET /job-visits/{id}.
func getJobVisitByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		http.Error(w, `{"error": "URL must be in the format /job-visits/{id}"}`, http.StatusBadRequest)
		return
	}
	contactID, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, `{"error": "Invalid contact ID"}`, http.StatusBadRequest)
		return
	}
	log.Printf("Fetching job visit for Contact ID: %d", contactID)
	getJobVisits(w, r, "jcm.id", contactID)
}

// getAllJobVisits handles GET /job-visits to retrieve all records.
func getAllJobVisits(w http.ResponseWriter, r *http.Request) {
	log.Println("Fetching all job visits")
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("DB Connection Error: %v", err)
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
        SELECT
            jcm.id, jcm.job_id, jcm.user_id, jcm.company_id, jcm.contact_date, jcm.contacted_status,
            jp.id, jp.company_id, jp.job_title, jp.job_description, jp.number_of_openings,
            jp.job_location, jp.job_type, jp.salary_range, jp.required_experience,
            jp.required_qualifications, jp.skills_required, jp.posting_date,
            jp.application_deadline, jp.job_posting_status
        FROM
            job_contacts_made jcm
        JOIN
            job_postings jp ON jcm.job_id = jp.id
        ORDER BY
            jcm.contact_date DESC`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Database query error: %v", err)
		http.Error(w, `{"message": "Failed to retrieve job visits"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var visits []JobVisitResponse
	for rows.Next() {
		var visit JobVisitResponse
		if err := rows.Scan(
			&visit.JobContactMade.ID, &visit.JobContactMade.JobID, &visit.JobContactMade.UserID, &visit.JobContactMade.CompanyID, &visit.JobContactMade.ContactDate, &visit.JobContactMade.ContactedStatus,
			&visit.JobPosting.ID, &visit.JobPosting.CompanyID, &visit.JobPosting.JobTitle, &visit.JobPosting.JobDescription, &visit.JobPosting.NumberOfOpenings,
			&visit.JobPosting.JobLocation, &visit.JobPosting.JobType, &visit.JobPosting.SalaryRange, &visit.JobPosting.RequiredExperience,
			&visit.JobPosting.RequiredQualifications, &visit.JobPosting.SkillsRequired, &visit.JobPosting.PostingDate,
			&visit.JobPosting.ApplicationDeadline, &visit.JobPosting.JobPostingStatus,
		); err != nil {
			log.Printf("Database scan error: %v", err)
			http.Error(w, `{"message": "Failed to process job visit data"}`, http.StatusInternalServerError)
			return
		}
		visits = append(visits, visit)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		http.Error(w, `{"message": "Error reading job visit results"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visits)
}
