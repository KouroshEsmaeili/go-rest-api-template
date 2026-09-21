# Go REST API Template

A compact, reusable starter for JSON REST APIs built with Go, Gin, and PostgreSQL. It demonstrates the infrastructure and engineering practices most APIs need while leaving product-specific decisions to the application that adopts it.

This repository is a starter/template, not a finished SaaS application or a claim of an all-purpose architecture.

## Features

- Versioned REST endpoints with one example `posts` resource
- PostgreSQL access through `database/sql` and pgx
- Explicit, reversible SQL migrations
- Typed environment configuration with development defaults
- Consistent JSON errors and explicit input validation
- Liveness and database-backed readiness checks
- Request size limits and HTTP server timeouts
- Graceful shutdown on `SIGINT` and `SIGTERM`
- Focused HTTP and validation tests
- Multi-stage, non-root container image
- Docker Compose development stack
- Formatting, test, vet, and build checks in GitHub Actions

## Architecture

The entry point composes concrete dependencies and passes them to the HTTP layer. The API depends on a small store contract so handlers remain easy to test; PostgreSQL is the production implementation. There is no dependency-injection framework, global database handle, or automatic schema mutation at startup.

```text
HTTP request
    -> Gin router and handlers
    -> posts.Store
    -> PostgreSQL (database/sql + pgx)
```

## Repository structure

```text
.
├── .github/workflows/ci.yml  # Continuous integration
├── cmd/api/main.go           # Process composition and lifecycle
├── internal/
│   ├── config/               # Typed environment configuration
│   ├── database/             # PostgreSQL pool setup
│   ├── health/               # Liveness and readiness handlers
│   ├── httpapi/              # Router, handlers, and API errors
│   └── posts/                # Example domain and PostgreSQL store
├── migrations/               # Versioned up/down SQL migrations
├── Dockerfile
└── docker-compose.yml
```

## Requirements

- Go 1.27 or newer
- PostgreSQL 14 or newer
- Docker with Docker Compose (optional, recommended for the quickest start)
- [`golang-migrate`](https://github.com/golang-migrate/migrate) when running migrations without Compose

## Quick start with Docker Compose

Start PostgreSQL, apply migrations, build the API image, and run the server:

```sh
docker compose up --build
```

Then verify it:

```sh
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

Compose stores database data in the `postgres-data` named volume. To stop the services, run `docker compose down`. Add `--volumes` only when you intentionally want to delete local database data.

## Local development

Copy the example configuration if you want a local environment file for reference:

```sh
cp .env.example .env
```

The application reads process environment variables directly; it does not load `.env` files. Export the values using your shell or an environment-loading tool, start PostgreSQL, apply the migration, and run:

```sh
go run ./cmd/api
```

With the default values, PostgreSQL is expected at `localhost:5432` with the username, password, and database all set to `postgres`.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_ENV` | `development` | Set to `production` to enable Gin release mode. |
| `PORT` | `8080` | HTTP listen port, from 1 through 65535. |
| `DATABASE_URL` | Local PostgreSQL URL | PostgreSQL connection string. Required explicitly when `APP_ENV=production`. |

Never commit `.env` or production credentials. The example values are only for local development.

## Database migrations

Migrations use the `golang-migrate` filename convention and are never run implicitly by the API. Compose runs all pending migrations before starting the API.

With Compose services available:

```sh
docker compose run --rm migrate
```

With a local `migrate` executable:

```sh
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" down 1
```

Create each schema change as a matching `.up.sql` and `.down.sql` pair. Review down migrations before using them against valuable data.

## API

| Method | Path | Success | Purpose |
| --- | --- | --- | --- |
| `GET` | `/health` | `200` | Process liveness |
| `GET` | `/ready` | `200` | PostgreSQL readiness |
| `GET` | `/api/v1/posts` | `200` | List posts |
| `GET` | `/api/v1/posts/:id` | `200` | Retrieve a post |
| `POST` | `/api/v1/posts` | `201` | Create a post |
| `PUT` | `/api/v1/posts/:id` | `200` | Replace a post's editable fields |
| `DELETE` | `/api/v1/posts/:id` | `204` | Delete a post |

Create or update requests accept:

```json
{
  "title": "A short required title",
  "content": "Optional content"
}
```

Errors use a stable envelope and do not expose database details:

```json
{
  "error": {
    "code": "validation_error",
    "message": "title is required"
  }
}
```

## Tests and quality checks

The handler suite exercises health behavior, malformed JSON, validation, the complete post lifecycle, missing records, invalid IDs, and database-error redaction. It does not require a live database.

```sh
gofmt -w .
go test ./...
go vet ./...
go build ./...
```

GitHub Actions runs formatting checks, race-enabled tests, vet, and build for every push and pull request.

## Using this repository as a template

1. Create a repository from the GitHub template or clone it without preserving this project's history.
2. Change the module path in `go.mod` and update imports with `go mod edit -module=your/module/path` followed by `go mod tidy`.
3. Replace `posts` with one feature from your own domain, preserving the route/domain/store boundary where it remains useful.
4. Add migrations for your schema and configure deployment secrets through your hosting platform.
5. Extend the tests alongside each endpoint and storage behavior.

## Deliberate non-goals

This starter intentionally does not include authentication, authorization, an ORM, background jobs, caching, code generation, a dependency-injection framework, or deployment manifests. Add those only when the application has a concrete requirement for them.

## License

MIT
