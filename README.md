<div align="center">

<img src="https://go.dev/images/gophers/motorcycle.svg" width="160" alt="Go Gopher" />

# go-service-starter-kit

**A production-ready Go backend template for services that do more than serve HTTP.**

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat)](LICENSE)
[![Build](https://img.shields.io/badge/build-passing-brightgreen?style=flat)](#)

*One binary. Multiple processes. Zero boilerplate.*

</div>

---

## Why this template exists

Most Go service templates stop at "here is how to start an HTTP server." Real production services are more complex: they run **background jobs**, consume **Kafka or RabbitMQ messages**, expose **gRPC endpoints** for internal traffic, and handle **one-off data migrations** — all sharing the same business logic and infrastructure.

This template was extracted from a real production service. The goal is to give you a solid, opinionated starting point that:

- Ships **all process types** (HTTP, gRPC, consumer, jobs) from a **single binary**, selected at runtime
- Enforces a **clean-layered architecture** — domain → business → endpoint → transport — so your code stays testable as it grows
- Provides **JWT authentication** baked in for HTTP, gRPC, and message consumers out of the box
- Wires **OpenTelemetry** tracing, structured logging, and metrics from day one, so observability is never an afterthought
- Keeps infrastructure concerns in `internal/pkg/` and your domain in `internal/app/`, making the boundary between _your code_ and _plumbing_ explicit

> The template is intentionally opinionated. It picks go-kit, gorilla/mux, sqlx, zerolog, and viper — a stack that has been proven at scale. You are free to swap any layer out.

---

## What's inside

```
go-service-starter-kit/
│
├── cmd/
│   ├── main.go               # Entry point — flag parsing, process dispatch
│   └── app/
│       ├── registry.go       # Long-running process registry
│       ├── http.go           # HTTP process wiring
│       ├── grpc.go           # gRPC process wiring
│       ├── consumer.go       # Message consumer wiring
│       └── job.go            # One-time job registry
│
├── internal/
│   ├── pkg/                  # Shared infrastructure — reuse across any domain
│   │   ├── auth/             # JWT (ClaimsParser) + bcrypt password hashing
│   │   ├── clients/db/       # OTel-traced PostgreSQL connection pool
│   │   ├── config/           # Viper loader — YAML + .env + env vars
│   │   ├── db/               # Pagination & safe SQL ordering helpers
│   │   ├── gokit/http/       # go-kit HTTP handler factory
│   │   ├── middleware/
│   │   │   ├── http.go       # JWT HTTP middleware (AuthRequired / AuthOptional / AuthMock)
│   │   │   ├── gokit.go      # Timeout + sliding-window rate limiter
│   │   │   ├── logging.go    # Transport-aware logging with sensitive-field masking
│   │   │   └── consumer.go   # JWT consumer middleware (works with any message broker)
│   │   ├── observability/
│   │   │   ├── logger/       # zerolog-backed structured, context-aware logger
│   │   │   ├── tracing/      # OTel trace provider (OTLP gRPC exporter)
│   │   │   └── metric/       # OTel metric interfaces + Reporter
│   │   ├── text/             # NonLoggable — redacts sensitive strings from logs and JSON
│   │   └── worker/
│   │       ├── http.go       # HTTP worker — graceful shutdown on SIGINT/SIGTERM
│   │       ├── grpc.go       # gRPC worker — graceful shutdown
│   │       └── consumer.go   # Consumer worker — graceful shutdown
│   │
│   └── app/                  # Your domain code lives here
│       ├── domain/           # Entities + sentinel errors
│       ├── business/         # Use cases — one file per operation
│       ├── repositories/     # Data access layer (sqlx + PostgreSQL)
│       ├── endpoint/v1/      # go-kit endpoint adapters
│       └── transport/
│           ├── http/         # HTTP — server, DI bootstrap, encode/decode, error encoder
│           ├── grpc/         # gRPC — server stub with logging + tracing interceptors
│           └── consumer/     # Consumer — broker-agnostic stub with JWT auth wiring
│
├── db/migrations/            # SQL migration files (golang-migrate)
├── docs/
│   ├── openapi/              # OpenAPI 3.1 spec + oapi-codegen config
│   ├── proto/                # proto3 service definitions
│   └── otel/                 # OTel collector config for local dev
├── docker-compose.yml        # postgres · kafka · otel-collector profiles
├── config.yaml               # Default configuration
├── .env.sample               # Environment variable template
├── Dockerfile                # Multi-stage build (builder → minimal runner)
└── Makefile                  # Developer commands
```

---

## Architecture overview

The design follows a strict dependency flow — outer layers depend on inner layers, never the reverse:

```
Request
   │
   ▼
Transport          (HTTP / gRPC / Consumer)
   │  encode / decode
   ▼
Endpoint           (go-kit adapter — applies middleware: auth, timeout, rate-limit, logging)
   │  typed request
   ▼
Business           (use-case handler — pure Go, no framework dependency)
   │  repository interface
   ▼
Repository         (sqlx + PostgreSQL — implements the interface)
   │
   ▼
Database
```

Each layer communicates through **interfaces**, which means every layer can be unit-tested in isolation with mocks — no database, no HTTP server required.

---

## Getting started

### Prerequisites

- Go 1.23+
- Docker + Docker Compose

### 1. Clone and configure

```bash
git clone https://github.com/yourorg/go-service-starter-kit.git
cd go-service-starter-kit

# Copy sample env and fill in your values
make env
```

### 2. Start the database

```bash
docker compose up postgres -d
docker compose up migrate      # runs all pending migrations
```

### 3. Build and run

```bash
make build

./bin/app http       # HTTP server  (default :8080)
./bin/app grpc       # gRPC server  (default :50051)
./bin/app consumer   # message consumer
./bin/app <job>      # one-time job
```

---

## Configuration

Configuration is merged from three sources (highest priority first):

| Priority | Source | Format |
|---|---|---|
| 1 (highest) | `.env` file | `KEY__NESTED=value` |
| 2 | OS environment variables | `KEY__NESTED=value` |
| 3 (lowest) | `config.yaml` | YAML |

The `__` double-underscore is the struct delimiter — `DB__DSN` maps to `Config.DB.DSN`.

**Required variables:**

| Variable | Description |
|---|---|
| `DB__DSN` | PostgreSQL connection string |
| `JWT__SECRET` | HMAC-SHA256 signing secret |
| `HTTP__PORT` | HTTP listen port |

---

## Extending the template

### Rename the domain

Replace `internal/app/` with your service name (e.g. `internal/orders/`) and update the import paths. The `internal/pkg/` packages stay unchanged.

### Add a new use case

```
internal/<domain>/
  domain/            ← add your entity + any new sentinel errors
  repositories/<x>/  ← add your SQL queries
  business/<op>/     ← add your handler (CreateXxx, UpdateXxx, …)
  endpoint/v1/       ← add your go-kit endpoint adapter
  transport/http/v1/ ← add your encode/decode codec
```

Then wire it in `internal/app/transport/http/bootstrap/`:
1. `handler_initializer.go` — instantiate the handler
2. `router_v1_register.go` — mount the route

### Add a new process

```go
// cmd/app/my_process.go
type MyProcess struct{}
func (MyProcess) Register(args ProcessArgs) (Process, error) { ... }

// cmd/app/registry.go
var RegistryProcessesMap = map[string]ProcessRegistry{
    "http":       NewHTTPServerProcess(),
    "my-process": MyProcess{},  // ← add here
}
```

### Add a one-time job

```go
// internal/app/jobs/my_job.go
type MyJob struct{}
func (MyJob) Schedule(args app.ProcessArgs) error { ... }

// cmd/app/job.go
var JobsMap = map[string]Scheduler{
    "my-job": MyJob{},
}
```

Run it with: `./bin/app my-job`

### Consumer JWT authentication

Any message that carries a user token in its headers is authenticated before your handler runs:

```go
// Required — message must carry a valid JWT
ctx, err = middleware.ConsumerAuthRequired(claimsParser)(ctx, msg.Headers["Authorization"])

// Optional — unauthenticated messages are allowed through
ctx, err = middleware.ConsumerAuthOptional(claimsParser)(ctx, msg.Headers["Authorization"])

// Retrieve the authenticated user anywhere downstream
user := auth.UserFromCtx(ctx)
```

---

## Key dependencies

| Package | Role |
|---|---|
| [`go-kit/kit`](https://github.com/go-kit/kit) | Endpoint / transport abstraction, middleware chaining |
| [`gorilla/mux`](https://github.com/gorilla/mux) | HTTP routing |
| [`jmoiron/sqlx`](https://github.com/jmoiron/sqlx) | PostgreSQL — ergonomic `database/sql` wrapper |
| [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) | Schema migrations |
| [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt) | JWT generation and validation |
| [`rs/zerolog`](https://github.com/rs/zerolog) | Zero-allocation structured logging |
| [`go.opentelemetry.io/otel`](https://opentelemetry.io) | Distributed tracing, metrics (OTLP gRPC) |
| [`spf13/viper`](https://github.com/spf13/viper) | Configuration loading |
| [`go.uber.org/mock`](https://github.com/uber-go/mock) | Mock generation |
| [`google.golang.org/grpc`](https://grpc.io) | gRPC server and client |

---

## Testing strategy

Tests are organised around the layer they cover:

| Layer | What to test | How |
|---|---|---|
| **Business** | Use-case logic — happy path, error cases, edge cases | Unit test with mocked repository interfaces (`go.uber.org/mock`) |
| **Repository** | SQL queries — correct results, constraint violations, not-found | Integration test against a real PostgreSQL instance (test container or Docker Compose) |
| **Endpoint** | Middleware behaviour — timeout fires, rate limit triggers, auth rejects | Unit test by calling the endpoint directly |
| **Transport** | Encode/decode — valid body accepted, bad body rejected with 400 | Unit test using `httptest.NewRecorder` |

`internal/pkg/` packages ship with their own unit tests (see `_test.go` files).

Use build tags to separate unit and integration tests:

```bash
make test              # unit tests only  (-tags unit)
make test-integration  # integration tests (-tags integration)
```

---

## Contributing

Contributions, bug reports, and feature requests are very welcome.

**To contribute:**

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-improvement`
3. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org)
4. Open a pull request with a clear description of the problem and solution

**Not sure where to start?** Open an [issue](https://github.com/yourorg/go-service-starter-kit/issues) first to discuss the idea — that is always appreciated before writing code.

**Feedback** on the architecture, tooling choices, or documentation is equally welcome. If something feels wrong or overly complex, please say so.

---

## License

MIT — see [LICENSE](LICENSE).

---

<div align="center">

Built with the [Go gopher](https://go.dev/blog/gopher) and real production pain.

*If this template saved you time, consider leaving a ⭐ on GitHub.*

</div>
