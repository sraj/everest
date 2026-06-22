# Everest Docs

A Google Docs-like document management application with real-time document editing, thumbnail previews, and cloud storage.

![](screenshot-1.png)![alt text](screenshot-2.png)

## Features

- Rich text document editor with TipTap
- Automatic thumbnail generation (Google Docs-style preview images)
- Document storage in MinIO (S3-compatible object storage)
- PostgreSQL for metadata and user management
- Responsive document grid with 3:4 aspect ratio previews

## Tech Stack

**Backend:**
- Go with Fiber web framework
- PostgreSQL for database
- MinIO for object storage
- Chromedp for headless Chrome thumbnail generation

**Frontend:**
- React 19 with TypeScript
- Vite for build tooling
- Tailwind CSS v4
- TipTap rich text editor
- Redux Toolkit for state management

## Quick Start

### Option 1: Full Docker Stack

Single command:

```bash
make docker-up
```

This starts all services:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (admin/minioadmin)
- **PostgreSQL**: localhost:5432

After services are up, apply migrations and seed:

```bash
make migrate-up
make seed
```

To stop:

```bash
make docker-down
```

To stop and remove all data:

```bash
make docker-down -v
```

### Option 2: Local Development (Hot Reload)

#### Prerequisites

- Go 1.26+
- Node.js 22+
- PostgreSQL 16
- MinIO (or S3-compatible storage)
- Chrome/Chromium (for thumbnail generation)

#### Setup

1. **Install Go development tools:**
   ```bash
   make install-tools
   ```

2. **Copy environment file:**
   ```bash
   cp .env.example .env
   ```

3. **Start infrastructure services only:**
   ```bash
   make docker-infra
   ```

4. **Run database migrations:**
   ```bash
   make migrate-up
   ```

5. **Seed the database:**
   ```bash
   make seed
   ```

6. **Start the backend (with hot reload):**
   ```bash
   make dev
   ```

7. **In another terminal, start the frontend:**
   ```bash
   make web-install
   make web-dev
   ```

## Database Operations

### Migrations

```bash
# Apply all pending migrations
make migrate-up

# Roll back all migrations
make migrate-down

# Show current migration version
make migrate-version

# Create a new migration
make migrate-create
# Then enter the migration name when prompted
```

### Seeding

```bash
# Run all pending seeds
make seed

# Show seed status
make seed-status

# Reset seed tracking (re-run all seeds)
make seed-reset
```

### Combined Operations

```bash
# Run migrations + seeds
make setup

# Fresh start (drop all, migrate, seed)
make fresh
```

## Makefile Commands

Run `make help` for a full list of available commands:

### Backend
| Command | Description |
|---------|-------------|
| `make build` | Build the server application |
| `make build-migrate` | Build the migrate CLI |
| `make run` | Build and run the application |
| `make dev` | Run with hot reload (requires air) |
| `make test` | Run tests |
| `make test-coverage` | Run tests with coverage report |
| `make lint` | Run linter |
| `make fmt` | Format code |
| `make tidy` | Tidy dependencies |

### Database
| Command | Description |
|---------|-------------|
| `make docker-up` | Start all services in Docker (full stack with `--build`) |
| `make docker-infra` | Start infrastructure only (PostgreSQL, MinIO) |
| `make docker-down` | Stop Docker services |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back all migrations |
| `make migrate-version` | Show current migration version |
| `make migrate-create` | Create a new migration file |
| `make seed` | Run all pending seeds |
| `make seed-status` | Show seed status |
| `make seed-reset` | Reset seed tracking |
| `make setup` | Run migrations + seeds |
| `make fresh` | Drop, migrate, and seed (fresh start) |

### Frontend
| Command | Description |
|---------|-------------|
| `make web-install` | Install frontend dependencies |
| `make web-dev` | Start frontend dev server |
| `make web-build` | Build frontend for production |

### Other
| Command | Description |
|---------|-------------|
| `make clean` | Clean build artifacts |
| `make install-tools` | Install development tools (air, golangci-lint, migrate) |

## Project Structure

```
everest/
├── .air.toml                 # Hot-reload configuration (air)
├── cmd/
│   ├── server/               # Main application entrypoint
│   └── migrate/              # Migration CLI
├── internal/
│   ├── config/               # App-specific configuration (uses configx)
│   ├── domain/
│   │   ├── model/            # Domain models (Document, User) + pagination types
│   │   └── repository/       # Repository interfaces (split by concern)
│   ├── handler/
│   │   ├── http/             # Fiber HTTP handlers
│   │   └── grpc/             # gRPC service stubs (optional)
│   ├── infrastructure/
│   │   ├── minio/            # MinIO content repository
│   │   └── postgres/         # PostgreSQL repositories
│   └── service/              # Business logic with interfaces
├── migrations/               # SQL migration files
├── seeds/                    # Database seed files
├── pkg/
│   ├── configx/              # Generic env config loader
│   ├── dbx/                  # PostgreSQL query builder (CTEs, pagination, tx)
│   ├── logger/               # slog handler backed by zerolog
│   └── server/               # HTTP/gRPC server lifecycle
├── web/                      # React frontend
│   ├── src/
│   │   ├── pages/            # Page components (Home, DocumentEditor)
│   │   ├── components/       # Reusable components (Editor, toolbar)
│   │   └── store/            # Redux store (documents slice)
│   └── Dockerfile            # Frontend Docker image
├── docker-compose.yml        # Services: postgres, minio, backend, frontend
├── Dockerfile                # Backend Docker image (with Chromium)
└── Makefile                  # Build and dev commands
```

## Environment Variables

See `.env.example` for all available configuration options:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Application name | `everest` |
| `PORT` | HTTP server port (empty to disable) | `8080` |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `CORS_ORIGINS` | Allowed CORS origins | `*` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `MINIO_ENDPOINT` | MinIO endpoint | `localhost:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` |
| `MINIO_SECRET_KEY` | MinIO secret key | `minioadmin` |
| `MINIO_BUCKET` | MinIO bucket name | `documents` |
| `MINIO_USE_SSL` | Use SSL for MinIO | `false` |
| `GRPC_PORT` | gRPC server port (empty to disable) | `""` |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/v1/documents?page=1&size=20` | List documents (paginated) |
| `POST` | `/api/v1/documents` | Create a document |
| `GET` | `/api/v1/documents/:id` | Get a document |
| `PUT` | `/api/v1/documents/:id` | Update a document |
| `DELETE` | `/api/v1/documents/:id` | Delete a document |
| `GET` | `/api/v1/documents/:id/download` | Download document content |
| `GET` | `/api/v1/documents/:id/thumbnail` | Get document thumbnail |

Pagination: list endpoint supports `?page=` (default 1) and `?size=` (default 20, max 100). Returns `total`, `page`, `page_size`, `total_pages` metadata.

## License

MIT
