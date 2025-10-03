# Ads Backend API

## Notes: This project was done with the help of AI, specifically Cursor editor. 
Important decisions made: 
Use of uber-fx in order to simplify dependency injections. 
Mockery library to test due to the familiarity of the user with it.
Basic clean architecture implemented with domain, service, handlers, persistence layers.

## Installation

```bash
git clone <repo>
cd ads_backend
go mod download
```

## Commands

### Run server
```bash
go run cmd/main.go
```
Server starts at `http://localhost:8080` if no PORT is specified as a .env variable

### Export .env variables
```bash
export $(grep -v '^#' .env | xargs)
```

### Unit tests
```bash
go test ./internal/http/... -v
```
### Coverage Checks
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### E2E test
```bash
go test -v -run TestE2E
```

### Generate mocks
```bash
mockery --all
```

## Environment Variables

```bash
HTTP_PORT=8080
```

