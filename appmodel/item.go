package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"encoding/json"
	"log"
	"strconv"
)

type Float64 float64

// Item struct represents a single product or item with its details.
type Item struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	Category           string  `json:"category"`
	MRP                Float64 `json:"mrp"`
	DiscountPercentage Float64 `json:"discount_percentage"`
	FinalPrice         Float64 `json:"final_price"`
	ItemImageURL       string  `json:"item_image_url,omitempty"`
}

func (f *Float64) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		// If it's not a string, try to unmarshal as a direct float64
		var num float64
		if err := json.Unmarshal(b, &num); err != nil {
			return err
		}
		*f = Float64(num)
		return nil
	}
	// Try to convert the string to float64
	var num float64
	num, err := strconv.ParseFloat(s, 64) // strconv imported as 's'
	if err != nil {
		return err
	}
	*f = Float64(num)
	return nil
}

func CreateItemTable() {
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS items (
    	id INT AUTO_INCREMENT PRIMARY KEY,
	    name VARCHAR(255) NOT NULL,
	    description TEXT,
		category VARCHAR(255),
	    mrp DECIMAL(10, 2) NOT NULL,
	    discount_percentage DECIMAL(5, 2) NOT NULL,
	    final_price DECIMAL(10, 2) NOT NULL,
		item_image_url VARCHAR(512),
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create items table: %v", err)
	}
	db.Close()

	log.Println("Successfully created table 'items'!")
}
