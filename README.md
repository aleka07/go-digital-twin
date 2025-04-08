# Go Digital Twin Framework

A Go-based framework for building and managing digital twins. This framework provides a robust foundation for creating digital twin applications with features for twin management, registry, messaging simulation, and API integration.

## Project Structure

```
go-digital-twin/
├── cmd/
│   └── dt_server/         # Main server application
├── pkg/
│   ├── api/              # API-related functionality
│   ├── messaging_sim/    # Messaging simulation components
│   ├── registry/         # Twin registry management
│   └── twin/            # Core digital twin functionality
└── tests/               # Test files
```

## Features

- Digital Twin Management
- Twin Registry System
- Messaging Simulation
- RESTful API Interface
- Chi Router Integration

## Requirements

- Go 1.21 or higher

## Installation

1. Clone the repository:
```bash
git clone https://github.com/aleka07/go-digital-twin.git
cd go-digital-twin
```

2. Install dependencies:
```bash
go mod download
```

## Usage

To run the digital twin server:

```bash
go run cmd/dt_server/main.go
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o dt_server cmd/dt_server/main.go
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 