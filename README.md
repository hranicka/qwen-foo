# foo

A minimal HTTP REST API with a PostgreSQL-backed counter, written in Go.

![Gopher on bicycle](gopher-on-bicycle.jpg)

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.26 |
| Database | PostgreSQL 17 |
| HTTP | net/http (stdlib) |
| DB driver | jackc/pgx/v5 |
| Migrations | golang-migrate |
| Config | YAML + env var substitution |
| Build | make + Docker |

## Business Logic

The application exposes a single endpoint:

- **GET `/hello`** — Increments a counter in PostgreSQL and returns the current value:
  ```json
  {"message": "Hello", "counter": 1}
  ```

The counter is stored in a `counter` table (id=1, value auto-increments on each `/hello` request).

## Configuration

Configuration is loaded from `config.yaml` with env var substitution via `${VAR_NAME}` syntax:

```yaml
database:
  url: "${DATABASE_URL}"
migration:
  dir: "migrations"
http:
  addr: ":8080"
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `CONFIG_PATH` | No | Path to config.yaml (default: `config.yaml`) |

### .env file

Place a `.env` file in the project root for local development:

```bash
DATABASE_URL=postgres://app:secret@localhost:5432/app?sslmode=disable
```

## Makefile Commands

All operations use `make` commands.

### Development

```bash
# Start services (background)
make up

# Start services (foreground, with hot reload)
make dev

# Stop services
make down
```

### Testing

```bash
# Run integration tests
make test
```

### Migrations

```bash
# Apply migrations
make migrate-up

# Roll back last migration
make migrate-down
```

### Build

```bash
# Build binary
make build

# Run go vet
make vet

# Clean build artifacts
make clean
```

## Testing the API

Start the services, then curl the `/hello` endpoint:

```bash
make up
curl http://localhost:8080/hello
# {"message":"Hello","counter":1}

curl http://localhost:8080/hello
# {"message":"Hello","counter":2}
```

## Project Structure

```
├── cmd/migrate/          # Migration CLI tool
├── config/               # Config loading package
├── internal/
│   ├── db/               # PostgreSQL connection + counter logic
│   └── server/           # HTTP handlers
├── migrations/           # SQL migration files
├── tests/                # Integration tests
├── openapi/              # OpenAPI spec (documentation)
└── compose.yml           # Docker Compose services
```
