package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver import
)

// PageData struct to pass data to the dashboard template
type PageData struct {
	Users []User
}

// LoginRequest defines the expected structure for the login request body.
// 'UserName' can be either an email or a mobile number.
type LoginRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	UserType string `json:"user_type"` // ENUM: 'JOB_SEEKER', 'RECRUITER', 'ADVERTISER'
}

// User represents the data structure for the 'users' table.
// It contains core information for all user types in the system.
type User struct {
	ID           int    `json:"id,omitempty"`
	UserName     string `json:"user_name"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Gender       string `json:"gender,omitempty"`
	DateOfBirth  string `json:"date_of_birth,omitempty"`
	Password     string `json:"password"`
	UserType     string `json:"user_type"` // ENUM: 'JOB_SEEKER', 'RECRUITER', 'ADVERTISER'
}

// JobSeekerProfile represents the data structure for the 'job_seeker_profiles' table.
type JobSeekerProfile struct {
	ID               int       `json:"id,omitempty"`
	UserId           StringInt `json:"user_id"`
	ProfileSummary   string    `json:"profile_summary"`
	Location         string    `json:"location"` // <-- change json tag from "locations" to "location"
	Currentcompany   string    `json:"current_company"`
	OpenForLocations string    `json:"open_for_locations"`
	SalaryRange      string    `json:"salary_range"`
	WorkEx           string    `json:"work_ex"`
	OverallWorkEx    string    `json:"overall_work_ex"`
	Skills           string    `json:"skills"`
	JobType          string    `json:"job_type"`
	JobTitle         string    `json:"job_title"`
}

// Company represents the data structure for the 'companies' table.
// This table holds data for both recruiters and advertisers.
type Company struct {
	ID             int       `json:"id,omitempty"`
	UserID         StringInt `json:"user_id"`
	CompanyName    string    `json:"company_name"`
	GSTNumber      string    `json:"gst_number,omitempty"`
	CompanyRating  string    `json:"company_rating,omitempty"`
	CompanyType    string    `json:"company_type,omitempty"`
	CompanyAddress string    `json:"company_address,omitempty"`
	CompanyProfile string    `json:"company_profile,omitempty"`
}

// RecruiterPayload is a composite struct for unmarshaling the combined JSON.
type CompanyPayload struct {
	User    // Embedded User struct
	Company // Embedded Company struct
}

// JobPosting represents the data structure for the 'job_postings' table.
type JobPosting struct {
	ID                     StringInt `json:"id,omitempty"`
	CompanyID              StringInt `json:"company_id"`
	JobTitle               string    `json:"job_title"`
	JobDescription         string    `json:"job_description"`
	NumberOfOpenings       int       `json:"number_of_openings"`
	JobLocation            string    `json:"job_location,omitempty"`
	JobType                string    `json:"job_type,omitempty"`
	SalaryRange            string    `json:"salary_range,omitempty"`
	RequiredExperience     string    `json:"required_experience,omitempty"`
	RequiredQualifications string    `json:"required_qualifications,omitempty"`
	SkillsRequired         string    `json:"skills_required,omitempty"`
	PostingDate            time.Time `json:"posting_date"`
	ApplicationDeadline    time.Time `json:"application_deadline,omitempty"`
	JobPostingStatus       string    `json:"job_posting_status"` // ENUM: 'OPEN', 'CLOSED', 'PAUSED'
}

// JobContactMade represents the data structure for the 'job_contacts_made' table.
type JobContactMade struct {
	ID              int       `json:"id,omitempty"`
	JobID           StringInt `json:"job_id"`
	CompanyID       StringInt `json:"company_id"`
	UserID          StringInt `json:"user_id"`
	ContactDate     time.Time `json:"contact_date"`
	ContactedStatus string    `json:"contacted_status"` // ENUM: 'APPLIED', 'PENDING', 'REJECTED', 'SELECTED'
}

type ResumeContactMade struct {
	ID          int       `json:"id,omitempty"`
	UserID      StringInt `json:"user_id"`
	ResumeID    StringInt `json:"resume_id"`
	CompanyID   StringInt `json:"company_id"`
	ContactDate time.Time `json:"contact_date"`
	Status      string    `json:"status"` // ENUM: 'RESUME_CHECKED','CONTACTED', 'PENDING', 'REJECTED', 'SELECTED'
}

// Advertisement represents the data structure for the 'advertisements' table.
type Advertisement struct {
	ID                int       `json:"id,omitempty"`
	CompanyID         StringInt `json:"company_id"`
	AdvertisementType string    `json:"advertisement_type,omitempty"`
	ActiveStatus      string    `json:"active_status"` // ENUM: 'ACTIVE', 'INACTIVE', 'PENDING'
	AdContent         string    `json:"ad_content,omitempty"`
	StartDate         time.Time `json:"start_date,omitempty"`
	EndDate           time.Time `json:"end_date,omitempty"`
	GeoTargeting      string    `json:"geo_targeting,omitempty"`
	Budget            float64   `json:"budget,omitempty"`
}

// CreateAllTables connects to the database and creates all the necessary tables.
func CreateAllTables() {
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close() // Ensure the database connection is closed

	// SQL statements to create all tables.
	// The statements are executed in a logical order to satisfy foreign key constraints.
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        user_name VARCHAR(255) NOT NULL,
        first_name VARCHAR(255),
        last_name VARCHAR(255),
        mobile_number VARCHAR(20) NOT NULL UNIQUE,
        email VARCHAR(255) NOT NULL UNIQUE,
        gender VARCHAR(50),
        date_of_birth DATE,
        password VARCHAR(255) NOT NULL,
        user_type ENUM('JOB_SEEKER', 'RECRUITER', 'ADVERTISER') NOT NULL DEFAULT 'JOB_SEEKER'
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	createTableSQL = `
    CREATE TABLE IF NOT EXISTS job_seeker_profiles (
        id INT AUTO_INCREMENT PRIMARY KEY,
        user_id INT NOT NULL UNIQUE,
        profile_summary TEXT,
        location VARCHAR(255),
        current_company VARCHAR(255),
        open_for_locations VARCHAR(255),
        salary_range VARCHAR(255),
        work_ex VARCHAR(255),
        overall_work_ex VARCHAR(255),
        skills VARCHAR(255),
        job_type VARCHAR(255),
        job_title VARCHAR(255),
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	createTableSQL = `
    CREATE TABLE IF NOT EXISTS companies (
        id INT AUTO_INCREMENT PRIMARY KEY,
        user_id INT NOT NULL UNIQUE,
        company_name VARCHAR(255) NOT NULL,
        gst_number VARCHAR(20),
        company_rating VARCHAR(50),
        company_type VARCHAR(255),
        company_address TEXT,
		company_profile TEXT,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create companies table: %v", err)
	}
	createTableSQL = `

    CREATE TABLE IF NOT EXISTS job_postings (
        id INT AUTO_INCREMENT PRIMARY KEY,
        company_id INT NOT NULL,
        job_title VARCHAR(255) NOT NULL,
        job_description TEXT NOT NULL,
        number_of_openings INT NOT NULL,
		job_location VARCHAR(255),
		job_type VARCHAR(100),
		salary_range VARCHAR(100),
		required_experience VARCHAR(100),
		required_qualifications VARCHAR(255),
		skills_required TEXT,
		posting_date DATE NOT NULL,
		application_deadline DATE,
		job_posting_status ENUM('OPEN', 'CLOSED', 'PAUSED') NOT NULL DEFAULT 'OPEN',
		FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
	);	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create job_postings table: %v", err)
	}

	createTableSQL = `
    CREATE TABLE IF NOT EXISTS job_contacts_made (
        id INT AUTO_INCREMENT PRIMARY KEY,
        job_id INT NOT NULL,
        user_id INT NOT NULL,
		company_id INT NOT NULL,
        contact_date DATE NOT NULL,
		contacted_status ENUM('CONTACTED', 'PENDING', 'REJECTED', 'SELECTED') NOT NULL DEFAULT 'PENDING',
        FOREIGN KEY (job_id) REFERENCES job_postings(id) ON DELETE CASCADE,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE KEY uq_job_user_company (job_id, user_id, company_id)
    );	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create job_contacts_made table: %v", err)
	}
	createTableSQL = `
	CREATE TABLE IF NOT EXISTS resume_contacts_made (
		id INT AUTO_INCREMENT PRIMARY KEY,
		resume_id BIGINT UNSIGNED NOT NULL,
		user_id INT NOT NULL,
		company_id INT NOT NULL,
		contact_date DATE NOT NULL,
		status ENUM('RESUME_CHECKED','CONTACTED', 'PENDING', 'REJECTED', 'SELECTED') NOT NULL DEFAULT 'PENDING',
		FOREIGN KEY (resume_id) REFERENCES user_resumes(id) ON DELETE CASCADE,
		FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE KEY uq_resume_user_company (resume_id, user_id, company_id)
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create resume_contacts_made table: %v", err)
	}

	createTableSQL = `

    CREATE TABLE IF NOT EXISTS advertisements (
        id INT AUTO_INCREMENT PRIMARY KEY,
        company_id INT NOT NULL,
        advertisement_type VARCHAR(255),
        active_status ENUM('ACTIVE', 'INACTIVE', 'PENDING') NOT NULL DEFAULT 'PENDING',
        ad_content TEXT,
        start_date DATE,
        end_date DATE,
		geo_targeting VARCHAR(255),
		budget DECIMAL(10, 2),
		FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create advertisements table: %v", err)
	}

	fmt.Println("Successfully created all tables!")
}
