# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- CI/CD pipeline with GitHub Actions
- Postman collection for API testing
- Contributing guidelines

## [1.0.0] - 2024-01-15

### Added

#### Core Features
- **Skills Management**: Load and manage CLI skills from GitLab repository
  - Support for skill versioning
  - Parameter validation
  - Custom prompt templates
- **Task Execution**: Execute tasks with selected skills
  - Concurrent task execution
  - Real-time status updates via WebSocket
  - Task cancellation support
- **Pipeline Orchestration**: Chain multiple CLI commands
  - Serial execution mode
  - Parallel execution mode
  - Dependency graph support
  - Error handling strategies (continue/stop/retry)

#### API Endpoints
- `GET /api/skills` - List all available skills
- `GET /api/skills/:id` - Get skill details
- `POST /api/skills/sync` - Sync skills from GitLab
- `POST /api/tasks` - Create a new task
- `GET /api/tasks` - List all tasks
- `GET /api/tasks/:id` - Get task details
- `GET /api/tasks/:id/result` - Get task result
- `DELETE /api/tasks/:id` - Cancel task
- `GET /api/pipelines` - List pipelines
- `POST /api/pipelines` - Create pipeline
- `GET /api/pipelines/:id` - Get pipeline details
- `POST /api/pipelines/:id/run` - Execute pipeline
- `GET /api/pipelines/:id/status` - Get pipeline status
- `GET /api/runs` - List all runs
- `GET /api/runs/:id` - Get run details
- `GET /api/status` - System status
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics

#### Enterprise Features
- **Authentication**: API key and Bearer token authentication
- **Rate Limiting**: Configurable rate limits per API key
- **Audit Logging**: Comprehensive audit trail for compliance
- **Webhook Support**: Callback URLs for task completion
- **Caching**: Redis-based caching with TTL support
- **Metrics**: Prometheus-compatible metrics endpoint

#### Deployment
- Docker support with multi-stage builds
- Docker Compose for local development
- Kubernetes manifests with HPA
- Helm chart for Kubernetes deployment
- Production-ready configurations

#### Frontend
- React + Vite frontend with Anthropic design style
- Dark/Light theme support
- Real-time task output via WebSocket
- Skill cards with parameter forms
- Pipeline visualization
- Responsive design

#### Multi-CLI Support
- Claude Code CLI
- npm
- git
- docker
- kubectl
- bash scripts
- python scripts
- go commands

#### Templates
- Pre-built pipeline templates
- Full code review workflow
- CI/CD pipeline template
- Parallel analysis template

### Security
- SQL injection protection
- XSS prevention
- Input validation
- Secure environment variable handling
- API key rotation support

### Performance
- Concurrent task execution
- Redis-based task queue
- Connection pooling
- Graceful shutdown

## [0.2.0] - 2024-01-01

### Added
- Basic CLI executor
- Redis integration
- Simple task queue
- HTTP API with Gin

### Changed
- Improved error handling
- Better logging with zap

## [0.1.0] - 2023-12-15

### Added
- Initial project structure
- Basic skill model
- Configuration management
- Docker support

---

## Release Notes Template

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes to existing features

### Deprecated
- Features to be removed

### Removed
- Features removed in this version

### Fixed
- Bug fixes

### Security
- Security improvements
```

## Migration Guide

### Upgrading from 0.2.x to 1.0.0

1. Update environment variables:
   - `ANTHROPIC_API_KEY` is now required
   - `REDIS_ADDR` format changed to `host:port`

2. API changes:
   - `/api/execute` renamed to `/api/tasks`
   - Response format updated for consistency

3. Configuration:
   - YAML configuration is now required
   - See `config/config.yaml.example` for reference

4. Deployment:
   - Update Kubernetes manifests
   - New Helm chart structure