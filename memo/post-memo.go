package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Parse command-line arguments
	tags := flag.String("tags", "", "Comma-separated list of tags for the memo (e.g., 'shell,commands')")
	flag.Parse()

	// Get API key and API URL from environment variables
	apiKey := os.Getenv("USEMEMOS_API_KEY")
	apiURL := os.Getenv("USEMEMOS_API_URL")

	// Ensure the API URL ends with `/api/memo`
	if !strings.HasSuffix(apiURL, "/api/v1/memos") {
		apiURL = strings.TrimRight(apiURL, "/") + "/api/v1/memos"
	}

	if apiKey == "" || apiURL == "" {
		fmt.Println("Environment variables USEMEMOS_API_KEY and USEMEMOS_API_URL must be set.")
		os.Exit(1)
	}

	// Get the last shell command executed
	lastCommand, err := getLastShellCommand()
	if err != nil {
		fmt.Printf("Error retrieving last shell command: %v\n", err)
		os.Exit(1)
	}

	// Trim any trailing newline
	lastCommand = strings.TrimSpace(lastCommand)

	// If the last command is empty, exit
	if lastCommand == "" {
		fmt.Println("No last shell command found.")
		os.Exit(1)
	}

	// Prompt for additional tags
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter additional tags (comma-separated): ")
	additionalTags, _ := reader.ReadString('\n')
	additionalTags = strings.TrimSpace(additionalTags)

	// Combine tags from flag and prompt
	var allTags []string
	if *tags != "" {
		allTags = append(allTags, strings.Split(*tags, ",")...)
	}
	if additionalTags != "" {
		allTags = append(allTags, strings.Split(additionalTags, ",")...)
	}

	// Prefix tags with # and format as Markdown
	var hashtags []string
	for _, tag := range allTags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			hashtags = append(hashtags, "#"+tag)
		}
	}

	tagsMarkdown := ""
	if len(hashtags) > 0 {
		tagsMarkdown = "\n\n**Tags:** " + strings.Join(hashtags, " ")
	}

	// Create Markdown content
	markdownContent := fmt.Sprintf("```shell\n%s\n```%s", lastCommand, tagsMarkdown)

	// Create memo payload
	memo := map[string]interface{}{
		"content":    markdownContent,
		"visibility": "PUBLIC", // Default visibility
	}
	if len(allTags) > 0 {
		memo["tags"] = allTags
	}

	payload, err := json.Marshal(memo)
	if err != nil {
		fmt.Printf("Error creating request payload: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		os.Exit(1)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
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

	// Output response
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Memo posted successfully!")
		fmt.Printf("Response: %s\n", string(body))
	} else {
		fmt.Printf("Failed to post memo. Status: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
	}
}

// getLastShellCommand retrieves the last executed command from the shell history
func getLastShellCommand() (string, error) {
	// Get the shell's history file (assuming Bash or Zsh)
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		historyFile = os.ExpandEnv("$HOME/.bash_history") // Default to Bash history
	}

	// Read the history file
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read history file: %w", err)
	}

	// Split the history file into lines and return the last non-empty line
	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i], nil
		}
	}

	return "", fmt.Errorf("no commands found in history file")
}
