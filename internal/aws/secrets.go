package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsClient struct {
	client *secretsmanager.Client
}

type AppSecrets struct {
	OddsAPIKey  string `json:"ODDS_API_KEY"`
	DatabaseURL string `json:"DATABASE_URL,omitempty"`
}

// NewSecretsClient creates a new AWS Secrets Manager client
func NewSecretsClient() (*SecretsClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return &SecretsClient{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}

// GetAppSecrets retrieves application secrets from AWS Secrets Manager
// Falls back to environment variables if running locally or if secret is not found
func (s *SecretsClient) GetAppSecrets(secretName string) (*AppSecrets, error) {
	// First try to get from environment variables (for local development)
	if oddsKey := os.Getenv("ODDS_API_KEY"); oddsKey != "" {
		log.Println("Using ODDS_API_KEY from environment variable (local development)")
		return &AppSecrets{
			OddsAPIKey: oddsKey,
		}, nil
	}

	// Try to get from AWS Secrets Manager
	log.Printf("Attempting to retrieve secrets from AWS Secrets Manager: %s", secretName)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := s.client.GetSecretValue(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secret %s: %w", secretName, err)
	}

	var secrets AppSecrets
	if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	log.Println("Successfully retrieved secrets from AWS Secrets Manager")
	return &secrets, nil
}

// GetOddsAPIKey is a convenience method to get just the Odds API key
func (s *SecretsClient) GetOddsAPIKey(secretName string) (string, error) {
	secrets, err := s.GetAppSecrets(secretName)
	if err != nil {
		return "", err
	}
	return secrets.OddsAPIKey, nil
}

// GetDatabaseURL is a convenience method to get the database URL if stored in secrets
func (s *SecretsClient) GetDatabaseURL(secretName string) (string, error) {
	secrets, err := s.GetAppSecrets(secretName)
	if err != nil {
		return "", err
	}

	// Return database URL from secrets if available, otherwise empty string
	return secrets.DatabaseURL, nil
}