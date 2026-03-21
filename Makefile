.PHONY: build run test clean docker docker-down logs redis-cli help

help:
	@echo "Available commands:"
	@echo "  make build       - Build the binary"
	@echo "  make run         - Run the server"
	@echo "  make test        - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker       - Build and run with Docker"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make logs         - View Docker logs"
	@echo "  make redis-cli    - Open Redis CLI"

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/

docker:
	docker-compose up -d --build

docker-down:
	docker-compose down

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