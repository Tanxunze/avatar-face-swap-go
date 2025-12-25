# Avatar Face Swap - Go Backend

A Go rewrite of the Flask backend for the Avatar Face Swap application.

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── handler/             # HTTP handlers (routes)
│   ├── middleware/          # Auth and other middleware
│   ├── model/               # Data models
│   ├── repository/          # Database operations
│   └── service/             # Business logic
├── configs/                 # Configuration files
├── go.mod
└── go.sum
```

## Development

```bash
# Run the server
go run cmd/server/main.go

# Build
go build -o bin/server.exe cmd/server/main.go
```

## API Endpoints

TBD - migrating from Flask backend.
