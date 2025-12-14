package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	model "achievesomethingbro/appmodel"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Helper function to get user_id from username
func getUserIDByUsername(db *sql.DB, username string) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE user_name = ?", username).Scan(&userID)
	return userID, err
}

// handleAddressRequests routes requests to the correct CRUD function.
func handleAddressRequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("Address request: %s %s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodPost:
		addAddress(w, r)
	case http.MethodGet:
		getAddresses(w, r)
	case http.MethodPut:
		updateAddress(w, r)
	case http.MethodDelete:
		deleteAddress(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// addAddress handles POST requests to create a new user address.
func addAddress(w http.ResponseWriter, r *http.Request) {
	log.Printf("Address request: %s %s", r.Method, r.URL.Path)
	var addr model.UserAddress
	if err := json.NewDecoder(r.Body).Decode(&addr); err != nil {
		log.Printf("Failed to decode address: %v", err)
		http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation: Check mandatory fields
	if addr.UserName == "" || addr.Name == "" || addr.PinCode == "" || addr.HouseNumber == "" || addr.City == "" || addr.State == "" {
		log.Printf("Failed to decode address: %v", addr.UserName)
		http.Error(w, "User Name, Name, House Number, Pin Code, City, and State are mandatory fields", http.StatusBadRequest)
		return
	}

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch user_id from username
	userID, err := getUserIDByUsername(db, addr.UserName)

	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// SQL statement to insert all detailed address fields
	query := `INSERT INTO user_addresses (
        user_id, name, floor, house_number, society_name, nearby_landmark, sector, pin_code, city, state, country, latitude, longitude
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := db.Exec(query,
		userID, addr.Name, addr.Floor, addr.HouseNumber, addr.SocietyName, addr.NearbyLandmark,
		addr.Sector, addr.PinCode, addr.City, addr.State, addr.Country, addr.Latitude, addr.Longitude)
	if err != nil {
		log.Printf("Failed to insert address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	lastID, _ := res.LastInsertId()
	addr.ID = int(lastID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(addr)
}

// getAddresses handles GET requests to retrieve user addresses.
func getAddresses(w http.ResponseWriter, r *http.Request) {
	// Extracts User Name from the URL path (e.g., /address/username)
	username := strings.TrimPrefix(r.URL.Path, "/address/")
	if username == "" {
		http.Error(w, "Bad request: a valid username is required", http.StatusBadRequest)
		return
	}

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Fetch user_id from username
	userID, err := getUserIDByUsername(db, username)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Select statement for all detailed address fields
	query := `SELECT 
        id, user_id, name, floor, house_number, society_name, nearby_landmark, sector, pin_code, city, state, country, latitude, longitude 
        FROM user_addresses WHERE user_id = ?`
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("Failed to query addresses: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var addresses []model.UserAddress
	for rows.Next() {
		var addr model.UserAddress
		// Scan all 13 columns into the struct fields
		if err := rows.Scan(
			&addr.ID, &addr.UserId, &addr.Name, &addr.Floor, &addr.HouseNumber, &addr.SocietyName, &addr.NearbyLandmark,
			&addr.Sector, &addr.PinCode, &addr.City, &addr.State, &addr.Country, &addr.Latitude, &addr.Longitude,
		); err != nil {
			log.Printf("Failed to scan address row: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		addresses = append(addresses, addr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addresses)
}

// updateAddress handles PUT requests to update an existing address.
func updateAddress(w http.ResponseWriter, r *http.Request) {
	// Extracts Address ID from the URL path (e.g., /address/456)
	addrIDStr := strings.TrimPrefix(r.URL.Path, "/address/")
	addrID, err := strconv.Atoi(addrIDStr)
	if err != nil || addrID == 0 {
		http.Error(w, "Bad request: a valid address ID is required", http.StatusBadRequest)
		return
	}

	var addr model.UserAddress
	if err := json.NewDecoder(r.Body).Decode(&addr); err != nil {
		http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	// Validation: Check essential fields for update (do not check UserId/UserName)
	if addr.Name == "" || addr.PinCode == "" {
		http.Error(w, "Name and Pin Code are mandatory fields for update", http.StatusBadRequest)
		return
	}

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Only update allowed fields, do not update user_id or username
	query := `UPDATE user_addresses SET 
        name = ?, floor = ?, house_number = ?, society_name = ?, nearby_landmark = ?, sector = ?, pin_code = ?, 
        city = ?, state = ?, country = ?, latitude = ?, longitude = ? 
    WHERE id = ?`
	_, err = db.Exec(query,
		addr.Name, addr.Floor, addr.HouseNumber, addr.SocietyName, addr.NearbyLandmark, addr.Sector, addr.PinCode,
		addr.City, addr.State, addr.Country, addr.Latitude, addr.Longitude, addrID)
	if err != nil {
		log.Printf("Failed to update address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	addr.ID = addrID
	// UserId and UserName are not modified, so do not set them in the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(addr)
}

// deleteAddress handles DELETE requests to remove an address.
func deleteAddress(w http.ResponseWriter, r *http.Request) {
	// Extracts Address ID from the URL path (e.g., /address/456)
	addrIDStr := strings.TrimPrefix(r.URL.Path, "/address/")
	addrID, err := strconv.Atoi(addrIDStr)
	if err != nil || addrID == 0 {
		http.Error(w, "Bad request: a valid address ID is required", http.StatusBadRequest)
		return
	}

	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM user_addresses WHERE id = ?", addrID)
	if err != nil {
		log.Printf("Failed to delete address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Address deleted successfully"})
}
