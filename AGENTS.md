# AGENTS.md - AI Agent Instructions for IPINFO Project

## Project Overview

IPINFO is a Go-based web service that provides IP geolocation information through multiple API endpoints.
The service returns IP details in various formats (JSON, XML, HTML, plain text)
and includes geographical, timezone, and network information.

## Architecture

- **Language**: Go 1.24+
- **Main Entry**: `ipinfo.go`
- **Configuration**: `conf/` package handles JSON configuration
- **HTTP Handlers**: `handle/` package manages all HTTP endpoints
- **Database**: MaxMind GeoLite2 City database for geolocation
- **Deployment**: Docker containerization with multi-stage builds

## Core Components

### 1. Configuration (`conf/`)

- Manages application settings via JSON configuration
- Handles database paths, server settings, and logging
- Validates configuration parameters on startup

### 2. HTTP Handlers (`handle/`)

- Implements multiple response formats: text, JSON, XML, HTML
- Provides endpoints: `/`, `/short`, `/compact`, `/json`, `/xml`, `/html`, `/full`, `/version`, `/health`
- Handles IP detection from headers (X-Forwarded-For, X-Real-IP)

### 3. Main Application (`ipinfo.go`)

- Sets up HTTP server with graceful shutdown
- Initializes MaxMind database connection
- Manages logging and error handling

## Development Guidelines

### Code Standards

- Follow Go conventions and use `gofmt`
- Use `golangci-lint` for code quality (config in `golangci.yml`)
- Maintain BSD 3-Clause license headers
- Write unit tests for all new functionality

### Testing

- Test files follow `*_test.go` pattern
- Use table-driven tests where appropriate
- Mock HTTP requests using `httptest` package
- Ensure >80% code coverage

### Error Handling

- Use structured error handling with context
- Log errors with appropriate severity levels
- Return meaningful HTTP status codes
- Handle MaxMind database errors gracefully

## API Endpoints Reference

| Endpoint   | Description              | Response Format |
|------------|--------------------------|-----------------|
| `/`        | Default detailed IP info | Plain text      |
| `/short`   | Concise IP information   | Plain text      |
| `/compact` | Minimal IP details       | Plain text      |
| `/json`    | Structured IP data       | JSON            |
| `/xml`     | Structured IP data       | XML             |
| `/html`    | Human-readable page      | HTML            |
| `/full`    | Enhanced HTML page       | HTML            |
| `/version` | Application version      | JSON            |
| `/health`  | Health check             | JSON            |

## Configuration Management

### Required Files

- `config.json` or `ipinfo.json`: Main configuration
- `GeoLite2-City.mmdb`: MaxMind database file

### Environment Variables

- `CONFIG`: Override default config file path
- `PORT`: Override default server port
- `DEBUG`: Enable debug logging

## Deployment Instructions

### Local Development

```bash
make build          # Build binary
make start          # Start service
make stop           # Stop service
make restart        # Restart service
```

### Docker Deployment

```bash
make docker         # Build Docker image
make docker_linux_amd64  # Build for Linux AMD64
```

### Production Considerations

- Use read-only volumes for configuration and database
- Run with non-root user (`-u $UID:$UID`)
- Configure reverse proxy for HTTPS
- Monitor disk space for log files

## Security Guidelines

### Input Validation

- Validate all HTTP headers for IP extraction
- Sanitize user inputs before processing
- Use proper Content-Type headers

### Network Security

- Bind to specific interfaces in production
- Use TLS termination at reverse proxy
- Implement rate limiting if needed

### Data Privacy

- Log minimal PII information
- Respect IP anonymization requirements
- Comply with applicable data protection laws

## Performance Optimization

### Database Access

- MaxMind database is memory-mapped for performance
- Cache database connections appropriately
- Handle database reload for updates

### HTTP Performance

- Use appropriate caching headers
- Implement request timeouts
- Monitor response times and error rates

## Monitoring and Observability

### Health Checks

- `/health` endpoint for load balancer checks
- Monitor database connectivity
- Track service startup and shutdown events

### Logging

- Structured logging with appropriate levels
- Include request IDs for traceability
- Log performance metrics

### Metrics

- Track request counts by endpoint
- Monitor response times
- Alert on error rate thresholds

## Common Tasks for AI Agents

1. **Adding New Endpoints**: Extend `handle/` package with new handlers
2. **Configuration Changes**: Modify `conf/` package structures
3. **Response Format Updates**: Update parsing and formatting logic
4. **Testing**: Add comprehensive test coverage
5. **Documentation**: Update API documentation in `api.md`
6. **Performance Tuning**: Optimize database queries and response generation

## Troubleshooting

### Common Issues

- MaxMind database not found: Check file path and permissions
- Port binding errors: Verify port availability
- Configuration errors: Validate JSON syntax and required fields

### Debug Mode

- Enable with `DEBUG=1` environment variable
- Provides detailed request/response logging
- Shows database query performance

## Dependencies

### Core Dependencies

- MaxMind GeoIP2 database
- Standard Go libraries
- No external HTTP frameworks (uses stdlib)

### Development Dependencies

- golangci-lint for code quality
- Docker for containerization
- Make for build automation

Remember to update this documentation when making significant changes to the project structure or adding new features.