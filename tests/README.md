# Integration Tests for Instagrano

This directory contains comprehensive integration tests for the Instagrano application using Testcontainers for authentic testing with real PostgreSQL, Redis, and LocalStack S3 containers.

## Overview

The integration tests verify:
- ✅ User registration and login with JWT authentication
- ✅ File upload functionality (both local files and URL-based)
- ✅ Feed access with Redis caching
- ✅ Post interactions (likes, comments, views)
- ✅ WebSocket real-time events
- ✅ S3 storage operations via LocalStack

## Test Architecture

### Testcontainers Integration
- **PostgreSQL Container**: Real database with automatic schema migrations
- **Redis Container**: Real cache for testing caching behavior
- **LocalStack Container**: S3-compatible storage for media uploads
- **Dynamic Ports**: All containers use dynamic port allocation
- **Automatic Cleanup**: Containers are automatically terminated after tests

### Mock Transport for Webclient
- **Custom HTTP Transport**: Intercepts test URLs (`http://localhost/test/image`)
- **No External Calls**: Returns predefined responses without real HTTP requests
- **Test Isolation**: Enables URL-based upload tests without external dependencies

## Running the Tests

### Prerequisites
- Docker and Docker Compose running
- Go 1.23+

### Quick Test Run
```bash
# Run all tests (uses Testcontainers - no manual setup needed)
make test-full
```

### Detailed Testing
```bash
# Run tests with verbose output
go test -race -v ./tests/...

# Run specific test suites
go test ./tests/ -run "TestAuth" -v
go test ./tests/ -run "TestFeed" -v
go test ./tests/ -run "TestPost" -v
go test ./tests/ -run "TestIntegration" -v

# Run with coverage
go test -cover ./tests/...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./tests/ -coverpkg=./...
go tool cover -html=coverage.out
```

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

## Test Categories

### 1. Authentication Tests
- User registration with unique emails
- Login functionality and JWT token generation
- JWT token validation and middleware
- Authentication edge cases and error handling

### 2. Post Creation Tests
- **File Upload**: Multipart form data handling
- **URL-based Upload**: Media download and S3 storage
- Post creation with media validation
- S3 integration via LocalStack

### 3. Feed and Caching Tests
- Feed generation with scoring algorithm
- Redis cache hit/miss behavior
- Cache invalidation on interactions
- Pagination with cursor-based navigation

### 4. Interaction Tests
- Like/unlike functionality with toggle behavior
- Comment creation and retrieval
- Post view time tracking
- Real-time WebSocket events

### 5. Integration Tests
- Complete API endpoint testing
- End-to-end user workflows
- Error handling and edge cases
- Performance characteristics

## Key Features Tested

### WebSocket Real-time Events
Tests verify:
- WebSocket connection establishment with JWT authentication
- Real-time event broadcasting (likes, comments)
- Client-side event handling and UI updates
- Connection resilience and reconnection

### S3 Storage Integration
Tests verify:
- LocalStack S3 container setup
- Media upload to S3 buckets
- URL generation and retrieval
- File type validation and processing

### Redis Caching Strategy
Tests verify:
- Feed cache performance improvements
- Cache invalidation on data changes
- Cache hit/miss behavior
- TTL configuration and expiration

## Test Data Management

Tests use Testcontainers to create isolated environments:
- **Fresh Database**: Each test run gets a clean PostgreSQL instance
- **Fresh Cache**: Each test run gets a clean Redis instance
- **Fresh Storage**: Each test run gets a clean LocalStack S3 bucket
- **Automatic Cleanup**: All containers are terminated after tests complete
- **No Shared State**: Complete isolation between test runs

## Performance Testing

See [Performance Results](../docs/PERFORMANCE.md) for K6 load testing details and comprehensive metrics.

## Troubleshooting

### Test Failures
- **Container Issues**: Ensure Docker is running and has sufficient resources
- **Port Conflicts**: Testcontainers use dynamic ports, conflicts are rare
- **Timeout Issues**: Increase test timeouts if containers are slow to start

### LocalStack Issues
- **S3 Bucket Creation**: Buckets are created automatically during tests
- **Endpoint Configuration**: Dynamic endpoints are configured automatically
- **Region Settings**: Tests use `us-east-1` region for LocalStack

### Database Issues
- **Migration Failures**: Schema migrations run automatically via golang-migrate
- **Connection Issues**: Database connections are managed by Testcontainers
- **Data Isolation**: Each test gets a fresh database instance

## Continuous Integration

The test suite is designed for CI/CD environments:
- **GitHub Actions**: Full Testcontainers support with Docker-in-Docker
- **Race Detection**: Concurrent testing for thread safety validation
- **Coverage Reporting**: Comprehensive test coverage metrics
- **No External Dependencies**: All services run in containers

## Contributing

When adding new tests:
1. Follow the existing test structure and naming conventions
2. Use Testcontainers for all external dependencies
3. Test both success and failure scenarios
4. Include performance characteristics where relevant
5. Update this README if adding new test categories

## Dependencies

- `github.com/onsi/ginkgo/v2` - BDD testing framework
- `github.com/onsi/gomega` - Matcher library for assertions
- `github.com/testcontainers/testcontainers-go` - Container orchestration
- `github.com/testcontainers/testcontainers-go/modules/postgres` - PostgreSQL container
- `github.com/testcontainers/testcontainers-go/modules/redis` - Redis container
- `github.com/testcontainers/testcontainers-go/modules/localstack` - LocalStack container
- `github.com/golang-migrate/migrate/v4` - Database migrations