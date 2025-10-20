# 🚀 Instagrano MVP

A mini Instagram API built with Go Fiber, PostgreSQL, LocalStack S3, and JWT authentication.

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

## Frontend

Open: http://localhost:3007/feed.html (after logging in at http://localhost:3007/)
