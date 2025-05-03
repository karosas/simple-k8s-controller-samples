package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	//"postgres://username:password@hostname/db_name?sslmode=disable"
	connStr := os.Getenv("CONNECTION_STRING")

	// Create a channel to listen for interrupt or termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Ticker for periodic health checks
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Starting PostgreSQL health checker... (press Ctrl+C to stop)")

	for {
		select {
		case <-ticker.C:
			err := checkDBHealth(connStr)
			if err != nil {
				log.Printf("[UNHEALTHY] Database ping failed: %v\n", err)
			} else {
				log.Println("[HEALTHY] Database is reachable")
			}
		case <-stop:
			log.Println("Received shutdown signal. Exiting gracefully...")
			return
		}
	}
}

func checkDBHealth(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Failed to close connection")
		}
	}(db)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
