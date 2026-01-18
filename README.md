<div align="center">

# 🚀 Golang Boilerplate

**A production-ready RESTful API boilerplate built with Go, following Clean Architecture principles.**

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Gin](https://img.shields.io/badge/Gin-v1.10.0-blue?style=for-the-badge)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-v1.25.12-red?style=for-the-badge)](https://gorm.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)

*Kickstart your next Go API project with battle-tested patterns and best practices.*

</div>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🏗️ **Clean Architecture** | Organized in layers (Handler → Service → Repository) for maintainability |
| 🔐 **JWT Authentication** | Secure token-based auth with role-based access control |
| 📦 **GORM + PostgreSQL** | Powerful ORM with auto-migration, soft deletes, and connection pooling |
| 📝 **Structured Logging** | Production-grade logging with Uber's Zap logger |
| ⚙️ **Viper Configuration** | Flexible config management via `.env` files and environment variables |
| 🔄 **Graceful Shutdown** | Handles OS signals for zero-downtime deployments |
| 📄 **Pagination** | Built-in pagination utilities for list endpoints |
| ✅ **Input Validation** | Request validation with human-readable error messages |
| 🐳 **Docker Ready** | Multi-stage Dockerfile + Docker Compose (App + PostgreSQL + Redis) |
| 🛡️ **Middleware Stack** | CORS, request logging, panic recovery, and JWT auth |

---

## 🏛️ Architecture

This project follows **Clean Architecture** principles, separating concerns into distinct layers:

```
┌──────────────────────────────────────────────────┐
│                   HTTP Layer                      │
│              (Gin Router + Middleware)             │
├──────────────────────────────────────────────────┤
│                  Handler Layer                    │
│           (Request/Response handling)             │
├──────────────────────────────────────────────────┤
│                  Service Layer                    │
│              (Business Logic)                     │
├──────────────────────────────────────────────────┤
│                Repository Layer                   │
│              (Data Access / GORM)                 │
├──────────────────────────────────────────────────┤
│                   Database                        │
│               (PostgreSQL)                        │
└──────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
golang-boilerplate/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration management (Viper)
│   ├── handler/
│   │   ├── health_handler.go       # Health check endpoint
│   │   ├── response.go             # Standardized API responses
│   │   └── user_handler.go         # User CRUD handlers
│   ├── middleware/
│   │   ├── auth.go                 # JWT authentication
│   │   ├── cors.go                 # CORS configuration
│   │   ├── logger.go               # Request logging (Zap)
│   │   └── recovery.go             # Panic recovery
│   ├── model/
│   │   ├── base.go                 # Base model (UUID + timestamps)
│   │   └── user.go                 # User model
│   ├── repository/
│   │   └── user_repository.go      # User data access layer
│   ├── router/
│   │   └── router.go               # Route registration
│   └── service/
│       └── user_service.go         # User business logic
├── pkg/
│   ├── pagination/
│   │   └── pagination.go           # Pagination utilities
│   └── validator/
│       └── validator.go            # Request validation
├── .env.example                    # Environment variables template
├── .gitignore
├── docker-compose.yml              # Docker Compose services
├── Dockerfile                      # Multi-stage Docker build
├── go.mod
├── go.sum
├── LICENSE
├── Makefile                        # Development commands
└── README.md
```

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.23 or higher
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/) (optional)
- [PostgreSQL](https://www.postgresql.org/) 15+ (if running without Docker)

### Quick Start with Docker

```bash
# Clone the repository
git clone https://github.com/tanawat011/golang-boilerplate.git
cd golang-boilerplate

# Start all services (App + PostgreSQL + Redis)
make docker-up

# The API is now running at http://localhost:8080
```

### Manual Setup

```bash
# Clone and navigate
git clone https://github.com/tanawat011/golang-boilerplate.git
cd golang-boilerplate

# Copy environment file
cp .env.example .env

# Install dependencies
go mod download

# Start PostgreSQL (make sure it's running on localhost:5432)

# Run the application
make run
```

---

## 📡 API Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/v1/health` | Health check | ❌ |
| `POST` | `/api/v1/users` | Create a new user | ❌ |
| `GET` | `/api/v1/users` | List all users (paginated) | ❌ |
| `GET` | `/api/v1/users/:id` | Get user by ID | ❌ |
| `PUT` | `/api/v1/users/:id` | Update user | ✅ |
| `DELETE` | `/api/v1/users/:id` | Delete user | ✅ |

### Example Requests

**Create User:**
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepass123",
    "full_name": "John Doe"
  }'
```

**List Users with Pagination:**
```bash
curl "http://localhost:8080/api/v1/users?page=1&limit=10"
```

**Response Format:**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com",
    "full_name": "John Doe",
    "role": "user",
    "is_active": true
  }
}
```

---

## ⚙️ Configuration

All configuration is managed through environment variables. Copy `.env.example` to `.env` and customize:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | Server port |
| `APP_MODE` | `development` | App mode (`development` / `production`) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `golang_boilerplate` | Database name |
| `JWT_SECRET` | — | JWT signing secret (change in production!) |
| `JWT_EXPIRATION_HOURS` | `24` | Token expiration time |

---

## 🛠️ Available Commands

```bash
make run              # Run the application
make build            # Build binary
make test             # Run tests with coverage
make lint             # Run linter (requires golangci-lint)
make fmt              # Format code
make vet              # Run go vet
make docker-up        # Start Docker Compose services
make docker-down      # Stop Docker Compose services
make docker-logs      # View logs
make clean            # Clean build artifacts
```

---

## 🧰 Tech Stack

| Technology | Purpose |
|-----------|---------|
| [Gin](https://github.com/gin-gonic/gin) | HTTP web framework |
| [GORM](https://gorm.io/) | ORM for database operations |
| [Viper](https://github.com/spf13/viper) | Configuration management |
| [Zap](https://github.com/uber-go/zap) | Structured logging |
| [JWT](https://github.com/golang-jwt/jwt) | Token-based authentication |
| [PostgreSQL](https://www.postgresql.org/) | Primary database |
| [Redis](https://redis.io/) | Caching (ready to integrate) |
| [Docker](https://www.docker.com/) | Containerization |

---

## 📝 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built with ❤️ and Go**

*If you found this helpful, give it a ⭐️!*

</div>
