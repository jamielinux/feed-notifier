package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jamielinux/feed-notifier/internal/config"
	"github.com/jamielinux/feed-notifier/internal/db"
	"github.com/jamielinux/feed-notifier/internal/logger"
	"github.com/jamielinux/feed-notifier/internal/service"
)

func main() {
	printUsage := func() {
		fmt.Fprintln(os.Stderr, "Usage: feed-notifier CONFIG_FILE")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printUsage()
	}

	configPath := os.Args[1]
	if _, err := os.Stat(configPath); err != nil {
		printUsage()
	}

	config, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	logger.DebugEnabled = config.Debug
	logger.Debug("Debug logging enabled")

	database, err := db.Open(config.Database)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer database.Close()

	service, err := service.New(config, database)
	if err != nil {
		log.Fatalf("Service initialization error: %v", err)
	}

	service.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	log.Printf("Received signal %v, shutting down...", sig)
	service.Stop()
}
