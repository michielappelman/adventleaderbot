package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	webex "github.com/jbogarin/go-cisco-webex-teams/sdk"
)

func register(ctx context.Context, client *datastore.Client, m webex.Message) {
	log.Printf("register")

	parameters := strings.Fields(m.Text)
	if len(parameters) != 3 {
		sendReply(m.RoomID, "Please specify the Advent of Code *leaderboard ID* and *API key*. See `help` for detailed usage")
		return
	}

	// Describe room and associated leaderboard
	roomKey := datastore.NameKey("Room", m.RoomID, nil)
	lbInt, err := strconv.Atoi(parameters[1])
	if err != nil {
		sendReply(m.RoomID, fmt.Sprintf("Leaderboard ID %s seems to be an invalid integer", parameters[1]))
		return
	}
	room := Room{
		Room:          m.RoomID,
		LeaderboardID: lbInt,
		APIKey:        parameters[2],
		APIKeySetBy:   m.PersonEmail,
		LastDataPoll:  time.Now(),
		LastStarCount: 0,
		Year:          2018,
	}

	// Save Room to datastore
	if _, err := client.Put(ctx, roomKey, &room); err != nil {
		log.Printf("failed to save room: %v", err)
	}

	sendReply(m.RoomID, "The leaderboard ID and API key were stored")
}

func year(ctx context.Context, client *datastore.Client, m webex.Message) {
	log.Printf("year")

	parameters := strings.Fields(m.Text)
	year, err := strconv.Atoi(parameters[1])
	if err != nil {
		sendReply(m.RoomID, fmt.Sprintf("%s is not a valid year (or at least integer).", parameters[1]))
		return
	}
	// Describe room and associated leaderboard
	roomKey := datastore.NameKey("Room", m.RoomID, nil)

	_, err = client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var room Room
		if err := tx.Get(roomKey, &room); err != nil {
			return err
		}
		room.Year = year
		_, err := tx.Put(roomKey, &room)
		return err
	})

	sendReply(m.RoomID, fmt.Sprintf("The year was changed to %d for this room.", year))
}

func poll(ctx context.Context, client *datastore.Client, m webex.Message) {
	log.Printf("poll")
	pollTime := time.Now()

	var room Room
	key := datastore.NameKey("Room", m.RoomID, nil)
	if err := client.Get(ctx, key, &room); err != nil {
		log.Printf("failed to get room %s", m.RoomID)
	}
	if pollTime.Sub(room.LastDataPoll) < 60000000000 {
		sendReply(m.RoomID, "Your last poll was less than 1 minute ago...")
		return
	}

	room.LastDataPoll = pollTime
	if _, err := client.Put(ctx, key, &room); err != nil {
		log.Printf("failed to store new poll time: %v", err)
	}

	pollLeaderboard(m.RoomID, true)
}

func help(ctx context.Context, client *datastore.Client, m webex.Message) {
	log.Printf("help")
	message := "Hi, this is the [Advent of Code](https://adventofcode.com/) chatbot! 🎄 " +
		"Say `register` to me, followed by the Advent of Code leaderboard ID and session cookie, to" +
		" register this room to a certain leaderboard. Eg. `@AdventLeaderBot register 12345 46c281...`" +
		" *You can delete the message afterwards if you're afraid someone will steal your cookie. 🍪*\n\n" +
		"The leaderboard will be checked every 5 minutes for changes, but you can poll the current" +
		" status earlier by saying `poll` to me. **Have fun!**"

	var room Room
	key := datastore.NameKey("Room", m.RoomID, nil)
	if err := client.Get(ctx, key, &room); err == nil {
		message = message + fmt.Sprintf("\n\nThe current leaderboard is "+
			"[%d](https://adventofcode.com/%d/leaderboard/private/view/%d), set by %s.",
			room.LeaderboardID, room.Year, room.LeaderboardID, room.APIKeySetBy)
	} else if err == datastore.ErrNoSuchEntity {
		message = message + " *(There are not settings found for this room, BTW.)*\n"
	}
	sendReply(m.RoomID, message)
}

func invalid(m webex.Message) {
	log.Printf("Invalid Command: %s", m.Text)
}
