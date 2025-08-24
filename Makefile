# Makefile
.PHONY: help build run test clean docker-build docker-run

help:
	@echo "Available commands:"
	@echo "  make build        - Build the binary"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run with docker-compose"

build:
	go build -o bin/go_runner ./cmd/go_runner

run:
	ADMIN_TOKEN=test123 go run ./cmd/go_runner

test:
	go test -v -cover ./...

clean:
	rm -rf bin/ data/

docker-build:
	docker build -t go_runner:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down
