# Variables
BINARY_NAME=api-server
GO_FILES=$(shell find . -name '*.go')
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOPATH=$(shell go env GOPATH)

# Use the Go modules
export GO111MODULE=on

# Colors for terminal output
CYAN=\033[0;36m
NC=\033[0m # No Color

.PHONY: all build clean test coverage lint fmt help run dev docker-build docker-run

# Default target
all: help

## build: Build the binary
build:
	@printf "${CYAN}Building binary...${NC}\n"
	@go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME} ./cmd/webhook

## clean: Clean build files
clean:
	@printf "${CYAN}Cleaning build cache...${NC}\n"
	@go clean
	@rm -rf ${GOBIN}

## test: Run tests
test:
	@printf "${CYAN}Running tests...${NC}\n"
	@go test -v ./...

## coverage: Run tests with coverage
coverage:
	@printf "${CYAN}Running tests with coverage...${NC}\n"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@printf "${CYAN}Coverage report generated: coverage.html${NC}\n"

## lint: Run linter
lint:
	@printf "${CYAN}Running linter...${NC}\n"
	@if [ ! -f $(GOPATH)/bin/golangci-lint ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin; \
	fi
	@golangci-lint run

## fmt: Format code
fmt:
	@printf "${CYAN}Formatting code...${NC}\n"
	@go fmt ./...

## run: Run the application
run: build
	@printf "${CYAN}Running application...${NC}\n"
	@${GOBIN}/${BINARY_NAME}

## dev: Run the application in development mode
dev:
	@printf "${CYAN}Running in development mode...${NC}\n"
	@air -c .air.toml

## mod: Download and tidy dependencies
mod:
	@printf "${CYAN}Downloading and tidying dependencies...${NC}\n"
	@go mod download
	@go mod tidy

## docker-build: Build docker image
docker-build:
	@printf "${CYAN}Building Docker image...${NC}\n"
	@docker build -t ${BINARY_NAME}:${VERSION} .

## docker-run: Run docker container
docker-run:
	@printf "${CYAN}Running Docker container...${NC}\n"
	@docker run -p 8080:8080 ${BINARY_NAME}:${VERSION}

## migrate: Run database migrations
migrate:
	@printf "${CYAN}Running database migrations...${NC}\n"
	@go run ./cmd/migrate

## generate: Generate Go code (mocks, etc.)
generate:
	@printf "${CYAN}Generating code...${NC}\n"
	@go generate ./...

## init-dev: Initialize development environment
init-dev: mod
	@printf "${CYAN}Initializing development environment...${NC}\n"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@printf "${CYAN}Development environment initialized${NC}\n"

## help: Show this help
help: Makefile
	@echo "\nUsage: make [target]\n"
	@echo "Targets:"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
