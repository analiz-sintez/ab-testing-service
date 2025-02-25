# A/B Testing Service

A Go-based service for splitting web traffic for A/B testing purposes. The service includes a REST API, proxy modules for traffic splitting, and comprehensive monitoring capabilities.

## Features

- Dynamic proxy configuration for A/B testing
- REST API for managing proxies and viewing statistics
- Cookie-based user session persistence
- Traffic splitting based on configurable weights
- Prometheus metrics for monitoring
- Kafka integration for statistics processing
- Redis caching for proxy configurations
- PostgreSQL for persistent storage

## Requirements

- Docker and Docker Compose
- Go 1.23 or later (for local development)

## Quick Start

1. Clone the repository

2. Login into Github container registry (GHCR):
   - acquire token for ghcr (Github settings -> developer settings -> personal access tokens)
   - use it as password for `docker login ghcr.io`

2. Start the services using Docker Compose (use `-d` to run program in the background):
   ```bash
   docker-compose up
   ```

3. The service will be available at (ports can be reassigned in `.env`):
   - UI: http://localhost:38000
   - REST API: http://localhost:38080
   - prometheus: http://localhost:38000/prom
   - kafka-ui: http://localhost:38000/kafka-ui
   - grafana metrics: http://localhost:38000/grafana

## API Endpoints

- `GET /api/proxies` - List all proxies
- `POST /api/proxies` - Create a new proxy
- `GET /api/proxies/:id` - Get proxy details
- `DELETE /api/proxies/:id` - Delete a proxy
- `GET /api/proxies/:id/stats` - Get proxy statistics
- `PUT /api/proxies/:id/targets` - Update proxy targets

## Frontend

Frontend devserver starts from the `web` directory. Install dependencies using `npm install` and Run `npm run dev` to start the devserver.

## Configuration

The service can be configured using the `config/config.yaml` file. Key configuration options include:

- Server port and host
- Database connection details
- Redis connection details
- Kafka configuration
- Prometheus settings

## Development

To run the service locally:

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the service:
   ```bash
   go run main.go
   ```

## Monitoring

The service exposes Prometheus metrics at `/metrics` endpoint, including:

- Request counts per target
- Request latencies
- Error rates

## Deployment

App will be built and pushed to Github Container Registry (GHCR) on adding tags. Tags can be added like this: `git tag v0.0.1; git push --tags`.

## License

MIT
