package main

import (
	"fmt"
	"log"
	"os"

	webex "github.com/jbogarin/go-cisco-webex-teams/sdk"
	resty "gopkg.in/resty.v1"
)

type Webhook struct {
	Message webex.Message `json:"data"`
}

func sendReply(room string, text string) {
	client := resty.New()
	client.SetAuthToken(authToken)
	webex_client := webex.NewClient(client)

	markDownMessage := &webex.MessageCreateRequest{
		Markdown: text,
		RoomID:   room,
	}
	_, _, err := webex_client.Messages.CreateMessage(markDownMessage)
	if err != nil {
		log.Printf("failed to send message: %s", text)
	}
}

func createWebhook() error {
	client := resty.New()
	client.SetAuthToken(authToken)
	webex_client := webex.NewClient(client)

	webhooksQueryParams := &webex.ListWebhooksQueryParams{
		Max: 10,
	}
	webhooks, _, err := webex_client.Webhooks.ListWebhooks(webhooksQueryParams)
	if err != nil {
		log.Fatal(err)
	}
	for id, webhook := range webhooks.Items {
		fmt.Println("GET:", id, webhook.ID, webhook.Name, webhook.TargetURL, webhook.Created)
	}

	webhookTargetURL := fmt.Sprintf("https://%s.appspot.com/webhook", os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if os.Getenv("DEBUG_BOT_WEBHOOK") != "" {
		webhookTargetURL = os.Getenv("DEBUG_BOT_WEBHOOK")
	}
	if len(webhooks.Items) == 0 {
		webhookRequest := &webex.WebhookCreateRequest{
			Name:      "AdventLeaderBot",
			TargetURL: webhookTargetURL,
			Resource:  "messages",
			Event:     "created",
		}
		_, _, err := webex_client.Webhooks.CreateWebhook(webhookRequest)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

func getMessageText(mID string) (string, error) {
	client := resty.New()
	client.SetAuthToken(authToken)
	webex_client := webex.NewClient(client)

	m, _, err := webex_client.Messages.GetMessage(mID)
	if err != nil {
		log.Printf("could not get message ID %s", mID)
		return "", err
	}
	return m.Text, nil
}
