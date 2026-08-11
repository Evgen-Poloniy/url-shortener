# URL Shortener Service

A RESTful URL shortener service written in Go. It converts original URLs into unique 10-character short aliases using an alphabet consisting of uppercase/lowercase Latin letters, digits, and underscores ([a-zA-Z0-9_]).

The service supports dual storage backends: an in-memory storage implementation (thread-safe using sync.RWMutex / sync.Map) and PostgreSQL, selected dynamically via environment variables or configuration flags.

## Features

- Unique 10-Character Short Keys: Guarantees a deterministic 1:1 mapping between original URLs and short aliases.
- Dual Storage Options: Switch seamlessly between memory and postgres.
- High Concurrency & Scalability: Optimized for high throughput with concurrent-safe memory structures and connection pooling for PostgreSQL.
- API Security & Middleware: Includes API Key authentication, CORS support, recovery middleware, secure headers, structured logging, and centralized error handling.
- Swagger Documentation: Automated OpenAPI specification and interactive API testing interface.
- Mocking & Database Migrations: Out-of-the-box support for golang/mock code generation and SQL schema migrations.
- Containerization & Tooling: Comprehensive Dockerfile, docker-compose.yaml, and Makefile for deployment and testing.

---

## Technical Design & Key Decisions

### 1. Short URL Generation Algorithm
- **Alphabet**: `[a-zA-Z0-9_]` (63 possible characters).
- **Format**: Exactly 10 characters long.
- **Capacity**: $63^{10} \approx 9.85 \times 10^{17}$ unique combinations, preventing hash collision issues at scale.
- **Implementation & Uniqueness**:
  - **Existing URLs**: A database or memory lookup checks if the original URL has already been shortened. If present, it returns the existing 10-character alias, satisfying the requirement that one original URL maps to exactly one short URL.
  - **New URLs**: A SHA-256 digest of the original URL is computed, taking the first 8 bytes as a `uint64` integer. This value is iteratively mapped onto the 63-character alphabet via modulo operations (`num % 63`) to derive a deterministic 10-character alias.

### 2. High Concurrency & Longevity
- In-Memory Store: Thread-safe implementation avoiding data races under concurrent reads/writes (-race clean).
- PostgreSQL Store: Configured with database connection pooling (SetMaxOpenConns, SetMaxIdleConns, SetConnMaxLifetime) to efficiently handle hundreds of simultaneous requests without leaking resources or exhausting connection limits.

---

## Project Structure

```text
.
├── cmd/
│   └── url-shortener/
│       └── main.go          # Application entry point
├── configs/
│   └── config.yaml          # Default configuration file
├── docs/                    # Generated Swagger API documentation
├── internal/
│   ├── config/              # Application configuration loader
│   ├── domain/              # Application errors
│   ├── middleware/          # HTTP middlewares (Auth, CORS, Logger, Errors, Security headers)
│   ├── repository/          # Storage implementations
│   │   ├── memory/          # Thread-safe in-memory store
│   │   └── postgres/        # PostgreSQL store
│   ├── service/
│   │   └── shortener/       # Business logic (URL generator and shortener service)
│   ├── transport/
│   │   └── http/
│   │        └── v1/         # REST API HTTP handlers
│   └── pkg/                 # Shared packages
│       ├── database/        # PostgreSQL initialization
│       └── logs/            # Logrus logger initialization
├── migrations/              # SQL migration files
├── tests/
│   └── integration/         # Integration test suites
├── Dockerfile               # Multi-stage Docker build configuration
├── docker-compose.yaml      # Orchestration for app and PostgreSQL container
├── Makefile                 # Development task automation commands
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.26.1
- Docker & Docker Compose
- make utility
- golang-migrate (optional, for running local migrations outside Docker)
- golangci-lint

---

## Configuration

The service can be configured via environment variables or a YAML configuration file.

| Environment Variable | Description |
| :--- | :--- |
| CONFIG_PATH | Path to YAML config file |
| API_HOST | Host address to bind HTTP server |
| API_PORT | Port to listen on |
| API_KEY | Secret API key for protected routes |
| DB_HOST | PostgreSQL database host |
| DB_PORT | PostgreSQL database port |
| DB_USER | PostgreSQL user |
| DB_PASSWORD | PostgreSQL user password |
| DB_NAME | PostgreSQL database name |
| SSL_MODE | PostgreSQL SSL connection mode |

---

## Running the Application

### Option 1: Docker Compose (Storage type "postgres")

To run the complete setup (App + Migrate + PostgreSQL) using Docker:
```bash
make up-postgres
```
To stop and tear down containers:
```bash
make down
```

### Option 2: Docker (Storage type "memory")

To run App with RAM memory storage:
```bash
make up-memory
```

To stop container:
```bash
make down
```
---

## API Documentation & Usage

Once the application is running, Swagger UI is available at:
http://localhost:8080/swagger/index.html

### Endpoints

#### 1. Create Short URL
- HTTP Method: POST
- Path: /api/v1/urls
- Headers:
  - Content-Type: application/json
  - X-API-Key: <YOUR_API_KEY> (if configured)
- Request Body:
```json
{
    "url": "https://example.com/very/long/url/path?param=value"
}
```
- Response (201 Created):
```json
{
    "data": {
        "short_url": "Gq7rmpLMN2"
    }
}
```

#### 2. Get Original URL
- HTTP Method: GET
- Path: /api/v1/urls/Gq7rmpLMN2
- Headers:
  - X-API-Key: <YOUR_API_KEY> (if configured)
- Response (200 OK):
```json
{
    "data": {
        "url": "https://example.com/very/long/url/path?param=value"
    }
}
```

#### 3. Health Checks
- HTTP Method: GET
- Path: /health or /healthz
- Response (200 OK):
```json
{
    "status": "ok"
}
```
---

## Code Generation & Database Migrations

### Mock Generation
Generates mock implementations for interfaces using go generate:
```bash
make mock
```
### Database Migrations
Creates a new pair of sequential SQL migration files (.up.sql and .down.sql) inside the migrations/ directory:
```bash
make migrate
```
---

## Testing

### Unit Tests
Run unit tests with race condition detection and coverage statistics:
```bash
make unit-test
```
Generate an HTML code coverage report:
```bash
make test-html
```
Launch linters:
```bash
make linter
```

### Integration Tests
Run integration tests against the API endpoints and external dependencies:
```bash
make integration-test
```
---

## Useful Makefile Commands

| Command | Action |
| :--- | :--- |
| make unit-test | Runs unit tests across /internal/... with -race and coverage |
| make integration-test | Runs integration test suites |
| make test-html | Opens HTML code coverage report |
| make mock | Triggers go generate ./... to update interface mocks |
| make swag-init | Regenerates OpenAPI/Swagger specification files |
| make up-postgres | Builds images and starts Docker Compose services with PostgreSQL storage type in detached mode |
| make up-memory | Builds images and starts Docker Compose services with RAM memory storage type in detached mode |
| make migrate | Generates a new SQL migration file pair in migrations/ |
| make down | Stops and removes Docker Compose resources |
