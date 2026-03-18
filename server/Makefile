.PHONY: build run test clean docker-build docker-run

BUILD_TIME=$(shell date -u +"%Y-%m-%d %H:%M:%S")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

build:
	go build -ldflags "-X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)" -o bin/game-server ./cmd/game-server
	go build -ldflags "-X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)" -o bin/auth-server ./cmd/auth-server
	go build -ldflags "-X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)" -o bin/gateway ./cmd/gateway

run: build
	./bin/game-server

test:
	go test -v -cover ./...

test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	go clean -cache

docker-build:
	docker build -t mmo-server:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

lint:
	golangci-lint run ./...

generate:
	go generate ./...
