<div align="center">

<img src="docs/img/golang.svg" width="160" alt="Go Gopher" />

# go-service-starter-kit

**A production-ready Go backend template for services that do more than serve HTTP.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat)](LICENSE)
[![CI](https://github.com/nawafswe/go-service-starter-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/nawafswe/go-service-starter-kit/actions/workflows/ci.yml)

*One binary. Multiple processes. Zero boilerplate.*

</div>

---

## Why this template exists

Most Go service templates stop at "here is how to start an HTTP server." Real production services are more complex: they run **background jobs**, consume **Kafka or RabbitMQ messages**, expose **gRPC endpoints** for internal traffic, and handle **one-off data migrations** — all sharing the same business logic and infrastructure.

This template distils patterns and lessons from multiple production Go services. Rather than reflecting a single codebase, it combines the architectural decisions that proved scalable across different domains and team sizes into one reusable starting point that:

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
├── api/                      # Proto module (github.com/nawafswe/go-service-starter-kit/api)
│   ├── go.mod                # Nested Go module — import with: go get …/api@v1.0.0
│   └── proto/grpc/v1/
│       ├── example.proto     # proto3 service definitions
│       └── gen/              # Generated Go stubs (pb.go + grpc.pb.go)
│
├── internal/
│   ├── pkg/                  # Shared infrastructure — reuse across any domain
│   │   ├── auth/             # JWT (ClaimsParser) + bcrypt password hashing
│   │   ├── clients/
│   │   │   └── db/
│   │   │       ├── postgres/ # OTel-traced PostgreSQL pool (sqlx + lib/pq)
│   │   │       ├── mysql/    # OTel-traced MySQL pool (sqlx + go-sql-driver)
│   │   │       └── mongodb/  # OTel-traced MongoDB client (mongo-driver)
│   │   ├── config/           # Viper loader — YAML + .env + env vars
│   │   ├── db/               # Database-agnostic pagination (Page, PageResult)
│   │   │   └── sqlorder/     # SQL ORDER BY builder with column sanitisation
│   │   ├── gokit/
│   │   │   ├── http/         # go-kit HTTP handler factory
│   │   │   ├── grpc/         # go-kit gRPC handler factory
│   │   │   └── consumer/     # go-kit endpoint wrapper for message consumers
│   │   ├── httperrors/       # Reusable HTTP error types (400, 401, 403, 404, 409, 500)
│   │   ├── httpx/            # Resilient HTTP client (retry, circuit breaker, OTel)
│   │   │   └── mock/         # MockDoer for unit tests
│   │   ├── grpcx/            # Resilient gRPC client (retry, circuit breaker, OTel)
│   │   │   └── mock/         # MockInvoker for unit tests
│   │   ├── middleware/
│   │   │   ├── http.go       # JWT HTTP middleware (AuthRequired / AuthOptional / AuthMock)
│   │   │   ├── grpc.go       # JWT gRPC interceptors + logging + tracing interceptors
│   │   │   ├── consumer.go   # JWT consumer middleware (works with any message broker)
│   │   │   ├── gokit.go      # Timeout + sliding-window rate limiter (go-kit)
│   │   │   └── logging.go    # Transport-aware logging with sensitive-field masking
│   │   ├── observability/
│   │   │   ├── logger/       # zerolog-backed structured, context-aware logger
│   │   │   ├── tracing/      # OTel trace provider (OTLP gRPC exporter)
│   │   │   └── metric/       # OTel metric Reporter
│   │   ├── text/             # NonLoggable — redacts sensitive strings from logs and JSON
│   │   └── worker/
│   │       ├── http.go       # HTTP worker — graceful shutdown on SIGINT/SIGTERM
│   │       ├── grpc.go       # gRPC worker — graceful shutdown
│   │       ├── consumer.go   # Consumer worker — graceful shutdown
│   │       └── mock/         # MockMessageConsumer for unit tests
│   │
│   └── app/                  # Your domain code lives here
│       ├── domain/           # Entities + sentinel errors
│       ├── business/         # Use cases — one file per operation
│       ├── repositories/     # Data access layer (sqlx + PostgreSQL)
│       ├── endpoint/v1/      # go-kit endpoint adapters
│       └── transport/
│           ├── http/         # HTTP — server, bootstrap, v1 encode/decode, JSON:API error encoder
│           ├── grpc/         # gRPC — server, bootstrap, v1 handler + encode/decode via go-kit
│           └── consumer/     # Consumer — bootstrap, v1 message decode, map-based endpoint routing
│
├── db/
│   ├── initdb.d/             # PostgreSQL init scripts (user/schema bootstrap)
│   ├── load/                 # Seed / fixture data
│   └── migrations/           # SQL migration files (golang-migrate)
│
├── docs/
│   ├── asyncapi/             # AsyncAPI 3.0 spec — event / message contracts
│   ├── img/                  # Assets used in documentation
│   └── openapi/              # OpenAPI 3.1 spec + oapi-codegen config
│
├── .github/workflows/ci.yml  # CI — build + unit tests + integration tests on every push / PR
├── docker-compose.yml        # postgres, kafka, otel-collector (profile-based)
├── otel-collector-config.yaml # OpenTelemetry Collector configuration
├── config.yaml               # Default configuration
├── .env.sample               # Environment variable template
├── Dockerfile                # Production multi-stage build
├── Dockerfile.dev            # Development build
├── Makefile                  # Developer commands
└── LICENSE
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

- Go 1.24+
- Docker + Docker Compose

### 1. Clone, rename, and configure

```bash
git clone https://github.com/nawafswe/go-service-starter-kit.git my-service
cd my-service

# Rename the Go module to match your repository
go mod edit -module github.com/<your-org>/<your-service>
go mod tidy

# Copy sample env and fill in your values
make env
```

> **Important:** update all import paths after renaming the module. A global search-and-replace of `github.com/nawafswe/go-service-starter-kit` with your new module path covers everything.

### 2. Start the database and run migrations

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

## Makefile commands

| Command | Description |
|---|---|
| `make build` | Build the binary to `./bin/app` |
| `make build-docker` | Build the production Docker image |
| `make run-http` | Build and run the HTTP server |
| `make run-grpc` | Build and run the gRPC server |
| `make run-consumer` | Build and run the message consumer |
| `make env` | Copy `.env.sample` to `.env` if it doesn't exist |
| `make clean` | Remove built binaries |
| `make migrate-up` | Run all pending database migrations |
| `make migrate-create name=<name>` | Create a new migration file |
| `make lint` | Run golangci-lint (via Docker) |
| `make test` | Run unit tests with coverage |
| `make test-integration` | Run integration tests |
| `make fmt` | Format code (gci + gofumpt) |
| `make generate` | Run `go generate` across all packages |
| `make generate-contracts` | Regenerate HTTP types from OpenAPI spec |
| `make docker-start` | Start the Docker environment |
| `make docker-stop` | Stop the Docker environment |
| `make docker-clean` | Remove Docker containers and volumes |
| `make docker-restart` | Restart the Docker environment |

---

## Configuration

Configuration is merged from three sources (highest priority first):

| Priority | Source | Format |
|---|---|---|
| 1 (highest) | OS environment variables | `KEY__NESTED=value` |
| 2 | `.env` file | `KEY__NESTED=value` |
| 3 (lowest) | `config.yaml` | YAML |

The `__` double-underscore is the struct delimiter — `DB__DSN` maps to `Config.DB.DSN`.

**Required variables:**

| Variable | Description |
|---|---|
| `DB__DSN` | PostgreSQL connection string |
| `JWT__SECRET` | HMAC-SHA256 signing secret |
| `HTTP__PORT` | HTTP listen port |

### Endpoint-level configuration

Each endpoint can be individually tuned for timeout and rate limiting in `config.yaml`:

```yaml
ENDPOINTS:
  EXAMPLE_CREATE:
    DEADLINE: 5s               # request timeout
    RATE_LIMITER:
      INTERVAL: 1m             # sliding window duration
      LIMIT: 100               # max requests per window
```

These values are applied as go-kit middleware in each transport's bootstrap layer.

---

## Docker Compose profiles

Services are organised into profiles so you only run what you need:

| Profile | Services | Command |
|---|---|---|
| *(default)* | `postgres`, `migrate` | `docker compose up` |
| `app` | `app-http`, `app-grpc`, `app-consumer` | `docker compose --profile app up` |
| `kafka` | `kafka` (single-node KRaft) | `docker compose --profile kafka up` |
| `observability` | `otel-collector` | `docker compose --profile observability up` |

Combine profiles as needed: `docker compose --profile app --profile kafka --profile observability up`

---

## Extending the template

### Rename the domain

Replace `internal/app/` with your service name (e.g. `internal/orders/`) and update the import paths. The `internal/pkg/` packages stay unchanged.

### Add a new use case

```
internal/<domain>/
  domain/            <- add your entity + any new sentinel errors
  repositories/<x>/  <- add your SQL queries
  business/<op>/     <- add your handler (CreateXxx, UpdateXxx, ...)
  endpoint/v1/       <- add your go-kit endpoint adapter
  transport/http/v1/ <- add your encode/decode codec
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
    "my-process": MyProcess{},  // <- add here
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

### Implement the message consumer

The consumer transport is **broker-agnostic** — it defines a `MessageRouter` that maps message types to go-kit endpoints with decode functions. You provide the broker integration:

```go
// internal/app/transport/consumer/consumer.go — Start()
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: c.cfg.Consumer.Brokers,
    GroupID: c.cfg.Consumer.GroupID,
    Topic:   c.cfg.Consumer.Topics[0],
})
defer reader.Close()

