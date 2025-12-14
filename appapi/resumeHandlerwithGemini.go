package signupapiv1

// import (
// 	"context"
// 	"database/sql"
// 	"encoding/base64"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"regexp"
// 	"strings"
// 	"time"

// 	"github.com/google/uuid"
// )

// // Define the directory where resumes will be saved
// const resumeStorageDir = "./resumes"

// // --- Data Transfer Objects (DTOs) and Models ---

// // Resume represents the metadata stored in the database
// type Resume struct {
// 	UserID      int    `json:"user_id"`
// 	Filename    string `json:"filename"`
// 	MimeType    string `json:"mime_type"`
// 	Filepath    string `json:"filepath"` // Path on the server's disk
// 	UploadedAt  string `json:"uploaded_at"`
// 	DownloadURL string `json:"download_url,omitempty"` // Used only for frontend response
// }

// // UploadRequest is the DTO received from the frontend when uploading a file
// type UploadRequest struct {
// 	UserID     int    `json:"user_id"`
// 	Filename   string `json:"filename"`
// 	MimeType   string `json:"mime_type"`
// 	Base64Data string `json:"base64_data"` // Contains the full data URI (e.g., data:application/pdf;base64,...)
// }

// // AnalyzeRequest is the DTO used for the AI analysis endpoint
// type AnalyzeRequest struct {
// 	UserID int `json:"user_id"`
// }

// // AnalyzeResponse is the DTO returned by the AI analysis endpoint
// type AnalyzeResponse struct {
// 	Summary string `json:"summary"`
// 	Message string `json:"message"`
// }

// // --- Handler Setup ---

// // NewResumeHandler creates a handler function for the /resume endpoint.
// func NewResumeHandler(db *sql.DB) http.HandlerFunc {
// 	// Ensure the storage directory exists on startup
// 	if err := os.MkdirAll(resumeStorageDir, 0755); err != nil {
// 		log.Fatalf("Failed to create resume storage directory: %v", err)
// 	}

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		switch r.Method {
// 		case http.MethodPost:
// 			uploadResume(w, r, db)
// 		case http.MethodGet:
// 			getResume(w, r, db)
// 		default:
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		}
// 	}
// }

// // NewDownloadHandler serves the physical file requested by the client.
// func NewDownloadHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		downloadResume(w, r, db)
// 	}
// }

// // NewAnalyzeResumeHandler sets up the handler for AI analysis.
// func NewAnalyzeResumeHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method == http.MethodPost {
// 			analyzeResume(w, r, db)
// 		} else {
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		}
// 	}
// }

// // --- Core Logic Functions ---

// var allowedMimeTypes = map[string]bool{
// 	"application/pdf":    true,
// 	"application/msword": true, // .doc
// 	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
// }

// // getResume handles GET requests to retrieve resume metadata and the download URL.
// func getResume(w http.ResponseWriter, r *http.Request, db *sql.DB) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// Expecting URL path format: /resume/{userID}
// 	parts := strings.Split(r.URL.Path, "/")
// 	if len(parts) < 3 || parts[2] == "" {
// 		http.Error(w, `{"error": "Missing user ID in URL path"}`, http.StatusBadRequest)
// 		return
// 	}
// 	userID := parts[2]

// 	var resume Resume
// 	query := "SELECT filename, mime_type, filepath, uploaded_at FROM user_resumes WHERE user_id = ?"
// 	err := db.QueryRow(query, userID).Scan(&resume.Filename, &resume.MimeType, &resume.Filepath, &resume.UploadedAt)

// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			// Resume not found: This is expected for new users, return 404
// 			http.Error(w, `{"error": "Resume not found for this user"}`, http.StatusNotFound)
// 			return
// 		}
// 		log.Printf("Database error fetching resume metadata: %v", err)
// 		http.Error(w, `{"error": "Internal database error"}`, http.StatusInternalServerError)
// 		return
// 	}

// 	// Construct the dynamic download URL for the client
// 	resume.DownloadURL = fmt.Sprintf("/resume/download/%s", userID)

// 	json.NewEncoder(w).Encode(resume)
// }

// // downloadResume handles streaming the actual file content to the client.
// func downloadResume(w http.ResponseWriter, r *http.Request, db *sql.DB) {
// 	// Expecting URL path format: /resume/download/{userID}
// 	parts := strings.Split(r.URL.Path, "/")
// 	if len(parts) < 4 || parts[3] == "" {
// 		http.Error(w, "Missing user ID in URL path", http.StatusBadRequest)
// 		return
// 	}
// 	userID := parts[3]

// 	var filepath string
// 	var mimeType string
// 	var filename string

