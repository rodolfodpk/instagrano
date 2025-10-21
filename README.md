# Instagrano MVP

[![CI/CD](https://github.com/rodolfodpk/instagrano/actions/workflows/ci.yml/badge.svg)](https://github.com/rodolfodpk/instagrano/actions/workflows/ci.yml)
[![Test Coverage](https://img.shields.io/badge/coverage-81.9%25-brightgreen)](https://github.com/rodolfodpk/instagrano)
[![Go Report Card](https://goreportcard.com/badge/github.com/rodolfodpk/instagrano)](https://goreportcard.com/report/github.com/rodolfodpk/instagrano)
[![Dependabot](https://img.shields.io/badge/dependabot-enabled-025e8c?style=flat&logo=dependabot)](https://github.com/rodolfodpk/instagrano/security/dependabot)

A mini Instagram API built with Go Fiber, PostgreSQL, Redis, LocalStack S3, Zap logging, Testcontainers, Gomega, Docker Compose, and JWT authentication.

## Features

- Redis caching with feed performance optimization
- Structured logging with Zap
- Cursor-based pagination
- S3-compatible storage via LocalStack
- JWT authentication
- Feed scoring algorithm (time decay + engagement)
- File upload (images/videos)
- Frontend with Alpine.js
- View time tracking

## Quick Start

```bash
# Run tests (uses Testcontainers - no setup needed)
make test-full

# For development with frontend:
make docker-up
make migrate
make start
```

## API Endpoints

- `GET /health` - Health check
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/me` - Get current user (requires JWT)
- `POST /api/posts` - Create post with file upload or URL (requires JWT)
- `GET /api/posts/:id` - Get specific post (requires JWT)
- `POST /api/posts/:id/like` - Like a post (requires JWT)
- `POST /api/posts/:id/comment` - Comment on a post (requires JWT)
- `POST /api/posts/:id/view/start` - Start tracking view time (requires JWT)
- `POST /api/posts/:id/view/end` - End tracking and record duration (requires JWT)
- `GET /api/feed` - Get user feed (requires JWT)

### Feed Pagination

The feed endpoint supports both cursor-based and page-based pagination with configurable page sizes:

**Cursor-based (recommended):**
```bash
# First page with default size (20 posts)
GET /api/feed

# Custom page size
GET /api/feed?limit=10

# Next page using cursor from previous response
GET /api/feed?limit=10&cursor=MTc2MDk3MjI5OF8xNQ==
```

**Page-based (legacy):**
```bash
# Default page size
GET /api/feed?page=1

# Custom page size
GET /api/feed?page=1&limit=50
```

**Configuration:**
- `DEFAULT_PAGE_SIZE`: Default posts per page (default: 20)
- `MAX_PAGE_SIZE`: Maximum allowed page size (default: 100)
- Frontend dropdown: 3, 5, 10, 20, 50 posts per page

**Response format (cursor-based):**
```json
{
  "posts": [...],
  "next_cursor": "MTc2MDk3MjI5OF8xNQ==",
  "has_more": true
}
```

## Frontend

Open: http://localhost:8080/feed.html (after logging in at http://localhost:8080/)

The frontend includes:
- **Login/Registration**: Tabbed interface for user authentication
- **Feed with Load More**: Cursor-based pagination with "Load More" button
- **Post Creation**: File upload with image/video support or URL-based media
- **Interactions**: Like and comment functionality
- **View Time Tracking**: Automatic tracking using Intersection Observer

## Redis Caching

The application uses Redis for high-performance feed caching with dramatic performance improvements:

**Performance Impact:**
- **Cache Hit**: ~1-2ms response time (10x faster)
- **Cache Miss**: ~13-22ms response time (database query)
- **Expected Hit Rate**: 70-80% in steady state

**Cache Strategy:**
- **Cache Key Format**: `feed:cursor:{cursor}:limit:{limit}`
- **TTL**: 5 minutes (configurable via `CACHE_TTL`)
- **Invalidation**: Time-based expiration (simple and predictable)

**Configuration:**
```bash
REDIS_ADDR=localhost:6379     # Redis connection address
REDIS_PASSWORD=               # Redis password (empty for dev)
REDIS_DB=0                    # Redis database number
CACHE_TTL=5m                  # Cache TTL (accepts: 30s, 5m, 1h, etc.)
```

**Cache Operations:**
```bash
# Connect to Redis CLI
make redis-cli

# View all cached feed keys
KEYS feed:*

# Get cache stats
INFO stats

# Monitor cache operations in real-time
MONITOR

# Flush all cached data
make redis-flush

# Check specific cache key
GET "feed:cursor::limit:20"
```

**Structured Logging for Cache:**

Cache operations are logged with structured fields for monitoring:
```json
{
  "level": "info",
  "msg": "cache hit",
  "cache_key": "feed:cursor::limit:20",
  "duration": "1.2ms"
}
```

**Trade-offs:**
- ✅ **Pros**: 10x faster responses, reduced DB load, horizontally scalable
- ⚠️ **Cons**: Data up to 5 minutes stale, additional service dependency, memory usage

## View Time Tracking

The application automatically tracks how long users spend viewing each post using Intersection Observer API.

**How it works:**
- **Automatic Detection**: Posts are tracked when 50%+ visible for 1+ second
- **Duration Calculation**: Time between view start and end is recorded
- **Database Storage**: View sessions stored in `post_views` table
- **Counter Updates**: `views_count` incremented for each view start

**API Endpoints:**
```bash
# Start tracking (called automatically by frontend)
POST /api/posts/:id/view/start
# Returns: {"id": 1, "user_id": 1, "post_id": 5, "started_at": "2024-01-15T10:30:00Z"}

# End tracking (called automatically by frontend)  
POST /api/posts/:id/view/end
# Body: {"started_at": "2024-01-15T10:30:00Z"}
# Returns: {"message": "view ended"}
```

**Frontend Implementation:**
- Uses Intersection Observer to detect viewport visibility
- Tracks active view sessions in Alpine.js state
- Automatically ends views on page unload
- Handles multiple views of same post gracefully

**Database Schema:**
```sql
CREATE TABLE post_views (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    post_id INT REFERENCES posts(id),
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    duration_seconds INT
);
```

## Structured Logging

The application uses Zap for high-performance structured logging:

**Configuration:**
- `LOG_LEVEL`: debug, info, warn, error (default: info)
- `LOG_FORMAT`: json, console (default: json)

**Log Examples:**
```bash
# View logs in console format
LOG_FORMAT=console make start

# Filter logs by user
grep '"user_id":12' logs/app.log

# Find slow requests
grep '"duration":"[5-9][0-9][0-9]ms"' logs/app.log

# Find cache hits
grep '"cache hit"' logs/app.log

# Find cache misses
grep '"cache miss"' logs/app.log
```

**Log Fields:**
- `request_id`: Unique identifier for request tracing
- `user_id`: Authenticated user ID (when available)
- `method`, `path`: HTTP method and path
- `status`: HTTP response status
- `duration`: Request processing time
- `response_size`: Response body size in bytes
- `cache_key`: Redis cache key (when caching)
- `cache_hit`/`cache_miss`: Cache operation result

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend       │    │   Go Fiber      │    │   PostgreSQL    │
│   (Alpine.js)    │◄──►│   HTTP Server   │◄──►│   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                       ┌────────┴────────┐
                       ▼                 ▼
              ┌─────────────────┐ ┌─────────────────┐
              │   LocalStack    │ │     Redis       │
              │   S3 Storage    │ │     Cache       │
              └─────────────────┘ └─────────────────┘
```

## Post Creation

You can create posts in two ways:

### File Upload
```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Authorization: Bearer $TOKEN" \
  -F "title=My Post" \
  -F "caption=Uploaded file" \
  -F "media_type=image" \
  -F "media=@/path/to/image.jpg"
```

### URL-Based Media
```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Authorization: Bearer $TOKEN" \
  -F "title=My Post" \
  -F "caption=From URL" \
  -F "media_url=https://example.com/image.jpg"
```

The system will:
1. Download the media from the URL
2. Upload it to S3/LocalStack
3. Create the post with the S3 URL

## Development

**Environment Variables:**
```bash
PORT=8080                    # API server port
SWAGGER_PORT=8081            # Swagger UI server port
JWT_SECRET=your-secret-key   # JWT signing secret
LOG_LEVEL=info              # Log level
LOG_FORMAT=json             # Log format
DATABASE_URL=postgres://... # PostgreSQL connection string
S3_ENDPOINT=http://localhost:4566  # LocalStack S3 endpoint
S3_BUCKET=instagrano-media  # S3 bucket name
REDIS_ADDR=localhost:6379   # Redis connection address
REDIS_PASSWORD=             # Redis password (empty for dev)
REDIS_DB=0                  # Redis database number (0-15)
CACHE_TTL=5m                # Cache TTL (e.g., 30s, 5m, 1h)
DEFAULT_PAGE_SIZE=20        # Default posts per page
MAX_PAGE_SIZE=100           # Maximum posts per page
```

**Testing:**
```bash
# Unit tests
go test ./...

# Integration tests (requires running server)
make itest

# Test coverage
go test -cover ./...
```

## Testing

The project includes comprehensive testing with **77.6% code coverage**:

### Test Architecture
- **Testcontainers**: Real PostgreSQL and Redis containers for integration tests
- **Gomega**: BDD-style assertions for readable test code
- **No Mocks**: Uses real dependencies for authentic testing
- **Structured Tests**: Organized by domain (auth, posts, feed, interactions)

### Test Categories

**Unit Tests:**
- Domain models (scoring algorithm, password validation)
- Service layer (auth, posts, feed, interactions)
- Configuration and logging utilities
- Pagination and caching logic

**Integration Tests:**
- Complete API endpoints with real database
- Authentication flows (register, login, JWT validation)
- Post creation and retrieval
- Feed generation with caching
- Like and comment interactions

**Test Commands:**
```bash
# Run all tests with coverage
make itest

# Run specific test categories
go test ./tests/ -run "TestAuth" -v
go test ./tests/ -run "TestFeed" -v
go test ./tests/ -run "TestPost" -v

# Generate coverage report
go test -coverprofile=coverage.out ./tests/ -coverpkg=./...
go tool cover -html=coverage.out
```

### Test Coverage Details
- **Domain Layer**: 100% coverage (scoring, validation)
- **Service Layer**: 95%+ coverage (auth, posts, feed, interactions)
- **Handler Layer**: 90%+ coverage (API endpoints)
- **Repository Layer**: 85%+ coverage (database operations)
- **Infrastructure**: 80%+ coverage (config, logging, caching)

### Testing Best Practices
- **Real Dependencies**: Uses Testcontainers for PostgreSQL and Redis
- **Isolation**: Each test gets fresh containers and database state
- **BDD Style**: Given-When-Then structure with Gomega assertions
- **Comprehensive**: Tests happy paths, error cases, and edge conditions
- **Performance**: Tests caching behavior and performance characteristics

## Performance Testing

The project includes K6 load tests to verify API performance under load and validate Redis caching improvements.

### K6 Test Scenarios

**Authentication Load Test:**
- Concurrent user registration and login
- JWT token generation throughput
- Authenticated endpoint access

**Feed Cache Performance Test:**
- Cold cache performance (database query)
- Warm cache performance (Redis hit)
- Validates 10x performance improvement

**Post Creation Load Test:**
- Concurrent post creation with file uploads
- S3/LocalStack integration under load

**Full User Journey Test:**
- Complete flow: Register → Login → Create Post → View Feed → Like → Comment

### Running K6 Tests

```bash
# Install K6 (if not already installed)
make k6-install

# Run all K6 tests
make k6-all

# Run specific test scenarios
make k6-auth      # Authentication load test
make k6-cache     # Feed cache performance test
make k6-posts     # Post creation load test
make k6-journey   # Full user journey test
```

### Performance Baselines

Based on testing with Redis caching enabled:

| Endpoint | Cold Cache | Warm Cache | Improvement |
|----------|------------|------------|-------------|
| GET /api/feed | ~200ms | ~20ms | 10x |
| POST /api/auth/login | ~50ms | N/A | N/A |
| POST /api/posts | ~300ms | N/A | N/A |

For detailed K6 test results and performance metrics, see [docs/PERFORMANCE.md](docs/PERFORMANCE.md).

For K6 test implementation details, see [tests/k6/README.md](tests/k6/README.md).

## API Documentation

Interactive Swagger UI available at: http://localhost:8081/swagger/

**Note:** The API runs on port 8080 and Swagger documentation runs on port 8081 for clean separation.

### Endpoints

**Authentication:**
- POST /api/auth/register - Register new user
- POST /api/auth/login - Login and get JWT token

**Posts:**
- POST /api/posts - Create post (multipart/form-data)
- GET /api/posts/:id - Get post by ID

**Feed:**
- GET /api/feed - Get paginated feed (cursor-based)
  - Query params: cursor (optional), limit (optional, default 20)

**Interactions:**
- POST /api/posts/:id/like - Like a post
- POST /api/posts/:id/comment - Comment on post

**System:**
- GET /health - Health check

### Swagger Generation

```bash
# Generate Swagger documentation
make swagger

# Start API server (port 8080)
make start

# Start Swagger UI server (port 8081) in another terminal
make swagger-ui

# Or start both together
make start-all

# Visit Swagger UI
open http://localhost:8081/swagger/
```

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

## Production Considerations

**Caching:**
- Use Redis Cluster for high availability and horizontal scaling
- Monitor cache hit rates and adjust TTL based on your data freshness requirements
- Consider event-based invalidation for critical real-time data
- Set up Redis persistence (AOF or RDB) for cache warm-up on restart

**Logging:**
- Use JSON format for log aggregation systems (ELK, Fluentd)
- Set appropriate log levels (info for production)
- Monitor request duration, error rates, and cache hit rates

**Pagination:**
- Cursor-based pagination prevents duplicate/skipped results
- More efficient for large datasets than OFFSET-based pagination
- Consistent ordering even with concurrent writes
- Caching works seamlessly with cursor-based pagination

**Performance:**
- Redis caching provides 10x speedup for cached feed requests
- Zap logger is 3-4x faster than standard library logging
- Zero-allocation logging reduces GC pressure
- Request correlation IDs enable distributed tracing
- Monitor cache memory usage and eviction policies