for {
    msg, err := reader.ReadMessage(ctx)
    if err != nil {
        if errors.Is(err, context.Canceled) { return nil }
        return fmt.Errorf("consumer: read message: %w", err)
    }
    if err := c.handleMessage(ctx, "example.create", msg.Headers["Authorization"], msg.Value); err != nil {
        c.lgr.Error(ctx, err, "failed to handle message")
    }
}
```

The `handleMessage` method authenticates the message (JWT), looks up the handler by message type from the router map, decodes the payload, and calls the go-kit endpoint — the same middleware chain (timeout, rate-limit, logging) applies as HTTP and gRPC.

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

## Resilient clients

`internal/pkg/httpx` and `internal/pkg/grpcx` wrap every outbound call with:

- **Retries with exponential back-off + jitter** — avoids thundering herd on transient errors
- **Circuit breaker** (sony/gobreaker) — opens automatically after N consecutive failures, giving downstream services time to recover
- **OTel tracing** — every call starts a client span named after the dependency and RPC method
- **OTel metrics** — `http_client_requests_total` / `grpc_client_requests_total` counters with `dependency`, `circuit_breaker`, `method`, and `result` labels

### Configuration

Both clients are configured under `CLIENTS` in `config.yaml`:

```yaml
CLIENTS:
  HTTP:
    JSON_EXAMPLE:                # key used in code: cfg.Clients.HTTP["JSON_EXAMPLE"]
      NAME: json-example         # used for span names and circuit-breaker naming
      BASE_URL: https://jsonplaceholder.typicode.com
      TIMEOUT: 10s
      MAX_RETRIES: 3
      RETRY_WAIT_MIN: 100ms      # min back-off before first retry
      RETRY_WAIT_MAX: 2s         # back-off is capped at this value
      CIRCUIT_BREAKER:
        MAX_REQUESTS: 5          # max calls in half-open state
        INTERVAL: 30s            # window for counting failures in closed state
        TIMEOUT: 60s             # how long the CB stays open before retrying
        THRESHOLD: 5             # consecutive failures that open the CB (0 = disabled)
  GRPC:
    EXAMPLE:
      NAME: example-grpc
      ADDRESS: localhost:50051
      TIMEOUT: 10s
      MAX_RETRIES: 3
      RETRY_WAIT_MIN: 100ms
      RETRY_WAIT_MAX: 2s
      CIRCUIT_BREAKER:
        MAX_REQUESTS: 5
        INTERVAL: 30s
        TIMEOUT: 60s
        THRESHOLD: 5
