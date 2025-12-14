package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	localmodel "achievesomethingbro/appmodel"
	localservice "achievesomethingbro/services"
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	"github.com/unidoc/unidoc/pdf/extractor"
	"github.com/unidoc/unidoc/pdf/model"
)

// Constants
const resumeStorageDir = "./resumes" // Directory where files will be stored relative to the Go application executable

func MakeDir() {
	// Ensure the storage directory exists on startup
	if err := os.MkdirAll(resumeStorageDir, 0755); err != nil {
		log.Fatalf("Failed to create resume storage directory %s: %v", resumeStorageDir, err)
	}
}

func handleResumeRequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("Resume request: %s %s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodPost:
		uploadResume(w, r)
	case http.MethodGet:
		getResume(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// NewDownloadHandler serves the physical file requested
func handleDownloadResumeRequests(w http.ResponseWriter, r *http.Request) {
	downloadResume(w, r)
}

// NewAnalyzeResumeHandler sets up the handler for AI analysis. func handleAnalyzeResumeRequests(w http.ResponseWriter, r *http.Request) {
func handleAnalyzeResumeRequests(w http.ResponseWriter, r *http.Request) {
	log.Print("Analyze Resume Request", r.Body)
	if r.Method == http.MethodPost {
		analyzeResume(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getResume handles GET requests to retrieve resume metadata and the download URL.
func getResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
	userID := parts[2]

	var resume localmodel.Resume
	query := "SELECT id,filename, mime_type, filepath, uploaded_at FROM user_resumes WHERE user_id = ?"
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, `{"error": "Internal database error"}`, http.StatusInternalServerError)
		return
	}
	err = db.QueryRow(query, userID).Scan(&resume.Id, &resume.Filename, &resume.MimeType, &resume.Filepath, &resume.UploadedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Resume not found: This is expected for new users, return 404
			http.Error(w, `{"error": "Resume not found for this user"}`, http.StatusNotFound)
			return
		}
		log.Printf("Database error fetching resume metadata: %v", err)
		http.Error(w, `{"error": "Internal database error"}`, http.StatusInternalServerError)
		return
	}

	// Construct the dynamic download URL for the client
	resume.DownloadURL = fmt.Sprintf("/resume/download/%s", userID)
	log.Print(resume)
	json.NewEncoder(w).Encode(resume)
	db.Close()
}

// downloadResume handles streaming the actual file content to the client.
func downloadResume(w http.ResponseWriter, r *http.Request) {
	// Expecting URL path format: /resume/download/{userID}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "Missing user ID in URL path", http.StatusBadRequest)
		return
	}
	userID := parts[3]

	var filepath string
	var mimeType string
	var filename string

	query := "SELECT filepath, mime_type, filename FROM user_resumes WHERE user_id = ?"
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = db.QueryRow(query, userID).Scan(&filepath, &mimeType, &filename)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error fetching file path: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set headers for file download/viewing
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename)) // 'inline' streams it in browser

	// Use http.ServeFile to efficiently stream the local file
	http.ServeFile(w, r, filepath)
	db.Close()
}

