# Claude Code CLI Pipeline Service

A Golang-based pipeline service that allows users to select different skills to run Claude Code CLI. All CLI instances run in the same service but execute different tasks based on skill selection.

## Features

- **Multi-Skill Support**: Choose from code-review, deploy, test-gen, refactor, and more
- **RESTful API**: Simple HTTP API to create and manage tasks
- **Real-time Updates**: WebSocket support for live task output
- **Redis-Based**: Uses Redis for queue, storage, and pub/sub
- **Docker Ready**: Includes Dockerfile and docker-compose

## Quick Start

### Prerequisites

- Go 1.22+
- Redis
- Claude Code CLI (`npm install -g @anthropic-ai/claude-code`)

### Run Locally

```bash
# Install dependencies
go mod tidy

# Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# Run the server
make run
```

### Run with Docker

```bash
# Build and run
make docker

# View logs
make logs
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/skills` | List all available skills |
| GET | `/api/skills/:id` | Get skill details |
| POST | `/api/skills/sync` | Sync skills from GitLab |
| POST | `/api/tasks` | Create a new task |
| GET | `/api/tasks` | List all tasks |
| GET | `/api/tasks/:id` | Get task details |
| GET | `/api/tasks/:id/result` | Get task result |
| DELETE | `/api/tasks/:id` | Cancel a task |
| GET | `/api/status` | Service status |
| GET | `/ws/tasks/:id/output` | WebSocket for real-time output |

## Usage Examples

### List Skills

```bash
curl http://localhost:8080/api/skills
```

### Create Task

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id": "code-review",
    "parameters": {
      "target": "src/",
      "depth": "deep"
    }
  }'
```

### Get Task Result

```bash
curl http://localhost:8080/api/tasks/task-abc123/result
```

## Available Skills

| Skill | Description | Parameters |
|-------|-------------|------------|
| `code-review` | Code quality analysis | `target`, `depth` |
| `deploy` | Deploy to environments | `environment`, `dry_run` |
| `test-gen` | Generate unit tests | `source`, `framework` |
| `refactor` | Refactor code | `target`, `type` |

## Project Structure

```
claude-pipeline/
├── cmd/server/          # Server entrypoint
├── internal/
│   ├── api/             # HTTP handlers and routes
│   ├── config/          # Configuration
│   ├── model/           # Data models
│   ├── repository/      # Redis storage
│   └── service/         # Business logic
├── pkg/
│   └── logger/          # Logging utilities
├── config/              # Config files
├── tests/               # Test files
├── scripts/             # Utility scripts
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Configuration

Edit `config/config.yaml`:

```yaml
server:
  port: 8080

redis:
  addr: localhost:6379

gitlab:
  url: https://gitlab.company.com
  token: ${GITLAB_TOKEN}

cli:
  max_concurrency: 5
```

## Testing

```bash
# Run unit tests
make test

# Run integration tests
./scripts/test.sh

# Run API demo
./scripts/demo.sh
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Anthropic API key for Claude |
| `GITLAB_TOKEN` | GitLab access token |

## License

MIT