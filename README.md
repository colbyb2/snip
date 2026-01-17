# Snip

A production-grade URL shortener built for learning.

## Project Structure

```
snip/
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── handler/          # HTTP handlers
│   ├── model/            # Domain models
│   ├── repository/       # Data persistence interfaces and implementations
│   └── service/          # Business logic
├── pkg/
│   └── shortcode/        # Short code generation (reusable package)
├── terraform/            # Infrastructure as code (coming soon)
└── docs/                 # Documentation
```

## Quick Start

### Prerequisites

- Go 1.25+

### Run Locally

```bash
# Run the server
go run cmd/api/main.go

# Or build and run
go build -o snip cmd/api/main.go
./snip
```

The server starts on `http://localhost:8080` by default.

### Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `BASE_URL` | `http://localhost:8080` | Base URL for generated short links |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

## API Endpoints

### Create Short Link

```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url"}'
```

Response:
```json
{
  "short_code": "abc1234",
  "short_url": "http://localhost:8080/abc1234",
  "original_url": "https://example.com/very/long/url"
}
```

### Redirect

```bash
curl -L http://localhost:8080/abc1234
```

### Get Stats

```bash
curl http://localhost:8080/api/links/abc1234/stats
```

Response:
```json
{
  "short_code": "abc1234",
  "original_url": "https://example.com/very/long/url",
  "click_count": 42,
  "created_at": "2025-01-17T12:00:00Z"
}
```

### Delete Link

```bash
curl -X DELETE http://localhost:8080/api/links/abc1234
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...
```

## Development Phases

- [x] Phase 1 Week 1: Core API with in-memory storage
- [ ] Phase 1 Week 2: Terraform + AWS infrastructure
- [ ] Phase 1 Week 3: CI/CD pipeline
- [ ] Phase 1 Week 4: Observability
- [ ] Phase 1 Week 5: Hardening and documentation

## License

MIT
