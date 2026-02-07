# Browser Query AI

A browser query AI service built with Go for learning distributed systems.

## Getting Started

Run the server:

```bash
go run ./cmd/server
```

Run the tests:
```bash
go test -v ./internal/session/...
```

Run the tests with race detection:

```bash
go test -race ./internal/session/...
```
## Environment Variables

The following environment variables can be set to configure the service:

### `ENV`
Sets the environment mode. Affects logging format.
- `production` - Uses JSON logging format
- `development` (default) - Uses human-readable text logging format

```bash
ENV=production go run ./cmd/server
```

### `CHROMIUM_PATH`
Optional. Path to the Chromium/Chrome binary. If not set, the service will automatically search common installation paths.

```bash
CHROMIUM_PATH="/path/to/chromium" go run ./cmd/server
```

### `SERVER_PORT`
Optional. Port number for the server to listen on.
- Default: `8080`

```bash
SERVER_PORT=3000 go run ./cmd/server
```

### `MAX_BROWSERS`
Optional. Maximum number of browser instances that can run concurrently.
- Default: `5`

```bash
MAX_BROWSERS=10 go run ./cmd/server
```

## Example with Multiple Environment Variables

```bash
ENV=production SERVER_PORT=3000 MAX_BROWSERS=10 go run ./cmd/server
```
