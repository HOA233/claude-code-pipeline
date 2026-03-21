# Contributing to Claude Pipeline

Thank you for your interest in contributing to Claude Pipeline! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow. Please be respectful and constructive in all interactions.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Set up the development environment
4. Create a feature branch
5. Make your changes
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.22 or later
- Node.js 18+ (for frontend development)
- Docker and Docker Compose
- Redis (can be run via Docker)
- Make (optional, for using Makefile commands)

### Local Setup

```bash
# Clone the repository
git clone https://github.com/HOA233/claude-code-pipeline.git
cd claude-code-pipeline

# Run setup script
./scripts/setup.sh

# Start Redis
docker-compose up -d redis

# Start the development server
make run
```

### Environment Variables

Copy `.env.example` to `.env` and configure:

```env
ANTHROPIC_API_KEY=your-api-key
GITLAB_TOKEN=your-gitlab-token
REDIS_ADDR=localhost:6379
```

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/HOA233/claude-code-pipeline/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Open a discussion in [GitHub Discussions](https://github.com/HOA233/claude-code-pipeline/discussions)
2. Describe the feature and its use case
3. Explain why it would benefit the project

### Contributing Code

1. Find an issue to work on or propose a new feature
2. Create a feature branch from `main`
3. Write code following our coding standards
4. Add tests for your changes
5. Update documentation if needed
6. Submit a pull request

## Pull Request Process

1. **Branch Naming**: Use descriptive branch names
   - `feature/add-webhook-support`
   - `fix/memory-leak-executor`
   - `docs/update-api-reference`

2. **Commit Messages**: Follow conventional commits
   ```
   feat: add webhook support for task completion
   fix: resolve memory leak in CLI executor
   docs: update API documentation
   test: add integration tests for pipeline orchestration
   ```

3. **PR Size**: Keep PRs focused and reasonably sized
   - One feature/fix per PR
   - If too large, consider splitting

4. **PR Description**: Include
   - What changes were made
   - Why they were made
   - How to test them
   - Any breaking changes

5. **Review Process**:
   - All PRs require at least one review
   - Address all review comments
   - Keep discussions constructive

## Coding Standards

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write meaningful comments for exported functions
- Handle errors properly (don't ignore them)

### Project Structure

```
├── cmd/                    # Application entry points
├── internal/
│   ├── api/               # HTTP handlers and routes
│   ├── service/           # Business logic
│   ├── model/             # Data models
│   ├── repository/        # Data access layer
│   └── config/            # Configuration
├── pkg/                    # Public packages
├── frontend/               # React frontend
├── deploy/                 # Deployment configs
├── docs/                   # Documentation
└── tests/                  # Test files
```

### Code Style

```go
// Good: Clear function with documentation
// CreateTask creates a new task with the given skill and parameters.
// It validates the skill exists and parameters are correct before creation.
func (s *TaskService) CreateTask(ctx context.Context, req *model.TaskCreateRequest) (*model.Task, error) {
    // Implementation
}

// Bad: No documentation, unclear naming
func (s *TaskService) Create(ctx context.Context, r *model.Req) (*model.Task, error) {
    // Implementation
}
```

## Testing Guidelines

### Running Tests

```bash
# Run all tests
make test

# Run specific test
go test -v ./tests/service_test.go

# Run with coverage
go test -cover ./...

# Run integration tests
./scripts/comprehensive_test.sh
```

### Writing Tests

1. **Unit Tests**: Test individual functions in isolation
   ```go
   func TestSkillService_GetSkill(t *testing.T) {
       // Setup
       mockRedis := &MockRedisClient{}
       svc := NewSkillService(mockRedis, nil)

       // Test
       skill, err := svc.GetSkill(context.Background(), "code-review")

       // Assert
       assert.NoError(t, err)
       assert.Equal(t, "code-review", skill.ID)
   }
   ```

2. **Integration Tests**: Test component interactions
   ```go
   func TestTaskExecution(t *testing.T) {
       if testing.Short() {
           t.Skip("Skipping integration test")
       }
       // Integration test code
   }
   ```

3. **Table-Driven Tests**: For multiple test cases
   ```go
   func TestValidateParameters(t *testing.T) {
       tests := []struct {
           name    string
           params  map[string]interface{}
           wantErr bool
       }{
           {"valid params", map[string]interface{}{"target": "src/"}, false},
           {"missing required", map[string]interface{}{}, true},
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test implementation
           })
       }
   }
   ```

### Test Coverage

- Aim for at least 80% coverage on new code
- Focus on critical paths and error handling
- Use `go test -coverprofile=coverage.out ./...`

## Documentation

### API Documentation

- Update OpenAPI spec in `docs/openapi.yaml` for new endpoints
- Include request/response examples
- Document error responses

### Code Documentation

- Add package comments for new packages
- Document exported functions and types
- Use godoc format

```go
// Package executor provides CLI execution capabilities.
//
// The executor package handles spawning and managing CLI subprocesses,
// capturing their output, and reporting results back to the task system.
package executor

// Executor manages CLI process execution.
type Executor struct {
    // config holds the executor configuration
    config Config
    // activeProcesses tracks running processes
    activeProcesses sync.Map
}
```

### README Updates

- Update README.md for new features
- Add usage examples
- Update API endpoints section

## Release Process

1. Version follows [Semantic Versioning](https://semver.org/)
2. Update CHANGELOG.md with release notes
3. Create a git tag: `git tag v1.0.0`
4. Push tag: `git push origin v1.0.0`
5. CI/CD will handle the rest

## Getting Help

- Open a [GitHub Discussion](https://github.com/HOA233/claude-code-pipeline/discussions) for questions
- Join our community chat (if available)
- Check existing issues before creating new ones

Thank you for contributing! 🎉