package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type ProxyStats struct {
	ProxyID      string   `json:"proxy_id"`
	TargetID     string   `json:"target_id"`
	Timestamp    int64    `json:"timestamp"`
	RequestCount int      `json:"request_count"`
	ErrorCount   int      `json:"error_count"`
	UniqueUsers  []string `json:"unique_users"`
}

// checkKafkaConnection attempts to establish a connection to Kafka
func checkKafkaConnection(ctx context.Context, kafkaURL string) error {
	// Create a dialer with timeout
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true, // Enable both IPv4 and IPv6
	}

	// Try to connect to Kafka
	conn, err := dialer.DialContext(ctx, "tcp", kafkaURL)
	if err != nil {
		// Check for specific error types
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("kafka connection timeout: deadline exceeded")
		} else if errors.As(err, &netErr) && netErr.Timeout() {
			return errors.New("kafka connection timeout: network timeout")
		}
		return err
	}
	defer conn.Close()

	return nil
}

// isTopicEmpty checks if the topic is empty (no messages to consume)
// This is a best-effort check that should be used as a hint, not a definitive answer
func isTopicEmpty(r *kafka.Reader) bool {
	// Check consumer lag - this is the most reliable indicator
	lag := r.Lag()
	return lag == 0
}

func main() {
	// Context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a WaitGroup to wait for the consumer to finish
	var wg sync.WaitGroup

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	kafkaURL := os.Getenv("kafkaURL")
	topic := os.Getenv("topic")
	groupID := os.Getenv("groupID")

	// Kafka configuration with more robust settings
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          []string{kafkaURL},
		Topic:            topic,
		GroupID:          groupID,
		MinBytes:         10e3,                   // 10KB
		MaxBytes:         10e6,                   // 10MB
		ReadLagInterval:  1 * time.Minute,        // How often to update lag info
		MaxWait:          500 * time.Millisecond, // Short wait for empty topics (prevents long blocking)
		ReadBackoffMin:   100 * time.Millisecond, // Minimum backoff time
		ReadBackoffMax:   5 * time.Second,        // Maximum backoff time
		CommitInterval:   1 * time.Second,        // How often to commit offsets
		SessionTimeout:   30 * time.Second,       // Consumer group session timeout
		RebalanceTimeout: 30 * time.Second,       // Consumer group rebalance timeout
		RetentionTime:    24 * time.Hour,         // Retention policy
		StartOffset:      kafka.FirstOffset,      // Start from the oldest message if no offset is committed
	})

	// Set up a connection check
	connCtx, connCancel := context.WithTimeout(ctx, 30*time.Second)
	defer connCancel()

	// Verify Kafka connection before starting consumer
	if err := checkKafkaConnection(connCtx, kafkaURL); err != nil {
		log.Printf("Warning: Initial Kafka connection check failed: %v. Will retry in consumer loop.", err)
	}

	// PostgreSQL connection
	connStr := "postgresql://abtest:abtest@postgres:5432/abtest?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Prepare insert statement
	stmt, err := db.Prepare(`
		INSERT INTO proxy_stats (
			proxy_id, 
			target_id, 
			timestamp, 
			request_count, 
			error_count, 
			unique_users
		) VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		log.Fatal("Error preparing statement:", err)
	}

	// Start the consumer in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer r.Close()
		defer db.Close()
		defer stmt.Close()

		emptyTopicBackoff := 5 * time.Second

		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down consumer...")
				return
			default:
				// Check if topic appears empty before trying to read
				if isTopicEmpty(r) {
					log.Println("Topic appears empty, waiting for new messages...")
					time.Sleep(emptyTopicBackoff)
					continue
				}

				// Create a context with timeout for reading messages
				msgCtx, msgCancel := context.WithTimeout(ctx, 5*time.Second)

				// Try to read a message
				m, err := r.ReadMessage(msgCtx)
				msgCancel()

				if err != nil {
					// Check if this is just a timeout due to empty topic
					if errors.Is(err, context.DeadlineExceeded) {
						log.Println("No messages available, waiting...")
						time.Sleep(emptyTopicBackoff)
						continue
					}

					// Handle other errors
					log.Printf("Error reading message: %v, retrying...", err)
					time.Sleep(1 * time.Second)
					continue
				}

				// Process the message
				var stats ProxyStats
				if err := json.Unmarshal(m.Value, &stats); err != nil {
					log.Println("Error unmarshaling message:", err)
					continue
				}

				// Convert unique users to JSONB
				uniqueUsersJSON, err := json.Marshal(stats.UniqueUsers)
				if err != nil {
					log.Println("Error marshaling unique users:", err)
					continue
				}

				// Convert Unix timestamp to time.Time
				timestamp := time.Unix(stats.Timestamp, 0)

				// Insert into database with context
				dbCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
				_, err = stmt.ExecContext(dbCtx,
					stats.ProxyID,
					stats.TargetID,
					timestamp,
					stats.RequestCount,
					stats.ErrorCount,
					uniqueUsersJSON,
				)
				dbCancel()

				if err != nil {
					log.Println("Error inserting into database:", err)
					continue
				}

				log.Printf("Successfully processed message for proxy %s\n", stats.ProxyID)
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal. Starting graceful shutdown...")

	// Trigger shutdown
	cancel()

	// Wait for consumer to finish with timeout
	shutdownChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownChan)
	}()

	// Wait for graceful shutdown with timeout
	select {
	case <-shutdownChan:
		log.Println("Graceful shutdown completed")
	case <-time.After(30 * time.Second):
		log.Println("Shutdown timed out after 30 seconds")
	}
}
