# K6 Performance Test Results

## Overview

This document contains comprehensive performance test results for the Instagrano API using K6 load testing. All tests validate API performance, caching effectiveness, and system reliability under load.

## Test Environment

- **Load Testing Tool**: K6 v0.47.0+
- **Test Duration**: 30-45 seconds per test (with graceful stop)
- **Virtual Users**: 3-5 VUs per test
- **API Base URL**: `http://localhost:8080`
- **Database**: PostgreSQL with Redis caching
- **Storage**: LocalStack S3-compatible storage

## Test Scenarios

### 1. Authentication Test (`auth.js`)

**Configuration**: 5 VUs, 30s duration

**Test Flow**: User registration → Login → Profile retrieval

**Results**:
```
✓ registration status is 200
✓ registration has user data
✓ login status is 200
✓ login returns token
✓ login returns user data
✓ me endpoint status is 200
✓ me endpoint returns user_id

checks.........................: 100.00% ✓ 350      ✗ 0
data_received..................: 48 kB   1.5 kB/s
data_sent......................: 39 kB   1.2 kB/s
✓ http_req_duration..............: avg=65.53ms min=832µs   med=78.6ms   max=223.41ms p(90)=109.29ms p(95)=125.21ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 150
http_reqs......................: 150     4.682497/s
iterations.....................: 50      1.560832/s
vus............................: 5      min=5      max=5
```

**Key Metrics**:
- **Success Rate**: 100% (350/350 checks passed)
- **Response Time**: p95 < 125ms ✅
- **Error Rate**: 0% ✅
- **Throughput**: 4.68 requests/second

---

### 2. Feed Cache Test (`feed-cache.js`)

**Configuration**: 3 VUs, 45s duration

**Test Flow**: User registration → Login → Post creation → Feed requests (cold/warm cache)

**Results**:
```
✓ cold feed status is 200
✓ cold feed has posts
✓ warm feed status is 200
✓ warm feed has posts

✓ cache_hit_duration.............: avg=5.555556 min=2       med=5        max=14        p(90)=10.2     p(95)=13.4
✓ cache_miss_duration............: avg=8.148148 min=1       med=5        max=23        p(90)=18.4     p(95)=22
checks.........................: 100.00% ✓ 108      ✗ 0
data_received..................: 261 kB  5.6 kB/s
data_sent......................: 58 kB   1.2 kB/s
✓ http_req_duration..............: avg=41.85ms  min=1.23ms  med=24.38ms  max=111.31ms  p(90)=89.22ms  p(95)=100.7ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 135
http_reqs......................: 135     2.875849/s
iterations.....................: 27      0.57517/s
```

**Cache Performance Analysis**:
- **Cache Hit Duration**: avg 5.56ms, p95 13.4ms ✅ (threshold: p95 < 100ms)
- **Cache Miss Duration**: avg 8.15ms, p95 22ms ✅ (threshold: p95 < 500ms)
- **Cache Effectiveness**: Both hit and miss times are very fast due to local Redis
- **Success Rate**: 100% (108/108 checks passed)

---

### 3. Post Creation (File Upload) Test (`posts.js`)

**Configuration**: 3 VUs, 30s duration

**Test Flow**: User registration → Login → File upload post creation

**Results**:
```
✓ post creation status is 201
✓ post has ID
✓ post has media_url

checks.........................: 100.00% ✓ 90       ✗ 0
data_received..................: 37 kB   1.1 kB/s
data_sent......................: 47 kB   1.5 kB/s
✓ http_req_duration..............: avg=65.71ms  min=3.98ms  med=68.8ms   max=115.27ms  p(90)=96.12ms  p(95)=100.48ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 90
http_reqs......................: 90      2.809534/s
iterations.....................: 30      0.936511/s
```

**Key Metrics**:
- **Success Rate**: 100% (90/90 checks passed)
- **Response Time**: p95 < 100ms ✅
- **File Upload Performance**: ~66ms average for multipart uploads
- **Throughput**: 2.81 requests/second

---

### 4. Post Creation (URL) Test (`posts-url.js`)

**Configuration**: 3 VUs, 30s duration

**Test Flow**: User registration → Login → URL-based post creation

**Results**:
```
✓ registration successful
✓ login successful
✓ post from URL status is 201
✓ post has ID
✓ post has media_url

checks.........................: 100.00% ✓ 150      ✗ 0
data_received..................: 35 kB   1.1 kB/s
data_sent......................: 36 kB   1.1 kB/s
✓ http_req_duration..............: avg=72.61ms min=5.94ms  med=68.34ms  max=237.11ms  p(90)=106.3ms   p(95)=120.55ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 90
http_reqs......................: 90      2.791945/s
iterations.....................: 30      0.930648/s
```

