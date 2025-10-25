#!/bin/bash

# Apollo AWS App Runner Deployment Script

set -e

# Configuration
AWS_REGION=${AWS_REGION:-us-east-1}
ECR_REPOSITORY_NAME="apollo-app"
APP_RUNNER_SERVICE_NAME="apollo-app"
SECRET_NAME="apollo/api-keys"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Apollo App Runner deployment...${NC}"

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}AWS CLI is not installed. Please install it first.${NC}"
    exit 1
fi

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Get AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_URI="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY_NAME}"

echo -e "${YELLOW}AWS Account ID: ${ACCOUNT_ID}${NC}"
echo -e "${YELLOW}ECR Repository: ${ECR_URI}${NC}"

# Create ECR repository if it doesn't exist
echo -e "${GREEN}Creating ECR repository if it doesn't exist...${NC}"
aws ecr describe-repositories --repository-names $ECR_REPOSITORY_NAME --region $AWS_REGION >/dev/null 2>&1 || \
    aws ecr create-repository --repository-name $ECR_REPOSITORY_NAME --region $AWS_REGION

# Get ECR login token
echo -e "${GREEN}Logging in to ECR...${NC}"
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_URI

# Build Docker image
echo -e "${GREEN}Building Docker image...${NC}"
docker build -t apollo-app .

# Tag image for ECR
echo -e "${GREEN}Tagging image for ECR...${NC}"
docker tag apollo-app:latest $ECR_URI:latest

# Push image to ECR
echo -e "${GREEN}Pushing image to ECR...${NC}"
docker push $ECR_URI:latest

echo -e "${GREEN}Docker image pushed successfully!${NC}"
echo -e "${YELLOW}ECR Image URI: ${ECR_URI}:latest${NC}"

echo -e "${GREEN}Deployment complete!${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Update your secrets in AWS Secrets Manager with name: $SECRET_NAME"
echo "2. Use the CloudFormation template in aws/apprunner-cloudformation.yaml to create the App Runner service"
echo "3. Set the ImageUri parameter to: $ECR_URI:latest"
echo ""
echo -e "${GREEN}To deploy using CloudFormation:${NC}"
echo "aws cloudformation deploy --template-file aws/apprunner-cloudformation.yaml \\"
echo "  --stack-name apollo-apprunner \\"
echo "  --parameter-overrides \\"
echo "    ImageUri=$ECR_URI:latest \\"
echo "    DatabaseURL='your-database-url' \\"
echo "    JWTSecret='your-jwt-secret' \\"
echo "  --capabilities CAPABILITY_NAMED_IAM \\"
echo "  --region $AWS_REGION"