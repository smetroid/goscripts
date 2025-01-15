package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"memo/sunbeam"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Memo struct {
	Content string `json:"content"`
}

type MemoResponse struct {
	NextPageToken string `json:"nextPageToken"`
	Memos         []Memo `json:"memos"`
}

// extractCommand parses the command from the shell code block
func extractCommand(content string) string {
	// Match the content inside the shell code block
	re := regexp.MustCompile("(?s)```shell\\n(.*?)\\n```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractTags parses the tags from the Tags section
func extractTags(content string) []string {
	// Match the line starting with **Tags:** and extract hashtags
	re := regexp.MustCompile(`(?i)\*\*Tags:\*\*\s*(#[a-zA-Z0-9-_]+(?:\s*#[a-zA-Z0-9-_]+)*)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		tagsLine := matches[1]
		tags := strings.Fields(tagsLine) // Split by spaces
		for i, tag := range tags {
			tags[i] = strings.TrimPrefix(tag, "#") // Remove leading '#'
		}
		return tags
	}
	return nil
}

// Function to filter commands based on a tag
func filterCommandsByTag(resultSlice []map[string]string, tag string) map[string]string {
	// Create a map to hold the filtered results
	filteredResults := make(map[string]string)

	// Iterate through the slice and check if the "tags" contain the specified tag
	for _, result := range resultSlice {
		// Check if the "tags" contain the provided tag (case-sensitive)
		if strings.Contains(result["tags"], tag) {
			// Add the command to the filtered map
			filteredResults[result["name"]] = result["tags"]
		}
	}

	return filteredResults
}

func main() {
	// Example path to the Sunbeam configuration file
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "sunbeam.json")
	// Retrieve memo preferences
	preferences, err := sunbeam.ReadSunbeamConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	apiKey := ""
	apiURL := ""
	if len(preferences.MemoToken) == 0 || len(preferences.MemoURL) == 0 {
		fmt.Printf("Error: no values found in sunbeam memo extension configuration ... trying environment variables")

		apiKey = os.Getenv("USEMEMOS_API_KEY")
		apiURL = os.Getenv("USEMEMOS_API_URL")
	} else {
		apiKey = preferences.MemoToken
		apiURL = preferences.MemoURL
	}

	if apiKey == "" || apiURL == "" {
		fmt.Println("Environment variables USEMEMOS_API_KEY and USEMEMOS_API_URL must be set. ... OR ")
		fmt.Println("add token and url in sunbeam memos configuration")
		os.Exit(1)
	}

	// Parse command-line arguments for additional filter tags
	tags := flag.String("tags", "cmd,shell,script", "Comma-separated list of tags to filter memos (e.g., 'cmd,shell,script')")
	flag.Parse()

	// Ensure the API URL ends with `/api/memos`
	if !strings.HasSuffix(apiURL, "/api/v1/memos") {
		apiURL = strings.TrimRight(apiURL, "/") + "/api/v1/memos"
	}

	// Format tags into query parameter
	tagFilter := fmt.Sprintf("tags=%s", *tags)

	// Construct the full URL with the tag filter
	url := fmt.Sprintf("%s?%s", apiURL, tagFilter)
	fmt.Println(url)

	// Create HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		os.Exit(1)
	}

	// Add headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Handle the response based on status code
	if resp.StatusCode == http.StatusOK {
		// Parse the response body as JSON
		var response MemoResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Printf("Error parsing response JSON: %v\n", err)
			os.Exit(1)
		}

		// Display the retrieved memos
		if len(response.Memos) == 0 {
			fmt.Println("No memos found with the specified tags.")
		} else {

			stringsMap := []map[string]string{}
			for _, memo := range response.Memos {
				//fmt.Println(memo.Content)
				codeBlock := extractCommand(memo.Content)
				//fmt.Println(codeBlock)
				tags := extractTags(memo.Content)
				itemMap := map[string]string{"name": codeBlock, "tags": strings.Join(tags, " ")}
				stringsMap = append(stringsMap, itemMap)
			}

			// Call the function with tag "docker"
			tag := "cmd"
			filteredCommands := filterCommandsByTag(stringsMap, tag)
			fmt.Println(filteredCommands)

			// Transform to desired structure
			var transformed []map[string]string
			for key, value := range filteredCommands {
				transformed = append(transformed, map[string]string{
					"name": key,
					"type": value,
				})
			}
			// Convert list to JSON
			jsonData, err := json.MarshalIndent(transformed, "", "  ")
			if err != nil {
				log.Fatalf("Error converting to JSON: %v", err)
			}

			// Print JSON to console
			fmt.Println(string(jsonData))

		}
	} else {
		fmt.Printf("Failed to retrieve memos. Status: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}