**Key Metrics**:
- **Success Rate**: 100% (150/150 checks passed) ✅
- **Response Time**: p95 < 121ms ✅
- **URL Processing**: ~73ms average for URL download + S3 upload
- **Throughput**: 2.79 requests/second

---

### 5. Post Retrieval Cache Test (`post-retrieval.js`)

**Configuration**: 5 VUs, 30s duration

**Test Flow**: User registration → Login → Post creation → Post retrieval (cold/warm cache)

**Results**:
```
✓ cold post status 200
✓ warm post status 200

✓ cache_hit_duration.............: avg=1.918736 min=0        med=1        max=29        p(90)=3        p(95)=7.75
✓ cache_miss_duration............: avg=5.557562 min=1        med=3        max=52        p(90)=13       p(95)=18.75
checks.........................: 100.00% ✓ 1772      ✗ 0
data_received..................: 1.9 MB  61 kB/s
data_sent......................: 1.6 MB  53 kB/s
✓ http_req_duration..............: avg=34.1ms   min=404µs    med=13.8ms   max=412.9ms   p(90)=74.89ms  p(95)=80.84ms
✓ http_req_failed................: 0.00%  ✓ 0          ✗ 4430
http_reqs......................: 4430    145.32451/s
iterations.....................: 886     29.064902/s
```

**Cache Performance Analysis**:
- **Cache Hit Duration**: avg 1.92ms, p95 7.75ms ✅ (threshold: p95 < 50ms)
- **Cache Miss Duration**: avg 5.56ms, p95 18.75ms ✅ (threshold: p95 < 200ms)
- **Cache Improvement**: **65% reduction** (1.92ms vs 5.56ms)
- **High Throughput**: 145 requests/second, 886 iterations
- **Success Rate**: 100% (1772/1772 checks passed)

---

### 6. User Journey Test (`user-journey.js`)

**Configuration**: 3 VUs, 45s duration

**Test Flow**: Complete user workflow - registration → login → post creation → feed viewing → interactions

**Results**:
```
✓ register status is 200
✓ login status is 200
✓ post creation status is 201
✓ feed status is 200
✓ feed has posts
✓ like status is 200
✓ comment status is 200

checks.........................: 100.00% ✓ 168      ✗ 0
data_received..................: 137 kB  2.7 kB/s
data_sent......................: 62 kB   1.2 kB/s
✓ http_req_duration..............: avg=49.61ms  min=1.87ms  med=33.38ms  max=207.2ms   p(90)=98.5ms   p(95)=103.2ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 144
http_reqs......................: 144     2.852502/s
iterations.....................: 24      0.475417/s
```

**Key Metrics**:
- **Success Rate**: 100% (168/168 checks passed)
- **Response Time**: p95 < 103ms ✅
- **Complete Workflow**: All user interactions working correctly
- **Throughput**: 2.85 requests/second

---

## Performance Summary

### Overall Results
- **Total Tests**: 6 scenarios
- **Success Rate**: 100% across all tests
- **Error Rate**: 0% across all tests
- **Response Time**: All p95 < 125ms ✅

### Cache Performance Highlights
- **Post Retrieval Cache**: 65% performance improvement (1.92ms vs 5.56ms)
- **Feed Cache**: Both hit and miss times under 22ms
- **Cache Thresholds**: All cache performance thresholds met

### Key Achievements
1. **Zero Failures**: All 6 test scenarios pass with 100% success rate
2. **Response Times**: p95 response times consistently under 125ms
3. **Effective Caching**: Significant performance improvements with Redis caching
4. **Dual Upload Support**: Both file upload and URL-based upload working correctly
5. **High Throughput**: Up to 145 requests/second for cached operations

### Test Files Reference
- [Authentication Test](../../tests/k6/scenarios/auth.js)
- [Feed Cache Test](../../tests/k6/scenarios/feed-cache.js)
- [Post Creation (File) Test](../../tests/k6/scenarios/posts.js)
- [Post Creation (URL) Test](../../tests/k6/scenarios/posts-url.js)
- [Post Retrieval Cache Test](../../tests/k6/scenarios/post-retrieval.js)
- [User Journey Test](../../tests/k6/scenarios/user-journey.js)

### Running Tests
```bash
# Run individual tests
make k6-auth
make k6-cache
make k6-posts
make k6-posts-url
make k6-post-retrieval
make k6-journey

# Run all tests
make k6-all
```

---

*Last Updated: October 2024*
*Test Results: All tests passing with 100% success rate*
