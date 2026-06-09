# ADR 003: Read-Through Cache with Redis

## Status

Accepted

## Context

We needed to optimize the `GET /:code` redirect path. This is the most latency-sensitive endpoint — it should respond as fast as possible (<5ms p99).

## Options

1. **Cache-Aside (read-through)**: Application checks Redis first, then DB on miss, writes to Redis on miss.

2. **Write-Through (write-around)**: Always write to DB, then to cache. Simpler but cache may contain stale data.

3. **No cache**: DB only. Simplest but slowest.

## Decision

We chose **Read-Through Cache** with Redis because:

- The `GET /:code` path is read-heavy (each redirect reads once, writes analytics asynchronously)
- Redis consistently delivers sub-millisecond reads
- On cache miss, we populate the cache synchronously, ensuring subsequent reads hit the cache
- TTL of 24 hours prevents stale data while keeping memory usage bounded
- Redis Streams additionally serve as a buffer for analytics writes (write-behind pattern)

## Performance Impact

| Scenario | Latency (p99) |
|---|---|
| No cache (DB only) | ~5-10ms |
| Read-through cache (Redis hit) | <1ms |
| Read-through cache (Miss + DB) | ~5-10ms + cache update |

## Consequences

- Adds Redis as a dependency (handled by Docker Compose)
- Cache invalidation is handled by TTL (not explicit invalidation)
- Cache hit/miss metrics are exposed via Prometheus for monitoring
- Redis streams provide durability for analytics events (survive restarts)