// uploadResume handles POST requests to save the resume file and update the database.
func uploadResume(w http.ResponseWriter, r *http.Request) {
	log.Printf("Upload Resume Request: %+v", r.Body)
	w.Header().Set("Content-Type", "application/json")

	var req localmodel.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if !localmodel.AllowedMimeTypes[req.MimeType] {
		http.Error(w, `{"error": "Invalid file type. Only PDF and DOCX/DOC are supported."}`, http.StatusBadRequest)
		return
	}

	// 1. Decode Base64 Data URI to Raw Bytes
	re := regexp.MustCompile(`^data:.*?;base64,`)
	base64String := re.ReplaceAllString(req.Base64Data, "")

	fileBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		http.Error(w, `{"error": "Invalid Base64 data provided"}`, http.StatusBadRequest)
		return
	}

	// 2. Determine File Path and Extension
	ext := strings.TrimPrefix(filepath.Ext(req.Filename), ".")
	if ext == "" {
		// Fallback for files without extension (e.g., sometimes DOCX types)
		if strings.Contains(req.MimeType, "pdf") {
			ext = "pdf"
		} else if strings.Contains(req.MimeType, "document") {
			ext = "docx"
		} else {
			ext = "dat" // generic binary fallback
		}
	}

	// Create a unique filename on the server (using UUID + time for safety)
	uniqueFileName := fmt.Sprintf("user_%d_%s_%s.%s", req.UserID, uuid.New().String(), time.Now().Format("20060102"), ext)
	newFilePath := filepath.Join(resumeStorageDir, uniqueFileName)

	// 3. Save the file to disk
	if err := os.WriteFile(newFilePath, fileBytes, 0644); err != nil {
		log.Printf("Error saving file to disk: %v", err)
		http.Error(w, `{"error": "Failed to save file on server"}`, http.StatusInternalServerError)
		return
	}

	// 4. Handle Deletion of Old File (Cleanup)
	var oldFilePath sql.NullString
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		// Attempt to delete the newly saved file to prevent orphaned files
		os.Remove(newFilePath)
		http.Error(w, `{"error": "Internal database error"}`, http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT filepath FROM user_resumes WHERE user_id = ?", req.UserID).Scan(&oldFilePath)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Database lookup error during cleanup: %v", err)
		// Non-critical, continue update but log error
	} else if oldFilePath.Valid && oldFilePath.String != "" {
		if err := os.Remove(oldFilePath.String); err != nil {
			log.Printf("Warning: Failed to delete old resume file %s: %v", oldFilePath.String, err)
		}
	}

	// 5. Update/Insert into Database
	_, err = db.Exec(
		`INSERT INTO user_resumes (user_id, filename, mime_type, filepath, uploaded_at)
		 VALUES (?, ?, ?, ?, NOW())
		 ON DUPLICATE KEY UPDATE filename=?, mime_type=?, filepath=?, uploaded_at=NOW()`,
		req.UserID, req.Filename, req.MimeType, newFilePath,
		req.Filename, req.MimeType, newFilePath,
	)

	if err != nil {
		// If DB update fails, attempt to delete the newly saved file to prevent orphaned files
		os.Remove(newFilePath)
		log.Printf("Database error saving resume metadata: %v", err)
		http.Error(w, `{"error": "Failed to save file metadata to database"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Resume uploaded successfully", "filepath": newFilePath})
	db.Close()
}

// analyzeResume handles POST requests to simulate sending resume content for parsing.
func analyzeResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// time.Sleep(5 * time.Second) // Simulate processing delay
	var req localmodel.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.UserID == 0 {
		http.Error(w, `{"error": "User ID is required"}`, http.StatusBadRequest)
		return
	}

	// 1. Retrieve resume metadata (only need to confirm existence for the mock)
	var resume localmodel.Resume
	query := "SELECT filepath FROM user_resumes WHERE user_id = ?"
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Database connection error: %v", err)
		http.Error(w, `{"error": "Internal database error"}`, http.StatusInternalServerError)
		return
	}
	err = db.QueryRow(query, req.UserID).Scan(&resume.Filepath)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Resume not found: Prevent analysis if no file exists
			http.Error(w, `{"error": "Resume not found. Please upload a resume first."}`, http.StatusNotFound)
			return
		}
		log.Printf("DB error getting resume path for analysis: %v", err)
		http.Error(w, `{"error": "Internal server error during data retrieval"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("Found resume for UserID %d: %s", req.UserID, resume.Filename)
	// --- Simple Mock Analysis ---
	// Generate a simulated, structured summary without using the Gemini API.
	extractedText, err := extractTextFromPDF(req.UserID, resume.Filepath)
	if err != nil {
		log.Printf("Error extracting text from PDF: %v", err)
		http.Error(w, `{"error": "Failed to extract text from resume PDF"}`, http.StatusInternalServerError)
		return
	}
	// eSResumeResponse, err := SearchResumeKeywords(dbpg.EsClient, "Kubernetes")
	eSSkillsResponse, err := SearchSkills(dbpg.EsClient, "Kubernetes")
	if err != nil {
		log.Printf("Error searching skills in Elasticsearch: %v", err)
		http.Error(w, `{"error": "Failed to search skills in resume"}`, http.StatusInternalServerError)
		return
	}
	// Loop through the slice of individual results
	// for i, hit := range eSResumeResponse.Hits.Hits {
	// The data is contained within the '_source' field, which is unmarshalled
	// into the 'Source' field of the hit struct (type ResumeAnalysisResult).

	// analysisResult := hit.Source

	// log.Printf("\n--- Result %d ---\n", i+1)
	// log.Printf("Years of Experience: %s\n", analysisResult.YearsExp)
	// log.Printf("Tech Stack Snippet: %s\n", analysisResult.KeyTechStack)
	// log.Printf("Summary: %s\n", analysisResult.Summary)
	// Example: Accessing a nested list of projects
	// for _, project := range analysisResult.Projects {
	// 	fmt.Printf("- Project Title: %s\n", project.Title)
	// }

	// 		curl -XGET "http://localhost:9200/resumes_analysis/_search?pretty" -H 'Content-Type: application/json' -d'
	// {
	//   "query": {
	//     "match": {
	//       "skills": "Kubernetes"
	//     }
	//   }
	// }'
	// }
	if err := localservice.NotifyStakeholders(db, req.UserID, eSSkillsResponse); err != nil {
		// Log the error but continue, as notification failure shouldn't fail the whole upload
		log.Printf("Warning: Failed to notify stakeholders: %v", err)
	}
	// log.Print("eSSkillsResponse: ", eSSkillsResponse)
	// 5. Return the mock summary
	json.NewEncoder(w).Encode(localmodel.AnalyzeResponse{
		Summary: extractedText,
		Message: "Resume analysis simulated successfully (Gemini integration bypassed).",
	})
	db.Close()
}

// fetchGeminiResumeAnalysis (REMOVED: Functionality is now simulated in analyzeResume)
// The function below is kept to avoid breaking the compilation if it was defined in a separate file,
// but since this is a monolithic file, it is removed.
// Since the prompt requires replacing the Gemini code, we remove the function body.
// We remove the entire function to clean up the code.
// func fetchGeminiResumeAnalysis(fileContent string) (string, error) {
// 	// Placeholder for actual Gemini API integration
// 	// This function is no longer used as per the latest requirements
// 	return "", nil
// }

// --- Text Extraction Helper Functions ---

// extractTextFromPDF attempts to use unidoc to get text from the PDF file.
func extractTextFromPDF(userId int, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("could not open PDF file: %w", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return "", fmt.Errorf("could not create PDF reader: %w", err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("could not get page count: %w", err)
	}

	var extractedText strings.Builder
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Printf("Warning: Failed to get page %d for PDF extraction: %v", i, err)
			continue
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Printf("Warning: Failed to create extractor for page %d: %v", i, err)
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			log.Printf("Warning: Failed to extract text from page %d: %v", i, err)
			continue
		}
		extractedText.WriteString(text)
		extractedText.WriteString("\n\n")
	}

	if extractedText.Len() == 0 {
		return "", errors.New("no text could be extracted from PDF")
	}
	log.Print("EsClient: ", dbpg.EsClient)
	IndexResumeKeywords(dbpg.EsClient, userId, extractedText.String())
	return extractedText.String(), nil
}

// // extractTextFromDOCX attempts to use go-docx to get text from the DOCX file.
// func extractTextFromDOCX(filePath string) (string, error) {
// 	doc, err := docx.Open(filePath)
// 	if err != nil {
// 		return "", fmt.Errorf("could not open DOCX file: %w", err)
// 	}

// 	text := doc.Text()
// 	if text == "" {
// 		return "", errors.New("no text could be extracted from DOCX")
// 	}
// 	return text, nil
// }

// --- Core Elasticsearch Functions ---

// IndexResumeKeywords takes extracted text and indexes it into Elasticsearch.
func IndexResumeKeywords(es *elasticsearch.Client, userId int, text string) error {
	// 1. Prepare the data payload
	// NOTE: In
	// a real app, this is where NLP and data transformation would happen
	// to derive 'skills' and create the 'summary' fields.

	data := localmodel.ResumeIndexData{
		UserID:        userId,
		ExtractedText: text,
		Skills:        "Go, Kubernetes, Docker, MySQL", // Simplified/derived for example
	}

	// 2. Marshal the struct into a JSON body
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	documentID := strconv.Itoa(userId)
	// 3. Create the Index Request
	req := esapi.IndexRequest{
		Index:      localmodel.ResumeIndexName,
		DocumentID: documentID, // Use UserID as the unique document ID
		Body:       bytes.NewReader(body),
		Refresh:    "true", // Make the document immediately searchable
	}
	log.Printf("Indexing resume for user ID %d into index %s request: %+v", userId, localmodel.ResumeIndexName, req)
	// 4. Execute the Request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Printf("ERROR: Failed to index document for user %d: %v", userId, err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New("elasticsearch indexing error: " + res.Status())
	}

	log.Printf("INFO: Successfully indexed resume for user ID %d. Status: %s", userId, res.Status())
	return nil
}