// 	query := "SELECT filepath, mime_type, filename FROM user_resumes WHERE user_id = ?"
// 	err := db.QueryRow(query, userID).Scan(&filepath, &mimeType, &filename)

// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			http.Error(w, "File not found", http.StatusNotFound)
// 			return
// 		}
// 		log.Printf("Database error fetching file path: %v", err)
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	// Set headers for file download/viewing
// 	w.Header().Set("Content-Type", mimeType)
// 	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename)) // 'inline' streams it in browser

// 	// Use http.ServeFile to efficiently stream the local file
// 	http.ServeFile(w, r, filepath)
// }

// // uploadResume handles POST requests to save the resume file and update the database.
// func uploadResume(w http.ResponseWriter, r *http.Request, db *sql.DB) {
// 	w.Header().Set("Content-Type", "application/json")

// 	var req UploadRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
// 		return
// 	}

// 	if !allowedMimeTypes[req.MimeType] {
// 		http.Error(w, `{"error": "Invalid file type. Only PDF and DOCX/DOC are supported."}`, http.StatusBadRequest)
// 		return
// 	}

// 	// 1. Decode Base64 Data URI to Raw Bytes
// 	re := regexp.MustCompile(`^data:.*?;base64,`)
// 	base64String := re.ReplaceAllString(req.Base64Data, "")

// 	fileBytes, err := base64.StdEncoding.DecodeString(base64String)
// 	if err != nil {
// 		http.Error(w, `{"error": "Invalid Base64 data provided"}`, http.StatusBadRequest)
// 		return
// 	}

// 	// 2. Determine File Path and Extension
// 	ext := strings.TrimPrefix(filepath.Ext(req.Filename), ".")
// 	if ext == "" {
// 		// Fallback for files without extension (e.g., sometimes DOCX types)
// 		if strings.Contains(req.MimeType, "pdf") {
// 			ext = "pdf"
// 		} else if strings.Contains(req.MimeType, "document") {
// 			ext = "docx"
// 		} else {
// 			ext = "dat" // generic binary fallback
// 		}
// 	}

// 	// Create a unique filename on the server (using UUID + time for safety)
// 	uniqueFileName := fmt.Sprintf("user_%d_%s_%s.%s", req.UserID, uuid.New().String(), time.Now().Format("20060102"), ext)
// 	newFilePath := filepath.Join(resumeStorageDir, uniqueFileName)

// 	// 3. Save the file to disk
// 	if err := os.WriteFile(newFilePath, fileBytes, 0644); err != nil {
// 		log.Printf("Error saving file to disk: %v", err)
// 		http.Error(w, `{"error": "Failed to save file on server"}`, http.StatusInternalServerError)
// 		return
// 	}

// 	// 4. Handle Deletion of Old File (Cleanup)
// 	var oldFilePath sql.NullString
// 	err = db.QueryRow("SELECT filepath FROM user_resumes WHERE user_id = ?", req.UserID).Scan(&oldFilePath)

// 	if err != nil && !errors.Is(err, sql.ErrNoRows) {
// 		log.Printf("Database lookup error during cleanup: %v", err)
// 		// Non-critical, continue update but log error
// 	} else if oldFilePath.Valid && oldFilePath.String != "" {
// 		if err := os.Remove(oldFilePath.String); err != nil {
// 			log.Printf("Warning: Failed to delete old resume file %s: %v", oldFilePath.String, err)
// 		}
// 	}

// 	// 5. Update/Insert into Database
// 	_, err = db.Exec(
// 		`INSERT INTO user_resumes (user_id, filename, mime_type, filepath, uploaded_at)
// 		 VALUES (?, ?, ?, ?, NOW())
// 		 ON DUPLICATE KEY UPDATE filename=?, mime_type=?, filepath=?, uploaded_at=NOW()`,
// 		req.UserID, req.Filename, req.MimeType, newFilePath,
// 		req.Filename, req.MimeType, newFilePath,
// 	)

// 	if err != nil {
// 		// If DB update fails, attempt to delete the newly saved file to prevent orphaned files
// 		os.Remove(newFilePath)
// 		log.Printf("Database error saving resume metadata: %v", err)
// 		http.Error(w, `{"error": "Failed to save file metadata to database"}`, http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Resume uploaded successfully", "filepath": newFilePath})
// }

// // analyzeResume handles POST requests to send resume content to the Gemini API for parsing.
// func analyzeResume(w http.ResponseWriter, r *http.Request, db *sql.DB) {
// 	w.Header().Set("Content-Type", "application/json")

// 	var req AnalyzeRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
// 		return
// 	}

// 	if req.UserID == 0 {
// 		http.Error(w, `{"error": "User ID is required"}`, http.StatusBadRequest)
// 		return
// 	}