```

### Usage

```go
// HTTP client
httpClient, err := httpx.New(
    cfg.Clients.HTTP["JSON_EXAMPLE"],
    &http.Client{Timeout: cfg.Clients.HTTP["JSON_EXAMPLE"].Timeout},
    otelMeter,
    otelTracerProvider,
)
req, _ := http.NewRequest(http.MethodGet, cfg.Clients.HTTP["JSON_EXAMPLE"].BaseURL+"/posts/1", nil)
resp, err := httpClient.Do(ctx, req)

// gRPC client
grpcConn, _ := grpc.NewClient(cfg.Clients.GRPC["EXAMPLE"].Address)
grpcClient, err := grpcx.New(
    cfg.Clients.GRPC["EXAMPLE"],
    grpcConn,
    otelMeter,
    otelTracerProvider,
)
var reply pb.ExampleResponse
err = grpcClient.Invoke(ctx, "/example.v1.ExampleService/GetExample", &req, &reply)
```

### Observability

| Signal | Name | Labels |
|---|---|---|
| Metric | `http_client_requests_total` | `dependency`, `circuit_breaker`, `method`, `result` |
| Metric | `grpc_client_requests_total` | `dependency`, `circuit_breaker`, `method`, `result` |
| Trace | span per call | named after the RPC method, `SpanKindClient` |

---

## Async event contracts

Event and message schemas are documented with **AsyncAPI 3.0** at [`docs/asyncapi/asyncapi.yaml`](docs/asyncapi/asyncapi.yaml).

The spec covers:

| Channel | Direction | Description |
|---|---|---|
| `example.created` | publish | Emitted when a new Example is created |
| `example.deleted` | publish | Emitted when an Example is deleted |
| `example.commands` | subscribe | Inbound commands consumed by the service |

Every message includes a `correlationId` header for distributed tracing and an optional `Authorization` header for JWT-authenticated producers.

Render the spec locally with the [AsyncAPI Studio](https://studio.asyncapi.com) or the CLI:

```bash
npx @asyncapi/cli preview docs/asyncapi/asyncapi.yaml
```

---

## Key dependencies

| Package | Role |
|---|---|
| [`go-kit/kit`](https://github.com/go-kit/kit) | Endpoint / transport abstraction, middleware chaining |
| [`gorilla/mux`](https://github.com/gorilla/mux) | HTTP routing |
| [`jmoiron/sqlx`](https://github.com/jmoiron/sqlx) | PostgreSQL — ergonomic `database/sql` wrapper |
| [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) | Schema migrations (runs via Docker) |
| [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt) | JWT generation and validation |
| [`rs/zerolog`](https://github.com/rs/zerolog) | Zero-allocation structured logging |
| [`go.opentelemetry.io/otel`](https://opentelemetry.io) | Distributed tracing, metrics (OTLP gRPC) |
| [`spf13/viper`](https://github.com/spf13/viper) | Configuration loading |
| [`sony/gobreaker`](https://github.com/sony/gobreaker) | Circuit breaker |
| [`go.uber.org/mock`](https://github.com/uber-go/mock) | Mock generation (`go generate ./...`) |
| [`google.golang.org/grpc`](https://grpc.io) | gRPC server and client |
| [`go.mongodb.org/mongo-driver`](https://github.com/mongodb/mongo-go-driver) | MongoDB client |
| [`stretchr/testify`](https://github.com/stretchr/testify) | Test assertions (`assert`, `require`) |

---

## Testing strategy

Tests are organized around the layer they cover:

| Layer | What to test | How |
|---|---|---|
| **Business** | Use-case logic — happy path, error cases, edge cases | Unit test with mocked repository interfaces (`go.uber.org/mock`) |
| **Repository** | SQL queries — correct results, constraint violations, not-found | Integration test against a real PostgreSQL instance (test container or Docker Compose) |
| **Endpoint** | Middleware behaviour — timeout fires, rate limit triggers, auth rejects | Unit test by calling the endpoint directly |
| **Transport** | Encode/decode — valid body accepted, bad body rejected with 400 | Unit test using `httptest.NewRecorder` |

`internal/pkg/` packages ship with their own unit and integration tests.

**Test conventions:**
- One `_test.go` file per source file (e.g. `jwt_test.go` for `jwt.go`)
- External test package (`package foo_test`) for black-box testing
- Table-driven tests with `tests := []struct{ name string; expectedErr error; ... }{...}` using `testify/assert` and `testify/require`
- `//go:build integration` tag on tests that start real servers or make real network connections

```bash
make test              # unit tests only
make test-integration  # integration tests (includes server lifecycle, DB connectivity)
make lint              # golangci-lint via Docker
```

CI runs both in sequence — unit tests must pass before integration tests run.

---

## Contributing

Contributions, bug reports, and feature requests are very welcome.

**To contribute:**

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-improvement`
3. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org)
4. Open a pull request with a clear description of the problem and solution

**Not sure where to start?** Open an [issue](https://github.com/nawafswe/go-service-starter-kit/issues) first to discuss the idea — that is always appreciated before writing code.

**Feedback** on the architecture, tooling choices, or documentation is equally welcome. If something feels wrong or overly complex, please say so.

---

## License

MIT — see [LICENSE](LICENSE).

---

<div align="center">

Built with the [Go gopher](https://go.dev/blog/gopher) and real production pain.

*If this template saved you time, consider leaving a star on GitHub.*

</div>
