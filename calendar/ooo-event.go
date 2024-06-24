package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Load client secrets from a file.
func getClient(credentialsFile string) (*http.Client, error) {
	// Read the credentials file
	//b, err := os.ReadFile(credentialsFile)
	//if err != nil {
	//	return nil, fmt.Errorf("unable to read client secret file: %v", err)
	//}

	// Parse the credentials file
	config, err := google.ConfigFromJSON([]byte(credentialsFile), calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %v", err)
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

// Request a token from the web, then returns the retrieved token.
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

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Event structure to unmarshal JSON data
type Event struct {
	Summary     string `json:"summary"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Start       struct {
		Date     string `json:"date"`
		TimeZone string `json:"timeZone"`
	} `json:"start"`
	End struct {
		Date     string `json:"date"`
		TimeZone string `json:"timeZone"`
	} `json:"end"`
	Attendees []struct {
		Email string `json:"email"`
	} `json:"attendees"`
}

func calList(calList *calendar.CalendarList, name string) (id string) {
	var calendarID string
	// Find the calendar ID by name
	if name == "primary" {
		return name
	}

	for _, item := range calList.Items {
		println(item.Description)
		println(item.Id)
		println(item.Summary)
		if item.Summary == name {
			calendarID = item.Id
			break
		}
	}
	//os.Exit(0)
	return calendarID
}

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	//config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	//if err != nil {
	//	log.Fatalf("Unable to parse client secret file to config: %v", err)
	//}

	// Set up the Google API client
	client, err := getClient(string(b))
	if err != nil {
		log.Fatalf("Failed to create Google Calendar client: %v", err)
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	//srv, err := calendar.NewService(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// Load the event details from the JSON file
	eventFile := "/tmp/ooo-event.json"
	eventData, err := os.ReadFile(eventFile)
	if err != nil {
		log.Fatalf("Unable to read event file: %v", err)
	}

	var event Event
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Fatalf("Unable to unmarshal event data: %v", err)
	}

	// Create the event
	calendarEvent := &calendar.Event{
		Summary:     event.Summary,
		Location:    event.Location,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			Date:     event.Start.Date,
			TimeZone: event.Start.TimeZone,
		},
		End: &calendar.EventDateTime{
			Date:     event.End.Date,
			TimeZone: event.End.TimeZone,
		},
		Attendees: []*calendar.EventAttendee{},
	}

	for _, attendee := range event.Attendees {
		calendarEvent.Attendees = append(calendarEvent.Attendees, &calendar.EventAttendee{Email: attendee.Email})
	}

	calendars, err := srv.CalendarList.List().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve calendar list: %v", err)
	}

	// TODO: primary and NOC are only for EverOps, Life360 does not have a NOC calendarId
	// calendarId needs to be passed on a per account basis
	calendarId := []string{"primary", "ECTest"}
	for _, v := range calendarId {
		fmt.Println(fmt.Printf("Calendar ID: %s\n", v))
		createdEvent, err := srv.Events.Insert(calList(calendars, v), calendarEvent).Do()
		if err != nil {
			log.Fatalf("Unable to create event: %v", err)
		}

		fmt.Printf("Event created: %s\n", createdEvent.HtmlLink)
	}
}
