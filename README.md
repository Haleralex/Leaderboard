# 🎮 Leaderboard Microservice

A production-ready, scalable leaderboard microservice built with Go for game applications. Features JWT authentication, Redis caching, PostgreSQL persistence, and comprehensive API endpoints for score submission and leaderboard queries.

## 📋 Features

- **Score Management**: Submit and update player scores with anti-cheat rate limiting
- **Leaderboard Queries**: Paginated, sorted leaderboards with multiple filtering options
- **Seasonal Support**: Multiple leaderboard seasons (e.g., global, weekly, monthly)
- **High Performance**: Redis caching with PostgreSQL fallback for optimal speed
- **Authentication**: JWT-based authentication middleware
- **Rate Limiting**: Per-IP rate limiting to prevent abuse
- **Health Checks**: Kubernetes-ready health, readiness, and liveness endpoints
- **Production-Ready**: Structured logging, graceful shutdown, CORS support

## 🏗️ Architecture

```
┌─────────────┐      ┌──────────────┐      ┌──────────────┐
│   Unity     │─────▶│ Leaderboard  │─────▶│  PostgreSQL  │
│   Client    │      │   Service    │      │  (Supabase)  │
└─────────────┘      │   (Go/Chi)   │      └──────────────┘
                     │              │
                     │              │─────▶│    Redis     │
                     └──────────────┘      │   (Cache)    │
                                           └──────────────┘
```

### Tech Stack

- **Language**: Go 1.25+
- **Router**: Chi v5
- **Database**: PostgreSQL (Supabase) with GORM
- **Cache**: Redis v7 with go-redis/v9
- **Auth**: JWT (golang-jwt/jwt/v5)
- **Logging**: Zerolog
- **Testing**: Testify
- **Containerization**: Docker & Docker Compose

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL (or Supabase account)
- Redis
- Docker (optional)

### 1. Clone & Setup

```bash
git clone https://github.com/Haleralex/leaderboard-service.git
cd leaderboard-service

# Copy environment variables
cp .env.example .env

# Edit .env with your credentials
# Required: DATABASE_URL, JWT_SECRET
```

### 2. Database Setup

Apply the schema to your PostgreSQL database:

```bash
# Using Supabase SQL Editor: paste contents of sql/schema.sql

# Or using psql:
psql $DATABASE_URL < sql/schema.sql
```

### 3. Run Locally

```bash
# Install dependencies
go mod download

# Run the service
go run cmd/server/main.go
```

The service will start on `http://localhost:8080`

### 4. Run with Docker

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f leaderboard

# Stop services
docker-compose down
```

## 📚 API Documentation

Base URL: `http://localhost:8080/api/v1`

### Authentication Endpoints

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "Player1",
  "email": "player1@example.com",
  "password": "secure_password"
}

Response: 201 Created
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Player1",
    "email": "player1@example.com",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "player1@example.com",
  "password": "secure_password"
}

Response: 200 OK
{
  "success": true,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "expires_at": 1704153600
  }
}
```

### Leaderboard Endpoints (Requires JWT)

All leaderboard endpoints require JWT token in header:
```
Authorization: Bearer <your_jwt_token>
```

#### Submit Score
```http
POST /api/v1/submit-score
Authorization: Bearer <token>
Content-Type: application/json

{
  "score": 1000,
  "season": "global",
  "metadata": {
    "level": "10",
    "time_played": "3600"
  }
}

