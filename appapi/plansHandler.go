package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	localmodel "achievesomethingbro/appmodel"
	"encoding/json"
	"log"
	"net/http"
)

func JobseekerPlansHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Print("jobseeker get")
		getJobseekerPlans(w, r)
	case http.MethodPost:
		log.Print("jobseeker post")
		createJobseekerPlan(w, r)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func CompanyPlansHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Print("Company get")
		getCompanyPlans(w, r)
	case http.MethodPost:
		log.Print("Company post")
		createCompanyPlan(w, r)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func AdvertiserPlansHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Print("Advertiser get")
		getAdvertiserPlans(w, r)
	case http.MethodPost:
		log.Print("Advertiser post")
		createAdvertiserPlan(w, r)
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func getJobseekerPlans(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed", http.StatusNotFound)
		return
	}
	query := "SELECT id, name, duration, amount, applies, admin_id, created_time, updated_time FROM jobseeker_plans"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("DB Query Error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var plans []localmodel.JobseekerPlan
	for rows.Next() {
		var plan localmodel.JobseekerPlan
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.Duration, &plan.Amount, &plan.Applies, &plan.AdminID, &plan.CreatedTime, &plan.UpdatedTime); err != nil {
			log.Printf("DB Scan Error: %v", err)
			http.Error(w, `{"error": "Failed to process plan data"}`, http.StatusInternalServerError)
			return
		}
		plans = append(plans, plan)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

func createJobseekerPlan(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	log.Print("Bhaiyaji yahan")
	if err != nil {
		log.Print("Bhaiyaji")
		http.Error(w, "Failed", http.StatusNotFound)
		return
	}
	var newPlan localmodel.JobseekerPlan
	if err := json.NewDecoder(r.Body).Decode(&newPlan); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	log.Print(newPlan.UserType)
	query := "INSERT INTO jobseeker_plans (name, duration, amount, applies, admin_id) VALUES (?, ?, ?, ?, ?)"
	res, err := db.Exec(query, newPlan.Name, newPlan.Duration, newPlan.Amount, newPlan.Applies, newPlan.AdminID)
	if err != nil {
		log.Printf("DB Insert Error: %v", err)
		http.Error(w, `{"error": "Could not create job seeker plan"}`, http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("DB LastInsertId Error: %v", err)
		http.Error(w, `{"error": "Could not retrieve created plan ID"}`, http.StatusInternalServerError)
		return
	}
	newPlan.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	log.Print("Bhaiyaji yahan", newPlan)
	json.NewEncoder(w).Encode(newPlan)
}

func getCompanyPlans(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed", http.StatusNotFound)
		return
	}
	query := "SELECT id, name, duration, amount, postings, resume_visits, admin_id, created_time, updated_time FROM company_plans"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("DB Query Error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var plans []localmodel.CompanyPlan
	for rows.Next() {
		var plan localmodel.CompanyPlan
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.Duration, &plan.Amount, &plan.Postings, &plan.ResumeVisits, &plan.AdminID, &plan.CreatedTime, &plan.UpdatedTime); err != nil {
			log.Printf("DB Scan Error: %v", err)
			http.Error(w, `{"error": "Failed to process plan data"}`, http.StatusInternalServerError)
			return
		}
		plans = append(plans, plan)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

func createCompanyPlan(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed", http.StatusNotFound)
		return
	}
	var newPlan localmodel.CompanyPlan
	if err := json.NewDecoder(r.Body).Decode(&newPlan); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	log.Print(newPlan)

	query := "INSERT INTO company_plans (name, duration, amount, postings, resume_visits, admin_id) VALUES (?, ?, ?, ?, ?, ?)"
	res, err := db.Exec(query, newPlan.Name, newPlan.Duration, newPlan.Amount, newPlan.Postings, newPlan.ResumeVisits, newPlan.AdminID)
	if err != nil {
		log.Printf("DB Insert Error: %v", err)
		http.Error(w, `{"error": "Could not create company plan"}`, http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("DB LastInsertId Error: %v", err)
		http.Error(w, `{"error": "Could not retrieve created plan ID"}`, http.StatusInternalServerError)
		return
	}
	newPlan.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPlan)
}

func getAdvertiserPlans(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed", http.StatusNotFound)
		return
	}
	query := "SELECT id, name, ad_radius, ad_duration, amount, admin_id, created_time, updated_time FROM advertiser_plans"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("DB Query Error: %v", err)
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var plans []localmodel.AdvertiserPlan
	for rows.Next() {
		var plan localmodel.AdvertiserPlan
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.AdRadius, &plan.AdDuration, &plan.Amount, &plan.AdminID, &plan.CreatedTime, &plan.UpdatedTime); err != nil {
			log.Printf("DB Scan Error: %v", err)
			http.Error(w, `{"error": "Failed to process advertiser plan data"}`, http.StatusInternalServerError)
			return
		}
		plans = append(plans, plan)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

func createAdvertiserPlan(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed", http.StatusNotFound)
		return
	}
	var newPlan localmodel.AdvertiserPlan
	if err := json.NewDecoder(r.Body).Decode(&newPlan); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	query := "INSERT INTO advertiser_plans (name, ad_radius, ad_duration, amount, admin_id) VALUES (?, ?, ?, ?, ?)"
	res, err := db.Exec(query, newPlan.Name, newPlan.AdRadius, newPlan.AdDuration, newPlan.Amount, newPlan.AdminID)
	if err != nil {
		log.Printf("DB Insert Error: %v", err)
		http.Error(w, `{"error": "Could not create advertiser plan"}`, http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("DB LastInsertId Error: %v", err)
		http.Error(w, `{"error": "Could not retrieve created plan ID"}`, http.StatusInternalServerError)
		return
	}
	newPlan.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPlan)
}
