package signupapiv1

import (
	"encoding/json"
	"log"
	"net/http"
)

// LoginRequest defines the structure of the incoming JSON from the frontend.
// The `json:"..."` tags map the JSON keys to the struct fields.
type LoginRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
	UserId   int    `json:"user_id"`
	UserType string `json:"user_type"`
}

// LoginResponse defines the structure of the JSON response sent back to the frontend.
// `omitempty` means a field will be excluded from the JSON if it's empty.
type LoginResponse struct {
	Status  string `json:"status"`
	Role    string `json:"role,omitempty"`
	UserId  int    `json:"user_id"`
	Message string `json:"message,omitempty"`
}

// adminLoginHandler processes the login request.
func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Ensure the request method is POST
	log.Print("i am here bro")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode the incoming JSON body into our LoginRequest struct
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	log.Print(req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// 4. **AUTHENTICATION LOGIC**
	// This is where you would normally check credentials against a database.
	// We are simulating the logic based on the frontend's expectations.
	var res LoginResponse
	isAuthenticated := false
	// db, err := dbpg.ConnectDB()
	// Check for a user with 'all' privileges
	if req.UserName == "admin" && req.Password == "password123" {
		isAuthenticated = true

		res = LoginResponse{Status: "success", Role: "all", UserId: 1}
		w.WriteHeader(http.StatusOK) // 200 OK
	}

	// Check for a user with 'view' privileges
	if req.UserName == "viewer" && req.Password == "password456" {
		isAuthenticated = true
		res = LoginResponse{Status: "success", Role: "view", UserId: 2}
		w.WriteHeader(http.StatusOK) // 200 OK
	}

	// If neither of the above, authentication fails
	if !isAuthenticated {
		res = LoginResponse{Status: "error", Message: "Invalid credentials"}
		w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized is appropriate for failed logins
	}

	// 5. Encode the LoginResponse struct into JSON and send it back
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
