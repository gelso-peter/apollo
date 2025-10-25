# Apollo React Frontend - AWS Production Setup

This README provides instructions for connecting your React frontend application to the Apollo GraphQL API deployed on AWS App Runner.

## 🚀 Quick Start

### Environment Variables

Create environment files for different environments:

#### `.env.local` (Local Development)
```bash
# Local development - connects to local server
REACT_APP_API_URL=http://localhost:8080
REACT_APP_GRAPHQL_URL=http://localhost:8080/query
REACT_APP_ENVIRONMENT=development
```

#### `.env.production` (AWS Production)
```bash
# Production - connects to AWS App Runner
REACT_APP_API_URL=https://your-app-runner-url.us-east-1.awsapprunner.com
REACT_APP_GRAPHQL_URL=https://your-app-runner-url.us-east-1.awsapprunner.com/query
REACT_APP_ENVIRONMENT=production
```

> **⚠️ Important**: Replace `your-app-runner-url.us-east-1.awsapprunner.com` with your actual App Runner service URL.

## 🔧 GraphQL Client Setup

### Using Apollo Client

Install Apollo Client:
```bash
npm install @apollo/client graphql
```

#### Configuration (`src/apollo/client.js`)
```javascript
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

// HTTP link for GraphQL endpoint
const httpLink = createHttpLink({
  uri: process.env.REACT_APP_GRAPHQL_URL,
  credentials: 'include', // Important for CORS cookies
});

// Auth link for JWT tokens
const authLink = setContext((_, { headers }) => {
  // Get JWT token from localStorage or your auth system
  const token = localStorage.getItem('apollo_jwt_token');

  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
      'Content-Type': 'application/json',
    }
  };
});

// Create Apollo Client
const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      errorPolicy: 'all',
    },
    query: {
      errorPolicy: 'all',
    },
  },
});

export default client;
```

#### App Setup (`src/App.js`)
```javascript
import React from 'react';
import { ApolloProvider } from '@apollo/client';
import client from './apollo/client';
import YourMainComponent from './components/YourMainComponent';

function App() {
  return (
    <ApolloProvider client={client}>
      <div className="App">
        <YourMainComponent />
      </div>
    </ApolloProvider>
  );
}

export default App;
```

### Using other GraphQL clients

#### With `graphql-request`
```bash
npm install graphql-request graphql
```

```javascript
import { GraphQLClient } from 'graphql-request';

const client = new GraphQLClient(process.env.REACT_APP_GRAPHQL_URL, {
  credentials: 'include',
  headers: {
    authorization: `Bearer ${localStorage.getItem('apollo_jwt_token')}`,
  },
});

export default client;
```

## 🔐 Authentication Setup

### JWT Token Management

#### Login Function
```javascript
const login = async (credentials) => {
  try {
    const response = await fetch(`${process.env.REACT_APP_API_URL}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(credentials),
    });

    const data = await response.json();

    if (data.token) {
      localStorage.setItem('apollo_jwt_token', data.token);
      // Refresh Apollo Client cache or redirect
      window.location.reload();
    }

    return data;
  } catch (error) {
    console.error('Login failed:', error);
    throw error;
  }
};
```

#### Logout Function
```javascript
const logout = () => {
  localStorage.removeItem('apollo_jwt_token');
  // Clear Apollo Client cache
  client.clearStore();
  // Redirect to login page
  window.location.href = '/login';
};
```

## 📡 API Endpoints Reference

### GraphQL Endpoint
- **URL**: `https://your-app-runner-url.us-east-1.awsapprunner.com/query`
- **Method**: POST
- **Headers**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <your-jwt-token>` (for protected queries)
- **CORS**: Enabled with credentials support

### REST API Endpoints
Base URL: `https://your-app-runner-url.us-east-1.awsapprunner.com/api`

Common endpoints (update based on your actual API):
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `GET /api/health` - Health check
- `GET /api/user/profile` - User profile (protected)

### Health Check Endpoint
- **URL**: `https://your-app-runner-url.us-east-1.awsapprunner.com/health`
- **Method**: GET
- **Response**: `OK`

## 🌐 CORS Configuration

### For Development
Local development (localhost:3000) is already whitelisted in the Apollo server.

### For Production
You need to update the Apollo server's CORS configuration to include your production domain:

1. In your Apollo server code (`server/main.go`), update the CORS origins:
```go
c := cors.New(cors.Options{
    AllowedOrigins: []string{
        "https://your-frontend-domain.com",
        "http://localhost:3000", // Keep for local dev
    },
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
})
```

2. Redeploy your Apollo server to App Runner

## 🏗️ Build Configuration

