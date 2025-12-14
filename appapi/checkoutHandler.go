package signupapiv1

import (
	dbpg "achievesomethingbro/appdb"
	model "achievesomethingbro/appmodel"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func handleCheckoutResource(w http.ResponseWriter, r *http.Request) {
	log.Printf("hereCheckout: %+v", r.Body)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("hereCheckout2: %+v", r.Body)
	var req model.CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("hereCheckout1: %+v", err)
		http.Error(w, "Bad request: invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.UserName == "" {
		http.Error(w, "User Name is a mandatory field", http.StatusBadRequest)
		return
	}
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Use a transaction to ensure all operations are atomic.
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // The rollback will be ignored if the transaction is committed.

	// Fetch user ID from user_name
	var userID int
	err = tx.QueryRow("SELECT id FROM users WHERE user_name = ?", req.UserName).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	var itemsToProcess []model.CartItem
	if req.CartCheckout {
		// Scenario 1: Checkout the entire cart.
		// Get all items from the user's cart.
		rows, err := tx.Query("SELECT item_id, quantity FROM carts WHERE user_id = ?", userID)
		if err != nil {
			log.Printf("Failed to query cart items: %v", err)
			http.Error(w, "Failed to retrieve cart items", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var item model.CartItem
			if err := rows.Scan(&item.ItemID, &item.Quantity); err != nil {
				log.Printf("Failed to scan cart item: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			itemsToProcess = append(itemsToProcess, item)
		}

		// If the cart is empty, return an error.
		if len(itemsToProcess) == 0 {
			http.Error(w, "Cart is empty", http.StatusBadRequest)
			return
		}

	} else {
		// Scenario 2: Direct checkout from items in the request body.
		if len(req.Items) == 0 {
			http.Error(w, "Items array is empty for direct checkout", http.StatusBadRequest)
			return
		}
		itemsToProcess = req.Items
	}
	newUUID, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated UUID:")
	fmt.Printf("%s", newUUID)
	orderUUID := strings.TrimSpace(string(newUUID))
	// Prepare a statement to insert into the orders table.
	stmt, err := tx.Prepare("INSERT INTO orders(order_id, user_id, product_name, quantity) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Printf("Failed to prepare order statement: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Insert each item into the orders table.
	for _, item := range itemsToProcess {
		var productName string
		err := tx.QueryRow("SELECT name FROM items WHERE id = ?", item.ItemID).Scan(&productName)
		if err != nil {
			log.Printf("Failed to fetch product name: %v", err)
			http.Error(w, "Failed to complete checkout", http.StatusInternalServerError)
			return
		}
		if _, err := stmt.Exec(orderUUID, userID, productName, item.Quantity); err != nil {
			log.Printf("Failed to insert order item: %v", err)
			http.Error(w, "Failed to complete checkout", http.StatusInternalServerError)
			return
		}
	}

	// If cart checkout was performed, clear the user's cart.
	if req.CartCheckout {
		if _, err := tx.Exec("DELETE FROM carts WHERE user_id = ?", userID); err != nil {
			log.Printf("Failed to clear cart: %v", err)
			http.Error(w, "Failed to clear cart after checkout", http.StatusInternalServerError)
			return
		}
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, "Internal server error during transaction commit", http.StatusInternalServerError)
		return
	}

	// Send a success response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Checkout successful"})
}
