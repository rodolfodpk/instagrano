# Testing Guide

Comprehensive testing strategy and CI/CD pipeline for Instagrano.

## Test Architecture

- **Testcontainers**: Real PostgreSQL and Redis containers for integration tests
- **Gomega**: BDD-style assertions for readable test code
- **No Mocks**: Uses real dependencies for authentic testing
- **Structured Tests**: Organized by domain (auth, posts, feed, interactions)

## Test Categories

### Unit Tests
- Domain models (scoring algorithm, password validation)
- Service layer (auth, posts, feed, interactions)
- Configuration and logging utilities
- Pagination and caching logic

### Integration Tests
- Complete API endpoints with real database
- Authentication flows (register, login, JWT validation)
- Post creation and retrieval
- Feed generation with caching
- Like and comment interactions

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

- **Real Dependencies**: Uses Testcontainers for PostgreSQL and Redis
- **Isolation**: Each test gets fresh containers and database state
- **BDD Style**: Given-When-Then structure with Gomega assertions
- **Comprehensive**: Tests happy paths, error cases, and edge conditions
- **Performance**: Tests caching behavior and performance characteristics

## CI/CD Pipeline

The project includes a comprehensive GitHub Actions workflow for continuous integration:

### Pipeline Features
- **Testcontainers Integration**: Real PostgreSQL and Redis containers for authentic testing
- **Docker-in-Docker**: Full container orchestration support in CI environment
- **Race Detection**: Concurrent testing for thread safety validation
- **Coverage Reporting**: Comprehensive test coverage metrics (77.6%)
- **Test Isolation**: Fresh database and cache containers for each test run

### Workflow Configuration
```yaml
# .github/workflows/ci.yml
- Docker-in-Docker service for Testcontainers
- PostgreSQL and Redis containers for integration tests
- Race condition testing with coverage reporting
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
├── setup_test.go            # Test application setup
├── auth_service_test.go     # Authentication tests
├── feed_service_test.go     # Feed and caching tests
├── post_service_test.go     # Post creation tests
├── integration_test.go      # API endpoint tests
├── cache_test.go           # Redis cache tests
├── handler_test.go         # HTTP handler tests
├── logger_test.go          # Logging tests
├── middleware_test.go      # Middleware tests
├── pagination_test.go      # Pagination tests
├── domain_test.go          # Domain model tests
├── config_test.go          # Configuration tests
└── mock_storage.go         # Mock S3 storage for tests
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
- Fresh PostgreSQL database for each test run
- Fresh Redis instance for cache testing
- Automatic cleanup after tests complete
- No shared state between test runs

## Performance Testing

See [Performance Results](PERFORMANCE.md) for K6 load testing details and metrics.