// SearchResumeKeywords performs a search query against the Elasticsearch resume index.
func SearchResumeKeywords(es *elasticsearch.Client, queryStr string) (*localmodel.ESResumeResponse, error) {
	// 1. Construct the Multi-Match Query DSL
	// This simulates searching for the query string across 'extracted_text' and 'skills' fields
	esQuery := localmodel.ESQueryContainer{
		Query: localmodel.ESQuery{
			Bool: localmodel.ESMust{
				Must: []localmodel.ESMultiMatch{
					{
						MultiMatch: map[string]interface{}{
							"query":  queryStr,
							"fields": []string{"extracted_text", "skills"},
						},
					},
				},
			},
		},
	}

	// 2. Marshal the query into JSON body
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, err
	}

	// 3. Create the Search Request
	req := esapi.SearchRequest{
		Index: []string{localmodel.ResumeIndexName},
		Body:  &buf,
	}

	// 4. Execute the Request
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Printf("ERROR: Failed to execute search query: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		// Log and return Elasticsearch search error
		return nil, errors.New("elasticsearch search error: " + res.Status())
	}

	// 5. Parse the Response
	var esResponse localmodel.ESResumeResponse
	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		log.Printf("ERROR: Failed to parse search response: %v", err)
		return nil, err
	}

	return &esResponse, nil
}

