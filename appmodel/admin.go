package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"log"
)

// CheckoutRequest represents the request body for a checkout.
type Admin struct {
	ID       int    `db:"id"`
	Username string `db:"admin_user_name"`
	Email    string `db:"email"`
	Password string `db:"password"`
	FullName string `db:"full_name"`
	Role     string `db:"role"`
	IsActive bool   `db:"is_active"`
}

func CreateAdminTable() {
	// Implementation for creating any necessary tables for checkout, if needed.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Example: Create a table for storing checkout sessions (if needed)
	_, err = db.Exec(`
	CREATE TABLE admins (
	id INT PRIMARY KEY AUTO_INCREMENT,
    admin_user_name VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    role ENUM('superadmin', 'editor', 'viewer') NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Failed to create checkouts table: %v", err)
	}
	db.Close()
	log.Println("Successfully created table 'checkouts'!")
}
