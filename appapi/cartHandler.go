package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	model "achievesomethingbro/appmodel"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// handleCartRequests routes requests to the appropriate cart handler.
func handleCartRequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("hereCart: %+v", r.Body)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	pathSegments := strings.Split(path, "/")

	switch r.Method {
	case "POST":

		addToCart(w, r)
	case "GET":
		if len(pathSegments) == 2 {
			http.Error(w, "User ID required in path", http.StatusBadRequest)
			return
		}
		userName := pathSegments[2]
		db, err := dbpg.ConnectDB()
		if err != nil {
			http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		var userID int
		err = db.QueryRow("SELECT id FROM users WHERE user_name = ?", userName).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
			return
		}
		getCart(w, r, userID)
	case "DELETE":
		if len(pathSegments) < 4 {
			http.Error(w, "User ID and Item ID required in path", http.StatusBadRequest)
			return
		}
		userName := pathSegments[2]
		db, err := dbpg.ConnectDB()
		if err != nil {
			http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		var userID int
		err = db.QueryRow("SELECT id FROM users WHERE user_name = ?", userName).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
			return
		}
		itemID, err := strconv.Atoi(pathSegments[3])
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}
		deleteFromCart(w, r, userID, itemID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// addToCart handles POST requests to add an item to a user's cart.
func addToCart(w http.ResponseWriter, r *http.Request) {
	var cartItem model.Cart
	log.Printf("Cart Item: %+v", cartItem)
	if err := json.NewDecoder(r.Body).Decode(&cartItem); err != nil {
		log.Printf("hereCart2: %+v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("Cart Item: %+v", cartItem)

	if cartItem.UserName == "" || cartItem.ItemID == 0 || cartItem.Quantity == 0 {
		http.Error(w, "User Name, Item ID, and Quantity are required", http.StatusBadRequest)
		return
	}

	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Fetch user ID from users table using user_name
	var userId int
	err = db.QueryRow("SELECT id FROM users WHERE user_name = ?", cartItem.UserName).Scan(&userId)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the item already exists in the user's cart
	var existingQuantity int
	err = db.QueryRow("SELECT quantity FROM carts WHERE user_id = ? AND item_id = ?", userId, cartItem.ItemID).Scan(&existingQuantity)

	if err == nil {
		// Update existing item
		newQuantity := existingQuantity + int(cartItem.Quantity)
		_, err := db.Exec("UPDATE carts SET quantity = ? WHERE user_id = ? AND item_id = ?", newQuantity, userId, cartItem.ItemID)
		if err != nil {
			http.Error(w, "Failed to update cart item: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Item quantity updated in cart"})
		return
	} else if err.Error() != "sql: no rows in result set" {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert new item
	_, err = db.Exec("INSERT INTO carts (user_id, item_id, quantity) VALUES (?, ?, ?)", userId, cartItem.ItemID, cartItem.Quantity)
	if err != nil {
		http.Error(w, "Failed to add item to cart: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item added to cart successfully"})
}

// getCart retrieves all items in a user's cart.
func getCart(w http.ResponseWriter, r *http.Request, userID int) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := db.Query("SELECT c.item_id, c.quantity, i.name, i.final_price FROM carts c JOIN items i ON c.item_id = i.id WHERE c.user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to get cart items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var cartItems []map[string]interface{}
	for rows.Next() {
		var itemID, quantity int
		var itemName string
		var finalPrice float64
		if err := rows.Scan(&itemID, &quantity, &itemName, &finalPrice); err != nil {
			http.Error(w, "Failed to scan cart item: "+err.Error(), http.StatusInternalServerError)
			return
		}
		cartItems = append(cartItems, map[string]interface{}{
			"itemID":     itemID,
			"name":       itemName,
			"quantity":   quantity,
			"finalPrice": finalPrice,
			"totalPrice": float64(quantity) * finalPrice,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cartItems)
	db.Close()
}

// deleteFromCart removes an item from a user's cart.
func deleteFromCart(w http.ResponseWriter, r *http.Request, userID, itemID int) {
	db, err := dbpg.ConnectDB()
	if err != nil {
		http.Error(w, "Failed to connect to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := db.Exec("DELETE FROM carts WHERE user_id = ? AND item_id = ?", userID, itemID)
	if err != nil {
		http.Error(w, "Failed to delete item from cart: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Item not found in cart", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item deleted from cart"})
	db.Close()
}
