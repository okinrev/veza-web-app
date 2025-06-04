.PHONY: build run test clean dev

BINARY_NAME=veza-api
BINARY_PATH=./
BUILD_DIR=./bin

build:
	@echo "Building application..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)/main.go

run:
	@echo "Running application..."
	go run $(BINARY_PATH)/main.go

dev:
	@echo "Running in development mode..."
	go run main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

migrate:
	@echo "Running migrations..."
	# TODO: Implement migration script

lint:
	@echo "Running linter..."
	golangci-lint run

docker-build:
	docker build -t veza-backend .

docker-run:
	docker-compose up -d
