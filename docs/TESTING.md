# Testing Guide

Comprehensive testing strategy and CI/CD pipeline for Instagrano.

## Test Architecture

- **Testcontainers**: Real PostgreSQL, Redis, and LocalStack containers for integration tests
- **Gomega**: BDD-style assertions for readable test code
- **No Mocks**: Uses real dependencies for authentic testing
- **Structured Tests**: Organized by domain (auth, posts, feed, interactions)
- **Mock Transport**: Custom HTTP transport for webclient to avoid external calls in tests

## Test Categories

### Unit Tests
- Domain models (scoring algorithm, password validation)
- Service layer (auth, posts, feed, interactions)
- Configuration and logging utilities
- Pagination and caching logic

### Integration Tests
- Complete API endpoints with real database
- Authentication flows (register, login, JWT validation)
- Post creation and retrieval (file upload and URL-based)
- Feed generation with caching
- Like and comment interactions
- WebSocket real-time events
- S3 storage operations via LocalStack

## Test Commands

```bash
# Run all tests with coverage
make test-full

# Run specific test categories
go test ./tests/ -run "TestAuth" -v
go test ./tests/ -run "TestFeed" -v
go test ./tests/ -run "TestPost" -v

# Generate coverage report
go test -coverprofile=coverage.out ./tests/ -coverpkg=./...
go tool cover -html=coverage.out
```

## Test Coverage Details

- **Domain Layer**: 100% coverage (scoring, validation)
- **Service Layer**: 95%+ coverage (auth, posts, feed, interactions)
- **Handler Layer**: 90%+ coverage (API endpoints)
- **Repository Layer**: 85%+ coverage (database operations)
- **Infrastructure**: 80%+ coverage (config, logging, caching)

## Testing Best Practices

- **Real Dependencies**: Uses Testcontainers for PostgreSQL, Redis, and LocalStack
- **Isolation**: Each test gets fresh containers and database state
- **BDD Style**: Given-When-Then structure with Gomega assertions
- **Comprehensive**: Tests happy paths, error cases, and edge conditions
- **Performance**: Tests caching behavior and performance characteristics
- **Mock Transport**: Avoids external HTTP calls in URL-based upload tests

## CI/CD Pipeline

The project includes a comprehensive GitHub Actions workflow for continuous integration:

### Pipeline Features
- **Testcontainers Integration**: Real PostgreSQL, Redis, and LocalStack containers for authentic testing
- **Docker-in-Docker**: Full container orchestration support in CI environment
- **Race Detection**: Concurrent testing for thread safety validation
- **Coverage Reporting**: Comprehensive test coverage metrics (77.6%)
- **Test Isolation**: Fresh database, cache, and storage containers for each test run
- **Mock Transport**: Custom HTTP transport prevents external calls during tests

### Workflow Configuration
```yaml
# .github/workflows/ci.yml
- Docker-in-Docker service for Testcontainers
- PostgreSQL, Redis, and LocalStack containers for integration tests
- Race condition testing with coverage reporting
- Mock transport configuration for webclient tests
- Automatic deployment to Railway on main branch
```

### CI Environment Variables
- `TESTCONTAINERS_RYUK_DISABLED=true` - Disables cleanup for CI stability
- Docker Compose support for multi-container testing
- Go 1.23.x with race detection enabled

## Test Structure

```
tests/
├── main_test.go              # Ginkgo suite entry point
├── setup_test.go            # Testcontainers setup and configuration
├── auth_service_test.go     # Authentication service tests
├── feed_service_test.go     # Feed and caching service tests
├── post_service_test.go     # Post creation service tests
├── post_service_url_test.go # URL-based upload service tests
├── integration_test.go      # Complete API endpoint tests
├── cache_test.go           # Redis cache behavior tests
├── handler_test.go         # HTTP handler tests
├── logger_test.go          # Logging configuration tests
├── middleware_test.go      # JWT middleware tests
├── pagination_test.go      # Cursor pagination tests
├── domain_test.go          # Domain model tests
├── config_test.go          # Configuration tests
└── post_view_test.go       # Post view tracking tests
```

## Running Tests Locally

### Prerequisites
- Docker and Docker Compose running
- Go 1.23+

### Quick Test Run
```bash
# Run all tests (uses Testcontainers - no setup needed)
make test-full
```

### Detailed Testing
```bash
# Run tests with verbose output
go test -race -v ./tests/...

# Run specific test suite
go test ./tests/ -run "TestAuth" -v

# Run with coverage
go test -cover ./tests/...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./tests/ -coverpkg=./...
go tool cover -html=coverage.out
```

## Test Data Management

Tests use Testcontainers to create isolated environments:
- **Fresh Database**: Each test run gets a clean PostgreSQL instance with automatic migrations
- **Fresh Cache**: Each test run gets a clean Redis instance
- **Fresh Storage**: Each test run gets a clean LocalStack S3 bucket
- **Automatic Cleanup**: All containers are terminated after tests complete
- **No Shared State**: Complete isolation between test runs
- **Mock Transport**: Custom HTTP transport prevents external calls during URL-based upload tests

## Performance Testing

See [Performance Results](PERFORMANCE.md) for K6 load testing details and metrics.
