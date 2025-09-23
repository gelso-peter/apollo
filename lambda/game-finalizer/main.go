package main

import (
	"apollo/db"
	"apollo/internal/aws"
	"apollo/services/cron"
	"apollo/services/odds.go"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LambdaHandler handles the Lambda function execution
func LambdaHandler(ctx context.Context) error {
	log.Println("Starting Lambda game finalization process...")

	// Initialize AWS Secrets Manager client
	secretsClient, err := aws.NewSecretsClient()
	if err != nil {
		return fmt.Errorf("failed to initialize AWS Secrets Manager client: %w", err)
	}

	// Get secrets from AWS Secrets Manager
	secretName := os.Getenv("AWS_SECRET_NAME")
	if secretName == "" {
		secretName = "apollo/api-keys" // Default secret name
	}

	secrets, err := secretsClient.GetAppSecrets(secretName)
	if err != nil {
		return fmt.Errorf("failed to retrieve secrets: %w", err)
	}

	// Initialize database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try to get from secrets if not in environment
		if secrets.DatabaseURL != "" {
			dbURL = secrets.DatabaseURL
		} else {
			return fmt.Errorf("DATABASE_URL not found in environment or secrets")
		}
	}

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer dbPool.Close()

	// Test database connection
	if err = dbPool.Ping(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	log.Println("Database connected successfully!")

	// Initialize odds service
	odds.InitOddsService(secrets.OddsAPIKey)
	oddsService := odds.GetOddsService()

	// Create game finalizer
	gameFinalizer := cron.NewGameFinalizer(dbPool, oddsService)

	// Execute game finalization
	if err := gameFinalizer.FinalizeGames(); err != nil {
		return fmt.Errorf("game finalization failed: %w", err)
	}

	log.Println("Lambda game finalization completed successfully!")
	return nil
}

func main() {
	lambda.Start(LambdaHandler)
}