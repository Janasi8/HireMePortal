package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

// Cart struct represents an item in a user's shopping cart.
type Cart struct {
	ID       int       `json:"id"`
	UserName string    `json:"user_name"`
	ItemID   StringInt `json:"item_id"`
	Quantity StringInt `json:"quantity"`
}
type StringInt int

func (f *StringInt) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		// If it's not a string, try to unmarshal as a direct int
		var num int
		if err := json.Unmarshal(b, &num); err != nil {
			return err
		}
		*f = StringInt(num)
		return nil
	}
	// Try to convert the string to int
	var num int
	num, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = StringInt(num)
	return nil
}

func CreateCartTable() {
	// SQL statement to create the 'carts' table.
	// It includes a primary key and foreign key constraints as per the model's requirements.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS carts (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		item_id INT NOT NULL,
		quantity INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
	);`

	// Execute the SQL statement.
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create carts table: %v", err)
	}
	db.Close()

	fmt.Println("Successfully created table 'carts'!")
}
