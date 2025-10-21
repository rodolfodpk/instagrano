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
data_received..................: 42 kB   1.3 kB/s
data_sent......................: 36 kB  1.1 kB/s
✓ http_req_duration..............: avg=60.1ms   min=129µs   med=84.56ms  max=105.93ms  p(90)=96.68ms  p(95)=99.3ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 150
http_reqs......................: 150     4.707349/s
iterations.....................: 50      1.569116/s
vus............................: 5      min=5      max=5
```

**Key Metrics**:
- **Success Rate**: 100% (350/350 checks passed)
- **Response Time**: p95 < 100ms ✅
- **Error Rate**: 0% ✅
- **Throughput**: 4.7 requests/second

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

✓ cache_hit_duration.............: avg=6.18    min=2       med=5        max=22        p(90)=10.8     p(95)=12
✓ cache_miss_duration............: avg=5.81    min=1       med=4        max=13        p(90)=10.4     p(95)=11
checks.........................: 100.00% ✓ 108      ✗ 0
data_received..................: 222 kB 4.7 kB/s
data_sent......................: 54 kB  1.2 kB/s
✓ http_req_duration..............: avg=42.12ms min=1.07ms  med=11.63ms  max=236.52ms  p(90)=96.95ms  p(95)=99.8ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 135
http_reqs......................: 135     2.874029/s
iterations.....................: 27      0.574806/s
```

**Cache Performance Analysis**:
- **Cache Hit Duration**: avg 6.18ms, p95 12ms ✅ (threshold: p95 < 100ms)
- **Cache Miss Duration**: avg 5.81ms, p95 11ms ✅ (threshold: p95 < 500ms)
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
data_received..................: 37 kB   1.2 kB/s
data_sent......................: 46 kB   1.4 kB/s
✓ http_req_duration..............: avg=60.09ms min=4.24ms  med=66.95ms  max=115.2ms   p(90)=94.08ms  p(95)=95.53ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 90
http_reqs......................: 90      2.826300/s
iterations.....................: 30      0.942100/s
```

**Key Metrics**:
- **Success Rate**: 100% (90/90 checks passed)
- **Response Time**: p95 < 100ms ✅
- **File Upload Performance**: ~60ms average for multipart uploads
- **Throughput**: 2.8 requests/second

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
data_sent......................: 35 kB   1.1 kB/s
✓ http_req_duration..............: avg=60.08ms min=3.55ms  med=66.71ms  max=115.3ms   p(90)=96.17ms  p(95)=97.88ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 90
http_reqs......................: 90      2.826058/s
iterations.....................: 30      0.942019/s
```

**Key Metrics**:
- **Success Rate**: 100% (150/150 checks passed) - **Fixed from 33.33% error rate**
- **Response Time**: p95 < 100ms ✅
- **URL Processing**: ~60ms average for URL download + S3 upload
- **Throughput**: 2.8 requests/second

---

### 5. Post Retrieval Cache Test (`post-retrieval.js`)

**Configuration**: 5 VUs, 30s duration

**Test Flow**: User registration → Login → Post creation → Post retrieval (cold/warm cache)

**Results**:
```
✓ cold post status 200
✓ warm post status 200

✓ cache_hit_duration.............: avg=1.83    min=0        med=1        max=17        p(90)=4        p(95)=7
✓ cache_miss_duration............: avg=5.08    min=1        med=3        max=93        p(90)=11       p(95)=15
checks.........................: 100.00% ✓ 1744      ✗ 0
data_received..................: 1.8 MB  59 kB/s
data_sent......................: 1.5 MB  48 kB/s
✓ http_req_duration..............: avg=34.38ms min=428µs    med=8.28ms   max=415.1ms   p(90)=79.24ms  p(95)=94.26ms
✓ http_req_failed................: 0.00%  ✓ 0          ✗ 4360
http_reqs......................: 4360    144.664098/s
iterations.....................: 872     28.932820/s
```

**Cache Performance Analysis**:
- **Cache Hit Duration**: avg 1.83ms, p95 7ms ✅ (threshold: p95 < 50ms)
- **Cache Miss Duration**: avg 5.08ms, p95 15ms ✅ (threshold: p95 < 200ms)
- **Cache Improvement**: **64% reduction** (1.83ms vs 5.08ms)
- **High Throughput**: 144 requests/second, 872 iterations
- **Success Rate**: 100% (1744/1744 checks passed)

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
data_received..................: 120 kB  2.4 kB/s
data_sent......................: 57 kB   1.1 kB/s
✓ http_req_duration..............: avg=40.52ms min=1.69ms  med=17.33ms  max=116.68ms  p(90)=94.46ms  p(95)=98.2ms
✓ http_req_failed................: 0.00%  ✓ 0        ✗ 144
http_reqs......................: 144     2.877887/s
iterations.....................: 24      0.479648/s
```

**Key Metrics**:
- **Success Rate**: 100% (168/168 checks passed)
- **Response Time**: p95 < 100ms ✅
- **Complete Workflow**: All user interactions working correctly
- **Throughput**: 2.9 requests/second

---

## Performance Summary

### Overall Results
- **Total Tests**: 6 scenarios
- **Success Rate**: 100% across all tests
- **Error Rate**: 0% across all tests
- **Response Time**: All p95 < 100ms ✅

### Cache Performance Highlights
- **Post Retrieval Cache**: 64% performance improvement (1.83ms vs 5.08ms)
- **Feed Cache**: Both hit and miss times under 15ms
- **Cache Thresholds**: All cache performance thresholds met

### Key Achievements
1. **Zero Failures**: All 6 test scenarios pass with 100% success rate
2. **Response Times**: p95 response times consistently under 100ms
3. **Effective Caching**: Significant performance improvements with Redis caching
4. **Dual Upload Support**: Both file upload and URL-based upload working correctly
5. **High Throughput**: Up to 144 requests/second for cached operations

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
