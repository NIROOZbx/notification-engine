package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/internal/utils"
)

type KeyInfo struct {
	APIKey         string `json:"api_key"`
	EventType      string `json:"event_type"`
	ExternalUserID string `json:"external_user_id"`
}

func main() {
	countFlag := flag.Int("count", 20, "Number of mock users/workspaces to seed")
	flag.Parse()

	count := *countFlag
	if count <= 0 {
		log.Fatalf("Count must be greater than 0")
	}

	fmt.Printf("Initializing seeding for %d test users...\n", count)

	// Load Viper config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}

	ctx := context.Background()

	// Connect to Database
	dbPool, err := pgxpool.New(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	var planID uuid.UUID
	err = dbPool.QueryRow(ctx, "SELECT id FROM plans WHERE name = 'Free' LIMIT 1").Scan(&planID)
	if err != nil {
		err = dbPool.QueryRow(ctx, "SELECT id FROM plans LIMIT 1").Scan(&planID)
		if err != nil {
			log.Fatalf("No plans found in database. Please run migrations first! Error: %v", err)
		}
	}
	fmt.Printf("Using Plan ID: %s\n", planID)

	var keyInfos []KeyInfo

	for i := 1; i <= count; i++ {
		userID := uuid.New()
		workspaceID := uuid.New()
		devEnvID := uuid.New()
		templateID := uuid.New()
		apiKeyID := uuid.New()

		email := fmt.Sprintf("k6_user_%d_%s@example.com", i, uuid.New().String()[:8])
		wspName := fmt.Sprintf("K6 Workspace %d", i)
		slug := fmt.Sprintf("k6-workspace-%d-%s", i, uuid.New().String()[:8])
		eventType := "k6.test.event"
		subscriberExtID := "k6_subscriber_user"
		subscriberEmail := fmt.Sprintf("k6_sub_%d@example.com", i)

		// 1. Insert User
		_, err = dbPool.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, full_name, is_verified, is_active)
			VALUES ($1, $2, $3, $4, true, true)
		`, userID, email, "$2a$10$8CgDkF0sZ2nQY5qB4p5oOeJ9.X8bZ3y9qZ3w8u1r7a6f5e4d3c2b1", "K6 Loader") // bcrypt for password123
		if err != nil {
			log.Fatalf("Failed to insert user %d: %v", i, err)
		}

		// 2. Insert Workspace
		_, err = dbPool.Exec(ctx, `
			INSERT INTO workspaces (id, name, slug, plan_id)
			VALUES ($1, $2, $3, $4)
		`, workspaceID, wspName, slug, planID)
		if err != nil {
			log.Fatalf("Failed to insert workspace %d: %v", i, err)
		}

		// 3. Insert Workspace Member
		_, err = dbPool.Exec(ctx, `
			INSERT INTO workspace_members (id, workspace_id, user_id, role)
			VALUES ($1, $2, $3, 'owner')
		`, uuid.New(), workspaceID, userID)
		if err != nil {
			log.Fatalf("Failed to insert workspace member %d: %v", i, err)
		}

		// 4. Insert Development Environment
		_, err = dbPool.Exec(ctx, `
			INSERT INTO environments (id, workspace_id, name)
			VALUES ($1, $2, 'development')
		`, devEnvID, workspaceID)
		if err != nil {
			log.Fatalf("Failed to insert environment %d: %v", i, err)
		}

		// 5. Insert Template (Event Type)
		_, err = dbPool.Exec(ctx, `
			INSERT INTO templates (id, workspace_id, environment_id, created_by, name, description, event_type, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'live')
		`, templateID, workspaceID, devEnvID, userID, "K6 Load Test Event", "Event for k6 tests", eventType)
		if err != nil {
			log.Fatalf("Failed to insert template %d: %v", i, err)
		}

		// 6. Insert Template Channel (Email)
		_, err = dbPool.Exec(ctx, `
			INSERT INTO template_channels (template_id, channel, content)
			VALUES ($1, 'email', '{"subject": "K6 Ingestion Test: {{ref_id}}", "body": "<p>Hello {{first_name}}! Status check active.</p>"}')
		`, templateID)
		if err != nil {
			log.Fatalf("Failed to insert template channel %d: %v", i, err)
		}

		// 7. Insert Subscriber
		_, err = dbPool.Exec(ctx, `
			INSERT INTO user_info (workspace_id, environment_id, external_user_id, channel, contact_value, metadata, verified)
			VALUES ($1, $2, $3, 'email', $4, '{}', true)
		`, workspaceID, devEnvID, subscriberExtID, subscriberEmail)
		if err != nil {
			log.Fatalf("Failed to insert subscriber %d: %v", i, err)
		}

		// 8. Generate and Insert API Key (development environment -> prefix "ne_test_")
		rawKey, hashedKey, hint := utils.GenerateAPIKey("development")
		_, err = dbPool.Exec(ctx, `
			INSERT INTO api_keys (id, workspace_id, environment_id, label, key_hash, key_hint, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, apiKeyID, workspaceID, devEnvID, "k6-load-test-key", hashedKey, hint, userID)
		if err != nil {
			log.Fatalf("Failed to insert API key %d: %v", i, err)
		}

		// Save the raw info for the load test
		keyInfos = append(keyInfos, KeyInfo{
			APIKey:         rawKey,
			EventType:      eventType,
			ExternalUserID: subscriberExtID,
		})
	}

	// 9. Write outputs to json file
	outputDir := "k6"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory %s: %v", outputDir, err)
	}

	outputPath := filepath.Join(outputDir, "api_keys.json")
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(keyInfos); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}

	fmt.Printf("\nSuccessfully seeded %d workspaces! Data written to: %s\n", count, outputPath)
}
