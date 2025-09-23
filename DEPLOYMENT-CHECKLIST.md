# Apollo AWS Deployment Checklist

Use this checklist to ensure all steps are completed for your AWS deployment.

## Pre-Deployment Setup

- [ ] AWS CLI installed and configured with appropriate permissions
- [ ] Docker installed and running
- [ ] Existing PostgreSQL database in AWS RDS accessible
- [ ] ODDS_API_KEY available
- [ ] JWT_SECRET generated for production

## AWS Secrets Manager

- [ ] Create secret `apollo/api-keys` in AWS Secrets Manager
- [ ] Add ODDS_API_KEY to the secret
- [ ] Verify secret is accessible in your target AWS region

## Docker Image Build & Push

- [ ] Run `./deploy-apprunner.sh` script
- [ ] Verify ECR repository created
- [ ] Verify Docker image pushed successfully
- [ ] Note down the ECR image URI

## App Runner Deployment

Choose one approach:

### Using CloudFormation (Recommended)
- [ ] Update parameters in the CloudFormation command
- [ ] Replace placeholder values:
  - [ ] ECR Image URI
  - [ ] Database URL with your actual RDS endpoint
  - [ ] JWT Secret
  - [ ] AWS region
- [ ] Run CloudFormation deploy command
- [ ] Verify stack creation successful

### Using AWS Console
- [ ] Navigate to App Runner console
- [ ] Create new service from container registry
- [ ] Configure image URI from ECR
- [ ] Set environment variables:
  - [ ] `PORT=8080`
  - [ ] `DATABASE_URL` (your RDS connection string)
  - [ ] `AWS_SECRET_NAME=apollo/api-keys`
  - [ ] `JWT_SECRET` (your production secret)
- [ ] Configure instance: 0.25 vCPU, 0.5 GB memory
- [ ] Deploy service

## Lambda Function Deployment

- [ ] Navigate to `lambda/game-finalizer/` directory
- [ ] Run `make build package`
- [ ] Update CloudFormation parameters:
  - [ ] Database URL
  - [ ] VPC ID (same as your RDS)
  - [ ] Subnet IDs (private subnets where Lambda should run)
  - [ ] Secret name
- [ ] Deploy Lambda stack with CloudFormation
- [ ] Upload function code using AWS CLI
- [ ] Verify EventBridge rule created for schedule

## Testing & Verification

- [ ] App Runner service shows "Running" status
- [ ] Visit App Runner URL - should show GraphQL playground
- [ ] Test health endpoint: `/health` returns "OK"
- [ ] Test GraphQL endpoint with a simple query
- [ ] Test Lambda function with manual invoke
- [ ] Check CloudWatch logs for both services

## Security & Configuration

- [ ] Update CORS settings for your production frontend domain
- [ ] Verify all secrets are stored in AWS Secrets Manager (not environment variables)
- [ ] Confirm RDS security groups allow access from App Runner and Lambda
- [ ] Ensure Lambda has proper VPC configuration if RDS is in VPC

## Monitoring (Optional but Recommended)

- [ ] Set up CloudWatch alarms for service health
- [ ] Configure log retention policies
- [ ] Set up notification for Lambda failures

## Production Readiness

- [ ] Remove or secure GraphQL playground for production
- [ ] Set up custom domain (optional)
- [ ] Configure SSL certificate (App Runner handles this automatically)
- [ ] Update DNS records if using custom domain

## Final Verification

- [ ] Game finalization schedule verified (Fri/Sun/Mon/Tue at 6:00 AM UTC)
- [ ] All services communicating with database successfully
- [ ] Secrets being retrieved from AWS Secrets Manager
- [ ] No plain-text secrets in environment variables
- [ ] Application logs showing in CloudWatch

## Post-Deployment

- [ ] Document your specific configuration (RDS endpoint, regions, etc.)
- [ ] Share App Runner service URL with your team
- [ ] Set up monitoring and alerting as needed
- [ ] Plan for regular security updates and patching

---

**Important Notes:**
- Save your ECR image URI - you'll need it for updates
- Keep track of your CloudFormation stack names
- Note your App Runner service URL
- Document any customizations you made

**For Updates:**
- Rebuild and push Docker image to ECR
- App Runner will automatically deploy new images if auto-deploy is enabled
- For Lambda updates, rebuild and re-upload the function code

**Emergency Rollback:**
- Use previous ECR image tag for App Runner rollback
- Lambda versions are automatically managed by AWS