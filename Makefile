.PHONY: build run test clean docker docker-down logs redis-cli help frontend frontend-dev frontend-build frontend-test e2e monitor

help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build        - Build the binary"
	@echo "  make run          - Run the server"
	@echo "  make deps         - Download dependencies"
	@echo "  make lint         - Format and lint code"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run Go tests"
	@echo "  make test-coverage- Run tests with coverage"
	@echo "  make e2e          - Run E2E tests"
	@echo "  make test-all     - Run all test scripts"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend     - Install frontend deps"
	@echo "  make frontend-dev - Run frontend dev server"
	@echo "  make frontend-build- Build frontend for production"
	@echo "  make frontend-test- Run frontend tests"
	@echo ""
	@echo "Docker:"
	@echo "  make docker       - Build and run with Docker"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make monitor      - Start with monitoring stack"
	@echo "  make logs         - View Docker logs"
	@echo "  make redis-cli    - Open Redis CLI"
	@echo ""
	@echo "Utility:"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make help         - Show this help message"

build:
	go build -o bin/server ./cmd/server
	go build -o bin/cli ./cmd/cli

run:
	go run ./cmd/server

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-all:
	@echo "Running all tests..."
	./scripts/test_all.sh

e2e:
	./scripts/e2e_test.sh

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

docker:
	docker-compose up -d --build

docker-down:
	docker-compose down

monitor:
	docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

logs:
	docker-compose logs -f api

redis-cli:
	docker-compose exec redis redis-cli

deps:
	go mod download
	go mod tidy

lint:
	go fmt ./...
	go vet ./...

# Frontend commands
frontend:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-test:
	cd frontend && npm run test:run