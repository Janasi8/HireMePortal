package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	model "achievesomethingbro/appmodel"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

// handleCreateOrder handles the creation of a new order.
func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("here0: %+v", r.Body)
	log.Printf("here00: %+v", r)
	var newOrder model.Order
	err := json.NewDecoder(r.Body).Decode(&newOrder)
	log.Printf("here0000: %+v", err)
	if err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	log.Printf("here1: %+v", newOrder)
	if newOrder.UserName == "" || newOrder.ProductName == "" || newOrder.Quantity == "" {
		http.Error(w, `{"error": "Mandatory fields (userName, productName, quantity) are required"}`, http.StatusBadRequest)
		return
	}
	newUUID, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated UUID:")
	fmt.Printf("%s", newUUID)
	newOrder.OrderID = strings.TrimSpace(string(newUUID))
	// Get the user ID from the username
	var userID int
	// Get all orders

	db, err := dbpg.ConnectDB()
	log.Printf("here2: %+v", newOrder)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("here3: %+v", newOrder)
	err = db.QueryRow("SELECT id FROM users WHERE user_name = ?", newOrder.UserName).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("Found user ID: %d for order: %s", userID, newOrder.OrderID)
	fmt.Printf("User ID: %d\n", newOrder.UserID)
	fmt.Printf("Product: %s\n", newOrder.ProductName)
	fmt.Printf("Quantity: %s\n", newOrder.Quantity)
	fmt.Printf("Status: %s\n", newOrder.Status)
	fmt.Printf("Created At: %s\n", newOrder.CreatedAt)
	// Insert the new order into the database
	_, err = db.Exec("INSERT INTO orders (user_id, order_id, product_name, quantity, status) VALUES (?, ?, ?, ?, ?)",
		userID, newOrder.OrderID, newOrder.ProductName, newOrder.Quantity, newOrder.Status)
	if err != nil {
		log.Printf("Failed to insert order: %v", err)
		http.Error(w, `{"error": "Failed to create order"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"message": "Order created successfully!"}
	json.NewEncoder(w).Encode(response)
	log.Printf("New order created for user: %s", newOrder.UserName)
	db.Close()
}

// handleGetOrders handles retrieving all orders or a specific order by ID from the URL path.
func handleGetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get all orders
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT o.order_id, u.user_name, o.product_name, o.quantity, o.status, o.created_at FROM orders o JOIN users u ON o.user_id = u.id")
	if err != nil {
		log.Printf("Database error retrieving orders: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.OrderID, &order.UserName, &order.ProductName, &order.Quantity, &order.Status, &order.CreatedAt); err != nil {
			log.Printf("Error scanning order row: %v", err)
			continue
		}
		orders = append(orders, order)
	}

	json.NewEncoder(w).Encode(orders)
	db.Close()
}

// handleGetSingleOrder handles retrieving a single order by ID from the URL path.
func handleGetSingleOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[2] == "" {
		http.Error(w, `{"error": "Invalid URL. Usage: /orders/{id}"}`, http.StatusBadRequest)
		return
	}
	orderIDStr := parts[2]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	var order model.Order
	// Get all orders
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	query := "SELECT o.order_id, u.user_name, o.product_name, o.quantity, o.status, o.created_at FROM orders o JOIN users u ON o.user_id = u.id WHERE o.order_id = ?"
	err = db.QueryRow(query, orderID).Scan(
		&order.OrderID,
		&order.UserName,
		&order.ProductName,
		&order.Quantity,
		&order.Status,
		&order.CreatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Database error retrieving order: %v", err)
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(order)
	db.Close()
}

// handleUpdateOrder handles updating an existing order.
func handleUpdateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[2] == "" {
		http.Error(w, `{"error": "Invalid URL. Usage: /orders/{id}"}`, http.StatusBadRequest)
		return
	}
	orderIDStr := parts[2]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	var updatedOrder model.Order
	err = json.NewDecoder(r.Body).Decode(&updatedOrder)
	if err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}
	// Get all orders
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	result, err := db.Exec("UPDATE orders SET product_name = ?, quantity = ?, status = ? WHERE order_id = ?",
		updatedOrder.ProductName, updatedOrder.Quantity, updatedOrder.Status, orderID)
	if err != nil {
		log.Printf("Failed to update order: %v", err)
		http.Error(w, `{"error": "Failed to update order"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error": "Order not found or no changes made"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Order updated successfully!"}
	json.NewEncoder(w).Encode(response)
	log.Printf("Order %d updated", orderID)
	db.Close()
}

// handleDeleteOrder handles deleting an order by ID from the URL path.
func handleDeleteOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[2] == "" {
		http.Error(w, `{"error": "Invalid URL. Usage: /orders/{id}"}`, http.StatusBadRequest)
		return
	}
	orderIDStr := parts[2]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}
	// Get all orders
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, `{"error": "Database connection error"}`, http.StatusInternalServerError)
		return
	}
	result, err := db.Exec("DELETE FROM orders WHERE order_id = ?", orderID)
	if err != nil {
		log.Printf("Failed to delete order: %v", err)
		http.Error(w, `{"error": "Failed to delete order"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error": "Order not found or no rows affected"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"message": "Order deleted successfully!"}
	json.NewEncoder(w).Encode(response)
	log.Printf("Order %d deleted", orderID)
	db.Close()
}
