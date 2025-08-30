# BinTraceBench

A secure backend API for binary analysis, benchmarking, and process inspection with persistent storage.

## Features

- Static and dynamic binary analysis with ELF parsing and ptrace syscall tracing
- Sandboxed benchmarking with Linux namespaces and resource limiting
- Live process inspection via /proc filesystem
- User authentication with bcrypt password hashing
- SQLite and PostgreSQL database support
- File hash-based result caching
- RESTful API with comprehensive endpoints

## Quick Start

### Installation

```bash
git clone https://github.com/ashborn3/BinTraceBench.git
cd BinTraceBench
go mod tidy
go build -o bintracebench ./cmd/bintracebench
```

### Configuration

```bash
# SQLite (default)
export DB_TYPE=sqlite
export SQLITE_PATH=./data/bintracebench.db

# PostgreSQL (optional)
export DB_TYPE=postgresql
export POSTGRES_HOST=localhost
export POSTGRES_USER=bintracebench
export POSTGRES_PASSWORD=password
export POSTGRES_DB=bintracebench
```

### Run Server

```bash
./bintracebench
# Server starts on http://localhost:8080
```

## API Endpoints

### Authentication
- POST `/auth/register` - Register new user
- POST `/auth/login` - Login and get access token
- GET `/auth/me` - Get current user info
- POST `/auth/logout` - Logout

### Binary Analysis (Protected)
- POST `/analyze` - Static analysis
- POST `/analyze?dynamic=true` - Dynamic tracing
- GET `/analyze` - List user's results
- GET `/analyze/{id}` - Get specific result
- DELETE `/analyze/{id}` - Delete result

### Benchmarking (Protected)
- POST `/bench` - Run benchmark
- POST `/bench?trace=true` - Benchmark with tracing
- GET `/bench` - List benchmark results
- GET `/bench/{id}` - Get specific benchmark
- DELETE `/bench/{id}` - Delete benchmark

### Process Inspection (Protected)
- GET `/proc/{pid}` - Process info
- GET `/proc/{pid}/files` - Open file descriptors
- GET `/proc/{pid}/net` - Network connections

## Usage Example

```bash
# Register user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass123","email":"user@example.com"}'

# Login
TOKEN=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass123"}' | jq -r '.token')

# Analyze binary
curl -X POST http://localhost:8080/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/bin/ls"

# Run benchmark
curl -X POST http://localhost:8080/bench \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/bin/ls"
```

## Testing

Run the automated test script:

```bash
python3 scripts/endpointTester.py
```

## Requirements

- Go 1.20+
- Linux OS
- SQLite3 or PostgreSQL

## Database Schema

The system uses four main tables:
- `users` - User accounts with bcrypt password hashing
- `sessions` - Authentication sessions with token expiry
- `analysis_results` - Binary analysis results with file hash caching
- `benchmark_results` - Benchmark results with execution metrics

## Security

- All endpoints except registration/login require authentication
- Bcrypt password hashing with salt
- Session-based token authentication with configurable expiry
- User data isolation and access control
- SQL injection protection with parameterized queries

## License

MIT License