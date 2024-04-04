package main

import (
	"fmt"
	"time"
)

// A message comprises a timestamp, a client, and text content.
type Message struct {
	time   time.Time
	client *Client
	text   string
}

// Generates a new message with the provided timestamp, client, and text.
func NewMessage(time time.Time, client *Client, text string) *Message {
	return &Message{
		time:   time,
		client: client,
		text:   text,
	}
}

// Generates a string representation of the message.
func (message *Message) String() string {
	return fmt.Sprintf("%s - %s: %s\n", message.time.Format(time.Kitchen), message.client.name, message.text)
}
