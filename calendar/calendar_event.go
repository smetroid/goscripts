package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func main() {
	// Path to your Google API credentials JSON file
	credentialsFile := "./credentials.json"

	// Set up the Google API client
	client, err := getClient(credentialsFile)
	if err != nil {
		log.Fatalf("Failed to create Google Calendar client: %v", err)
	}

	// Create a new calendar service
	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Failed to create calendar service: %v", err)
	}

	// Create the OOO event
	event := &calendar.Event{
		Summary:     "Out of Office",
		Location:    "Home",
		Description: "I'll be out of the office.",
		Start: &calendar.EventDateTime{
			Date:     "2023-05-25", // Replace with your start time
			TimeZone: "America/Denver",
		},
		End: &calendar.EventDateTime{
			Date:     "2023-05-26", // Replace with your end time
			TimeZone: "America/Denver",
		},
	}

	// Insert the event
	event, err = srv.Events.Insert("primary", event).Do()
	if err != nil {
		log.Fatalf("Failed to create event: %v", err)
	}

	fmt.Printf("OOO event created: %s\n", event.HtmlLink)
}

// getClient retrieves a valid OAuth2 token based on the provided credentials file.
func getClient(credentialsFile string) (*http.Client, error) {
	// Read the credentials file
	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// Parse the credentials file
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file: %v", err)
	}

	// Modify config to request offline access and refresh token
	config = configWithOffline(config)

	// Obtain a token
	client := getClientToken(config)
	return client, nil
}

// configWithOffline modifies the provided OAuth2 config to request offline access and refresh token.
func configWithOffline(config *oauth2.Config) *oauth2.Config {
	config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	//config.Scopes = append(config.Scopes, "offline_access")
	return config
}

// getClientToken retrieves a valid OAuth2 token using the provided config.
func getClientToken(config *oauth2.Config) *http.Client {
	// Check if a token file already exists
	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err == nil && token.Valid() {
		return config.Client(context.Background(), token)
	}

	// If token is not valid, request a new one
	authURL := config.AuthCodeURL("state-token")
	fmt.Printf("Go to the following link in your browser, then enter the "+
		"authorization code:\n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err = config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	// Save the token for future use
	saveToken(tokenFile, token)

	return config.Client(context.Background(), token)
}

// tokenFromFile retrieves a token from a file path.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

// saveToken saves a token to a file path.
func saveToken(file string, token *oauth2.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
