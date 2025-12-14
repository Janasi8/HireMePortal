package localmodel

import (
	dbpg "achievesomethingbro/appdb"
	"fmt"
	"log"
)

type ResumeSummary struct {
	ID        int    `json:"id,omitempty"`
	UserId    int    `json:"user_id"`
	ScanData  string `json:"scan_data"` // JSON data as string
	CreatedAt string `json:"created_at,omitempty"`
}

func CreateAiResumeSummaryTable() {
	// SQL statement to create the 'carts' table.
	// It includes a primary key and foreign key constraints as per the model's requirements.
	db, err := dbpg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS user_resume_summary (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL unique,
			resume_id BIGINT UNSIGNED NOT NULL,
			scan_data JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			foreign key (user_id) references users(id) on delete cascade,
			foreign key (resume_id) references user_resumes(id) on delete cascade
		);`

	// Execute the SQL statement.
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create user_resume_summary table: %v", err)
	}
	db.Close()

	fmt.Println("Successfully created table 'user_resume_summary'!")
}
