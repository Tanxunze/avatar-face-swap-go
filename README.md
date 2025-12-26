# Avatar Face Swap - Go Backend

A Go version backend for the [Avatar Face Swap](https://github.com/ZhaoKuanhong/avatar-face-swap) project.

## Project Structure

```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # Database initialization
│   ├── handler/         # HTTP handlers (routes)
│   ├── middleware/      # Auth middleware
│   ├── model/           # Data models
│   ├── repository/      # Database operations
│   ├── service/         # Business logic
│   └── storage/         # File storage utilities
├── pkg/response/        # Shared response utilities
├── nginx/               # Nginx configuration
└── data/                # SQLite DB + file storage (gitignored)
```

## Development

```bash
# Copy environment file
cp .env.example .env

# Run with hot reload (requires air)
air

# Or run directly
go run cmd/server/main.go

# Build
go build -o bin/server cmd/server/main.go
```
