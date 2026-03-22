# apollo
### Social Sports Line App

A backend service for tracking and comparing NFL spread picks among users in leagues, backed by The Odds API and stored in PostgreSQL. Built with Go, GraphQL, and React, deployed on AWS.

Currently only availble during the american football seasons and for select users.  

### Overview

Tracks sports betting picks for groups. Users join leagues, record picks against NFL spreads, and compare results. Odds data is fetched on-demand from The Odds API and game finalization runs on a schedule via AWS Lambda.

### Tech Stack
- **Backend**: Go 1.24, GraphQL (gqlgen), REST (gorilla/mux)
- **Database**: PostgreSQL, migrations (golang-migrate)
- **Auth**: JWT, bcrypt
- **Frontend**: React
- **Infrastructure**: AWS App Runner, Lambda, RDS, Secrets Manager, ECR, Docker
- **Tooling**: Docker Compose

### Architecture
Go server exposes GraphQL for picks, seasons, weekly spreads, and leaderboards. REST for auth and league helpers.
Odds fetched on-demand; no continuous polling.
Game finalization runs via AWS Lambda on a schedule.
PostgreSQL stores user data, picks, and odds.

### Getting Started
Prerequisites
Go 1.24+, Node.js, Docker Compose
The Odds API key
AWS CLI

### Run with Docker Compose
```
docker compose up --build
```

API listens on port 8080. GraphQL Playground: http://localhost:8080/

### Usage

REST Authentication

POST /api/signup → register
POST /api/login → login, returns JWT
Include Authorization: Bearer <token> for protected routes

### GraphQL

Queries: GetWeeklyNflGameSpreads, mySeasonPicks, seasonLeaderboard
Mutations: CreateGamePick, CreateLeagueSeason, FinalizeGames
Playground available at http://localhost:8080/

### Development
Dependencies: go mod tidy
Regenerate GraphQL code:
go run github.com/99designs/gqlgen generate

### Future Improvements
Add real-time updates via SSE/WebSockets
User accounts and authentication improvements
Optimize odds fetching and caching
Self service season and league creation
Finalize weekly games on aws lambda job
