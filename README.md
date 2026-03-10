# CTC DB API

## Table of Contents

- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Development](#development)
- [Observability & Monitoring](#observability--monitoring)
- [Database & Migrations](#database--migrations)
- [Contributing](#contributing)
- [Contact](#contact)

## Quick Start

### Prerequisites

- **Go** 1.24 or higher
- **PostgreSQL** (local or via Docker)
- **Docker & Docker Compose** (optional, for containerized setup)

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/lizobly/ctc-db-api.git
   cd ctc-db-api
   ```

2. **Configure environment**
   ```bash
   cp example.env config.env
   # Edit config.env with your local database credentials
   ```

3. **Run the application**

   **Option A: Locally**
   ```bash
   go run main.go
   ```

   **Option B: With Docker**
   ```bash
   docker-compose --profile with-app up -d --build
   ```

4. **Access the API**

   **Option A: Locally**
   - **Swagger UI**: http://localhost:9080/swagger/index.html
   - **API Base**: http://localhost:9080/api/v1

   **Option B: With Docker**
   - **Swagger UI**: http://localhost:8080/swagger/index.html
   - **API Base**: http://localhost:8080/api/v1

## Project Structure

```
internal/          # Core business logic (clean architecture layer)
├── user/         # User management and authentication
├── traveller/    # Traveller data operations
├── accessory/    # Accessory/equipment management
└── jwt/          # JWT token service

pkg/               # Shared utilities and packages
├── controller/   # HTTP controller (routes, request handling)
├── domain/       # Domain models (User, Traveller, Accessory)
├── helpers/      # Utility functions (env, pagination, caching, etc.)
├── logging/      # Structured logging with Zap
├── middleware/   # HTTP middleware (JWT, request ID, tracing, etc.)
├── telemetry/    # OpenTelemetry setup and utilities
└── validator/    # Input validation

docs/              # Auto-generated Swagger documentation
testdata/          # Test fixtures and seed data
configs/           # External service configurations (Grafana, Loki, Promtail)
```

### Architecture Notes

- **Clean Architecture**: Separation of concerns with domain, service, handler, and repository layers
- **Dependency Injection**: Services are injected into handlers for testability
- **Test Coverage**: Unit tests alongside implementations (`*_test.go` files)
- **Integration Tests**: Uses Testcontainers for isolated database testing

## API Documentation

The API documentation is auto-generated and available via **Swagger UI** at:
```
http://localhost:9080/swagger/index.html
```

### Authentication

All protected endpoints require a **Bearer token** in the `Authorization` header:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
  http://localhost:9080/api/v1/users
```

### Main Endpoints

- **Users**: `/api/v1/users` - User registration, login, profile management
- **Travellers**: `/api/v1/travellers` - CRUD operations for traveller entities
- **Accessories**: `/api/v1/accessories` - CRUD operations for accessories

For detailed endpoint specifications, request/response schemas, and examples, see the **Swagger UI**.

## Configuration

### Environment Variables

Copy `example.env` to `config.env` and configure:

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_HOST` | PostgreSQL host | `localhost` |
| `DATABASE_PORT` | PostgreSQL port | `5432` |
| `DATABASE_USER` | Database user | `postgres` |
| `DATABASE_PASSWORD` | Database password | `password` |
| `DATABASE_NAME` | Database name | `ctc_db` |
| `JWT_SECRET` | Secret key for signing JWT tokens | (use a strong secret) |
| `ENVIRONMENT` | Execution mode (`development` or `production`) | `development` |
| `LOG_LEVEL` | Logging level (`debug`, `info`, `warn`, `error`) | `info` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint | `http://localhost:4318` |

### Environment Modes

- **development**: Text-based logging, relaxed validation
- **production**: JSON logging (compatible with Loki), stricter validation

## Development

### Running Tests

**Unit tests only** (fast, no external dependencies):
```bash
make test-unit
```

**All tests** (includes integration tests with Testcontainers):
```bash
make test
```

### Running with Docker

Build and run the entire stack (API + observability):
```bash
docker-compose --profile with-app up -d --build
```

The compose file includes:
- **App**: CTC DB API on port 9080
- **Jaeger**: Distributed tracing UI on port 16686
- **Loki**: Log aggregation on port 3100
- **Grafana**: Dashboards (if configured in configs/)

### Code Style

This project follows standard Go conventions:
- Run `go fmt ./...` to format code
- Use meaningful variable and function names
- Write tests alongside your implementations

## Observability & Monitoring

This project includes a complete observability stack:

### Stack Components

| Component | Purpose | Port | Access |
|-----------|---------|------|--------|
| **Jaeger** | Distributed tracing | 16686 | http://localhost:16686 |
| **Loki** | Log aggregation | 3100 | http://localhost:3100 |
| **Promtail** | Log shipper | (internal) | - |
| **Grafana** | Dashboards | 3000 | http://localhost:3000 (if enabled) |

### What Gets Tracked

- **Traces**: All HTTP requests with automatically generated span IDs
- **Logs**: Structured logs with request context, error details, and performance metrics
- **Metrics**: Request latency, error rates, and database performance

### Debugging

To view traces for a specific request:
1. Copy the **Request ID** from the response header or logs
2. Open **Jaeger UI** (http://localhost:16686)
3. Search for traces matching that request ID
4. Inspect spans to see latency and error details

To view logs:
1. Open **Loki** (http://localhost:3100)
2. Query logs by labels (service, level, request_id, etc.)

## Database & Migrations

Database schema and migrations are managed in a separate repository:

**[ctc-db](https://github.com/lizobly/ctc-db)** - Database setup, schema, and seed data

Follow the instructions in that repository to:
- Set up the PostgreSQL database
- Run migrations
- Seed initial data

This API assumes the database tables already exist and are properly migrated.

## Contributing

1. Read the [pull request template](./pull_request_template.md) for PR guidelines
2. Ensure all tests pass: `make test`
3. Write tests for new features
4. Follow Go code style conventions
5. Use clear, descriptive commit messages

## Contact

- **Email**: j2qgehn84@mozmail.com

---

For more information on the database schema and setup, see [ctc-db](https://github.com/lizobly/ctc-db).
