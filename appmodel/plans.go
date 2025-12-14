package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"log"
)

// JobseekerPlan defines the structure for a job seeker's subscription plan.
type JobseekerPlan struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Duration    int     `json:"duration"` // Duration in days
	Amount      float64 `json:"amount"`
	Applies     int     `json:"applies"` // -1 signifies unlimited.
	AdminID     int64   `json:"admin_id"`
	UserType    string  `json:"user_type"`
	Status      string  `json:"status"` // ENUM: 'OPEN', 'CLOSED', 'PAUSED'
	CreatedTime string  `json:"created_time,omitempty"`
	UpdatedTime string  `json:"updated_time,omitempty"`
}

// CompanyPlan defines the structure for a company's subscription plan.
type CompanyPlan struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Duration     int     `json:"duration"` // Duration in days
	Amount       float64 `json:"amount"`
	Postings     int     `json:"postings"`     // -1 signifies unlimited.
	ResumeVisits int     `json:"resumeVisits"` // -1 signifies unlimited.
	AdminID      int64   `json:"admin_id"`
	UserType     string  `json:"user_type"`
	Status       string  `json:"status"` // ENUM: 'OPEN', 'CLOSED', 'PAUSED'
	CreatedTime  string  `json:"created_time,omitempty"`
	UpdatedTime  string  `json:"updated_time,omitempty"`
}

// AdvertiserPlan defines the structure for an advertiser's subscription plan.
type AdvertiserPlan struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	AdRadius    int     `json:"ad_radius"`   // Ad visibility radius in km
	AdDuration  int     `json:"ad_duration"` // Ad visibility duration in days
	Amount      float64 `json:"amount"`
	AdminID     int64   `json:"admin_id"`
	Status      string  `json:"status"` // ENUM: 'OPEN', 'CLOSED', 'PAUSED'
	CreatedTime string  `json:"created_time,omitempty"`
	UpdatedTime string  `json:"updated_time,omitempty"`
}

// createTables executes SQL statements to create the necessary tables.
func CreatePlanTables() error {
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Print("error in db")
		return nil
	}
	jobseekerTableQuery := `
    CREATE TABLE IF NOT EXISTS jobseeker_plans (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE,
        duration INT NOT NULL,
        amount DECIMAL(10, 2) NOT NULL,
        applies INT NOT NULL,
        admin_id INT NOT NULL,
		status ENUM('OPEN', 'CLOSED', 'PAUSED') NOT NULL DEFAULT 'OPEN',
        created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uq_jobsk_sub (duration, amount, applies, status)
    );`
	if _, err := db.Exec(jobseekerTableQuery); err != nil {
		return err
	}

	companyTableQuery := `
    CREATE TABLE IF NOT EXISTS company_plans (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE,
        duration INT NOT NULL,
        amount DECIMAL(10, 2) NOT NULL,
        postings INT NOT NULL,
        resume_visits INT NOT NULL,
        admin_id INT NOT NULL,
		status ENUM('OPEN', 'CLOSED', 'PAUSED') NOT NULL DEFAULT 'OPEN',
        created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uq_compsk_sub (duration, amount, postings, resume_visits, status)
    );`
	if _, err := db.Exec(companyTableQuery); err != nil {
		return err
	}

	advertiserTableQuery := `
	CREATE TABLE IF NOT EXISTS advertiser_plans (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		ad_radius INT NOT NULL,
		ad_duration INT NOT NULL,
		amount DECIMAL(10, 2) NOT NULL,
		admin_id INT NOT NULL,
		status ENUM('OPEN', 'CLOSED', 'PAUSED') NOT NULL DEFAULT 'OPEN',
        created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uq_compsk_sub (ad_radius, ad_duration, amount, status)
	);`
	if _, err := db.Exec(advertiserTableQuery); err != nil {
		return err
	}

	return nil
}
