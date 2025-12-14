package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"log"
)

// CartItem represents an item in the cart or for direct checkout.
type CartItem struct {
	ItemID   StringInt `json:"itemId"`
	Quantity StringInt `json:"quantity"`
}

func CreateCartItemsTable() {
	// SQL statement to create the 'carts' table.
	// It includes a primary key and foreign key constraints as per the model's requirements.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Implementation for creating the cart_items table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cart_items (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			item_id INT NOT NULL,
			quantity INT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (item_id) REFERENCES items(id)
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create cart_items table: %v", err)
	}
	db.Close()
	log.Println("Successfully created table 'cart_items'!")
}
