package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"fmt"
)

// Define the directory where resumes will be saved
const resumeStorageDir = "./resumes"

// --- Data Transfer Objects (DTOs) and Models ---

// Resume represents the metadata stored in the database
type Resume struct {
	Id          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Filename    string `json:"filename"`
	MimeType    string `json:"mime_type"`
	Filepath    string `json:"filepath"` // Path on the server's disk
	UploadedAt  string `json:"uploaded_at"`
	DownloadURL string `json:"download_url,omitempty"` // Used only for frontend response
}

// UploadRequest is the DTO received from the frontend when uploading a file
type UploadRequest struct {
	UserID     int    `json:"user_id"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	Base64Data string `json:"base64_data"` // Contains the full data URI (e.g., data:application/pdf;base64,...)
}

// AnalyzeRequest is the DTO used for the AI analysis endpoint
type AnalyzeRequest struct {
	UserID int `json:"user_id"`
}

// AnalyzeResponse is the DTO returned by the AI analysis endpoint
type AnalyzeResponse struct {
	Summary string `json:"summary"`
	Message string `json:"message"`
}

// --- Core Logic Functions ---

var AllowedMimeTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true, // .doc
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
}

func CreateUserResumeTable() error {
	db, err := dbpg.ConnectDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	query := `
	CREATE TABLE IF NOT EXISTS user_resumes (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL UNIQUE,
		filename VARCHAR(255) NOT NULL,
		mime_type VARCHAR(100) NOT NULL,
		filepath VARCHAR(255) NOT NULL,
		uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create user_resumes table: %w", err)
	}
	return nil
}
