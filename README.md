# go-auth-template

Production-ready authentication API template built with **Gin**, **GORM**, and **PostgreSQL**. Ships with Docker Compose, automated curl tests, and GitHub Actions CI.

A high-performance Go starting point for any backend that needs user authentication, RBAC, and email flows.

## Features

- **Core Auth** — Register, login, logout, token refresh with JWT (HS256)
- **Email Verification** — Verify-email flow with expiring tokens via SMTP
- **Password Reset** — Forgot-password / reset-password with one-time tokens
- **RBAC** — Role-based access control with permissions, user management, and admin endpoints
- **Docker** — Single `docker compose up` to spin up app + PostgreSQL
- **CI** — GitHub Actions workflow that builds Docker, runs curl tests, and reports status
- **Structured Logging** — `slog` for JSON-formatted logs
- **Configuration** — Viper for environment-based config

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL 16 |
| Migrations | golang-migrate |
| Auth | golang-jwt/v5 + bcrypt |
| Validation | go-playground/validator |
| Config | Viper |
| Logging | slog |
| Email | net/smtp (MailHog for dev) |
| Container | Docker Compose |

## Quick Start

```bash
# Clone
git clone https://github.com/vidwadeseram/go-auth-template.git
cd go-auth-template

# Configure
cp .env.example .env
# Edit .env — set JWT_SECRET for production

# Launch
docker compose up --build

# API available at http://localhost:8000
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `db` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `authdb` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | `change-me-in-production` | HMAC secret for JWT signing |
| `JWT_ACCESS_EXPIRE_MINUTES` | `15` | Access token lifetime |
| `JWT_REFRESH_EXPIRE_DAYS` | `7` | Refresh token lifetime |
| `SMTP_HOST` | `mailhog` | SMTP server hostname |
| `SMTP_PORT` | `1025` | SMTP server port |
| `APP_PORT` | `8000` | Application port |
| `MAIL_FROM` | `no-reply@go-auth-template.local` | Sender email address |

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/v1/auth/register` | Register a new user | No |
| `POST` | `/api/v1/auth/login` | Login with email + password | No |
| `POST` | `/api/v1/auth/refresh` | Refresh access token | No (send refresh token) |
| `POST` | `/api/v1/auth/logout` | Logout (invalidates refresh token) | Yes |
| `GET` | `/api/v1/auth/me` | Get current user profile | Yes |
| `POST` | `/api/v1/auth/verify-email` | Verify email with token | No |
| `POST` | `/api/v1/auth/forgot-password` | Request password reset email | No |
| `POST` | `/api/v1/auth/reset-password` | Reset password with token | No |

### Admin & RBAC

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `GET` | `/api/v1/admin/roles` | List all roles | `super_admin` |
| `GET` | `/api/v1/admin/users` | List all users with roles | `super_admin` |
| `POST` | `/api/v1/admin/users/{id}/roles` | Assign role to user | `super_admin` |
| `DELETE` | `/api/v1/admin/users/{id}/roles` | Remove role from user | `super_admin` |
| `POST` | `/api/v1/admin/roles/{id}/permissions` | Assign permission to role | `super_admin` |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |

## Project Structure

```
go-auth-template/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── config/              # Viper configuration
│   ├── database/            # GORM connection + migrations
│   ├── handlers/
│   │   ├── auth.go          # Auth endpoints
│   │   └── admin.go         # Admin/RBAC endpoints
│   ├── middleware/           # JWT auth middleware
│   ├── models/              # GORM models (User, Role, Permission…)
│   ├── repositories/        # Data access layer
│   ├── services/            # Business logic layer
│   └── utils/               # JWT, email, password helpers
├── tests/
│   └── test_api.sh          # Curl-based integration tests
├── docker/
│   └── Dockerfile
├── docker-compose.yml
├── .github/workflows/ci.yml
├── .env.example
├── go.mod
└── README.md
```

## Testing

```bash
# Run curl test suite against running instance
bash tests/test_api.sh http://localhost:8000/api/v1
```

The test script covers:
- User registration and login
- Token refresh and logout
- Email verification flow
- Password reset flow
- RBAC (403 without role, 200 with `super_admin`)
- Admin user/role management

## Response Format

All responses follow a consistent structure:

```json
// Success
{
  "data": {
    "user": { "id": "...", "email": "..." },
    "access_token": "...",
    "refresh_token": "..."
  }
}

// Error
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired token."
  }
}
```

## License

MIT
