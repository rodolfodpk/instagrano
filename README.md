# 🚀 Instagrano MVP

A mini Instagram API built with Go Fiber, PostgreSQL, LocalStack S3, and JWT authentication.

## Features

- **Structured Logging**: High-performance Zap logger with JSON output and request correlation IDs
- **Cursor-Based Pagination**: Efficient pagination for large datasets with consistent results
- **S3 Integration**: LocalStack S3-compatible storage for media files with CORS support
- **JWT Authentication**: Secure token-based authentication
- **Real-time Feed**: Hybrid scoring algorithm combining time decay and engagement metrics
- **File Upload**: Support for images and videos with proper content type validation
- **Interactive Frontend**: Alpine.js-powered UI with "Load More" functionality

## Quick Start

```bash
# Start database AND LocalStack S3
make docker-up

# Verify LocalStack is running
curl http://localhost:4566/_localstack/health

# Run migrations
make migrate

# Start server (will connect to LocalStack)
make start

# Run integration tests (in another terminal)
make itest
```

## API Endpoints

- `GET /health` - Health check
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/me` - Get current user (requires JWT)
- `POST /api/posts` - Create post with file upload (requires JWT)
- `GET /api/posts/:id` - Get specific post (requires JWT)
- `POST /api/posts/:id/like` - Like a post (requires JWT)
- `POST /api/posts/:id/comment` - Comment on a post (requires JWT)
- `GET /api/feed` - Get user feed (requires JWT)

### Feed Pagination

The feed endpoint supports both cursor-based and page-based pagination:

**Cursor-based (recommended):**
```bash
# First page
GET /api/feed?limit=10

# Next page using cursor from previous response
GET /api/feed?limit=10&cursor=MTc2MDk3MjI5OF8xNQ==
```

**Page-based (legacy):**
```bash
GET /api/feed?page=1&limit=10
```

**Response format (cursor-based):**
```json
{
  "posts": [...],
  "next_cursor": "MTc2MDk3MjI5OF8xNQ==",
  "has_more": true
}
```

## Frontend

Open: http://localhost:3007/feed.html (after logging in at http://localhost:3007/)

The frontend includes:
- **Login/Registration**: Tabbed interface for user authentication
- **Feed with Load More**: Cursor-based pagination with "Load More" button
- **Post Creation**: File upload with image/video support
- **Interactions**: Like and comment functionality

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
```

**Log Fields:**
- `request_id`: Unique identifier for request tracing
- `user_id`: Authenticated user ID (when available)
- `method`, `path`: HTTP method and path
- `status`: HTTP response status
- `duration`: Request processing time
- `response_size`: Response body size in bytes

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend       │    │   Go Fiber      │    │   PostgreSQL    │
│   (Alpine.js)    │◄──►│   HTTP Server   │◄──►│   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   LocalStack    │
                       │   S3 Storage    │
                       └─────────────────┘
```

## Development

**Environment Variables:**
```bash
PORT=3007                    # Server port
JWT_SECRET=your-secret-key   # JWT signing secret
LOG_LEVEL=info              # Log level
LOG_FORMAT=json             # Log format
DATABASE_URL=postgres://... # PostgreSQL connection string
S3_ENDPOINT=http://localhost:4566  # LocalStack S3 endpoint
S3_BUCKET=instagrano-media  # S3 bucket name
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

## Production Considerations

**Logging:**
- Use JSON format for log aggregation systems (ELK, Fluentd)
- Set appropriate log levels (info for production)
- Monitor request duration and error rates

**Pagination:**
- Cursor-based pagination prevents duplicate/skipped results
- More efficient for large datasets than OFFSET-based pagination
- Consistent ordering even with concurrent writes

**Performance:**
- Zap logger is 3-4x faster than standard library logging
- Zero-allocation logging reduces GC pressure
- Request correlation IDs enable distributed tracing
