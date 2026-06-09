# ADR 002: PostgreSQL over MongoDB / SQLite

## Status

Accepted

## Context

We needed a primary datastore for URL mappings. The options were:

1. **PostgreSQL**: Relational, ACID-compliant, mature, excellent JSON support, GIN indexes for future full-text search, strong ecosystem.

2. **MongoDB**: Document store, flexible schema, good for read-heavy workloads, but eventual consistency and less mature tooling.

3. **SQLite**: Embedded, simple, zero-config, but cannot handle concurrent writes at scale.

## Decision

We chose **PostgreSQL** because:

- ACID compliance ensures no duplicate short codes are ever persisted
- The unique index on `short_code` is critical for our collision strategy
- GORM's PostgreSQL driver is mature and well-tested
- Postgres handles concurrent writes well (unlike SQLite)
- For a URL shortener, the relational model (URL mapping with click counter) is inherently tabular
- Future features (full-text search, analytics) are well-supported by Postgres

## Consequences

- Requires a running Postgres instance (Docker Compose handles this)
- Schema migrations via GORM AutoMigrate
- Connection pooling is handled by GORM / pgx driver
