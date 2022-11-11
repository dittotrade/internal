package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// SendMessage sends slack message to group defined by SLACK_NOTIFICATION_URL variable
func SendMessage(text string) error {
	// url for scoring trigger for calculate new trader point
	slackNotificationURL := os.Getenv("SLACK_NOTIFICATION_URL")
	if slackNotificationURL == "" {
		return fmt.Errorf("env SLACK_NOTIFICATION_URL is not set!")
	}

	var message struct {
		Text string `json:"text"`
	}
	message.Text = text
	jsonBody, err := json.MarshalIndent(&message, "  ", "  ")
	if err != nil {
		return err
	}

	r := bytes.NewReader(jsonBody)

	if _, err = http.Post(slackNotificationURL, "application/json", r); err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	return nil
}
