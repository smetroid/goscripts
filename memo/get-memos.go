package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

//// Helper function to find the index of the line corresponding to a character offset
//func findLineIndex(lines []string, offset int) int {
//	currentOffset := 0
//	for i, line := range lines {
//		currentOffset += len(line) + 1 // Add 1 for the newline character
//		if currentOffset > offset {
//			return i
//		}
//	}
//	return -1
//}
//
//// extractMdBlocks extracts fenced code blocks and their associated tags from markdown text.
//func extractCodeBlock(text string) []map[string]string {
//	// Define the regular expression pattern for fenced code blocks
//	codeBlockPattern := regexp.MustCompile("(?s)```(?:\\w+\\s+)?(.*?)```")
//
//	// Find all matches of code blocks in the text
//	matches := codeBlockPattern.FindAllStringSubmatchIndex(string(text), -1)
//
//	var results []map[string]string
//	lines := strings.Split(string(text), "\n") // Split the text into lines for tag detection
//
//	for _, match := range matches {
//		if len(match) >= 4 {
//			// Extract the code block
//			codeBlock := strings.TrimSpace(string(text[match[2]:match[3]]))
//
//			// Locate the potential tag line
//			tags := ""
//			endOfBlockIndex := match[1] // End of the matched code block
//			tagStartLine := findLineIndex(lines, endOfBlockIndex)
//
//			if tagStartLine >= 0 && tagStartLine+1 < len(lines) {
//				possibleTagLine := strings.TrimSpace(lines[tagStartLine+1])
//				if strings.HasPrefix(possibleTagLine, "#") { // Ensure it's a valid tag line
//					tags = possibleTagLine
//				}
//			}
//
//			// Add the code block and tag to the results
//			results = append(results, map[string]string{
//				"code": codeBlock,
//				"tags": tags,
//			})
//		}
//	}
//
//	return results
//}

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
	// Parse command-line arguments for additional filter tags
	tags := flag.String("tags", "cmd,shell,script", "Comma-separated list of tags to filter memos (e.g., 'cmd,shell,script')")
	flag.Parse()

	// Get API key and API URL from environment variables
	apiKey := os.Getenv("USEMEMOS_API_KEY")
	apiURL := os.Getenv("USEMEMOS_API_URL")

	// Ensure the API URL ends with `/api/memos`
	if !strings.HasSuffix(apiURL, "/api/v1/memos") {
		apiURL = strings.TrimRight(apiURL, "/") + "/api/v1/memos"
	}

	if apiKey == "" || apiURL == "" {
		fmt.Println("Environment variables USEMEMOS_API_KEY and USEMEMOS_API_URL must be set.")
		os.Exit(1)
	}

	// Format tags into query parameter
	tagFilter := fmt.Sprintf("tags=%s", *tags)

	// Construct the full URL with the tag filter
	url := fmt.Sprintf("%s?%s", apiURL, tagFilter)

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

	fmt.Println(string(body))
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
				tags := extractTags(memo.Content)
				//fmt.Println("content")
				//fmt.Println(codeBlock)
				//fmt.Println(tags)
				//fmt.Println("content")
				itemMap := map[string]string{"name": codeBlock, "tags": strings.Join(tags, ", ")}
				stringsMap = append(stringsMap, itemMap)
			}

			// Call the function with tag "docker"
			tag := "shell"
			filteredCommands := filterCommandsByTag(stringsMap, tag)

			// Convert list to JSON
			jsonData, err := json.MarshalIndent(filteredCommands, "", "  ")
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
