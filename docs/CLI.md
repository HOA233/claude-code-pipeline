# CLI Tool

The Claude Pipeline CLI tool (`pipeline-cli`) allows you to interact with the Claude Pipeline service from the command line.

## Installation

### From Release

Download the latest release from [GitHub Releases](https://github.com/HOA233/claude-code-pipeline/releases):

```bash
# Linux/macOS
curl -sL https://github.com/HOA233/claude-code-pipeline/releases/latest/download/pipeline-cli-$(uname -s)-$(uname -m) -o pipeline-cli
chmod +x pipeline-cli
sudo mv pipeline-cli /usr/local/bin/

# Or via Go install
go install github.com/HOA233/claude-code-pipeline/cmd/cli@latest
```

### From Source

```bash
git clone https://github.com/HOA233/claude-code-pipeline.git
cd claude-code-pipeline
make build-cli
```

## Configuration

Set up your environment:

```bash
# Required
export PIPELINE_API_URL="http://localhost:8080/api"
export PIPELINE_API_KEY="your-api-key"

# Optional
export PIPELINE_OUTPUT_FORMAT="json"  # json, yaml, table
export PIPELINE_TIMEOUT="60"          # default timeout in seconds
```

Or use a configuration file at `~/.pipeline/config.yaml`:

```yaml
api_url: http://localhost:8080/api
api_key: your-api-key
output_format: table
timeout: 60
```

## Commands

### Global Options

```
  --config string     Config file path (default "~/.pipeline/config.yaml")
  --output string     Output format: json, yaml, table (default "table")
  --timeout int       Request timeout in seconds (default 60)
  --help             Show help for command
```

### Skills

#### List Skills

```bash
pipeline-cli skills list

# Output formats
pipeline-cli skills list --output json
pipeline-cli skills list --output yaml
```

#### Get Skill Details

```bash
pipeline-cli skills get code-review
```

#### Sync Skills

```bash
pipeline-cli skills sync
```

### Tasks

#### Create Task

```bash
# Basic task
pipeline-cli tasks create --skill code-review --param target=src/ --param depth=deep

# With context
pipeline-cli tasks create \
  --skill deploy \
  --param environment=staging \
  --param dry_run=true \
  --context repository=https://github.com/company/project.git \
  --context branch=main

# With timeout and callback
pipeline-cli tasks create \
  --skill test-gen \
  --param source=src/utils/ \
  --timeout 1200 \
  --callback https://webhook.example.com/callback
```

#### List Tasks

```bash
# All tasks
pipeline-cli tasks list

# Filter by status
pipeline-cli tasks list --status running
pipeline-cli tasks list --status completed
pipeline-cli tasks list --status failed

# Limit results
pipeline-cli tasks list --limit 10
```

#### Get Task Details

```bash
pipeline-cli tasks get task-abc123
```

#### Get Task Result

```bash
pipeline-cli tasks result task-abc123

# Save to file
pipeline-cli tasks result task-abc123 --output-file result.json
```

#### Watch Task Output

Real-time streaming of task output:

```bash
pipeline-cli tasks watch task-abc123
```

#### Cancel Task

```bash
pipeline-cli tasks cancel task-abc123
```

### Pipelines

#### Create Pipeline

```bash
# From JSON file
pipeline-cli pipelines create --file pipeline.json

# From YAML file
pipeline-cli pipelines create --file pipeline.yaml

# Inline definition
pipeline-cli pipelines create --name "my-pipeline" --steps '[
  {"id": "analyze", "cli": "claude", "action": "analyze", "params": {"target": "src/"}},
  {"id": "test", "cli": "claude", "action": "test-gen", "params": {"source": "src/"}, "depends_on": ["analyze"]}
]'
```

#### List Pipelines

```bash
pipeline-cli pipelines list
```

#### Get Pipeline

```bash
pipeline-cli pipelines get pipeline-001
```

#### Run Pipeline

```bash
pipeline-cli pipelines run pipeline-001

# With parameters
pipeline-cli pipelines run pipeline-001 --param target=src/api/
```

#### Pipeline Status

```bash
pipeline-cli pipelines status pipeline-001
```

#### Delete Pipeline

```bash
pipeline-cli pipelines delete pipeline-001
```

### Runs

#### List Runs

```bash
pipeline-cli runs list

# Filter by pipeline
pipeline-cli runs list --pipeline pipeline-001

# Filter by status
pipeline-cli runs list --status completed
```

#### Get Run Details

```bash
pipeline-cli runs get run-abc123
```

### System

#### Health Check

```bash
pipeline-cli health

# JSON output
pipeline-cli health --output json
```

#### System Status

```bash
pipeline-cli status
```

#### Metrics

```bash
pipeline-cli metrics
```

### Templates

#### List Templates

```bash
pipeline-cli templates list
```

#### Apply Template

```bash
pipeline-cli templates apply full-code-review --param target=src/
```

## Examples

### Code Review Workflow

```bash
# 1. Create and run a code review task
pipeline-cli tasks create \
  --skill code-review \
  --param target=src/ \
  --param depth=deep \
  --output json > task.json

# 2. Extract task ID
TASK_ID=$(jq -r '.id' task.json)

# 3. Watch progress
pipeline-cli tasks watch $TASK_ID

# 4. Get results
pipeline-cli tasks result $TASK_ID --output-file review-report.json
```

### Full Pipeline Execution

```bash
# Create pipeline from template
cat > pipeline.yaml << 'EOF'
name: full-review
mode: serial
steps:
  - id: analyze
    cli: claude
    action: analyze
    params:
      target: src/
  - id: security
    cli: claude
    action: security-scan
    params:
      target: src/
    depends_on: [analyze]
  - id: tests
    cli: claude
    action: test-gen
    params:
      source: src/
    depends_on: [analyze]
EOF

# Create and run
pipeline-cli pipelines create --file pipeline.yaml
pipeline-cli pipelines run full-review --watch
```

### Batch Operations

```bash
# Run multiple reviews in parallel
for dir in src/*/; do
  pipeline-cli tasks create \
    --skill code-review \
    --param target="$dir" \
    --param depth=quick &
done
wait

# Check all running tasks
pipeline-cli tasks list --status running
```

## Shell Completion

Enable shell completion for bash, zsh, or fish:

```bash
# Bash
pipeline-cli completion bash > /etc/bash_completion.d/pipeline-cli
source ~/.bashrc

# Zsh
pipeline-cli completion zsh > "${fpath[1]}/_pipeline-cli"
source ~/.zshrc

# Fish
pipeline-cli completion fish > ~/.config/fish/completions/pipeline-cli.fish
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | API error |
| 4 | Authentication error |
| 5 | Not found |
| 6 | Timeout |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PIPELINE_API_URL` | API server URL | `http://localhost:8080/api` |
| `PIPELINE_API_KEY` | API key for authentication | - |
| `PIPELINE_OUTPUT_FORMAT` | Default output format | `table` |
| `PIPELINE_TIMEOUT` | Request timeout (seconds) | `60` |
| `PIPELINE_CONFIG` | Config file path | `~/.pipeline/config.yaml` |
| `NO_COLOR` | Disable colored output | `false` |

## Configuration File

`~/.pipeline/config.yaml`:

```yaml
# API Configuration
api_url: http://localhost:8080/api
api_key: your-api-key-here
timeout: 60

# Output Configuration
output_format: table
no_color: false

# Aliases for commonly used commands
aliases:
  review: tasks create --skill code-review
  deploy: tasks create --skill deploy
  test: tasks create --skill test-gen

# Default parameters for skills
defaults:
  code-review:
    depth: standard
  deploy:
    dry_run: true
```

## Troubleshooting

### Connection Issues

```bash
# Check API health
pipeline-cli health

# Verbose output
pipeline-cli tasks list --verbose
```

### Authentication Errors

```bash
# Verify API key
pipeline-cli status

# Check if API key is set
echo $PIPELINE_API_KEY
```

### Timeout Issues

```bash
# Increase timeout
pipeline-cli tasks create --skill code-review --timeout 1200
```

## Version

```bash
pipeline-cli version
pipeline-cli version --verbose
```