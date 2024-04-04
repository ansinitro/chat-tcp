package main

import (
	"bufio"
	"net"
	"strings"
	"time"
)

// A client simplifies the concept of a connection by using incoming and
// outgoing channels, and retains certain details about the client's status,
// such as their present name and chat room.
type Client struct {
	name     string
	chatRoom *ChatRoom
	incoming chan *Message
	outgoing chan string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
}

// This function creates a new client using the provided connection.
// Additionally, it initializes and starts separate goroutines for reading
// from and writing to the socket associated with the connection.
func NewClient(conn net.Conn) *Client {
	client := &Client{
		name:     CLIENT_NAME,
		chatRoom: nil,
		incoming: make(chan *Message),
		outgoing: make(chan string),
		conn:     conn,
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
	}

	client.Listen()
	return client
}

// This code initiates two concurrent operations. The first operation reads
// from the outgoing channel of the client and writes to the client's socket
// connection. The second operation reads from the client's socket connection
// and writes to the client's incoming channel.
func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

// This code reads strings from the socket connected to the client, formats them
// into messages, and then places these messages into the client's incoming channel.
func (client *Client) Read() {
	for {
		str, err := client.reader.ReadString('\n')
		if err != nil {
			logger.Errorf("error reading in client: %s", err.Error())
			break
		}
		message := NewMessage(time.Now(), client, strings.ReplaceAll(str, "\r\n", ""))
		client.incoming <- message
	}
	close(client.incoming)
	logger.Info("Closed client's incoming channel read thread")
}

// This function reads messages from the outgoing channel of the client and writes
// them to the socket associated with the client.
func (client *Client) Write() {
	for str := range client.outgoing {
		_, err := client.writer.WriteString(str)
		if err != nil {
			logger.Errorf("error when WriteString() in clien.Write(): %s", err.Error())
			break
		}
		err = client.writer.Flush()
		if err != nil {
			logger.Errorf("error when Flush() in clien.Write(): %s", err.Error())
			break
		}
	}
	logger.Info("Closed client's write thread")
}

// This function closes the connection for the client. It relies on error handling
// to close the socket, simplifying the code and ensuring that all associated threads
// are properly cleaned up.
func (client *Client) Quit() {
	client.conn.Close()
}
