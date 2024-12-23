package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"singdns/api"
	"singdns/api/auth"
	"singdns/api/proxy"
	"singdns/api/storage"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Parse command line flags
	var (
		configPath string
		dbPath     string
		logLevel   string
		adminUser  string
		adminPass  string
		jwtSecret  string
	)

	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.StringVar(&dbPath, "db", "data/singdns.db", "Path to SQLite database file")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&adminUser, "admin", "admin", "Admin username")
	flag.StringVar(&adminPass, "password", "admin", "Admin password")
	flag.StringVar(&jwtSecret, "secret", "your-secret-key", "JWT secret key")
	flag.Parse()

	// Initialize logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Fatal(err)
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create data directory if not exists
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Fatal(err)
	}

	// Initialize storage
	db, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Initialize proxy manager
	proxyManager := proxy.NewManager(logger, configPath, db)

	// Hash admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		logger.Fatal(err)
	}

	// Initialize API server
	config := &api.Config{
		UpdateInterval: 24 * time.Hour, // Default update interval
		Auth:           auth.NewManager(adminUser, string(hashedPassword), []byte(jwtSecret)),
	}
	server := api.NewServer(db, proxyManager, logger, config)

	// Start server
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal(err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown server
	server.Stop()
	fmt.Println("Server stopped")
}
