package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	userModel "achievesomethingbro/appmodel"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// handleGetUser is the HTTP handler for the /user endpoint.
// It retrieves a user's details by their username.
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests.
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get the username from the URL query parameter.
	userName := r.URL.Query().Get("username")
	if userName == "" {
		http.Error(w, `{"error": "Username query parameter is required"}`, http.StatusBadRequest)
		return
	}

	var user userModel.User
	// Note: We do NOT select the password for security reasons.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	query := "SELECT user_name, first_name, last_name, mobile_number, email, gender, date_of_birth FROM users WHERE user_name = ?"
	err = db.QueryRow(query, userName).Scan(
		&user.UserName,
		&user.FirstName,
		&user.LastName,
		&user.MobileNumber,
		&user.Email,
		&user.Gender,
		&user.DateOfBirth,
	)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Database error retrieving user: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	db.Close()
	// Respond with the user data.
	json.NewEncoder(w).Encode(user)
}

// handleDeleteUser is the HTTP handler for the /user endpoint.
// It deletes a user by their username.
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE requests.
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get the username from the URL path.
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[2] == "" {
		http.Error(w, `{"error": "Invalid URL. Usage: /user/{username}"}`, http.StatusBadRequest)
		return
	}
	userName := parts[2]
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	// Execute the DELETE statement.
	result, err := db.Exec("DELETE FROM users WHERE user_name = ?", userName)
	if err != nil {
		log.Printf("Database error deleting user: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error": "User not found or no rows affected"}`, http.StatusNotFound)
		return
	}

	// Respond with a success message.
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "User deleted successfully!"}
	json.NewEncoder(w).Encode(response)
	log.Printf("User deleted: %s", userName)
	db.Close()
}
