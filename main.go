package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func getClient(config *oauth2.Config) *http.Client {
	// Local webserver flow with redirect to localhost
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Unable to start local server: %v", err)
	}
	defer ln.Close()

	state := "state-token"
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Open the following URL in the browser:\n%v\n", url)

	// Wait for redirect with code
	codeCh := make(chan string)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("state") != state {
				http.Error(w, "State does not match", http.StatusBadRequest)
				return
			}
			code := r.URL.Query().Get("code")
			fmt.Fprint(w, "Authorization complete. You can close this window.")
			codeCh <- code
		})
		http.Serve(ln, nil)
	}()

	code := <-codeCh
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to exchange code for token: %v", err)
	}

	saveToken(token)
	return config.Client(context.Background(), token)
}

func tokenCacheFile() string {
	usr, _ := user.Current()
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, "calendar-token.json")
}

func saveToken(token *oauth2.Token) {
	file := tokenCacheFile()
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	fmt.Printf("Token saved to %s\n", file)
}
func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials.json: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// Example: Replace with your actual event ID from listEvents()
	eventId := "PUT_YOUR_EVENT_ID_HERE"

	// Use the functions below to test
	updateEvent(srv, "primary", eventId)
	// deleteEvent(srv, "primary", eventId)
	deleteEvent(srv, "primary", eventId)
}
func updateEvent(srv *calendar.Service, calendarId, eventId string) {
	event, err := srv.Events.Get(calendarId, eventId).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve event: %v", err)
	}

	// Modify event fields
	event.Summary = "Updated Event Title"
	event.Location = "Updated Location"
	event.Description = "Updated Description"
	startTime := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(49 * time.Hour).Format(time.RFC3339)
	event.Start = &calendar.EventDateTime{DateTime: startTime, TimeZone: "Asia/Kolkata"}
	event.End = &calendar.EventDateTime{DateTime: endTime, TimeZone: "Asia/Kolkata"}

	updatedEvent, err := srv.Events.Update(calendarId, event.Id, event).Do()
	if err != nil {
		log.Fatalf("Unable to update event: %v", err)
	}
	fmt.Printf("Event updated: %s\n", updatedEvent.HtmlLink)
}
func deleteEvent(srv *calendar.Service, calendarId, eventId string) {
	err := srv.Events.Delete(calendarId, eventId).Do()
	if err != nil {
		log.Fatalf(" Failed to delete event: %v", err)
	}
	fmt.Println(" Event deleted successfully.")
}