// SearchSkills constructs and executes a multi-match query against Elasticsearch
// for a specific skill and returns the raw JSON response or an error message.
// The raw JSON is returned as a string to be parsed later or sent directly.
func SearchSkills(es *elasticsearch.Client, querySkill string) (string, error) {
	if es == nil {
		return "", errors.New("elasticsearch client is nil")
	}
	if querySkill == "" {
		return "", errors.New("search query skill cannot be empty")
	}

	// 1. Build the Elasticsearch Query Body
	// This replicates the cURL JSON query: { "query": { "match": { "skills": "Kubernetes" } } }
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"skills": querySkill,
			},
		},
	}

	// Marshal the Go map into a JSON byte array
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query to JSON: %w", err)
	}

	// 2. Prepare the Search Request
	req := esapi.SearchRequest{
		Index:  []string{localmodel.ResumeIndexName},
		Body:   bytes.NewReader(queryJSON),
		Pretty: true, // Optional: makes the output readable (like in the curl example)
	}

	// 3. Execute the Search
	res, err := req.Do(context.Background(), es)
	if err != nil {
		return "", fmt.Errorf("failed to execute search request: %w", err)
	}
	defer res.Body.Close()

	// 4. Handle the Response
	if res.IsError() {
		// Read the error response body
		body, _ := io.ReadAll(res.Body)
		log.Printf("Elasticsearch search error [%s]: %s", res.Status(), body)
		return string(body), fmt.Errorf("elasticsearch search error: %s", res.Status())
	}

	// 5. Read the successful response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read successful response body: %w", err)
	}

	// Return the raw JSON string
	return string(body), nil
}
