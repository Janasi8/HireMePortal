package localservice

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// --- Simulated Database Models for Stakeholders and Criteria ---

// StakeholderCriteria represents the job requirements a stakeholder is interested in.
type StakeholderCriteria struct {
	StakeholderEmail string
	RequiredSkills   []string // e.g., ["Go", "Kubernetes", "Python"]
	MinExperience    int
}

// --- Notification Logic ---

// NotifyStakeholders simulates querying stakeholder criteria and sending notifications
// based on keywords found in the extracted resume text.
func NotifyStakeholders(db *sql.DB, userID int, extractedText string) error {
	log.Printf("INFO: Starting stakeholder notification process for User ID %d.", userID)

	// 1. Simulate fetching all active stakeholder criteria from the database
	criteria := simulateFetchAllCriteria()

	// 2. Normalize the extracted text for simple keyword matching
	normalizedText := strings.ToLower(extractedText)

	// 3. Iterate through criteria and check for matches
	for _, c := range criteria {
		matchFound := true
		matchedSkills := []string{}

		// Check if ALL required skills are present in the resume text
		for _, requiredSkill := range c.RequiredSkills {
			if !strings.Contains(normalizedText, strings.ToLower(requiredSkill)) {
				matchFound = false
				break
			}
			matchedSkills = append(matchedSkills, requiredSkill)
		}

		// (NOTE: Simulating experience matching without actual NLP)
		// If a match is found based on skills, simulate sending a notification.
		if matchFound {
			// Simulate finding an analysis result that meets the MinExperience criteria
			if c.MinExperience > 0 && strings.Contains(normalizedText, fmt.Sprintf("%d+ years", c.MinExperience)) {

				// --- 4. Trigger Notifications via Email and Kafka ---
				matchDetails := fmt.Sprintf("Candidate (User ID %d) matches role for %s. Skills: %v.",
					userID, c.StakeholderEmail, matchedSkills)

				// 4a. Send Email Notification
				sendEmailNotification(c.StakeholderEmail, matchDetails)

				// 4b. Publish Kafka Event
				publishKafkaNotification(userID, c.StakeholderEmail, matchedSkills)
			}
		}
	}

	log.Printf("INFO: Stakeholder notification process finished for User ID %d.", userID)
	// Return nil even if no matches were found, as the process completed successfully.
	return nil
}

// sendEmailNotification simulates sending an email to the stakeholder.
func sendEmailNotification(email string, details string) {
	sendTime := time.Now().Format("15:04:05")
	log.Printf("[EMAIL SUCCESS %s]: Sending detailed match alert to %s. Details: %s",
		sendTime, email, details)
	// In a real application, this would use a library like net/smtp or a service like SendGrid.
}

// publishKafkaNotification simulates publishing a structured event to a Kafka topic.
func publishKafkaNotification(userID int, email string, skills []string) {
	// In a real application, you would serialize this data (e.g., to JSON) and send it.
	log.Printf("[KAFKA SUCCESS]: Publishing event to topic 'resume_matches'. UserID: %d, Recipient: %s, Matched Skills Count: %d",
		userID, email, len(skills))
	// In a real application, this would use a library like confluent-kafka-go or segmentio/kafka-go.
}

// simulateFetchAllCriteria mocks fetching the list of required job criteria.
func simulateFetchAllCriteria() []StakeholderCriteria {
	// In a real application, this would involve a complex SQL query joining
	// 'jobs' and 'subscriptions' tables.
	return []StakeholderCriteria{
		{
			StakeholderEmail: "recruiter_go_dev@example.com",
			RequiredSkills:   []string{"Go", "Kubernetes", "Microservices"},
			MinExperience:    3,
		},
		{
			StakeholderEmail: "product_manager@example.com",
			RequiredSkills:   []string{"Jira", "Figma", "Agile"},
			MinExperience:    5,
		},
		{
			StakeholderEmail: "database_admin@example.com",
			RequiredSkills:   []string{"MySQL", "AWS DynamoDB", "SQL"},
			MinExperience:    2,
		},
	}
}
