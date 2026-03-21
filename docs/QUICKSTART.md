# Quick Start Guide

This guide will help you get the Claude Pipeline service up and running quickly.

## Prerequisites

- Docker and Docker Compose
- Anthropic API key
- (Optional) GitLab token for skill sync

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/HOA233/claude-code-pipeline.git
cd claude-code-pipeline
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env and add your API keys
```

### 3. Start Services

```bash
# Start with Docker Compose
docker-compose up -d

# Or start with monitoring
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d
```

### 4. Verify the Service

```bash
# Check health
curl http://localhost:8080/health

# List available skills
curl http://localhost:8080/api/skills
```

## Using the API

### Create a Task

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id": "code-review",
    "parameters": {
      "target": "src/",
      "depth": "standard"
    }
  }'
```

### Create a Pipeline

```bash
# Using the quickstart example
curl -X POST http://localhost:8080/api/pipelines \
  -H "Content-Type: application/json" \
  -d @examples/quickstart.json

# Or create inline
curl -X POST http://localhost:8080/api/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-pipeline",
    "mode": "serial",
    "steps": [
      {"id": "step1", "cli": "claude", "action": "review", "params": {"target": "src/"}}
    ]
  }'
```

### Run a Pipeline

```bash
# Replace <pipeline-id> with the ID from creation
curl -X POST http://localhost:8080/api/pipelines/<pipeline-id>/run
```

## Available Skills

| Skill ID | Description |
|----------|-------------|
| `code-review` | Analyze code quality and detect issues |
| `deploy` | Automated deployment to environments |
| `test-gen` | Generate unit tests |
| `refactor` | Intelligent code refactoring |
| `docs-gen` | Generate API documentation |

## API Endpoints

### Skills
- `GET /api/skills` - List all skills
- `GET /api/skills/:id` - Get skill details

### Tasks
- `POST /api/tasks` - Create a task
- `GET /api/tasks` - List all tasks
- `GET /api/tasks/:id` - Get task details
- `GET /api/tasks/:id/result` - Get task result
- `DELETE /api/tasks/:id` - Cancel task

### Pipelines
- `POST /api/pipelines` - Create pipeline
- `GET /api/pipelines` - List pipelines
- `GET /api/pipelines/:id` - Get pipeline
- `DELETE /api/pipelines/:id` - Delete pipeline
- `POST /api/pipelines/:id/run` - Execute pipeline

### Schedules
- `POST /api/schedules` - Create schedule
- `GET /api/schedules` - List schedules
- `PUT /api/schedules/:id` - Update schedule
- `DELETE /api/schedules/:id` - Delete schedule
- `POST /api/schedules/:id/trigger` - Trigger manually

### System
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /api/status` - System status

## Monitoring

When running with the monitoring stack:

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

## Development

### Run Tests

```bash
# Run all tests
./scripts/test_all.sh

# Run comprehensive tests
./scripts/comprehensive_test.sh

# Run load tests
./scripts/load_test.sh
```

### Build

```bash
# Build binary
./scripts/build.sh

# Build Docker image
docker build -t claude-pipeline:latest .
```

## Troubleshooting

### Redis Connection Issues

```bash
# Check Redis status
docker-compose exec redis redis-cli ping

# Check Redis logs
docker-compose logs redis
```

### API Not Responding

```bash
# Check API logs
docker-compose logs api

# Check container status
docker-compose ps
```

### Skill Sync Failed

Make sure your GitLab token has the correct permissions to access the skills repository.

## Next Steps

- Read the full [README.md](README.md)
- Check [examples/](examples/) for pipeline examples
- See [docs/CLI.md](docs/CLI.md) for CLI usage
- Review [docs/openapi.yaml](docs/openapi.yaml) for API reference