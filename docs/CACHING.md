# Caching Guide

Redis caching strategy and operations for Instagrano.

## Overview

The application uses Redis for feed caching with performance improvements:

**Performance Impact:**
- **Cache Hit**: ~1-2ms response time
- **Cache Miss**: ~13-22ms response time (database query)
- **Expected Hit Rate**: 70-80% in steady state

## Cache Strategy

**Cache Key Format**: `feed:cursor:{cursor}:limit:{limit}`
- `cursor`: Base64-encoded cursor for pagination (empty string for first page)
- `limit`: Number of posts per page

**TTL**: 5 minutes (configurable via `CACHE_TTL`)
**Invalidation**: Time-based expiration (simple and predictable)

## Configuration

```bash
REDIS_ADDR=localhost:6379     # Redis connection address
REDIS_PASSWORD=               # Redis password (empty for dev)
REDIS_DB=0                    # Redis database number
CACHE_TTL=5m                  # Cache TTL (accepts: 30s, 5m, 1h, etc.)
```

## Cache Operations

### Connect to Redis CLI
```bash
make redis-cli
```

### View Cached Data
```bash
# View all cached feed keys
KEYS feed:*

# Get cache stats
INFO stats

# Monitor cache operations in real-time
MONITOR

# Flush all cached data
make redis-flush

# Check specific cache key
GET "feed:cursor::limit:20"
```

### Example Cache Keys
```bash
# First page, default size
feed:cursor::limit:20

# First page, custom size
feed:cursor::limit:10

# Subsequent pages
feed:cursor:MTc2MDk3MjI5OF8xNQ==:limit:20
```

## Structured Logging

Cache operations are logged with structured fields for monitoring:

```json
{
  "level": "info",
  "msg": "cache hit",
  "cache_key": "feed:cursor::limit:20",
  "duration": "1.2ms"
}
```

```json
{
  "level": "info",
  "msg": "cache miss - fetching from database",
  "cache_key": "feed:cursor::limit:20"
}
```

```json
{
  "level": "info",
  "msg": "cached result",
  "cache_key": "feed:cursor::limit:20",
  "ttl": 300
}
```

## Trade-offs

**Pros:**
- Reduced response time
- Reduced DB load
- Horizontally scalable

**Cons:**
- Data up to 5 minutes stale
- Additional service dependency
- Memory usage

## Monitoring

### Cache Hit Rate
Monitor the ratio of cache hits to total requests:
```bash
# Count cache hits
grep '"cache hit"' logs/app.log | wc -l

# Count cache misses
grep '"cache miss"' logs/app.log | wc -l
```

### Performance Metrics
```bash
# Find slow cache operations
grep '"duration":"[5-9][0-9]ms"' logs/app.log

# Monitor cache key patterns
grep '"cache_key"' logs/app.log | cut -d'"' -f4 | sort | uniq -c
```

## Production Considerations

- Use Redis Cluster for high availability and horizontal scaling
- Monitor cache hit rates and adjust TTL based on data freshness requirements
- Consider event-based invalidation for critical real-time data
- Set up Redis persistence (AOF or RDB) for cache warm-up on restart
- Monitor cache memory usage and eviction policies
