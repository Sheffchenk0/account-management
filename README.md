# Account Management API

- [How to Run](#how-to-run)
- [Design Decisions](#design-decisions)
- [Test Suite](#test-suite)

---

## How to run

### Requirements

- Docker and Docker Compose
- PostgreSQL 15+
- RabbitMQ 3+

### Environment

Create a `.env` file in the root directory with the following variables:

```env
APP_NAME=account-manager
APP_VERSION=1.0.0
HTTP_PORT=8080
PG_POOL_MAX=10
PG_URL=postgres://user:pass@localhost:5432/account_db?sslmode=disable
RMQ_URL=amqp://guest:guest@localhost:5672/
RMQ_EXCHANGE_NAME=account_events
```

### Running with Docker Compose

1. Build and start:

```bash
docker-compose up --build
```

This will:
- Start PostgreSQL
- Start RabbitMQ
- Run migrations
- Start the API
- Start the outbox worker

2. Access the application:
- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- RabbitMQ Management UI: `http://localhost:15672` (guest/guest)

#### API Documentation

The API includes interactive Swagger/OpenAPI documentation that provides:
- Detailed descriptions of all endpoints
- Request/response schemas
- Interactive API testing from the browser
- Information about authentication, parameters, and possible errors

Access the Swagger UI to explore and test the API endpoints directly from your browser.

### Local Dev Setup

1. Start dependencies:

```bash
docker-compose up postgres rabbitmq
```

2. Run migrations:

```bash
goose postgres "postgres://user:pass@localhost:5432/account_db?sslmode=disable" up
```

3. Install dependencies:

```bash
go mod download
```

4. Run application:

```bash
go run cmd/app/main.go
```

## Design Decisions

### Architecture Overview

The application follows "Clean Architecture" principles with clear separation of concerns:

### Key Architectural Patterns

#### 1. Outbox Pattern

The system implements the "Outbox Pattern" to ensure reliable event publishing.

#### 2. Transaction Management

The application uses a "TransactionManager" to handle database transactions.

#### 3. Clean Architecture Layers

Entity Layer (`internal/entity/`)

Repository Layer (`internal/repo/`)

Service Layer (`internal/service/`)

Controller Layer (`internal/controller/`)

Worker Layer (`internal/worker/`)

## Test Suite

- **Unit tests**: Entity and service layer with table-driven tests and mocks (`internal/entity/`, `internal/service/`)
- **Integration tests**: e2e tests using testcontainers with PostgreSQL (`tests/integration_test.go`)

```bash
go test ./...              
go test -short ./...       
go test -cover ./...       
```

