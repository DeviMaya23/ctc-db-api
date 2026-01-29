docker-run-app:
	docker-compose --profile with-app up -d --build

.PHONY: test-unit test

test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

test:
	@echo "Running all tests..."
	go test -v ./...
