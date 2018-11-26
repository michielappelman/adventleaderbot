package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

type Room struct {
	Room          string
	LeaderboardID int
	APIKey        string
	APIKeySetBy   string
	LastDataPoll  time.Time
	LastStarCount int
	Year          int
}

var authToken string
var projectID string

func main() {
	authToken = os.Getenv("WEBEX_TEAMS_TOKEN")
	if authToken == "" {
		fmt.Fprintf(os.Stderr, "WEBEX_TEAMS_TOKEN environment variable must be set.\n")
		os.Exit(1)
	}
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
		os.Exit(1)
	}

	http.HandleFunc("/poll/", pollHandler)
	http.HandleFunc("/webhook", webhookHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func pollHandler(w http.ResponseWriter, r *http.Request) {
	if err := createWebhook(); err != nil {
		log.Printf("failed to create webhook: %v", err)
	}

	if r.Method != "GET" {
		http.Error(w, "Only GETs allowed here.", http.StatusBadRequest)
		return
	}
	pathParts := strings.Split(r.URL.Path, "/")
	roomID := pathParts[len(pathParts)-1]

	if roomID == "all" {
		rooms, err := getRoomKeys()
		if err != nil {
			log.Printf("Failed to get rooms: %v", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		for _, r := range rooms {
			pollLeaderboard(r, false)
		}
	} else {
		if known, err := isKnownRoom(roomID); err == nil && known {
			pollLeaderboard(roomID, true)
		}
	}
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POSTs allowed here.", http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ReadAll: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	var hook Webhook
	if err := json.Unmarshal(body, &hook); err != nil {
		log.Printf("JSON Unmarshal: %v", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}
	// Ignore the bots' own messages
	if hook.Message.PersonEmail == "AdventLeaderBot@webex.bot" {
		http.Error(w, "Success", http.StatusOK)
		return
	}
	messageText, err := getMessageText(hook.Message.ID)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
	}

	fields := strings.Fields(messageText)
	if fields[0] == "AdventLeaderBot" {
		messageText = strings.Join(fields[1:], " ")
	}
	hook.Message.Text = messageText

	// Not an error, but success.
	http.Error(w, "Success", http.StatusOK)

	// Open Datastore connection
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}

	switch strings.Fields(hook.Message.Text)[0] {
	case "register":
		register(ctx, client, hook.Message)
	case "poll":
		poll(ctx, client, hook.Message)
	case "help":
		help(ctx, client, hook.Message)
	case "year":
		year(ctx, client, hook.Message)
	default:
		invalid(hook.Message)
	}
}
