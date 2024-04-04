package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// The Lobby listens for messages on its channels and manages the list of
// currently connected clients and active chat rooms.
type Lobby struct {
	clients   []*Client
	chatRooms map[string]*ChatRoom
	incoming  chan *Message
	join      chan *Client
	leave     chan *Client
	delete    chan *ChatRoom
}

// Establishes a lobby that initiates listening on its channels.
func NewLobby() *Lobby {
	lobby := &Lobby{
		clients:   make([]*Client, 0),
		chatRooms: make(map[string]*ChatRoom),
		incoming:  make(chan *Message),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		delete:    make(chan *ChatRoom),
	}
	lobby.Listen()

	return lobby
}

// Initiates a new goroutine to listen over the various channels of the Lobby.
func (lobby *Lobby) Listen() {
	go func() {
		for {
			select {
			case message := <-lobby.incoming:
				lobby.Parse(message)
			case client := <-lobby.join:
				lobby.Join(client)
			case client := <-lobby.leave:
				lobby.Leave(client)
			case chatRoom := <-lobby.delete:
				lobby.DeleteChatRoom(chatRoom)
			}
		}
	}()
}

// Manages the process of clients connecting to the lobby.
func (lobby *Lobby) Join(client *Client) {
	if len(lobby.clients) >= MAX_CLIENTS {
		client.Quit()
		return
	}
	lobby.clients = append(lobby.clients, client)
	fmt.Println(len(lobby.clients))
	client.outgoing <- MSG_CONNECT
	go func() {
		for message := range client.incoming {
			lobby.incoming <- message
		}
		lobby.leave <- client
	}()
}

// Manages the process of clients disconnecting from the lobby.
func (lobby *Lobby) Leave(client *Client) {
	if client.chatRoom != nil {
		client.chatRoom.Leave(client)
	}
	for i, otherClient := range lobby.clients {
		if client == otherClient {
			lobby.clients = append(lobby.clients[:i], lobby.clients[i+1:]...)
			break
		}
	}
	close(client.outgoing)
	logger.Info("Closed client's outgoing channel")
}

// Examines whether a channel has expired. If it has, the chat room is removed.
// Otherwise, a signal is dispatched to the delete channel for its updated expiry time.
func (lobby *Lobby) DeleteChatRoom(chatRoom *ChatRoom) {
	if chatRoom.expiry.After(time.Now()) {
		go func() {
			time.Sleep(chatRoom.expiry.Sub(time.Now()))
			lobby.delete <- chatRoom
		}()
		log.Println("attempted to delete chat room")
	} else {
		chatRoom.Delete()
		delete(lobby.chatRooms, chatRoom.name)
		log.Println("deleted chat room")
	}
}

// Manages incoming messages directed to the lobby. If the message includes a command,
// the lobby executes the command. Otherwise, it forwards the message to the sender's
// active chat room.
func (lobby *Lobby) Parse(message *Message) {
	switch {
	default:
		lobby.SendMessage(message)
	case strings.HasPrefix(message.text, CMD_CREATE):
		name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_CREATE+" "), "\n")
		lobby.CreateChatRoom(message.client, name)
	case strings.HasPrefix(message.text, CMD_LIST):
		lobby.ListChatRooms(message.client)
	case strings.HasPrefix(message.text, CMD_JOIN):
		name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_JOIN+" "), "\n")
		lobby.JoinChatRoom(message.client, name)
	case strings.HasPrefix(message.text, CMD_LEAVE):
		lobby.LeaveChatRoom(message.client)
	case strings.HasPrefix(message.text, CMD_NAME):
		name := strings.TrimSuffix(strings.TrimPrefix(message.text, CMD_NAME+" "), "\n")
		lobby.ChangeName(message.client, name)
	case strings.HasPrefix(message.text, CMD_HELP):
		lobby.Help(message.client)
	case strings.HasPrefix(message.text, CMD_QUIT):
		message.client.Quit()
	}
}

// Tries to dispatch the provided message to the client's current chat room. If the client
// is not currently in a chat room, an error message is sent back to the client.
func (lobby *Lobby) SendMessage(message *Message) {
	if message.client.chatRoom == nil {
		message.client.outgoing <- ERROR_SEND
		logger.Info("client tried to send message in lobby")
		return
	}
	message.client.chatRoom.Broadcast(message.String())
	logger.Info("client sent message")
}

// Tries to establish a chat room with the specified name, assuming it doesn't already exist.
func (lobby *Lobby) CreateChatRoom(client *Client, name string) {
	if lobby.chatRooms[name] != nil {
		client.outgoing <- ERROR_CREATE
		logger.Info("client tried to create chat room with a name already in use")
		return
	}
	chatRoom := NewChatRoom(name)
	lobby.chatRooms[name] = chatRoom
	go func() {
		time.Sleep(EXPIRY_TIME)
		lobby.delete <- chatRoom
	}()
	client.outgoing <- fmt.Sprintf(NOTICE_PERSONAL_CREATE, chatRoom.name)
	logger.Info("client created chat room")
}

// Tries to include the client into the chat room with the specified name, assuming the chat room exists.
func (lobby *Lobby) JoinChatRoom(client *Client, name string) {
	if lobby.chatRooms[name] == nil {
		client.outgoing <- ERROR_JOIN
		logger.Info("client tried to join a chat room that does not exist")
		return
	}
	if client.chatRoom != nil {
		lobby.LeaveChatRoom(client)
	}
	lobby.chatRooms[name].Join(client)
	logger.Info("client joined chat room")
}

// Excludes the specified client from their current chat room.
func (lobby *Lobby) LeaveChatRoom(client *Client) {
	if client.chatRoom == nil {
		client.outgoing <- ERROR_LEAVE
		logger.Info("client tried to leave the lobby")
		return
	}
	client.chatRoom.Leave(client)
	logger.Info("client left chat room")
}

// Updates the client's name to the provided name.
func (lobby *Lobby) ChangeName(client *Client, name string) {
	if client.chatRoom == nil {
		client.outgoing <- fmt.Sprintf(NOTICE_PERSONAL_NAME, name)
	} else {
		client.chatRoom.Broadcast(fmt.Sprintf(NOTICE_ROOM_NAME, client.name, name))
	}
	client.name = name
	logger.Info("client changed their name")
}

// Dispatches to the client the list of currently available chat rooms.
func (lobby *Lobby) ListChatRooms(client *Client) {
	client.outgoing <- "\n"
	client.outgoing <- "Chat Rooms:\n"
	for name := range lobby.chatRooms {
		client.outgoing <- fmt.Sprintf("%s\n", name)
	}
	client.outgoing <- "\n"
	logger.Info("client listed chat rooms")
}

// Transmits to the client the list of available commands.
func (lobby *Lobby) Help(client *Client) {
	client.outgoing <- "\n"
	client.outgoing <- "Commands:\n"
	client.outgoing <- "/help - lists all commands\n"
	client.outgoing <- "/list - lists all chat rooms\n"
	client.outgoing <- "/create foo - creates a chat room named foo\n"
	client.outgoing <- "/join foo - joins a chat room named foo\n"
	client.outgoing <- "/leave - leaves the current chat room\n"
	client.outgoing <- "/name foo - changes your name to foo\n"
	client.outgoing <- "/quit - quits the program\n"
	client.outgoing <- "\n"
	logger.Info("client requested help")
}
