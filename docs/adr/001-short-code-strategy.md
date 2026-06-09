# ADR 001: Short Code Strategy - Random + Retry Loop

## Status

Accepted

## Context

We needed a strategy for generating short codes for URLs. The options considered were:

1. **Base62 encoding of auto-increment ID**: Simple, no collisions, but leaks URL count (security info leak) and creates hotspots on the DB primary key index (page contention in write-heavy loads).

2. **SHA256 + Base58 encoding**: The original implementation. Produces 8-char codes, but has collision risk (hash collisions), non-deterministic length, and is computationally heavier than necessary.

3. **Random 6-character alphanumeric + retry loop + unique index**: No sequence leaks, no hash collisions, distributed-friendly, and the unique index at the DB level provides a final safety net.

## Decision

We chose **random 6-character alphanumeric + retry loop** because:

- No sequence information leaks (unlike auto-increment ID)
- No hotspots on the primary key index
- The unique constraint on `short_code` at the DB level prevents any collision from being persisted
- The retry loop (up to 10 attempts) handles the collision probability effectively
- 6 characters from our 62-character alphabet gives 62^6 ≈ 56 billion possible codes, making collisions extremely unlikely

## Consequences

- The `CodeExists()` check against the DB unique index adds one extra query per code generation attempt
- In the extremely rare case of 10 consecutive collisions, the user gets an error
- This strategy scales horizontally without coordination (no need for a centralized sequence generator)
