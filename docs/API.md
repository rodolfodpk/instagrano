# API Reference

Complete API documentation for Instagrano endpoints.

## Authentication Endpoints

### Register User
```bash
POST /api/auth/register
Content-Type: application/json

{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

### Login
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepassword"
}

# Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com"
  }
}
```

### Get Current User
```bash
GET /api/auth/me
Authorization: Bearer <token>
```

## Real-time Events

### WebSocket Connection
```bash
GET /api/events/ws?token=<jwt_token>
```

**Connection**: WebSocket upgrade with JWT authentication

**Event Types**:
- `like` - When a post is liked
- `unlike` - When a post is unliked  
- `comment` - When a comment is added

**Example JavaScript Client**:
```javascript
const token = localStorage.getItem('jwt_token');
const ws = new WebSocket(`ws://localhost:8080/api/events/ws?token=${token}`);

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received event:', data);
  
  switch(data.type) {
    case 'like':
      updateLikeCount(data.post_id, data.likes_count);
      break;
    case 'comment':
      addCommentToPost(data.post_id, data.comment);
      break;
  }
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
  // Implement reconnection logic
};
```

## Post Endpoints

### Create Post (File Upload)
```bash
POST /api/posts
Authorization: Bearer <token>
Content-Type: multipart/form-data

title: "My Post"
caption: "Uploaded file"
media_type: "image"
media: <file>
```

### Create Post (URL-based)
```bash
POST /api/posts
Authorization: Bearer <token>
Content-Type: multipart/form-data

title: "My Post"
caption: "From URL"
media_url: "https://example.com/image.jpg"
```

### Get Post
```bash
GET /api/posts/:id
Authorization: Bearer <token>
```

## Feed Endpoints

### Get Feed (Cursor-based - Recommended)
```bash
# First page with default size (20 posts)
GET /api/feed
Authorization: Bearer <token>

# Custom page size
GET /api/feed?limit=10
Authorization: Bearer <token>

# Next page using cursor from previous response
GET /api/feed?limit=10&cursor=MTc2MDk3MjI5OF8xNQ==
Authorization: Bearer <token>
```

### Get Feed (Page-based - Legacy)
```bash
# Default page size
GET /api/feed?page=1
Authorization: Bearer <token>

# Custom page size
GET /api/feed?page=1&limit=50
Authorization: Bearer <token>
```

**Response Format (cursor-based):**
```json
{
  "posts": [
    {
      "id": 1,
      "title": "My Post",
      "caption": "Post caption",
      "media_url": "http://localhost:4566/instagrano-media/posts/1234567890-image.jpg",
      "media_type": "image",
      "user_id": 1,
      "created_at": "2024-01-15T10:30:00Z",
      "likes_count": 5,
      "comments_count": 2,
      "views_count": 10
    }
  ],
  "next_cursor": "MTc2MDk3MjI5OF8xNQ==",
  "has_more": true
}
```

**Configuration:**
- `DEFAULT_PAGE_SIZE`: Default posts per page (default: 20)
- `MAX_PAGE_SIZE`: Maximum allowed page size (default: 100)
- Frontend dropdown: 3, 5, 10, 20, 50 posts per page

## Interaction Endpoints

### Like Post
```bash
POST /api/posts/:id/like
Authorization: Bearer <token>
```

### Comment on Post
```bash
POST /api/posts/:id/comment
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Great post!"
}
```

### Start View Tracking
```bash
POST /api/posts/:id/view/start
Authorization: Bearer <token>

# Response:
{
  "id": 1,
  "user_id": 1,
  "post_id": 5,
  "started_at": "2024-01-15T10:30:00Z"
}
```

### End View Tracking
```bash
POST /api/posts/:id/view/end
Authorization: Bearer <token>
Content-Type: application/json

{
  "started_at": "2024-01-15T10:30:00Z"
}

# Response:
{
  "message": "view ended"
}
```

## System Endpoints

### Health Check
```bash
GET /health
```

## Swagger Documentation

Interactive Swagger UI is available at: http://localhost:8081/swagger/

**Note:** The API runs on port 8080 and Swagger documentation runs on port 8081 for clean separation.

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

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created (for new resources)
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (invalid/missing token)
- `404` - Not Found
- `500` - Internal Server Error
