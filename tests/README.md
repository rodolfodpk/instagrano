# Integration Tests for Instagrano

This directory contains comprehensive integration tests for the Instagrano application, specifically testing the JWT authentication and file upload functionality that was recently fixed.

## Overview

The integration tests verify:
- ✅ User registration and login
- ✅ JWT token generation and validation
- ✅ File upload functionality
- ✅ Feed access with authentication
- ✅ JWT token validation edge cases

## Running the Tests

### Prerequisites

1. **Start the server** with the correct JWT secret:
   ```bash
   JWT_SECRET='super-secret-key-for-testing' PORT=3007 go run cmd/api/main.go
   ```

2. **Ensure PostgreSQL is running** (via Docker Compose):
   ```bash
   docker-compose up -d
   ```

### Running Tests

#### Option 1: Using the test runner script
```bash
./run_integration_tests.sh
```

#### Option 2: Using go test directly
```bash
go test -v ./tests/
```

#### Option 3: Running specific test cases
```bash
go test -v ./tests/ -run "TestJWTIntegration/User_Registration_and_Login"
go test -v ./tests/ -run "TestJWTIntegration/File_Upload_and_Post_Creation"
go test -v ./tests/ -run "TestJWTIntegration/Feed_Access"
go test -v ./tests/ -run "TestJWTIntegration/JWT_Token_Validation_Edge_Cases"
```

## Test Structure

### Test Cases

1. **User Registration and Login**
   - Tests user registration with unique emails
   - Verifies login functionality
   - Tests JWT token generation
   - Validates `/me` endpoint authentication

2. **File Upload and Post Creation**
   - Tests multipart form data handling
   - Verifies file upload to local storage
   - Tests post creation with media
   - Validates response format

3. **Feed Access**
   - Tests authenticated feed access
   - Verifies feed response structure
   - Tests pagination parameters

4. **JWT Token Validation Edge Cases**
   - Tests invalid token handling
   - Tests malformed Authorization headers
   - Tests missing Authorization headers

### Test Data

- **Unique Test Users**: Each test run generates unique email addresses using timestamps
- **Test Files**: Creates temporary text files for upload testing
- **Error Handling**: Tests both success and failure scenarios

## Key Features Tested

### JWT Authentication Bug Fix
The tests specifically verify the fix for the JWT authentication bug where:
- Authorization headers are properly parsed
- JWT tokens are correctly validated
- The `strings.Split()` approach works correctly for token extraction

### File Upload Functionality
Tests verify:
- Multipart form data parsing
- File storage in `web/public/uploads/`
- Media URL generation
- Post creation with uploaded files

## Dependencies

- `github.com/stretchr/testify` - For assertions and test utilities
- Standard Go testing package
- HTTP client for API testing

## Test Output

Successful test run output:
```
=== RUN   TestJWTIntegration
=== RUN   TestJWTIntegration/User_Registration_and_Login
=== RUN   TestJWTIntegration/File_Upload_and_Post_Creation
=== RUN   TestJWTIntegration/Feed_Access
=== RUN   TestJWTIntegration/JWT_Token_Validation_Edge_Cases
--- PASS: TestJWTIntegration (0.31s)
    --- PASS: TestJWTIntegration/User_Registration_and_Login (0.15s)
    --- PASS: TestJWTIntegration/File_Upload_and_Post_Creation (0.07s)
    --- PASS: TestJWTIntegration/Feed_Access (0.07s)
    --- PASS: TestJWTIntegration/JWT_Token_Validation_Edge_Cases (0.01s)
PASS
```

## Troubleshooting

### Server Not Running
If you see "Server is not running", ensure:
1. The server is started with the correct JWT secret
2. The server is running on port 3007
3. PostgreSQL is running via Docker Compose

### Test Failures
- Check server logs for errors
- Verify database connectivity
- Ensure JWT secret matches between server and tests

## Contributing

When adding new tests:
1. Follow the existing test structure
2. Use unique test data to avoid conflicts
3. Test both success and failure scenarios
4. Update this README if adding new test categories
