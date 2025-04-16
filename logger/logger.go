package logger

import (
	"log"
	"os"
	"path/filepath"
)

var (
	Logger *log.Logger
)

func Init() {
	// Check if directory exists
	logDir := "./logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("Could not create directory: " + err.Error())
	}

	// Create log files
	logFile := filepath.Join(logDir, "app.log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open log file: " + err.Error())
	}

	// Set up logger to write to both file and stdout
	Logger = log.New(file, "", log.LstdFlags)
	Logger.SetOutput(os.Stdout)
}
