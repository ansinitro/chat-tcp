package main

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func initLogger(file *os.File) {
	// Set the formatter for the logger
	logger.SetFormatter(new(logrus.JSONFormatter))

	// Create a MultiWriter with the file and os.Stdout
	multiWriter := io.MultiWriter(file, os.Stdout)

	// Set the output to the MultiWriter
	logger.SetOutput(multiWriter)
}

func initLogFile() (*os.File, error) {
	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Handle error opening the file
		logger.Fatal("Unable to open log file:", err)
	}

	return file, err
}