// 	// 1. Retrieve resume metadata (filepath and mime_type)
// 	var resume Resume
// 	query := "SELECT mime_type, filepath FROM user_resumes WHERE user_id = ?"
// 	err := db.QueryRow(query, req.UserID).Scan(&resume.MimeType, &resume.Filepath)

// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			http.Error(w, `{"error": "Resume not found. Please upload a resume first."}`, http.StatusNotFound)
// 			return
// 		}
// 		log.Printf("DB error getting resume path for analysis: %v", err)
// 		http.Error(w, `{"error": "Internal server error during data retrieval"}`, http.StatusInternalServerError)
// 		return
// 	}

// 	// 2. Read the file content from disk
// 	fileBytes, err := os.ReadFile(resume.Filepath)
// 	if err != nil {
// 		log.Printf("Error reading file from disk for analysis: %v", err)
// 		http.Error(w, `{"error": "Could not read resume file"}`, http.StatusInternalServerError)
// 		return
// 	}

// 	// 3. Encode file content to Base64 (required for Gemini API inline data)
// 	base64Data := base64.StdEncoding.EncodeToString(fileBytes)

// 	// 4. Define the prompt and system instruction for the AI
// 	systemInstruction := `
// 	You are a world-class HR Analyst specializing in quick resume screening.
// 	Your task is to analyze the provided document content (which is a resume) and extract key professional data points.

// 	Provide a structured summary, ideal for a company to quickly evaluate the candidate.
// 	Focus only on the following sections:
// 	1. Years of Experience (Estimate based on dates or keywords)
// 	2. Key Technical Skills/Tech Stack (List as bullet points)
// 	3. Summary of Projects Worked On (Provide brief 1-2 sentence descriptions for the top 3 projects)

// 	Format the output using clear markdown structure and be extremely concise.
// 	`

// 	userQuery := "Analyze the attached resume and generate the structured HR summary."

// 	// --- Gemini API Call Setup ---

// 	summary, err := fetchGeminiResumeAnalysis(base64Data, resume.MimeType, userQuery, systemInstruction)
// 	if err != nil {
// 		log.Printf("Gemini API call failed: %v", err)
// 		http.Error(w, `{"error": "AI analysis failed: Check API Key/Network"}`, http.StatusInternalServerError)
// 		return
// 	}

// 	// 5. Return the AI-generated summary
// 	json.NewEncoder(w).Encode(AnalyzeResponse{
// 		Summary: summary,
// 		Message: "Resume analysis successful.",
// 	})
// }

// // fetchGeminiResumeAnalysis performs the necessary network call to the Gemini API.
// func fetchGeminiResumeAnalysis(base64Data, mimeType, userQuery, systemInstruction string) (string, error) {
// 	// IMPORTANT: API_KEY should be loaded securely from environment variables, not hardcoded.
// 	// We use an empty string here as per instructions for the runtime environment.
// 	// NOTE: If running locally, ensure GEMINI_API_KEY is set in your environment.
// 	apiKey := os.Getenv("GEMINI_API_KEY")

// 	if apiKey == "" {
// 		return "", errors.New("GEMINI_API_KEY environment variable is not set. Cannot perform AI analysis.")
// 	}

// 	apiUrl := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-05-20:generateContent?key=" + apiKey

// 	payload := map[string]interface{}{
// 		"contents": []map[string]interface{}{
// 			{
// 				"role": "user",
// 				"parts": []map[string]string{
// 					{"text": userQuery},
// 					{
// 						"inlineData": map[string]string{
// 							"mimeType": mimeType,
// 							"data":     base64Data,
// 						},
// 					},
// 				},
// 			},
// 		},
// 		"systemInstruction": map[string]interface{}{
// 			"parts": []map[string]string{{"text": systemInstruction}},
// 		},
// 	}

// 	payloadBytes, _ := json.Marshal(payload)

// 	req, err := http.NewRequestWithContext(context.Background(), "POST", apiUrl, strings.NewReader(string(payloadBytes)))
// 	if err != nil {
// 		return "", err
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{Timeout: 30 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	body, _ := io.ReadAll(resp.Body)

// 	if resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
// 	}

// 	var result struct {
// 		Candidates []struct {
// 			Content struct {
// 				Parts []struct {
// 					Text string `json:"text"`
// 				} `json:"parts"`
// 			} `json:"content"`
// 		} `json:"candidates"`
// 	}

// 	if err := json.Unmarshal(body, &result); err != nil {
// 		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
// 	}

// 	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
// 		return result.Candidates[0].Content.Parts[0].Text, nil
// 	}

// 	return "Analysis failed to produce structured output.", nil
// }
