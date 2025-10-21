# K6 Performance Load Tests

This directory contains K6 load tests for the Instagrano API to verify performance under load and validate Redis caching improvements.

## Prerequisites

### Install K6

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```bash
choco install k6
```

### Create Test Fixture

Create a test image file for file upload tests:

```bash
mkdir -p tests/k6/fixtures
# Create a simple test image (or use an existing one)
# You can download a small test image or create one
```

## Running Tests

### Using Makefile

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

### Direct K6 Commands

```bash
# Authentication test
k6 run tests/k6/scenarios/auth.js

# Cache performance test
k6 run tests/k6/scenarios/feed-cache.js

# Post creation test
k6 run tests/k6/scenarios/posts.js

# User journey test
k6 run tests/k6/scenarios/user-journey.js
```

### Custom API URL

```bash
API_URL=http://localhost:8080 k6 run tests/k6/scenarios/auth.js
```

## Test Scenarios

### 1. Authentication Load Test (`auth.js`)

Tests concurrent user registration and login flows:
- Multiple concurrent user registrations
- JWT token generation
- Authenticated `/me` endpoint access

**Expected Results:**
- Response time p95 < 500ms
- Error rate < 1%
- Throughput > 50 RPS

### 2. Feed Cache Performance Test (`feed-cache.js`)

Validates Redis caching performance improvements:
- **Cold Cache**: First request (database query)
- **Warm Cache**: Subsequent requests (Redis hit)
- Compares response times to verify 10x improvement

**Expected Results:**
- Cold cache: p95 < 500ms
- Warm cache: p95 < 100ms
- Cache improvement: > 5x speedup

### 3. Post Creation Load Test (`posts.js`)

Tests concurrent post creation with file uploads:
- Multiple concurrent post creations
- S3/LocalStack integration under load
- File upload performance

**Expected Results:**
- Response time p95 < 1000ms
- Error rate < 1%
- Throughput > 20 RPS

### 4. Full User Journey Test (`user-journey.js`)

Complete user flow simulation:
- Register → Login → Create Post → View Feed → Like → Comment

**Expected Results:**
- All steps complete successfully
- Error rate < 1%
- Average journey time < 5s

## Test Results

Results are saved in `tests/k6/results/` directory:
- `auth-summary.json` - Authentication test results
- `cache-summary.json` - Cache performance results
- `posts-summary.json` - Post creation results
- `journey-summary.json` - User journey results

## Performance Baselines

Based on testing with Redis caching enabled:

| Endpoint | Cold Cache | Warm Cache | Improvement |
|----------|------------|------------|-------------|
| GET /api/feed | ~200ms | ~20ms | 10x |
| POST /api/auth/login | ~50ms | N/A | N/A |
| POST /api/posts | ~300ms | N/A | N/A |

## Troubleshooting

### Tests failing with connection errors

Ensure the API server is running:
```bash
make start
```

### Cache tests not showing improvement

Verify Redis is running:
```bash
docker-compose exec redis redis-cli ping
```

### File upload tests failing

Ensure test fixture exists:
```bash
ls -la tests/k6/fixtures/test-image.png
```

## Continuous Integration

K6 tests can be integrated into CI/CD pipelines:

```yaml
# .github/workflows/performance.yml
- name: Run K6 Performance Tests
  run: |
    make docker-up
    make migrate
    make start &
    sleep 5
    make k6-all
```

