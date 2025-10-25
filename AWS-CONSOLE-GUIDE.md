# AWS Console Setup Guide for Apollo Application

This guide provides step-by-step instructions for setting up your Apollo application in the AWS Console manually, without using CLI commands or CloudFormation templates.

## 📋 Prerequisites

- AWS Account with appropriate permissions
- Your Apollo application code ready
- Docker installed locally
- Your ODDS_API_KEY
- A secure JWT secret for production

---

## 🗄️ Part 1: Database Setup (RDS PostgreSQL)

### Step 1: Create RDS PostgreSQL Instance

1. **Navigate to RDS**:
   - Go to [AWS RDS Console](https://console.aws.amazon.com/rds/)
   - Click **"Create database"**

2. **Choose Database Creation Method**:
   - Select **"Standard create"**

3. **Engine Options**:
   - Engine type: **PostgreSQL**
   - Engine version: **PostgreSQL 15.4** (or latest stable)

4. **Templates**:
   - Choose **"Free tier"** (if eligible) or **"Production"** based on your needs

5. **Settings**:
   - DB instance identifier: `apollo-db`
   - Master username: `postgres` (or your preferred username)
   - Master password: Create a strong password and **save it securely**
   - Confirm password

6. **DB Instance Class**:
   - For development/small production: `db.t3.micro` or `db.t4g.micro`
   - For larger production: `db.t3.small` or higher

7. **Storage**:
   - Storage type: **General Purpose SSD (gp3)**
   - Allocated storage: **20 GB** (minimum, can be increased later)
   - Enable storage autoscaling if desired

8. **Availability & Durability**:
   - For cost savings: **Don't create a standby instance**
   - For production: **Create a standby instance in a different AZ**

9. **Connectivity**:
   - Virtual Private Cloud (VPC): Use **default VPC** or create a custom one
   - Subnet group: **default**
   - Public access: **Yes** (for easier setup; can be changed later for security)
   - VPC security group: **Create new**
   - New VPC security group name: `apollo-db-sg`

10. **Database Authentication**:
    - Database authentication options: **Password authentication**

11. **Additional Configuration**:
    - Initial database name: `apollo`
    - DB parameter group: **default**
    - Backup retention period: **7 days**
    - Monitoring: Enable **Enhanced monitoring** if desired

12. **Click "Create database"**

### Step 2: Configure Security Group for Database

1. **Wait for database creation** (takes 5-10 minutes)

2. **Go to EC2 Console** → **Security Groups**

3. **Find your database security group** (`apollo-db-sg`)

4. **Edit inbound rules**:
   - Click **"Edit inbound rules"**
   - Click **"Add rule"**
   - Type: **PostgreSQL**
   - Protocol: **TCP**
   - Port: **5432**
   - Source: **Anywhere IPv4 (0.0.0.0/0)** (temporarily; we'll restrict this later)
   - Click **"Save rules"**

### Step 3: Get Database Connection String

1. **Go back to RDS Console**
2. **Click on your database** (`apollo-db`)
3. **In the "Connectivity & security" tab**:
   - Copy the **Endpoint** (e.g., `apollo-db.abc123.us-east-1.rds.amazonaws.com`)
   - Note the **Port** (should be `5432`)

4. **Create your connection string**:
   ```
   postgres://postgres:YOUR_PASSWORD@apollo-db.abc123.us-east-1.rds.amazonaws.com:5432/apollo?sslmode=require
   ```
   Replace `YOUR_PASSWORD` with the password you set earlier.

---

## 🔐 Part 2: Secrets Manager Setup

### Step 1: Create Secret for API Keys

1. **Navigate to Secrets Manager**:
   - Go to [AWS Secrets Manager Console](https://console.aws.amazon.com/secretsmanager/)

2. **Create Secret**:
   - Click **"Store a new secret"**

3. **Secret Type**:
   - Choose **"Other type of secret"**

4. **Key/Value Pairs**:
   - Click **"Plaintext"** tab
   - Enter the following JSON:
   ```json
   {
     "ODDS_API_KEY": "your-actual-odds-api-key-here"
   }
   ```
   Replace with your actual ODDS API key

5. **Encryption**:
   - Use **default encryption key** or choose a custom KMS key

6. **Secret Name and Description**:
   - Secret name: `apollo/api-keys`
   - Description: `Apollo application API keys`

7. **Configure Rotation** (optional):
   - Skip automatic rotation for now
   - Click **"Next"**

8. **Review and Store**:
   - Review your settings
   - Click **"Store secret"**

---

## 🐳 Part 3: Container Registry Setup (ECR)

### Step 1: Create ECR Repository

1. **Navigate to ECR**:
   - Go to [Amazon ECR Console](https://console.aws.amazon.com/ecr/)

2. **Create Repository**:
   - Click **"Create repository"**

3. **Repository Settings**:
   - Visibility settings: **Private**
   - Repository name: `apollo-app`

4. **Image Scanning**:
   - Enable **scan on push** (recommended)

5. **Create Repository**:
   - Click **"Create repository"**

### Step 2: Push Your Docker Image

1. **Click on your repository** (`apollo-app`)

2. **Click "View push commands"**

3. **Follow the commands shown**:
   ```bash
   # 1. Get login token (replace REGION and ACCOUNT_ID)
   aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

   # 2. Build your image (run from your Apollo project root)
   docker build -t apollo-app .

   # 3. Tag your image
   docker tag apollo-app:latest ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apollo-app:latest

   # 4. Push to ECR
   docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apollo-app:latest
   ```

4. **Save your ECR URI**:
   - Copy the full URI: `ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apollo-app:latest`
   - You'll need this for App Runner

---

## 🚀 Part 4: App Runner Setup

### Step 1: Create App Runner Service

1. **Navigate to App Runner**:
   - Go to [AWS App Runner Console](https://console.aws.amazon.com/apprunner/)

2. **Create Service**:
   - Click **"Create service"**

3. **Source**:
   - Repository type: **Container registry**
   - Provider: **Amazon ECR**
   - Container image URI: Paste your ECR URI from Step 3
   - Deployment trigger: **Manual** (you can change this later)

4. **Deployment Settings**:
   - ECR access role: **Create new role** (App Runner will create `AppRunnerECRAccessRole`)

5. **Service Settings**:
   - Service name: `apollo-app`
   - Virtual CPU: **0.25 vCPU**
   - Virtual memory: **0.5 GB**

6. **Environment Variables**:
   Click **"Add environment variable"** for each:
   - `PORT` = `8080`
   - `DATABASE_URL` = Your RDS connection string from Part 1
   - `AWS_SECRET_NAME` = `apollo/api-keys`
   - `JWT_SECRET` = Your secure JWT secret

7. **Auto Scaling**:
   - Min size: **1**
   - Max size: **3** (adjust based on your needs)

8. **Health Check**:
   - Health check protocol: **HTTP**
   - Health check path: `/health`

9. **Security**:
   - Instance role: **Create new role** (App Runner will create it)

### Step 2: Configure Service Role Permissions

1. **Wait for service creation** (takes 3-5 minutes)

2. **Go to IAM Console** → **Roles**

3. **Find the App Runner instance role** (usually named like `AppRunnerInstanceRole...`)

4. **Attach Policy for Secrets Manager**:
   - Click **"Add permissions"** → **"Attach policies"**
   - Search for and attach: `SecretsManagerReadWrite` (or create a custom policy)

5. **Alternative: Create Custom Policy**:
   - Click **"Add permissions"** → **"Create inline policy"**
   - Use the policy editor to add:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "secretsmanager:GetSecretValue"
         ],
         "Resource": "arn:aws:secretsmanager:us-east-1:YOUR_ACCOUNT_ID:secret:apollo/api-keys*"
       }
     ]
   }
   ```

### Step 3: Test Your App Runner Service

1. **Get your service URL**:
   - In App Runner console, click on your service
   - Copy the **Default domain** URL (e.g., `https://abc123.us-east-1.awsapprunner.com`)

2. **Test endpoints**:
   - Health check: `https://your-url.com/health` (should return "OK")
   - GraphQL playground: `https://your-url.com/` (should show GraphQL playground)

---

## ⚡ Part 5: Lambda Function Setup

### Step 1: Create Lambda Function

1. **Navigate to Lambda**:
   - Go to [AWS Lambda Console](https://console.aws.amazon.com/lambda/)

2. **Create Function**:
   - Click **"Create function"**
   - Choose **"Author from scratch"**

3. **Basic Information**:
   - Function name: `apollo-game-finalizer`
   - Runtime: **Go 1.x** or **Amazon Linux 2023**
   - Architecture: **x86_64**

4. **Permissions**:
   - Execution role: **Create a new role with basic Lambda permissions**

### Step 2: Upload Your Lambda Code

1. **Build your Lambda function locally**:
   ```bash
   cd lambda/game-finalizer
   GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
   zip lambda-function.zip bootstrap
   ```

2. **Upload the zip file**:
   - In Lambda console, click **"Upload from"** → **".zip file"**
   - Select your `lambda-function.zip`
   - Click **"Save"**

### Step 3: Configure Lambda Settings

1. **Environment Variables**:
   - Click **"Configuration"** → **"Environment variables"**
   - Add:
     - `DATABASE_URL` = Your RDS connection string
     - `AWS_SECRET_NAME` = `apollo/api-keys`

2. **Runtime Settings**:
   - Click **"Configuration"** → **"Runtime settings"**
   - Handler: `bootstrap` (for Go)

3. **Timeout and Memory**:
   - Click **"Configuration"** → **"General configuration"**
   - Timeout: **15 minutes**
   - Memory: **256 MB**

### Step 4: Configure VPC (if database is in VPC)

1. **VPC Configuration**:
   - Click **"Configuration"** → **"VPC"**
   - VPC: Select the same VPC as your RDS database
   - Subnets: Select private subnets (at least 2)
   - Security groups: Create or select one that allows database access

2. **Create Lambda Security Group**:
   - Go to **EC2** → **Security Groups**
   - Create new security group: `apollo-lambda-sg`
   - Add outbound rule for PostgreSQL (port 5432) to database security group

3. **Update Database Security Group**:
   - Edit your `apollo-db-sg` security group
   - Add inbound rule:
     - Type: PostgreSQL
     - Source: `apollo-lambda-sg` (the Lambda security group)

### Step 5: Configure Lambda Permissions

1. **Go to IAM** → **Roles**

2. **Find your Lambda execution role** (e.g., `apollo-game-finalizer-role-...`)

3. **Add Secrets Manager permissions**:
   - Click **"Add permissions"** → **"Attach policies"**
   - Attach `SecretsManagerReadWrite` or create custom policy:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "secretsmanager:GetSecretValue"
         ],
         "Resource": "arn:aws:secretsmanager:us-east-1:YOUR_ACCOUNT_ID:secret:apollo/api-keys*"
       }
     ]
   }
   ```

4. **Add VPC permissions** (if using VPC):
   - Attach `AWSLambdaVPCAccessExecutionRole`

### Step 6: Test Lambda Function

1. **Create test event**:
   - Click **"Test"** → **"Create new event"**
   - Event name: `test-event`
   - Template: `Hello World`
   - Click **"Save"**

2. **Run test**:
   - Click **"Test"**
   - Check the execution results and logs

---

## ⏰ Part 6: EventBridge Schedule Setup

### Step 1: Create EventBridge Rule

1. **Navigate to EventBridge**:
   - Go to [Amazon EventBridge Console](https://console.aws.amazon.com/events/)

2. **Create Rule**:
   - Click **"Create rule"**

3. **Rule Details**:
   - Name: `apollo-game-finalizer-schedule`
   - Description: `Runs Apollo game finalizer on Fri/Sun/Mon/Tue at 6:00 AM UTC`

4. **Rule Type**:
   - Select **"Schedule"**

5. **Schedule Pattern**:
   - Choose **"Cron expression"**
   - Cron expression: `0 6 ? * FRI,SUN,MON,TUE *`
   - This runs at 6:00 AM UTC on Friday, Sunday, Monday, and Tuesday

6. **Target**:
   - Target type: **AWS service**
   - Service: **Lambda function**
   - Function: `apollo-game-finalizer`

7. **Create Rule**:
   - Review settings and click **"Create rule"**

### Step 2: Grant EventBridge Permission to Invoke Lambda

1. **Go back to Lambda Console**

2. **Click on your function** (`apollo-game-finalizer`)

3. **Add Trigger**:
   - Click **"Add trigger"**
   - Source: **EventBridge**
   - Rule: Select `apollo-game-finalizer-schedule`
   - Click **"Add"**

---

## 📊 Part 7: Monitoring and Logging Setup

### Step 1: CloudWatch Logs

**App Runner Logs**:
- Automatically created at: `/aws/apprunner/apollo-app/application`

**Lambda Logs**:
- Automatically created at: `/aws/lambda/apollo-game-finalizer`

### Step 2: CloudWatch Alarms (Optional)

1. **Navigate to CloudWatch** → **Alarms**

2. **Create Alarm for App Runner Health**:
   - Metric: App Runner → Service Name → `apollo-app` → `2XXStatusCodeCount`
   - Statistic: Sum
   - Period: 5 minutes
   - Threshold: Less than 1 request per 5 minutes
   - Actions: Send SNS notification

3. **Create Alarm for Lambda Errors**:
   - Metric: Lambda → Function Name → `apollo-game-finalizer` → `Errors`
   - Statistic: Sum
   - Period: 1 minute
   - Threshold: Greater than 0 errors
   - Actions: Send SNS notification

---

## 🔧 Part 8: Final Configuration

### Step 1: Update CORS Settings

Since your frontend will be deployed to a different domain, you need to update the CORS settings:

1. **Modify your Apollo server code** (locally in `server/main.go`)
2. **Update the CORS configuration**:
   ```go
   c := cors.New(cors.Options{
       AllowedOrigins: []string{
           "https://your-frontend-domain.com",  // Add your frontend domain
           "http://localhost:3000",  // Keep for local development
       },
       AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
       AllowedHeaders:   []string{"Content-Type", "Authorization"},
       AllowCredentials: true,
   })
   ```

3. **Rebuild and push to ECR**:
   ```bash
   docker build -t apollo-app .
   docker tag apollo-app:latest ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apollo-app:latest
   docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/apollo-app:latest
   ```

4. **Redeploy App Runner**:
   - Go to App Runner console
   - Click on your service
   - Click **"Deploy"** to pull the latest image

### Step 2: Security Improvements

1. **Restrict Database Access**:
   - Edit your RDS security group (`apollo-db-sg`)
   - Remove the "Anywhere" rule
   - Add specific rules for:
     - App Runner (you may need to get IP ranges)
     - Lambda security group (if using VPC)

2. **Create Custom Domain** (Optional):
   - In App Runner console, go to your service
   - Click **"Custom domains"**
   - Add your domain and follow verification steps

---

## ✅ Part 9: Testing Your Complete Setup

### Step 1: Test Database Connection

Use any PostgreSQL client to connect to your RDS instance and verify it's accessible.

### Step 2: Test App Runner Service

1. **Health Check**: `https://your-app-runner-url.com/health`
2. **GraphQL Playground**: `https://your-app-runner-url.com/`
3. **Test a GraphQL query**

