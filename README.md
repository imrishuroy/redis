# Redis

A Redis server implementation in Go.

## Features

- Synchronous TCP server
- Configurable host and port via CLI flags
- Client connection tracking

## Getting Started

### Prerequisites

- Go 1.26.1 or later

### Installation

```bash
git clone https://github.com/imrishuroy/redis.git
cd redis
go build
```

### Usage

Run the server with default settings (0.0.0.0:6379):

```bash
./redis
```

Or customize the host and port:

```bash
./redis -host 127.0.0.1 -port 6380
```

### CLI Flags

| Flag   | Default   | Description             |
|--------|-----------|-------------------------|
| -host  | 0.0.0.0   | Host address to bind to |
| -port  | 6379      | Port to listen on       |

## Project Structure

```
.
├── main.go           # Entry point and CLI flag setup
├── config/
│   └── config.go     # Global configuration variables
├── server/
│   └── sync_tcp.go   # Synchronous TCP server implementation
└── docs/
    └── ABOUT_REDIS.md
```

## Testing

Connect using netcat or redis-cli:

```bash
nc localhost 6379
```
