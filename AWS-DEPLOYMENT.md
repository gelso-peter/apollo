# Apollo AWS Deployment Guide

This guide will walk you through deploying the Apollo application to AWS using App Runner for the main service and Lambda for the game finalization cron job.

## Architecture Overview

- **AWS App Runner**: Hosts the main Go application (GraphQL API and REST endpoints)
- **AWS Lambda**: Handles the game finalization cron job (runs Fri/Sun/Mon/Tue at 6:00 AM UTC)
- **AWS Secrets Manager**: Stores API keys securely
- **Amazon RDS PostgreSQL**: Database (you mentioned you already created this)
- **Amazon ECR**: Container registry for Docker images
- **Amazon EventBridge**: Triggers Lambda function on schedule

## Prerequisites

1. AWS CLI installed and configured
2. Docker installed
3. Your existing Apollo PostgreSQL database in AWS
4. Your ODDS_API_KEY

## Step 1: Set Up AWS Secrets Manager

Create a secret in AWS Secrets Manager to store your API keys:

```bash
aws secretsmanager create-secret \
    --name "apollo/api-keys" \
    --description "Apollo application API keys" \
    --secret-string '{"ODDS_API_KEY":"your-actual-odds-api-key-here"}' \
    --region us-east-1
```

## Step 2: Build and Push Docker Image to ECR

Use the provided deployment script:

```bash
# Make the script executable
chmod +x deploy-apprunner.sh

# Set your AWS region if different
export AWS_REGION=us-east-1

# Run the deployment script
./deploy-apprunner.sh
```

This script will:
- Create an ECR repository
- Build your Docker image
- Push it to ECR
- Provide the ECR URI for the next step

## Step 3: Deploy App Runner Service

### Option A: Using CloudFormation (Recommended)

1. Update the parameters in the CloudFormation template:

```bash
aws cloudformation deploy \
    --template-file aws/apprunner-cloudformation.yaml \
    --stack-name apollo-apprunner \
    --parameter-overrides \
        ImageUri="ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apollo-app:latest" \
        DatabaseURL="postgres://username:password@your-rds-endpoint.region.rds.amazonaws.com:5432/apollo?sslmode=require" \
        JWTSecret="your-production-jwt-secret" \
        SecretName="apollo/api-keys" \
    --capabilities CAPABILITY_NAMED_IAM \
    --region us-east-1
```

Replace:
- `ACCOUNT_ID` with your AWS account ID
- `username:password@your-rds-endpoint.region.rds.amazonaws.com` with your actual RDS connection details
- `your-production-jwt-secret` with a secure JWT secret

### Option B: Using AWS Console

1. Go to AWS App Runner console
2. Create new service
3. Choose "Container registry" as source
4. Select your ECR image
5. Configure environment variables:
   - `PORT`: `8080`
   - `DATABASE_URL`: Your RDS connection string
   - `AWS_SECRET_NAME`: `apollo/api-keys`
   - `JWT_SECRET`: Your JWT secret
6. Set instance configuration:
   - CPU: 0.25 vCPU
   - Memory: 0.5 GB
7. Configure auto-scaling (can keep defaults for 20 users)

## Step 4: Deploy Lambda Function for Game Finalization

### Build Lambda Function

```bash
cd lambda/game-finalizer
make build package
```

### Deploy Lambda using CloudFormation

```bash
aws cloudformation deploy \
    --template-file cloudformation.yaml \
    --stack-name apollo-lambda-finalizer \
    --parameter-overrides \
        DatabaseURL="postgres://username:password@your-rds-endpoint.region.rds.amazonaws.com:5432/apollo?sslmode=require" \
        SecretName="apollo/api-keys" \
        VpcId="your-vpc-id" \
        SubnetIds="subnet-123,subnet-456" \
    --capabilities CAPABILITY_NAMED_IAM \
    --region us-east-1
```

Then upload the function code:

```bash
aws lambda update-function-code \
    --function-name apollo-game-finalizer \
    --zip-file fileb://lambda-function.zip \
    --region us-east-1
```

## Step 5: Configure CORS for Production

Update the CORS settings in your App Runner deployment to include your frontend domain:

In `server/main.go`, the CORS configuration currently allows `http://localhost:3000`. You'll need to update this for production:

```go
c := cors.New(cors.Options{
    AllowedOrigins:   []string{"https://your-frontend-domain.com", "http://localhost:3000"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
})
```

## Step 6: Verify Deployment

1. **Test App Runner Service**:
   - Visit the App Runner service URL
   - Check `/health` endpoint returns "OK"
   - Test GraphQL playground at the root URL

2. **Test Lambda Function**:
   ```bash
   aws lambda invoke \
       --function-name apollo-game-finalizer \
       --region us-east-1 \
       output.json
   ```

3. **Monitor Logs**:
   - App Runner logs are in CloudWatch under `/aws/apprunner/apollo-app/application`
   - Lambda logs are in CloudWatch under `/aws/lambda/apollo-game-finalizer`

## Step 7: Set Up Custom Domain (Optional)

1. In App Runner console, go to your service
2. Click "Custom domains"
3. Add your domain and follow verification steps
4. Update your DNS to point to the App Runner service

## Environment Variables Reference

### App Runner Environment Variables
- `PORT`: `8080`
- `DATABASE_URL`: Your RDS PostgreSQL connection string
- `AWS_SECRET_NAME`: `apollo/api-keys` (or your chosen secret name)
- `JWT_SECRET`: Your JWT signing secret

### Lambda Environment Variables
- `DATABASE_URL`: Same as App Runner
- `AWS_SECRET_NAME`: Same as App Runner

## Security Considerations

1. **Database Security**:
   - Ensure RDS is in private subnets
   - Use security groups to limit database access
   - Use SSL connections (`sslmode=require`)

2. **Secrets Management**:
   - Never put secrets in environment variables in plain text
   - Use AWS Secrets Manager for sensitive data
   - Rotate secrets periodically

3. **Network Security**:
   - Lambda function should be in VPC if accessing RDS in VPC
   - Use private subnets for Lambda
   - Configure security groups appropriately

## Monitoring and Troubleshooting

### CloudWatch Alarms (Recommended)
Set up alarms for:
- App Runner service health
- Lambda function errors
- Database connections
- High latency

### Common Issues

1. **Lambda can't connect to database**:
   - Check VPC configuration
   - Verify security group rules
   - Ensure Lambda is in same VPC as RDS

2. **App Runner service unhealthy**:
   - Check CloudWatch logs
   - Verify environment variables
   - Test database connectivity

3. **Secrets Manager access denied**:
   - Check IAM roles have correct permissions
   - Verify secret name matches configuration

## Costs Estimation (for 20 users)

- **App Runner**: ~$25-50/month (0.25 vCPU, 0.5GB RAM with minimal traffic)
- **Lambda**: ~$1-5/month (runs 4 times per week, minimal execution time)
- **ECR**: ~$1/month (storage for Docker images)
- **Secrets Manager**: ~$0.40/month (1 secret)
- **RDS**: Cost depends on instance size (you already have this)

Total additional cost: ~$27-56/month

## Next Steps

1. Set up monitoring and alerting
2. Configure log aggregation
3. Set up backup strategy for RDS
4. Consider implementing CI/CD pipeline for automated deployments
5. Set up staging environment

## Support

If you encounter issues:
1. Check CloudWatch logs first
2. Verify all environment variables are set correctly
3. Test database connectivity from both App Runner and Lambda
4. Ensure IAM permissions are correctly configured