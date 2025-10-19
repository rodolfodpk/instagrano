# Instagrano Tests

This directory contains comprehensive tests for the Instagrano API.

## Test Structure

- **Unit Tests**: Basic functionality tests that don't require a running server
- **Integration Tests**: Full API endpoint tests that require the server to be running

## Running Tests

### Unit Tests Only
```bash
go test ./tests/... -v -ginkgo.focus="Unit"
```

### Integration Tests Only
```bash
# First, start the server in one terminal:
go run ./cmd/app

# Then run integration tests in another terminal:
go test ./tests/... -v -ginkgo.focus="Integration"
```

### All Tests
```bash
go test ./tests/... -v
```

## Test Coverage

### Unit Tests
- ✅ Basic functionality validation
- ✅ Math operations
- ✅ String operations

### Integration Tests
- ✅ **Home endpoint** (`/`)
- ✅ **Authentication**
  - User registration (`/register`)
  - Duplicate user rejection
  - User login (`/login`)
  - Invalid credentials rejection
- ✅ **Post Management**
  - Image upload (`/upload`)
  - Upload validation
  - Feed retrieval (`/feed`)
- ✅ **Social Features**
  - Post liking (`/posts/:id/like`)
  - Non-existent post handling
  - Comment addition (`/posts/:id/comments`)
  - Comment validation
  - Invalid post ID handling
- ✅ **API Documentation**
  - Swagger UI (`/docs`)
  - Swagger JSON (`/docs/doc.json`)

## Test Framework

Tests use:
- **Ginkgo**: BDD testing framework
- **Gomega**: Matcher library
- **HTTP Client**: For integration testing

## Notes

- Integration tests require the server to be running on `localhost:3000`
- Tests include proper error handling validation
- All API endpoints are covered with both success and failure scenarios
- Tests validate response codes, content types, and response bodies
