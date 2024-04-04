package main

import (
	"net"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// Initiates a lobby, listens for client connections, and links them to the lobby.
func main() {
	// Create a new file for logging
	file, err := initLogFile()
	defer file.Close()

	// Initialization logger
	initLogger(file)

	lobby := NewLobby()

	listener, err := net.Listen(CONN_TYPE, CONN_PORT)
	if err != nil {
		logrus.Fatalf("error creating tcp listener: %s", err.Error())
	}
	defer listener.Close()
	logrus.Info("Listening on port: " + CONN_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("error listener.Accept(): %s", err.Error())
			continue
		}
		lobby.Join(NewClient(conn))
	}
}
