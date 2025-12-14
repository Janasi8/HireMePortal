package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"fmt"
	"log"
)

type UserAddress struct {
	ID             int       `json:"id,omitempty"`
	UserName       string    `json:"user_name"`
	UserId         StringInt `json:"user_id,omitempty"`
	Name           string    `json:"name"` // Recipient Name
	Floor          string    `json:"floor"`
	HouseNumber    string    `json:"house_number"`
	SocietyName    string    `json:"society_name"`
	NearbyLandmark string    `json:"nearby_landmark"`
	Sector         string    `json:"sector"`
	PinCode        string    `json:"pin_code"`
	City           string    `json:"city"`
	State          string    `json:"state"`
	Country        string    `json:"country"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
}

func CreateUserAddressTable() {
	// SQL statement to create the 'user_addresses' table.
	// It includes a primary key and foreign key constraints as per the model's requirements.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	createTableSQL := `CREATE TABLE IF NOT EXISTS user_addresses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    floor VARCHAR(50),
    house_number VARCHAR(50) NOT NULL,
    society_name VARCHAR(255),
    nearby_landmark VARCHAR(255),
    sector VARCHAR(50),
    pin_code VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    latitude DOUBLE NOT NULL,
    longitude DOUBLE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// Execute the SQL statement.
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create user_addresses table: %v", err)
	}
	db.Close()

	fmt.Println("Successfully created table 'user_addresses'!")
}
