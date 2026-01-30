# CTC DB API

## Features

- **RESTful API** with proper HTTP methods and status codes
- **JWT Authentication** for secure access
- **OpenAPI/Swagger Documentation** for interactive API exploration
- **HTTP Caching** with ETags and Last-Modified headers
- **Optimistic Locking** via If-Match headers to prevent lost updates
- **Pagination** for list endpoints
- **Filtering and Sorting** capabilities
- **Soft Deletes** for data safety
- **Request Validation** with detailed error messages
- **Observability** with structured logging, distributed tracing, and request IDs

## Table of Contents

- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
- [Request Examples](#request-examples)
- [Response Formats](#response-formats)
- [Error Handling](#error-handling)
- [Caching](#caching)
- [Development](#development)

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Docker & Docker Compose (optional)

### Environment Variables

Copy `example.env` to `config.env` and configure:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ctc_db

# JWT
JWT_SECRET_KEY=your-secret-key-here
JWT_TIMEOUT=10m
AUTH_IS_ENABLED=true

# Server
PORT=9080
```

### Running with Docker

```bash
docker-compose up -d
```

### Running Locally

```bash
# Install dependencies
go mod download

# Run the server
go run main.go
```

The API will be available at `http://localhost:9080`

## Authentication

This API uses JWT (JSON Web Token) for authentication.

### Login

**Endpoint:** `POST /api/v1/login`

**Request Body:**
```json
{
  "username": "your-username",
  "password": "your-password"
}
```

**Response:**
```json
{
  "data": {
    "username": "your-username",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### Using the Token

Include the token in the `Authorization` header for all protected endpoints:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  http://localhost:9080/api/v1/travellers
```

## API Endpoints

### Travellers

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/travellers` | Get paginated list of travellers with optional filters | Yes |
| GET | `/api/v1/travellers/:id` | Get a specific traveller by ID | Yes |
| POST | `/api/v1/travellers` | Create a new traveller | Yes |
| PUT | `/api/v1/travellers/:id` | Update an existing traveller | Yes |
| DELETE | `/api/v1/travellers/:id` | Soft delete a traveller | Yes |

### Accessories

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/accessories` | Get paginated list of accessories with optional filters | Yes |

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/login` | Authenticate and get JWT token | No |

## Request Examples

### Get All Travellers

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:9080/api/v1/travellers?page=1&page_size=10"
```

**Query Parameters:**
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 10, max: 100)
- `name` - Filter by traveller name (case insensitive)
- `influence` - Filter by influence name
- `job` - Filter by job name

### Get Specific Traveller

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:9080/api/v1/travellers/1
```

### Create Traveller

```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ophilia",
    "rarity": 5,
    "influence": "Wealth",
    "job": "Cleric",
    "banner": "Limited",
    "release_date": "2024-01-15",
    "accessory": {
      "name": "Sacred Staff",
      "hp": 100,
      "sp": 50,
      "patk": 80,
      "pdef": 60,
      "eatk": 120,
      "edef": 70,
      "spd": 90,
      "crit": 15,
      "effect": "Increases healing power"
    }
  }' \
  http://localhost:9080/api/v1/travellers
```

### Update Traveller (with Optimistic Locking)

```bash
# First, get the current ETag
ETAG=$(curl -s -I -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:9080/api/v1/travellers/1 | grep -i etag | awk '{print $2}')

# Then update with If-Match header
curl -X PUT \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -H "If-Match: $ETAG" \
  -d '{
    "name": "Ophilia (Updated)",
    "rarity": 5
  }' \
  http://localhost:9080/api/v1/travellers/1
```

### Delete Traveller

```bash
curl -X DELETE \
  -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:9080/api/v1/travellers/1
```

### Get Accessories with Filtering

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:9080/api/v1/accessories?owner=Ophilia&order_by=eatk&order_direction=desc"
```

**Query Parameters:**
- `page` - Page number
- `page_size` - Items per page
- `owner` - Filter by traveller name
- `effect` - Filter by effect description
- `order_by` - Sort by: hp, sp, patk, pdef, eatk, edef, spd, crit
- `order_direction` - Sort direction: asc, desc

## Response Formats

### Success Response

```json
{
  "data": {
    "id": 1,
    "name": "Ophilia",
    "rarity": 5,
    "influence": "Wealth",
    "job": "Cleric"
  }
}
```

### Paginated Response

```json
{
  "data": [...],
  "page": 1,
  "page_size": 10,
  "total": 50,
  "total_pages": 5
}
```

### Error Response

```json
{
  "message": "validation failed",
  "errors": [
    {
      "field": "name",
      "message": "name is required"
    },
    {
      "field": "rarity",
      "message": "rarity must be between 1 and 5"
    }
  ]
}
```

## Error Handling

The API uses standard HTTP status codes:

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (successful deletion) |
| 400 | Bad Request (validation errors) |
| 401 | Unauthorized (invalid/missing token) |
| 404 | Not Found |
| 409 | Conflict (duplicate resource) |
| 412 | Precondition Failed (ETag mismatch) |
| 500 | Internal Server Error |

## Caching

### ETags

All GET endpoints for individual resources return an `ETag` header:

```bash
curl -I -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:9080/api/v1/travellers/1

# Response includes:
# ETag: "1706601234"
# Last-Modified: Tue, 30 Jan 2024 12:00:34 GMT
```

### Conditional Requests

Use `If-None-Match` to avoid downloading unchanged resources:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  -H 'If-None-Match: "1706601234"' \
  http://localhost:9080/api/v1/travellers/1

# Returns 304 Not Modified if unchanged
```

### Optimistic Locking

Use `If-Match` to prevent lost updates:

```bash
curl -X PUT \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -H 'If-Match: "1706601234"' \
  -d '{"name": "Updated"}' \
  http://localhost:9080/api/v1/travellers/1

# Returns 412 Precondition Failed if resource was modified
```

## Development

### Interactive API Documentation

Swagger UI is available at: `http://localhost:9080/swagger/index.html`

You can explore all endpoints, view schemas, and test API calls directly from the browser.

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/traveller/...
```

### Regenerating Swagger Docs

After modifying API handlers:

```bash
swag init
```

### Project Structure

```
├── internal/           # Private application code
│   ├── accessory/     # Accessory domain (handler, service, repository)
│   ├── traveller/     # Traveller domain
│   ├── user/          # User domain
│   └── jwt/           # JWT token service
├── pkg/               # Public libraries
│   ├── controller/    # HTTP response helpers
│   ├── domain/        # Domain models and DTOs
│   ├── helpers/       # Utility functions
│   ├── middleware/    # HTTP middleware
│   ├── logging/       # Structured logging
│   └── validator/     # Request validation
├── docs/              # Generated Swagger documentation
└── testdata/          # Test fixtures
```

## License

This project is licensed under the MIT License.

REST API for managing CTC game database, built with Go (Echo framework) and PostgreSQL. Provides endpoints for managing travellers (game characters), accessories (equipment), and user authentication.

## 🚀 Features

- **RESTful API** with OpenAPI/Swagger documentation
- **JWT Authentication** with Bearer token support
- **HTTP Caching** with ETags and conditional requests
- **Optimistic Locking** via If-Match headers
- **Pagination** for list endpoints
- **Advanced Filtering & Ordering** for accessories
- **Request Validation** with detailed error messages
- **Observability** with distributed tracing (OpenTelemetry), structured logging (Zap), and Grafana/Loki integration

## 📋 Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Docker & Docker Compose (optional)
- [Swag](https://github.com/swaggo/swag) for Swagger docs generation

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd ctc-db-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install Swag CLI** (for regenerating documentation)
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

4. **Set up environment variables**
   
   Copy `example.env` to `config.env` and configure:
   ```bash
   cp example.env config.env
   ```
   
   Key variables:
   ```env
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=yourpassword
   DB_NAME=ctc_db
   
   # JWT
   JWT_SECRET_KEY=your-secret-key
   JWT_TIMEOUT=10m
   AUTH_IS_ENABLED=true
   
   # Server
   PORT=9080
   ENVIRONMENT=development
   ```

5. **Run database migrations**
   ```bash
   # Using the testdata seed file
   psql -U postgres -d ctc_db -f testdata/db-seed.sql
   ```

6. **Run the application**
   ```bash
   # Development
   go run main.go
   
   # Or with hot reload (using air)
   air
   
   # Production build
   go build -o ctc-db-api
   ./ctc-db-api
   ```

## 🐳 Docker Setup

```bash
# Start all services (API + PostgreSQL + Monitoring)
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## 📚 API Documentation

### Swagger UI

Interactive API documentation is available at:
```
http://localhost:9080/swagger/index.html
```

### Base URL

```
http://localhost:9080/api/v1
```

### Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

**Get a token:**
```bash
curl -X POST http://localhost:9080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "yourpassword"
  }'
```

Response:
```json
{
  "username": "admin",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## 🔗 API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/login` | Authenticate user and receive JWT token | ❌ |

### Travellers

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/travellers` | Get paginated list with filters | ✅ |
| GET | `/travellers/:id` | Get traveller by ID | ✅ |
| POST | `/travellers` | Create new traveller | ✅ |
| PUT | `/travellers/:id` | Update traveller (supports optimistic locking) | ✅ |
| DELETE | `/travellers/:id` | Soft delete traveller | ✅ |

### Accessories

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/accessories` | Get paginated list with filters and ordering | ✅ |

## 💡 Usage Examples

### 1. Login

```bash
curl -X POST http://localhost:9080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

### 2. Get All Travellers (with pagination)

```bash
curl -X GET "http://localhost:9080/api/v1/travellers?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. Get Travellers by Filter

```bash
# Filter by influence
curl -X GET "http://localhost:9080/api/v1/travellers?influence=Wind&page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Filter by job
curl -X GET "http://localhost:9080/api/v1/travellers?job=Dancer" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Filter by name (case insensitive)
curl -X GET "http://localhost:9080/api/v1/travellers?name=viola" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. Get Traveller by ID (with caching)

```bash
curl -X GET http://localhost:9080/api/v1/travellers/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -i  # Show headers to see ETag and Last-Modified
```

Response headers include:
```
ETag: "1706621234"
Last-Modified: Wed, 30 Jan 2026 10:30:45 GMT
Cache-Control: public, max-age=300
```

### 5. Create Traveller (with accessory)

```bash
curl -X POST http://localhost:9080/api/v1/travellers \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Viola",
    "rarity": 5,
    "banner": "Standard Banner",
    "release_date": "01-10-2024",
    "influence": "Wind",
    "job": "Dancer",
    "accessory": {
      "name": "Crimson Cloak",
      "hp": 500,
      "sp": 50,
      "patk": 120,
      "pdef": 80,
      "eatk": 150,
      "edef": 100,
      "spd": 45,
      "crit": 25,
      "effect": "Increases elemental damage by 15%"
    }
  }'
```

### 6. Update Traveller (with optimistic locking)

```bash
# First, get the current ETag
ETAG=$(curl -s -X GET http://localhost:9080/api/v1/travellers/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -I | grep -i etag | awk '{print $2}' | tr -d '\r')

# Then update with If-Match header
curl -X PUT http://localhost:9080/api/v1/travellers/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -H "If-Match: $ETAG" \
  -d '{
    "name": "Viola",
    "rarity": 5,
    "banner": "Limited Banner",
    "release_date": "01-10-2024",
    "influence": "Wind",
    "job": "Dancer"
  }'
```

If the resource was modified by another client, you'll receive `412 Precondition Failed`.

### 7. Delete Traveller

```bash
curl -X DELETE http://localhost:9080/api/v1/travellers/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 8. Get Accessories with Ordering

```bash
# Order by elemental attack descending
curl -X GET "http://localhost:9080/api/v1/accessories?order_by=eatk&order_dir=desc" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Filter by effect and order by speed
curl -X GET "http://localhost:9080/api/v1/accessories?effect=damage&order_by=spd&order_dir=asc" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Filter by owner (traveller name)
curl -X GET "http://localhost:9080/api/v1/accessories?owner=Viola" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🎯 Response Format

### Success Response

```json
{
  "message": "success",
  "data": { }
}
```

### Error Response

```json
{
  "message": "validation failed",
  "errors": [
    {
      "field": "rarity",
      "message": "rarity must be no greater than 5"
    }
  ]
}
```

### Paginated Response

```json
{
  "data": [...],
  "page": 1,
  "page_size": 10,
  "total": 150,
  "total_pages": 15
}
```

## 🏗️ Project Structure

```
.
├── cmd/                    # Application entrypoints
├── configs/               # Configuration files (Grafana, Loki, Promtail)
├── docs/                  # Generated Swagger documentation
├── internal/              # Private application code
│   ├── accessory/        # Accessory domain (handler, service, repository)
│   ├── jwt/              # JWT token service
│   ├── traveller/        # Traveller domain
│   └── user/             # User/authentication domain
├── pkg/                   # Public libraries
│   ├── constants/        # Application constants
│   ├── controller/       # HTTP response helpers
│   ├── domain/           # Domain models and DTOs
│   ├── helpers/          # Utility functions
│   ├── logging/          # Structured logging
│   ├── middleware/       # HTTP middlewares
│   ├── telemetry/        # Tracing and monitoring
│   └── validator/        # Request validation
├── testdata/             # Test data and fixtures
├── docker-compose.yml    # Docker orchestration
├── Dockerfile            # Container image
├── main.go               # Application entry point
└── Makefile              # Build and deployment scripts
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/traveller/...

# Run integration tests (requires database)
go test -tags=integration ./...
```

## 🔄 Regenerate Swagger Documentation

After modifying handler annotations:

```bash
swag init
```

Or using make:
```bash
make swagger
```

This generates updated files in the `docs/` directory.

## 🧪 Using Swagger UI

Access the interactive API documentation at `http://localhost:9080/swagger/index.html`

### Authenticating in Swagger UI

1. First, use the `POST /api/v1/login` endpoint to get your JWT token
2. Copy the token value from the response (just the token string, not the entire JSON)
3. Click the **"Authorize"** button (lock icon) at the top right of the Swagger UI
4. In the "Value" field, type `Bearer ` followed by your token
   - **Important**: Include the word "Bearer" and a space before pasting your token
   - Example: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
5. Click **"Authorize"** and then **"Close"**
6. Now you can use any protected endpoint - the Authorization header will be added automatically

### Troubleshooting Swagger Authentication

If you get "401 missing or malformed jwt":
- Make sure you included the word `Bearer ` (with a space) before your token
- Verify the token hasn't expired (default timeout is 10 minutes)
- Check that you clicked "Authorize" after entering the token

## 📊 Monitoring

Access monitoring dashboards (when using Docker Compose):

- **Grafana**: http://localhost:3000 (admin/admin)
- **Loki**: http://localhost:3100

## 🔐 Security Notes

- **JWT Secrets**: Use strong, randomly generated secrets in production
- **Password Hashing**: Passwords are hashed using bcrypt
- **SQL Injection**: Protected via GORM parameterized queries
- **Input Validation**: All inputs validated using go-playground/validator
- **Rate Limiting**: Consider implementing rate limiting for production

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👤 Contact

**Liz** - j2qgehn84@mozmail.com

## 🙏 Acknowledgments

- [Echo Framework](https://echo.labstack.com/)
- [GORM](https://gorm.io/)
- [Swaggo](https://github.com/swaggo/swag)
- [Zap Logger](https://github.com/uber-go/zap)
- [OpenTelemetry](https://opentelemetry.io/)
