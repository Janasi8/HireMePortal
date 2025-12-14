package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"log"
)

// CheckoutRequest represents the request body for a checkout.
type CheckoutRequest struct {
	UserName     string     `json:"user_name"`
	CartCheckout bool       `json:"cart_checkout"`
	Items        []CartItem `json:"items,omitempty"`
}

func CreateCheckoutTable() {
	// Implementation for creating any necessary tables for checkout, if needed.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Example: Create a table for storing checkout sessions (if needed)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS checkouts (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			total_amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create checkouts table: %v", err)
	}
	db.Close()
	log.Println("Successfully created table 'checkouts'!")
}
