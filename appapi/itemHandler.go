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

// handleItemCollection handles GET and POST requests for the /api/items endpoint.
func handleItemCollection(w http.ResponseWriter, r *http.Request) {
	log.Printf("here0: %+v", r.Body)
	// This simple CORS header is needed for the frontend to work.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Pre-flight request for CORS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getAllItems(w, r)
	case "POST":
		log.Printf("Post: %+v", r.Body)
		addItem(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleItemResource handles GET, PUT, and DELETE requests for individual items by ID.
func handleItemResource(w http.ResponseWriter, r *http.Request) {
	log.Printf("here1: %+v", r.Body)
	// This simple CORS header is needed for the frontend to work.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Pre-flight request for CORS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Split the URL path to get the ID.
	path := strings.Trim(r.URL.Path, "/")
	pathSegments := strings.Split(path, "/")
	idStr := pathSegments[len(pathSegments)-1]

	switch r.Method {
	case "GET":
		getItemByID(w, r, idStr)
	case "PUT":
		updateItem(w, r, idStr)
	case "DELETE":
		deleteItem(w, r, idStr)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getAllItems handles GET requests to retrieve all items from the database.
func getAllItems(w http.ResponseWriter, r *http.Request) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, name, description, category, mrp, discount_percentage, final_price, item_image_url FROM items")
	if err != nil {
		http.Error(w, "Failed to get items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Category, &item.MRP, &item.DiscountPercentage, &item.FinalPrice, &item.ItemImageURL); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
	db.Close()
}

// getItemByID handles GET requests for a single item by its ID.
func getItemByID(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var item model.Item
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	row := db.QueryRow("SELECT id, name, description, category, mrp, discount_percentage, final_price, item_image_url FROM items WHERE id = ?", id)
	if err := row.Scan(&item.ID, &item.Name, &item.Description, &item.Category, &item.MRP, &item.DiscountPercentage, &item.FinalPrice, &item.ItemImageURL); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
	db.Close()
}

// addItem handles POST requests to add a new item to the database.
func addItem(w http.ResponseWriter, r *http.Request) {
	var newItem model.Item
	log.Printf("here33: %+v", r.Body)
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		log.Printf("here33212: %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("here333: %+v", newItem)
	// Calculate the final price before insertion.
	newItem.FinalPrice = newItem.MRP * (1 - newItem.DiscountPercentage/100)
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("here2: %+v", newItem)
	stmt, err := db.Prepare("INSERT INTO items(name, description, category, mrp, discount_percentage, final_price, item_image_url) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, "Failed to prepare statement: "+err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := stmt.Exec(newItem.Name, newItem.Description, newItem.Category, newItem.MRP, newItem.DiscountPercentage, newItem.FinalPrice, newItem.ItemImageURL)
	if err != nil {
		http.Error(w, "Failed to insert item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get last insert ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	newItem.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
	db.Close()
}

// updateItem handles PUT requests to update an existing item by ID.
func updateItem(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var updatedItem model.Item
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate the final price before updating.
	updatedItem.FinalPrice = updatedItem.MRP * (1 - updatedItem.DiscountPercentage/100)
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("UPDATE items SET name=?, description=?, category=?, mrp=?, discount_percentage=?, final_price=?, item_image_url=? WHERE id=?")
	if err != nil {
		http.Error(w, "Failed to prepare statement: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := stmt.Exec(updatedItem.Name, updatedItem.Description, updatedItem.Category, updatedItem.MRP, updatedItem.DiscountPercentage, updatedItem.FinalPrice, updatedItem.ItemImageURL, id)
	if err != nil {
		http.Error(w, "Failed to update item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Item not found or no changes were made", http.StatusNotFound)
		return
	}

	updatedItem.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedItem)
	db.Close()
}

// deleteItem handles DELETE requests to remove an item by ID.
func deleteItem(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	stmt, err := db.Prepare("DELETE FROM items WHERE id=?")
	if err != nil {
		http.Error(w, "Failed to prepare statement: "+err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := stmt.Exec(id)
	if err != nil {
		http.Error(w, "Failed to delete item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	db.Close()
	w.WriteHeader(http.StatusNoContent)
}
