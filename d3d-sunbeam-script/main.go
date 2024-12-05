package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/itchyny/gojq"
)

// Define a struct to represent the JSON data
type Sunbeam struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Diagram     string `json:"diagram"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

type Node struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Parent  string `json:"parent"`
	Options []struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	} `json:"options"`
}

func getNodeId(workflow Sunbeam, queryString string) (id []string) {

	// Parse the query
	query, err := gojq.Parse(queryString)
	if err != nil {
		log.Fatalf("Failed to parse jq query: %v", err)
	}

	// Unmarshal the JSON input into a map
	var jsonObject interface{}
	err = json.Unmarshal([]byte(workflow.Diagram), &jsonObject)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Execute the query
	labelID := []string{}
	iter := query.Run(jsonObject)
	for {
		result, ok := iter.Next()
		if !ok {
			// End of results
			break
		}
		if err, ok := result.(error); ok {
			log.Fatalf("Query execution error: %v", err)
		}

		// Print the matching IDs
		matchingIDs, ok := result.([]interface{})
		if !ok {
			log.Fatalf("Unexpected result type: %T", result)
		}

		// Iterate over the matching IDs
		for _, id := range matchingIDs {
			fmt.Println("Matching ID:", id)
			labelID = append(labelID, id.(string))
		}
	}
	return labelID
}

func main() {
	// API endpoint URL
	apiURL := "http://localhost:3000/dag/b10a2f02-f35d-4e93-9756-a8cf4a500bb8" // Replace with your actual API URL

	// JWT token
	jwtToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzM1NDM1NzYsImlzcyI6ImxkYXAiLCJqdGkiOiJzYW11cyIsIm5hbWUiOiJzYW11cyIsInJvbGUiOiJhZG1pbiJ9.mWUbYCuusbEs9DdQA0TLUdDvIXlLHK9pSD4jiQ9q2IQ"

	// Create a new HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}

	// Add the Authorization header
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Parse the JSON into a slice of Diagram
	var workflow Sunbeam
	if err := json.Unmarshal(body, &workflow); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	labelToFind := "options"
	queryString := fmt.Sprintf(`.nodes | map(select(.value.label == "%s").value.id)`, labelToFind)
	id := getNodeId(workflow, queryString)
	fmt.Printf("id:%s\n", id)

	queryString = fmt.Sprintf(`.nodes | map(select(.parent == "%s").value.id)`, id)
	ids := getNodeId(workflow, queryString)
	fmt.Printf("id:%s\n", ids)

	// Build the jq query dynamically
	//optionsChildren := fmt.Sprintf(`.nodes | map(select(.parent == "%s").value.id)`, optionId)

}
