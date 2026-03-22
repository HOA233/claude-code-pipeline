# ADR-001: Use Go as Primary Language

## Status
Accepted

## Context
We need to build a high-performance CLI orchestration service that can handle multiple concurrent CLI executions, manage state, and provide real-time updates.

## Decision
We will use Go (Golang) as the primary programming language for the backend service.

## Consequences

### Positive
- Excellent concurrency support with goroutines and channels
- Fast compilation and execution
- Strong standard library
- Built-in testing framework
- Single binary deployment
- Good ecosystem for web services (Gin, Redis clients)

### Negative
- Less mature frontend ecosystem (requires separate frontend)
- Verbose error handling

## Alternatives Considered
- **Node.js**: Good for I/O but weaker concurrency
- **Rust**: Better performance but steeper learning curve
- **Python**: Slower execution, GIL limitations