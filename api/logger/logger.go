package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

// InitLogger initializes the logger
func InitLogger(logDir string) error {
	// Create log directory if not exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	logFile := filepath.Join("logs", fmt.Sprintf("singdns_%s.log", time.Now().Format("2006-01-02")))
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Initialize loggers
	infoLogger = log.New(f, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger = log.New(f, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(f, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

// LogInfo logs an info message
func LogInfo(format string, v ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf(format, v...)
	}
}

// LogDebug logs a debug message
func LogDebug(format string, v ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf("DEBUG: "+format, v...)
	}
}

// LogWarning logs a warning message
func LogWarning(format string, v ...interface{}) {
	if warningLogger != nil {
		warningLogger.Printf(format, v...)
	}
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, v...)
	}
}

// LogRequest logs an HTTP request
func LogRequest(method, path, remoteAddr string, statusCode int, latency time.Duration) {
	LogInfo("Request: %s %s from %s, status: %d, latency: %v",
		method, path, remoteAddr, statusCode, latency)
}

// LogPanic logs a panic
func LogPanic(r interface{}) {
	if errorLogger != nil {
		errorLogger.Printf("Panic: %v", r)
	}
}