### Step 3: Test Lambda Function

1. **Manual test**: Go to Lambda console and test the function
2. **Check logs**: Verify it can connect to database and retrieve secrets
3. **Wait for scheduled run**: Check if it runs according to schedule

### Step 4: Test End-to-End

1. Deploy your React frontend with the App Runner URL
2. Test authentication and GraphQL queries
3. Verify the cron job runs on schedule

---

## 🚨 Troubleshooting

### Common Issues:

1. **App Runner service unhealthy**:
   - Check CloudWatch logs
   - Verify environment variables
   - Test database connectivity

2. **Lambda timeout or errors**:
   - Check execution time
   - Verify VPC configuration
   - Check database security groups

3. **Database connection failures**:
   - Verify connection string format
   - Check security group rules
   - Ensure SSL mode is correct

4. **Secrets Manager access denied**:
   - Verify IAM roles have correct permissions
   - Check secret name matches configuration

---

## 📝 Summary

After completing this guide, you'll have:

- ✅ PostgreSQL database in RDS
- ✅ API keys stored in Secrets Manager
- ✅ Docker image in ECR
- ✅ Apollo app running on App Runner
- ✅ Lambda function for game finalization
- ✅ Scheduled execution via EventBridge
- ✅ CloudWatch logging and monitoring

Your application is now fully deployed and ready for production use!