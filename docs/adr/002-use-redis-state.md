# ADR-002: Use Redis for State Management

## Status
Accepted

## Context
The CLI orchestration service needs to manage:
- Task state and results
- Pipeline configurations
- Real-time output streaming
- Pub/Sub for WebSocket updates

## Decision
We will use Redis as the primary data store for all state management.

## Consequences

### Positive
- Fast in-memory operations
- Native pub/sub support for real-time updates
- Data structure flexibility (strings, lists, hashes, sets)
- Built-in TTL for automatic cleanup
- Simple horizontal scaling
- Excellent Go client library (go-redis)

### Negative
- Data persistence requires configuration
- Memory-based storage limits
- Additional infrastructure dependency

## Alternatives Considered
- **PostgreSQL**: Better for complex queries, but slower for real-time
- **MongoDB**: Document store but less suitable for pub/sub
- **In-memory**: No persistence, doesn't scale across instances