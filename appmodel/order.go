package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"fmt"
	"log"
)

// Order represents the data structure for an order.
type Order struct {
	OrderID     string `json:"orderId"`
	UserID      int    `json:"userId,omitempty"`
	UserName    string `json:"userName,omitempty"`
	ProductName string `json:"productName"`
	Quantity    string `json:"quantity"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

func CreateOrderTable() {
	// SQL statement to create the 'orders' table.
	// It includes a primary key and unique constraints as per the model's requirements.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
		id INT AUTO_INCREMENT PRIMARY KEY,
		order_id VARCHAR(255) NOT NULL,
		user_id INT NOT NULL,
		product_name VARCHAR(255) NOT NULL,
		quantity VARCHAR(255) NOT NULL,
		status ENUM('created','pending', 'shipped', 'delivered') NOT NULL DEFAULT 'created',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// Execute the SQL statement.
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create orders table: %v", err)
	}
	db.Close()

	fmt.Println("Successfully created table 'users'!")
}
