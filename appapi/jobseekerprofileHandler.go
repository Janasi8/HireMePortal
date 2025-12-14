package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	localmodel "achievesomethingbro/appmodel"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func HandleJobSeekerProfileRequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("JobSeekerProfile request: %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 || parts[3] == "all" {
			GetAllProfiles(w, r)
		} else if len(parts) < 4 || parts[3] != "" {
			GetProfile(w, r)
		}
	case http.MethodPost:
		CreateProfile(w, r)
	case http.MethodPut:
		UpdateProfile(w, r)
	case http.MethodDelete:
		DeleteProfile(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetAllProfiles(w http.ResponseWriter, r *http.Request) {
	var profiles []localmodel.JobSeekerProfile
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := db.Query("SELECT id, user_id, profile_summary, location, current_company, open_for_locations, salary_range, work_ex, overall_work_ex, skills, job_type, job_title FROM job_seeker_profiles")
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		return
	}
	for rows.Next() {
		var profile localmodel.JobSeekerProfile
		// Scan logic is the same as in GetJobDescriptions
		err := rows.Scan(
			&profile.ID, &profile.UserId, &profile.ProfileSummary, &profile.Location,
			&profile.Currentcompany, &profile.OpenForLocations, &profile.SalaryRange,
			&profile.WorkEx, &profile.OverallWorkEx, &profile.Skills, &profile.JobType,
			&profile.JobTitle)
		if err != nil {
			log.Printf("Database error scanning row: %v", err)
			continue // Continue to next row on scan error
		}
		profiles = append(profiles, profile)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Database iteration error: %v", err)
	}

	if err == sql.ErrNoRows {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print(profiles)
	json.NewEncoder(w).Encode(profiles)
	db.Close()
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[3] == "" {
		http.Error(w, `{"error": "Missing user ID in URL path"}`, http.StatusBadRequest)
		return
	}
	userID := parts[3]
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	var profile localmodel.JobSeekerProfile
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print(profile)
	err = db.QueryRow("SELECT id, user_id, profile_summary, location, current_company, open_for_locations, salary_range, work_ex, overall_work_ex, skills, job_type, job_title FROM job_seeker_profiles WHERE user_id = ?", userID).Scan(
		&profile.ID, &profile.UserId, &profile.ProfileSummary, &profile.Location,
		&profile.Currentcompany, &profile.OpenForLocations, &profile.SalaryRange,
		&profile.WorkEx, &profile.OverallWorkEx, &profile.Skills, &profile.JobType,
		&profile.JobTitle)

	if err == sql.ErrNoRows {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print(profile)
	json.NewEncoder(w).Encode(profile)
	db.Close()
}

func CreateProfile(w http.ResponseWriter, r *http.Request) {
	var profile localmodel.JobSeekerProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var userProfileID int
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print(profile)
	err = db.QueryRow("SELECT id FROM job_seeker_profiles WHERE user_id = ?", profile.UserId).Scan(
		&userProfileID)

	if err == sql.ErrNoRows {
		log.Print(profile)

		// Remove RETURNING id for MySQL, use Exec and LastInsertId
		res, err := db.Exec(`
			INSERT INTO job_seeker_profiles 
			(user_id, profile_summary, location, current_company, open_for_locations, 
			salary_range, work_ex, overall_work_ex, skills, job_type, job_title)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			profile.UserId, profile.ProfileSummary, profile.Location, profile.Currentcompany,
			profile.OpenForLocations, profile.SalaryRange, profile.WorkEx, profile.OverallWorkEx,
			profile.Skills, profile.JobType, profile.JobTitle)

		if err != nil {
			log.Printf("Error inserting profile: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lastID, _ := res.LastInsertId()
		profile.ID = int(lastID)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(profile)
	} else {
		_, err := db.Exec(`
			UPDATE job_seeker_profiles SET 
			profile_summary=?, location=?, current_company=?, open_for_locations=?,
			salary_range=?, work_ex=?, overall_work_ex=?, skills=?, job_type=?,
			job_title=? WHERE id=?`,
			profile.ProfileSummary, profile.Location, profile.Currentcompany,
			profile.OpenForLocations, profile.SalaryRange, profile.WorkEx,
			profile.OverallWorkEx, profile.Skills, profile.JobType, profile.JobTitle,
			userProfileID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(profile)
	}
	db.Close()
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var profile localmodel.JobSeekerProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result, err := db.Exec(`
		UPDATE job_seeker_profiles SET 
		profile_summary=?, location=?, current_company=?, open_for_locations=?,
		salary_range=?, work_ex=?, overall_work_ex=?, skills=?, job_type=?,
		job_title=? WHERE user_id=?`,
		profile.ProfileSummary, profile.Location, profile.Currentcompany,
		profile.OpenForLocations, profile.SalaryRange, profile.WorkEx,
		profile.OverallWorkEx, profile.Skills, profile.JobType, profile.JobTitle,
		profile.UserId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
	db.Close()
}

func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("DELETE FROM job_seeker_profiles WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	db.Close()
}

// func ShowResumeCheckedHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodGet {
// 		ShowResumeChecked(w, r)
// 	} else {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// func JobAccessedHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodPost {
// 		RecordJobProfileAccessed(w, r)
// 	} else {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }
// func RecordJobProfileAccessed(w http.ResponseWriter, r *http.Request) {
// 	var access localmodel.ResumeContactMade
// 	if err := json.NewDecoder(r.Body).Decode(&access); err != nil {
// 		log.Printf("Error decoding request body: %v", err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	log.Print(access)
// 	db, err := dbpg.ConnectDB()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	var userId int
// 	db.QueryRow(`
// 		select user_id from job_seeker_profiles where id = ?`, access.ResumeID).Scan(&userId)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	// Remove RETURNING id for MySQL, use Exec and LastInsertId
// 	res, err := db.Exec(`
// 		INSERT INTO resume_contact_made
// 		(resume_id, user_id, company_id, contact_date, status)
// 		VALUES (?, ?, ?, ?, ?)`,
// 		access.ResumeID, userId, access.CompanyID, access.ContactDate, access.Status)

// 	if err != nil {
// 		log.Printf("Error inserting profile: %v", err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	lastID, _ := res.LastInsertId()
// 	access.ID = int(lastID)

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(access)
// 	db.Close()
// // }
// func ShowResumeChecked(w http.ResponseWriter, r *http.Request) {
// 	parts := strings.Split(r.URL.Path, "/")
// 	if len(parts) < 3 || parts[3] == "" {
// 		http.Error(w, `{"error": "Missing user ID in URL path"}`, http.StatusBadRequest)
// 		return
// 	}
// 	userID := parts[3]
// 	if userID == "" {
// 		http.Error(w, "user_id is required", http.StatusBadRequest)
// 		return
// 	}

// 	var resumes []localmodel.ResumeContactMade
// 	db, err := dbpg.ConnectDB()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	rows, err := db.Query("SELECT id, resume_id, company_id, contact_date, status FROM resume_contact_made WHERE user_id = ? and contact_date > ?", userID, time.Now()-N days.Format("2006-01-02"))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var resume localmodel.ResumeContactMade
// 		if err := rows.Scan(&resume.ID, &resume.ResumeID, &resume.CompanyID, &resume.ContactDate, &resume.Status); err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		resumes = append(resumes, resume)
// 	}

// 	if err := rows.Err(); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(resumes)
// 	db.Close()
// }

// findContactsOlderThanNDaysForUser retrieves contact records for a specific user made more than n days ago.
// func findContactsOlderThanNDaysForUser(db *sql.DB, userID int, daysAgo int) ([]ResumeContact, error) {
// 	cutoffDate := time.Now().AddDate(0, 0, -daysAgo)
// 	query := `
//         SELECT id, resume_id, company_id, contact_date, status
//         FROM resume_contact_made
//         WHERE user_id = $1 AND contact_date < $2;
//     `
// 	rows, err := db.Query(query, userID, cutoffDate)
// 	if err != nil {
// 		return nil, fmt.Errorf("database query failed: %w", err)
// 	}
// 	defer rows.Close()

// 	var contacts []ResumeContact
// 	for rows.Next() {
// 		var contact ResumeContact
// 		if err := rows.Scan(&contact.ID, &contact.ResumeID, &contact.CompanyID, &contact.ContactDate, &contact.Status); err != nil {
// 			return nil, fmt.Errorf("failed to scan row: %w", err)
// 		}
// 		contacts = append(contacts, contact)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error during row iteration: %w", err)
// 	}

// 	return contacts, nil
// }

// // findContactsNewerThanNDaysForUser retrieves contact records for a specific user made within the last n days.
// func findContactsNewerThanNDaysForUser(db *sql.DB, userID int, daysAgo int) ([]ResumeContact, error) {
// 	cutoffDate := time.Now().AddDate(0, 0, -daysAgo)
// 	query := `
//         SELECT id, resume_id, company_id, contact_date, status
//         FROM resume_contact_made
//         WHERE user_id = $1 AND contact_date > $2;
//     `
// 	rows, err := db.Query(query, userID, cutoffDate)
// 	if err != nil {
// 		return nil, fmt.Errorf("database query failed: %w", err)
// 	}
// 	defer rows.Close()

// 	var contacts []ResumeContact
// 	for rows.Next() {
// 		var contact ResumeContact
// 		if err := rows.Scan(&contact.ID, &contact.ResumeID, &contact.CompanyID, &contact.ContactDate, &contact.Status); err != nil {
// 			return nil, fmt.Errorf("failed to scan row: %w", err)
// 		}
// 		contacts = append(contacts, contact)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error during row iteration: %w", err)
// 	}

// 	return contacts, nil
// }

// // --- NEW DEMONSTRATION CODE ---

// // filterContactsOlderThanNDays is a mock version of the DB function that works on a slice.
// func filterContactsOlderThanNDays(allContacts []ResumeContact, userID int, daysAgo int) []ResumeContact {
// 	cutoffDate := time.Now().AddDate(0, 0, -daysAgo)
// 	var results []ResumeContact
// 	for _, contact := range allContacts {
// 		if contact.UserID == userID && contact.ContactDate.Before(cutoffDate) {
// 			results = append(results, contact)
// 		}
// 	}
// 	return results
// }

// // filterContactsNewerThanNDays is a mock version of the DB function that works on a slice.
// func filterContactsNewerThanNDays(allContacts []ResumeContact, userID int, daysAgo int) []ResumeContact {
// 	cutoffDate := time.Now().AddDate(0, 0, -daysAgo)
// 	var results []ResumeContact
// 	for _, contact := range allContacts {
// 		if contact.UserID == userID && contact.ContactDate.After(cutoffDate) {
// 			results = append(results, contact)
// 		}
// 	}
// 	return results
// }
