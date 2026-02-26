# Go Boilerplate

A production-ready Go REST API boilerplate with clean architecture, JWT authentication, PostgreSQL, and Redis.

## Features

- **Clean Architecture** - Separation of concerns with domain, service, repository, and HTTP layers
- **JWT Authentication** - RSA-signed access and refresh tokens with automatic rotation
- **PostgreSQL** - Type-safe SQL with sqlc code generation
- **Redis** - Refresh token storage and distributed rate limiting
- **Password Hashing** - Argon2id with configurable parameters
- **Request Validation** - go-playground/validator with custom phone validation
- **Database Migrations** - Goose for version-controlled schema changes
- **Docker Ready** - Multi-stage Dockerfile and docker-compose setup
- **Soft Delete** - Notes support soft delete with restore capability

## Quick Start

```bash
# Clone the repository
git clone https://github.com/nate/go-boilerplate.git
cd go-boilerplate

# Copy environment file
cp .env.example .env

# Generate RSA keys
make generate-keys

# Start services
make docker-up

# Run migrations (if not using docker)
make migrate-up

# Start the server
make run
```

The API will be available at `http://localhost:8080`.

## Prerequisites

- Go 1.23+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose (optional)
- sqlc (for code generation)
- goose (for migrations)

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── db/                      # Generated sqlc code
│   ├── domain/                  # Domain models (User, Note)
│   │   ├── user.go
│   │   ├── note.go
│   │   └── common.go
│   ├── errors/                  # Custom error types with codes
│   ├── http/
│   │   ├── handler/             # HTTP handlers by domain
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── note_handler.go
│   │   │   ├── helpers.go
│   │   │   └── responses.go
│   │   ├── middleware/          # Chi middleware
│   │   │   ├── auth.go
│   │   │   ├── rate_limit.go
│   │   │   └── common.go
│   │   └── router.go            # Route definitions
│   ├── repository/              # Data access layer
│   │   ├── interfaces.go
│   │   ├── user_repository.go
│   │   ├── note_repository.go
│   │   └── redis.go
│   └── service/                 # Business logic layer
│       ├── auth_service.go
│       ├── user_service.go
│       └── note_service.go
├── pkg/                         # Reusable packages
│   ├── jwt/
│   │   ├── jwt.go               # JWT token management
│   │   └── keys.go              # RSA key handling
│   ├── password/
│   │   └── argon2id.go          # Password hashing
│   ├── response/
│   │   └── response.go          # JSON response utilities
│   └── validation/
│       └── validation.go        # Request validation
├── migrations/                  # Database migrations
├── sqlc/
│   └── queries/
│       └── queries.sql          # SQL queries for sqlc
├── scripts/
│   └── generate_keys.sh         # RSA key generation script
├── keys/                        # RSA key pairs (generated)
├── Makefile
├── docker-compose.yml
├── Dockerfile
├── sqlc.yaml
└── .env.example
```

## Architecture

### Layered Architecture

```
HTTP Layer (Handlers)
       ↓
Service Layer (Business Logic)
       ↓
Repository Layer (Data Access)
       ↓
Database (PostgreSQL/Redis)
```

- **Handlers** - Parse requests, validate input, call services, format responses
- **Services** - Business logic, orchestration, domain rules
- **Repositories** - Database operations, data persistence

### Domain-Driven Structure

Each domain (auth, user, note) has its own:
- Handler file(s)
- Service file
- Repository file
- Domain model(s)

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/register` | Register new user |
| POST | `/api/v1/login` | Login with email/phone |
| POST | `/api/v1/refresh` | Refresh access token |
| POST | `/api/v1/logout` | Logout (revoke refresh token) |

### Authenticated Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/logout-all` | Logout from all devices |
| GET | `/api/v1/notes` | List user's notes |
| POST | `/api/v1/notes` | Create note |
| GET | `/api/v1/notes/{id}` | Get note by ID |
| PUT | `/api/v1/notes/{id}` | Update note |
| DELETE | `/api/v1/notes/{id}` | Soft delete note |
| POST | `/api/v1/notes/{id}/restore` | Restore deleted note |

### Admin Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users` | List all users |
| DELETE | `/api/v1/users/{id}` | Delete user |

## Authentication Flow

### Registration

```bash
# Register with email
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Register with phone
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"phone": "+1234567890", "password": "password123"}'
```

### Login

```bash
# Login with email or phone
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"identifier": "user@example.com", "password": "password123"}'
```

### Using Access Token

```bash
curl http://localhost:8080/api/v1/notes \
  -H "Authorization: Bearer <access_token>"
```

### Refreshing Tokens

```bash
curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token>"}'
```

## Configuration

