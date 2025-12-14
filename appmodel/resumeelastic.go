package localmodel

// Define the index name where resume data will be stored in Elasticsearch
const ResumeIndexName = "resumes_analysis"

// --- DTOs from es_models.go (required for request/response structure) ---

// ResumeIndexData represents the minimal data structure sent to Elasticsearch for indexing
type ResumeIndexData struct {
	UserID        int    `json:"user_id"`
	ExtractedText string `json:"extracted_text"`
	Skills        string `json:"skills"` // Analyzed skills/keywords
}

// ESQueryContainer encapsulates the full Elasticsearch query structure
type ESQueryContainer struct {
	Query ESQuery `json:"query"`
}

// ESQuery defines the type of query (e.g., must contain a multi_match)
type ESQuery struct {
	Bool ESMust `json:"bool"`
}

// ESMust defines the 'must' clause within the bool query
type ESMust struct {
	Must []ESMultiMatch `json:"must"`
}

// ESMultiMatch defines the standard Elasticsearch multi_match query
type ESMultiMatch struct {
	MultiMatch map[string]interface{} `json:"multi_match"`
}

// ESResumeResponse models the simplified structure of a successful Elasticsearch search response
type ESResumeResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source ResumeAnalysisResult `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// ResumeAnalysisResult models the document source retrieved from Elasticsearch
type ResumeAnalysisResult struct {
	UserID       int    `json:"user_id"`
	Summary      string `json:"summary"` // Generated summary
	YearsExp     int    `json:"years_exp"`
	KeyTechStack string `json:"key_tech_stack"`
}
