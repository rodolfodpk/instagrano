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
- WebSocket real-time updates
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

# Clean all data (Redis, Database, S3)
make clean-all
```

## Documentation

- [Getting Started Guide](docs/GETTING_STARTED.md) - Setup and first steps
- [API Reference](docs/API.md) - Endpoints and Swagger
- [Caching Guide](docs/CACHING.md) - Redis strategy and operations
- [Testing Guide](docs/TESTING.md) - Tests and CI/CD
- [Features Guide](docs/FEATURES.md) - View tracking, logging, pagination
- [Performance Results](docs/PERFORMANCE.md) - K6 load tests
- [Production Guide](docs/PRODUCTION.md) - Deployment best practices

## API Endpoints

- `GET /health` - Health check
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/auth/me` - Get current user (requires JWT)
- `GET /api/events/ws` - WebSocket connection for real-time events (requires JWT)
- `POST /api/posts` - Create post with file upload or URL (requires JWT)
- `GET /api/posts/:id` - Get specific post (requires JWT)
- `POST /api/posts/:id/like` - Like a post (requires JWT)
- `POST /api/posts/:id/comment` - Comment on a post (requires JWT)
- `POST /api/posts/:id/view/start` - Start tracking view time (requires JWT)
- `POST /api/posts/:id/view/end` - End tracking and record duration (requires JWT)
- `GET /api/feed` - Get user feed (requires JWT)

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend       │    │   Go Fiber      │    │   PostgreSQL    │
│   (Alpine.js)    │◄──►│   HTTP Server   │◄──►│   Database      │
│   WebSocket      │◄──►│   WebSocket     │    └─────────────────┘
└─────────────────┘    └─────────────────┘
                                │
                       ┌────────┴────────┐
                       ▼                 ▼
              ┌─────────────────┐ ┌─────────────────┐
              │   LocalStack    │ │     Redis       │
              │   S3 Storage    │ │     Cache       │
              └─────────────────┘ └─────────────────┘
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API server port | `8080` |
| `SWAGGER_PORT` | Swagger UI server port | `8081` |
| `JWT_SECRET` | JWT signing secret | Required |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `LOG_FORMAT` | Log format (json, console) | `json` |
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `S3_ENDPOINT` | LocalStack S3 endpoint | `http://localhost:4566` |
| `S3_BUCKET` | S3 bucket name | `instagrano-media` |
| `REDIS_ADDR` | Redis connection address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password (empty for dev) | `` |
| `REDIS_DB` | Redis database number (0-15) | `0` |
| `CACHE_TTL` | Cache TTL (e.g., 30s, 5m, 1h) | `5m` |
| `DEFAULT_PAGE_SIZE` | Default posts per page | `20` |
| `MAX_PAGE_SIZE` | Maximum posts per page | `100` |