.PHONY: swagger
swagger:
	swag init -g cmd/main.go -o docs

.PHONY: build
build:
	go build -o bin/server ./cmd/main.go

.PHONY: run
run:
	go run ./cmd/main.go api

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf docs/

# Load environment variables from .env file
include .env
export

.PHONY: migrate-apply
migrate-apply:
	@if [ -z "$(POSTGRES_CONNECTION_STRING)" ]; then \
		echo "Error: POSTGRES_CONNECTION_STRING not set in .env file"; \
		exit 1; \
	fi
	@echo "Applying migrations..."
	@atlas migrate apply --env local --url "$(POSTGRES_CONNECTION_STRING)" || \
		(echo "Migration command exited with error. Checking status..." && \
		 atlas migrate status --env local --url "$(POSTGRES_CONNECTION_STRING)" && \
		 exit 1)

.PHONY: migrate-status
migrate-status:
	@if [ -z "$(POSTGRES_CONNECTION_STRING)" ]; then \
		echo "Error: POSTGRES_CONNECTION_STRING not set in .env file"; \
		exit 1; \
	fi
	atlas migrate status --env local --url "$(POSTGRES_CONNECTION_STRING)"

.PHONY: migrate-validate
migrate-validate:
	@if [ -z "$(POSTGRES_CONNECTION_STRING)" ]; then \
		echo "Error: POSTGRES_CONNECTION_STRING not set in .env file"; \
		exit 1; \
	fi
	atlas migrate validate --env local --url "$(POSTGRES_CONNECTION_STRING)"

.PHONY: migrate-new
migrate-new:
	@if [ -z "$(POSTGRES_CONNECTION_STRING)" ]; then \
		echo "Error: POSTGRES_CONNECTION_STRING not set in .env file"; \
		exit 1; \
	fi
	@read -p "Enter migration name: " name; \
	atlas migrate new --env local --dir file://migrations $$name