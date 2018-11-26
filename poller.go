package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"github.com/michielappelman/leaderboard"
)

func pollLeaderboard(room string, postWhenUnchanged bool) error {
	// Open Datastore connection
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
		return err
	}

	var r Room
	key := datastore.NameKey("Room", room, nil)
	if err = client.Get(ctx, key, &r); err != nil {
		return err
	}

	members, err := leaderboard.GetMembers(r.LeaderboardID, r.APIKey, r.Year, leaderboard.SortByLocalScore)
	if err != nil {
		log.Printf("get leaderboard members failed: %v", err)
		sendReply(room, "Retrieving the leaderboard failed. Please check the validity of the session cookie.")
		return err
	}
	oldStarCount := r.LastStarCount
	newStarCount := leaderboard.CountTotalStars(members)
	r.LastStarCount = newStarCount

	if _, err := client.Put(ctx, key, &r); err != nil {
		log.Printf("failed to store new star count: %v", err)
		return err
	}

	if newStarCount != oldStarCount || postWhenUnchanged {
		updateRoom(&r, members)
	}
	return nil
}

func isKnownRoom(roomID string) (bool, error) {
	// Open Datastore connection
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
		return false, err
	}

	var room Room
	key := datastore.NameKey("Room", roomID, nil)
	err = client.Get(ctx, key, &room)
	switch {
	case err == datastore.ErrNoSuchEntity:
		return false, nil
	case err != nil:
		return false, err
	}
	return true, nil
}

func getRoomKeys() ([]string, error) {
	// Open Datastore connection
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
		return nil, err
	}

	// Request all Room keys
	query := datastore.NewQuery("Room").KeysOnly()
	keys, err := client.GetAll(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	var keyStrings []string
	for _, k := range keys {
		keyStrings = append(keyStrings, k.Name)
	}
	return keyStrings, nil
}

func updateRoom(r *Room, members []leaderboard.Member) error {
	message := fmt.Sprintf("### Leaderboard 🎄 %d\n\n", r.Year)
	for _, m := range members {
		var name string
		if m.Name == "" {
			name = m.ID
		} else {
			name = m.Name
		}
		message += fmt.Sprintf(" 1. %s – **%d** ⭐ %d",
			name, m.LocalScore, m.Stars)
		if m.GlobalScore > 0 {
			message += fmt.Sprintf(" (🌍 _%d_!)", m.GlobalScore)
		}
		message += "\n"
	}
	sendReply(r.Room, message)
	return nil
}