Response: 200 OK
{
  "success": true,
  "message": "score submitted successfully",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "score": 1000,
    "season": "global",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

#### Get Leaderboard
```http
GET /api/v1/leaderboard?season=global&limit=50&page=0&sort=desc
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "entries": [
      {
        "rank": 1,
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "user_name": "Player1",
        "score": 1000,
        "season": "global",
        "timestamp": "2024-01-01T12:00:00Z"
      }
    ],
    "total_count": 100,
    "page": 0,
    "limit": 50,
    "has_next": true,
    "next_cursor": "1:1000"
  }
}
```

Query Parameters:
- `season` (string, default: "global"): Leaderboard season
- `limit` (int, default: 50, max: 100): Results per page
- `page` (int, default: 0): Page number
- `sort` (string, default: "desc"): Sort order ("asc" or "desc")
- `cursor` (string, optional): Cursor for cursor-based pagination

#### Get User Rank
```http
GET /api/v1/leaderboard/user/{userID}?season=global
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "rank": 5,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_name": "Player1",
    "score": 750,
    "season": "global",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### Health Endpoints (No Auth Required)

```http
GET /health         # Overall health check
GET /ready          # Readiness probe (Kubernetes)
GET /live           # Liveness probe (Kubernetes)
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v ./internal/leaderboard/handler -run TestLeaderboardHandler

# Run with race detector
go test -race ./...
```

## 🔧 Configuration

Environment variables (`.env` file):

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | 8080 | No |
| `ENV` | Environment (development/production) | development | No |
| `DATABASE_URL` | PostgreSQL connection string | - | **Yes** |
| `REDIS_ADDR` | Redis address | localhost:6379 | No |
| `REDIS_PASSWORD` | Redis password | - | No |
| `JWT_SECRET` | Secret key for JWT signing | - | **Yes** |
| `JWT_EXPIRY_HOURS` | JWT token expiry in hours | 24 | No |
| `RATE_LIMIT_REQUESTS` | Max requests per window | 100 | No |
| `RATE_LIMIT_WINDOW_SECONDS` | Rate limit window in seconds | 60 | No |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | info | No |

## 📁 Project Structure

```
leaderboard-service/
├── cmd/
│   ├── server/
│   │   └── main.go              # Application entry point
│   ├── seed/
│   │   └── main.go              # Database seeding
│   └── simulator/
│       └── main.go              # Load testing simulator
├── internal/
│   ├── auth/                    # Auth module
│   │   ├── domain/              # Domain models & events
│   │   ├── handler/             # HTTP handlers
│   │   ├── repository/          # Data access
│   │   ├── service/             # Business logic
│   │   └── models/              # DTOs
│   ├── leaderboard/             # Leaderboard module
│   │   ├── domain/              # Domain models & events
│   │   ├── handler/             # HTTP handlers
│   │   ├── repository/          # Data access
│   │   ├── service/             # Business logic
│   │   └── models/              # DTOs
│   ├── shared/                  # Shared components
│   │   ├── config/              # Configuration management
│   │   ├── database/            # PostgreSQL & Redis connections
│   │   ├── middleware/          # JWT, rate limiting, CORS, logger
│   │   ├── repository/          # Base repository, specifications, UoW, decorators
│   │   ├── eventbus/            # Event bus for domain events
│   │   └── utils/               # Helpers, validators, errors
│   ├── factory/                 # Factory pattern implementations
│   ├── strategy/                # Strategy pattern for ranking
│   ├── websocket/               # WebSocket hub & clients
│   └── handlers/                # Shared handlers (health, websocket)
├── sql/
│   └── schema.sql               # Database schema
├── scripts/                     # Utility scripts
├── .env.example                 # Example environment variables
├── Dockerfile                   # Docker image definition
├── docker-compose.yml           # Docker Compose configuration
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## 🚢 Deployment

### Docker Deployment

```bash
# Build production image
docker build -t leaderboard-service:latest .

# Run with environment variables
docker run -p 8080:8080 \
  -e DATABASE_URL="postgresql://..." \
  -e REDIS_ADDR="redis:6379" \
  -e JWT_SECRET="your-secret" \
  leaderboard-service:latest
```

### Kubernetes Deployment

Example Kubernetes manifests:

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: leaderboard-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: leaderboard
  template:
    metadata:
      labels:
        app: leaderboard
    spec:
      containers:
      - name: leaderboard
        image: leaderboard-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: leaderboard-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: leaderboard-secrets
              key: jwt-secret
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 🔄 CI/CD

### GitHub Actions Example

Create `.github/workflows/ci.yml`:

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Install dependencies
      run: go mod download

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Build Docker image
      run: docker build -t leaderboard-service:${{ github.sha }} .

    - name: Push to registry
      run: |
        echo "${{ secrets.DOCKER_PASSWORD }}" | docker login -u "${{ secrets.DOCKER_USERNAME }}" --password-stdin
        docker push leaderboard-service:${{ github.sha }}
```

## 🌟 Extensions & Future Features

### 1. Real-time Updates with WebSockets

Add WebSocket support for live leaderboard updates:

```go
// internal/handlers/websocket_handler.go
func (h *LeaderboardHandler) WebSocketLeaderboard(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()

    // Subscribe to Redis pub/sub for score updates
    pubsub := h.redis.Subscribe(ctx, "leaderboard:updates")

    for msg := range pubsub.Channel() {
        conn.WriteJSON(msg.Payload)
    }
}
```

### 2. Advanced Anti-Cheat

Implement score validation and anomaly detection:

- Score velocity checks
- Pattern recognition
- Statistical outlier detection
- Machine learning models

### 3. Seasonal Rotation

Auto-rotate seasons with cron jobs:

```go
// Create new season weekly/monthly
func (s *Service) RotateSeason() {
    newSeason := fmt.Sprintf("season_%s", time.Now().Format("2006_01"))
    // Archive old season, create new
}
```

### 4. Analytics Dashboard

Track metrics with Prometheus/Grafana:

- Score submission rates
- API response times
- Cache hit/miss ratios
- User activity patterns

### 5. OAuth2 Integration

Add Supabase OAuth2 support:

```go
// internal/middleware/oauth2.go
import "golang.org/x/oauth2"

func SupabaseOAuth2() {
    oauth2Config := &oauth2.Config{
        ClientID:     os.Getenv("SUPABASE_CLIENT_ID"),
        ClientSecret: os.Getenv("SUPABASE_CLIENT_SECRET"),
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://your-project.supabase.co/auth/v1/authorize",
            TokenURL: "https://your-project.supabase.co/auth/v1/token",
        },
    }
    // Implement OAuth2 flow
}
```

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 👨‍💻 Author

Built with ❤️ by Haler

## 🙏 Acknowledgments

- Chi router for the excellent HTTP framework
- Supabase for managed PostgreSQL
- Redis for blazing-fast caching
- The Go community for amazing libraries

---

**Need help?** Open an issue or contact [your-email@example.com]
