package aws

import (
	"context"
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
// Returns nil during local development to force fallback to environment variables
func NewSecretsClient() (*SecretsClient, error) {
	// Check if we're running in an AWS execution environment
	awsExecutionEnv := os.Getenv("AWS_EXECUTION_ENV")
	useAwsSecrets := os.Getenv("USE_AWS_SECRETS")
	awsRegion := os.Getenv("AWS_REGION")

	// Only create AWS client if we're in AWS execution environment or explicitly enabled
	if awsExecutionEnv == "" && useAwsSecrets != "true" {
		// For local development, return nil to force fallback to env vars
		log.Println("Local development detected - AWS Secrets Manager disabled, using environment variables")
		return nil, fmt.Errorf("local development mode: AWS Secrets Manager disabled")
	}

	// Additional check: if no AWS region is set and not explicitly enabled, assume local
	if awsRegion == "" && useAwsSecrets != "true" {
		log.Println("No AWS region detected - AWS Secrets Manager disabled, using environment variables")
		return nil, fmt.Errorf("no AWS region detected: AWS Secrets Manager disabled")
	}

	log.Printf("AWS execution environment detected (%s) - initializing AWS Secrets Manager", awsExecutionEnv)

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return &SecretsClient{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}

// GetIndividualSecret retrieves a single secret value by ARN or name
// This is used when secrets are stored individually rather than in a JSON object
func (s *SecretsClient) GetIndividualSecret(secretName string) (string, error) {
	log.Printf("Attempting to retrieve individual secret from AWS Secrets Manager: %s", secretName)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := s.client.GetSecretValue(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret %s: %w", secretName, err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret %s has no string value", secretName)
	}

	log.Printf("Successfully retrieved individual secret from AWS Secrets Manager: %s", secretName)
	return *result.SecretString, nil
}