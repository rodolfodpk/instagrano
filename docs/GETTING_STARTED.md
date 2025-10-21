# Getting Started Guide

This guide will help you set up and run Instagrano locally for development.

## Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Make (optional, for convenience commands)

## Quick Setup

1. **Start services:**
   ```bash
   make docker-up
   ```

2. **Run database migrations:**
   ```bash
   make migrate
   ```

3. **Start the application:**
   ```bash
   make start
   ```

The API will be available at `http://localhost:8080` and the frontend at `http://localhost:8080/feed.html`.

## Frontend Access

Open: http://localhost:8080/feed.html (after logging in at http://localhost:8080/)

The frontend includes:
- **Login/Registration**: Tabbed interface for user authentication
- **Feed with Load More**: Cursor-based pagination with "Load More" button
- **Post Creation**: File upload with image/video support or URL-based media
- **Interactions**: Like and comment functionality
- **View Time Tracking**: Automatic tracking using Intersection Observer

## Post Creation Examples

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

## First Steps

1. **Register a new user** via the frontend or API
2. **Login** to get your JWT token
3. **Create your first post** using either file upload or URL
4. **View your feed** to see posts
5. **Interact** with posts by liking and commenting

## Troubleshooting

- **Port conflicts**: Make sure ports 8080, 8081, 5432, 6379, and 4566 are available
- **Docker issues**: Run `make clean` to reset all containers and volumes
- **Database connection**: Ensure PostgreSQL container is running with `docker-compose ps`
- **Redis connection**: Check Redis container status and logs

For more detailed information, see the [API Reference](API.md) and [Features Guide](FEATURES.md).
