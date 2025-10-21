# Features Guide

Detailed guide to Instagrano's key features including view tracking, pagination, logging, and frontend capabilities.

## View Time Tracking

The application automatically tracks how long users spend viewing each post using Intersection Observer API.

### How it Works
- **Automatic Detection**: Posts are tracked when 50%+ visible for 1+ second
- **Duration Calculation**: Time between view start and end is recorded
- **Database Storage**: View sessions stored in `post_views` table
- **Counter Updates**: `views_count` incremented for each view start

### API Endpoints
```bash
# Start tracking (called automatically by frontend)
POST /api/posts/:id/view/start
# Returns: {"id": 1, "user_id": 1, "post_id": 5, "started_at": "2024-01-15T10:30:00Z"}

# End tracking (called automatically by frontend)  
POST /api/posts/:id/view/end
# Body: {"started_at": "2024-01-15T10:30:00Z"}
# Returns: {"message": "view ended"}
```

### Frontend Implementation
- Uses Intersection Observer to detect viewport visibility
- Tracks active view sessions in Alpine.js state
- Automatically ends views on page unload
- Handles multiple views of same post gracefully

### Database Schema
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

## Feed Pagination

The feed endpoint supports both cursor-based and page-based pagination with configurable page sizes:

### Cursor-based Pagination (Recommended)
```bash
# First page with default size (20 posts)
GET /api/feed

# Custom page size
GET /api/feed?limit=10

# Next page using cursor from previous response
GET /api/feed?limit=10&cursor=MTc2MDk3MjI5OF8xNQ==
```

### Page-based Pagination (Legacy)
```bash
# Default page size
GET /api/feed?page=1

# Custom page size
GET /api/feed?page=1&limit=50
```

### Configuration
- `DEFAULT_PAGE_SIZE`: Default posts per page (default: 20)
- `MAX_PAGE_SIZE`: Maximum allowed page size (default: 100)
- Frontend dropdown: 3, 5, 10, 20, 50 posts per page

### Response Format (cursor-based)
```json
{
  "posts": [...],
  "next_cursor": "MTc2MDk3MjI5OF8xNQ==",
  "has_more": true
}
```

### Benefits of Cursor-based Pagination
- Prevents duplicate/skipped results
- More efficient for large datasets than OFFSET-based pagination
- Consistent ordering even with concurrent writes
- Caching works seamlessly with cursor-based pagination

## Structured Logging

The application uses Zap for structured logging:

### Configuration
- `LOG_LEVEL`: debug, info, warn, error (default: info)
- `LOG_FORMAT`: json, console (default: json)

### Log Examples
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

### Log Fields
- `request_id`: Unique identifier for request tracing
- `user_id`: Authenticated user ID (when available)
- `method`, `path`: HTTP method and path
- `status`: HTTP response status
- `duration`: Request processing time
- `response_size`: Response body size in bytes
- `cache_key`: Redis cache key (when caching)
- `cache_hit`/`cache_miss`: Cache operation result

### Example Log Entries
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "caller": "middleware/logger.go:26",
  "msg": "request started",
  "request_id": "abc123",
  "method": "GET",
  "path": "/api/feed",
  "ip": "127.0.0.1",
  "user_agent": "Mozilla/5.0..."
}
```

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "caller": "middleware/logger.go:62",
  "msg": "request completed",
  "request_id": "abc123",
  "method": "GET",
  "path": "/api/feed",
  "status": 200,
  "duration": "15.2ms",
  "response_size": 1024
}
```

## Frontend Features

The frontend is built with Alpine.js and includes:

### Authentication
- **Login/Registration**: Tabbed interface for user authentication
- **JWT Token Management**: Automatic token storage and refresh
- **Protected Routes**: Redirects to login when not authenticated

### Feed Interface
- **Load More Pagination**: Cursor-based pagination with "Load More" button
- **Page Size Selection**: Dropdown to choose posts per page (3, 5, 10, 20, 50)
- **Real-time Updates**: SSE connection for live feed updates
- **Responsive Design**: Works on desktop and mobile devices

### Post Creation
- **Dual Upload Support**: 
  - File upload with drag-and-drop
  - URL-based media upload
- **Media Types**: Support for images and videos
- **Form Validation**: Client-side validation for required fields
- **Preview**: Image preview before upload

### Interactions
- **Like System**: One-click like/unlike with real-time count updates
- **Comments**: Add comments with real-time display
- **View Tracking**: Automatic view time tracking using Intersection Observer
- **Real-time Updates**: Live updates via Server-Sent Events

### User Experience
- **Loading States**: Visual feedback during API calls
- **Error Handling**: User-friendly error messages
- **Auto-save**: Form data persistence during session
- **Keyboard Shortcuts**: Quick navigation and actions

## Performance Features

- **Redis Caching**: Feed responses cached for improved performance
- **Lazy Loading**: Images loaded as they enter viewport
- **Efficient Pagination**: Cursor-based pagination prevents performance degradation
- **Optimized Queries**: Database queries optimized with proper indexing
- **CDN Ready**: S3-compatible storage for media files