### Build Scripts (`package.json`)
```json
{
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build",
    "build:staging": "env-cmd -f .env.staging react-scripts build",
    "build:production": "env-cmd -f .env.production react-scripts build"
  }
}
```

Install env-cmd for environment-specific builds:
```bash
npm install --save-dev env-cmd
```

### Production Build
```bash
# Build for production
npm run build:production

# The build folder will contain your production-ready files
```

## 🔍 Testing Connection

### Test GraphQL Connection
```javascript
// Simple test query to verify connection
import { gql } from '@apollo/client';

const TEST_QUERY = gql`
  query TestConnection {
    __typename
  }
`;

const TestConnection = () => {
  const { data, loading, error } = useQuery(TEST_QUERY);

  if (loading) return <p>Testing connection...</p>;
  if (error) return <p>Connection failed: {error.message}</p>;

  return <p>✅ Connected successfully to Apollo GraphQL!</p>;
};
```

### Test REST API Connection
```javascript
const testRestConnection = async () => {
  try {
    const response = await fetch(`${process.env.REACT_APP_API_URL}/health`);
    const result = await response.text();
    console.log('Health check:', result); // Should log "OK"
  } catch (error) {
    console.error('REST API connection failed:', error);
  }
};
```

## 🚨 Troubleshooting

### Common Issues

#### 1. CORS Errors
**Problem**: Browser blocks requests due to CORS policy
**Solutions**:
- Ensure your domain is added to Apollo server's CORS allowed origins
- Check that credentials are properly configured
- Verify the App Runner URL is correct

#### 2. Authentication Failures
**Problem**: 401 Unauthorized errors
**Solutions**:
- Check JWT token is being sent in Authorization header
- Verify token hasn't expired
- Ensure token is properly formatted with "Bearer " prefix

#### 3. GraphQL Endpoint Not Found
**Problem**: 404 errors on GraphQL queries
**Solutions**:
- Verify GraphQL endpoint URL ends with `/query`
- Check App Runner service is running and healthy
- Test the endpoint directly in browser or Postman

#### 4. Network Errors
**Problem**: Connection timeouts or network errors
**Solutions**:
- Verify App Runner service URL is correct
- Check AWS App Runner service status in console
- Test health endpoint: `https://your-url.com/health`

### Debug Commands

#### Check Environment Variables
```javascript
console.log('API URL:', process.env.REACT_APP_API_URL);
console.log('GraphQL URL:', process.env.REACT_APP_GRAPHQL_URL);
console.log('Environment:', process.env.REACT_APP_ENVIRONMENT);
```

#### Test API Endpoints
```bash
# Test health endpoint
curl https://your-app-runner-url.us-east-1.awsapprunner.com/health

# Test GraphQL endpoint
curl -X POST https://your-app-runner-url.us-east-1.awsapprunner.com/query \
  -H "Content-Type: application/json" \
  -d '{"query": "{ __typename }"}'
```

## 📦 Deployment

### Frontend Deployment Options

#### 1. AWS Amplify (Recommended)
- Automatic builds from Git repository
- Built-in SSL certificates
- Global CDN distribution
- Easy custom domain setup

#### 2. AWS S3 + CloudFront
- Static site hosting
- Global CDN
- Custom domain support
- More manual setup but cheaper

#### 3. Netlify/Vercel
- Easy deployment from Git
- Automatic builds
- Built-in SSL
- Good for quick deployments

### Build and Deploy Steps
```bash
# 1. Build the production version
npm run build:production

# 2. Deploy the build folder to your hosting service
# (specific steps depend on your chosen deployment platform)
```

## 🔗 Useful Links

- **GraphQL Playground**: `https://your-app-runner-url.us-east-1.awsapprunner.com/` (in development mode)
- **Health Check**: `https://your-app-runner-url.us-east-1.awsapprunner.com/health`
- **App Runner Console**: [AWS App Runner Console](https://console.aws.amazon.com/apprunner)

## 💡 Best Practices

1. **Environment Variables**: Never commit `.env` files with production URLs to Git
2. **Error Handling**: Always handle network errors gracefully
3. **Loading States**: Show loading indicators for GraphQL queries
4. **Token Management**: Implement proper token refresh logic
5. **Caching**: Use Apollo Client cache effectively for better performance
6. **Security**: Always use HTTPS in production
7. **Monitoring**: Set up error tracking (Sentry, LogRocket, etc.)

## 🆘 Need Help?

1. Check the browser developer console for error messages
2. Verify all environment variables are set correctly
3. Test API endpoints independently using curl or Postman
4. Check AWS App Runner service logs in CloudWatch
5. Ensure your domain is whitelisted in Apollo server CORS settings

---

**Next Steps**: Once your frontend is connected, consider setting up error monitoring, analytics, and performance tracking for your production application.