Environment variables (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment | `development` |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `REDIS_URL` | Redis connection string | Required |
| `JWT_ACCESS_PRIVATE_KEY_PATH` | Access token private key path | `keys/access_private.pem` |
| `JWT_ACCESS_PUBLIC_KEY_PATH` | Access token public key path | `keys/access_public.pem` |
| `JWT_REFRESH_PRIVATE_KEY_PATH` | Refresh token private key path | `keys/refresh_private.pem` |
| `JWT_REFRESH_PUBLIC_KEY_PATH` | Refresh token public key path | `keys/refresh_public.pem` |
| `JWT_ACCESS_TTL` | Access token validity | `15m` |
| `JWT_REFRESH_TTL` | Refresh token validity | `168h` (7 days) |
| `ARGON2_MEMORY` | Argon2 memory (KB) | `65536` |
| `ARGON2_ITERATIONS` | Argon2 iterations | `3` |
| `ARGON2_PARALLELISM` | Argon2 parallelism | `2` |
| `RATE_LIMIT_REQUESTS` | Rate limit requests per window | `100` |
| `RATE_LIMIT_WINDOW` | Rate limit window | `1m` |

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255),
    phone VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_or_phone CHECK (email IS NOT NULL OR phone IS NOT NULL)
);
```

### Notes Table

```sql
CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

## Development

### Available Make Commands

```bash
make build              # Build the binary
make run                 # Run the application
make test                # Run tests
make test-coverage       # Run tests with coverage report
make lint                # Run golangci-lint
make fmt                 # Format code
make vet                 # Run go vet
make deps                # Download dependencies
make sqlc-generate       # Generate sqlc code
make generate-keys       # Generate RSA key pairs
make migrate-up          # Run database migrations
make migrate-down        # Rollback migrations
make migrate-status      # Check migration status
make docker-up           # Start Docker services
make docker-down         # Stop Docker services
make docker-logs         # View Docker logs
make docker-build        # Build Docker image
make dev                 # Start services and run app
```

### Adding a New Migration

```bash
make migrate-create
# Enter migration name when prompted
```

### Generating sqlc Code

After modifying `sqlc/queries/queries.sql`:

```bash
make sqlc-generate
```

### Generating RSA Keys

```bash
make generate-keys
# Or manually:
./scripts/generate_keys.sh keys
```

## Request Validation

The validation package uses `go-playground/validator` with custom validators:

### Validation Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field is required | `validate:"required"` |
| `email` | Valid email format | `validate:"email"` |
| `phone` | Valid E.164 phone | `validate:"phone"` |
| `min` | Minimum length | `validate:"min=8"` |
| `max` | Maximum length | `validate:"max=255"` |
| `omitempty` | Skip if empty | `validate:"omitempty,email"` |

### Validation Error Response

```json
[
  {
    "field": "Email",
    "message": "Email must be a valid email address"
  },
  {
    "field": "Password",
    "message": "Password must be at least 8 characters"
  }
]
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "Bad Request",
  "code": "VALIDATION_ERROR",
  "message": "Validation failed"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INTERNAL_ERROR` | 500 | Server error |
| `NOT_FOUND` | 404 | Resource not found |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `BAD_REQUEST` | 400 | Invalid request |
| `CONFLICT` | 409 | Resource conflict |
| `VALIDATION_ERROR` | 400 | Validation failed |
| `RATE_LIMITED` | 429 | Too many requests |
| `INVALID_TOKEN` | 401 | Invalid/expired token |
| `INVALID_CREDENTIALS` | 401 | Wrong email/phone or password |
| `EMAIL_EXISTS` | 409 | Email already registered |

## Rate Limiting

- Redis-backed distributed rate limiting
- Rate limit headers in response:
  - `X-RateLimit-Limit` - Maximum requests per window
  - `X-RateLimit-Remaining` - Remaining requests
  - `X-RateLimit-Reset` - Seconds until window resets

## Deployment

### Docker

```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f app
```

### Manual Deployment

1. Build the binary:
   ```bash
   make build
   ```

2. Generate keys:
   ```bash
   make generate-keys
   ```

3. Set environment variables

4. Run migrations:
   ```bash
   make migrate-up
   ```

5. Run the binary:
   ```bash
   ./bin/api
   ```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

## Security Considerations

- **RSA JWT Signing** - Uses asymmetric encryption (RS256)
- **Argon2id Password Hashing** - Memory-hard algorithm resistant to GPU attacks
- **Refresh Token Rotation** - Old refresh tokens are revoked on use
- **Rate Limiting** - Prevents brute force attacks
- **CORS** - Configurable allowed origins
- **Soft Delete** - Notes can be recovered if deleted accidentally

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linters
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.