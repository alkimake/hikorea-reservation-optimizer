package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Struct for the message payload
type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// Function to send a message using the Telegram API
func sendTelegramMessage(botToken, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	// Create the message payload
	msg := TelegramMessage{
		ChatID: chatID,
		Text:   message,
	}

	// Convert the payload to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Send the HTTP request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}

func SendGenericMessage(message string) {
	// Retrieve bot token and chat ID from environment variables or replace with actual values
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN") // or replace with your token
	chatID := os.Getenv("TELEGRAM_CHAT_ID")     // or replace with your chat ID

	// Send the message
	err := sendTelegramMessage(botToken, chatID, message)
	if err != nil {
		fmt.Printf("Error sending message: %s\n", err)
		return
	}

	log.Println("Message sent successfully!")
}

func SendToTelegram(day string) {
	message := "A new date is available: " + day
	SendGenericMessage(message)
}
