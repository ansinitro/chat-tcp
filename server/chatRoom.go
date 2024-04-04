package main

import (
	"fmt"
	"time"
)

// A ChatRoom includes the name of the chat, a roster of presently connected clients,
// a record of messages previously shared with users in the chat, and the current time
// when the ChatRoom is set to expire.
type ChatRoom struct {
	name     string
	clients  []*Client
	messages []string
	expiry   time.Time
}

// Generates a new chat room with the provided name and initializes its expiration time to
// the current time plus the predefined duration EXPIRY_TIME.
func NewChatRoom(name string) *ChatRoom {
	return &ChatRoom{
		name:     name,
		clients:  make([]*Client, 0),
		messages: make([]string, 0),
		expiry:   time.Now().Add(EXPIRY_TIME),
	}
}

// Includes the specified Client into the ChatRoom and forwards all messages sent
// since the ChatRoom was created to the Client.
func (chatRoom *ChatRoom) Join(client *Client) {
	client.chatRoom = chatRoom
	for _, message := range chatRoom.messages {
		client.outgoing <- message
	}
	chatRoom.clients = append(chatRoom.clients, client)
	chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_JOIN, client.name))
}

// Eliminates the specified Client from the ChatRoom.
func (chatRoom *ChatRoom) Leave(client *Client) {
	chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_LEAVE, client.name))
	for i, otherClient := range chatRoom.clients {
		if client == otherClient {
			chatRoom.clients = append(chatRoom.clients[:i], chatRoom.clients[i+1:]...)
			break
		}
	}
	client.chatRoom = nil
}

// Distributes the provided message to all Clients presently present in the ChatRoom.
func (chatRoom *ChatRoom) Broadcast(message string) {
	chatRoom.expiry = time.Now().Add(EXPIRY_TIME)
	chatRoom.messages = append(chatRoom.messages, message)
	for _, client := range chatRoom.clients {
		client.outgoing <- message
	}
}

// Informs the clients inside the chat room about its deletion and ejects them back to the lobby.
func (chatRoom *ChatRoom) Delete() {
	chatRoom.Broadcast(NOTICE_ROOM_DELETE)
	for _, client := range chatRoom.clients {
		client.chatRoom = nil
	}
}
