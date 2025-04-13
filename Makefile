DOCKER_COMPOSE = $(shell command -v docker-compose >/dev/null 2>&1 && echo docker-compose || echo docker compose)

deps:
	@echo "📦 Installing Go dependencies..."
	go mod tidy
	go mod download
	@echo "✅ Dependencies installed."

mockgen: deps
	mockgen -source=internal/blockchain/blockchain.go -destination=internal/blockchain/mock/mock_client.go -package=blockchain
	mockgen -source=internal/database/database.go -destination=internal/database/mock/mock_database.go -package=database

docker-build:
	docker build -t glif-app .

docker-run: docker-build
	$(DOCKER_COMPOSE) up -d

docker-stop:
	$(DOCKER_COMPOSE) down

up-postgres:
	@echo "🐘 Starting only the 'postgres' service from docker-compose.yml..."
	$(DOCKER_COMPOSE) up -d postgres

test: deps
	go test ./... -v


CHAIN_ID ?= testnet
POSTGRES_PORT ?= 5433
DATABASE_DSN ?= postgres://postgres:password@localhost:$(POSTGRES_PORT)/postgres?sslmode=disable
SERVER_LISTEN_ADDR ?= :8080

local-run: up-postgres deps
	@echo "🔧 Building app for your OS..."
	@mkdir -p bin
	go build -o bin/main ./main.go
	@echo "🚀 Running app with environment variables..."
	CHAIN_ID=$(CHAIN_ID) \
	POSTGRES_PORT=$(POSTGRES_PORT) \
	DATABASE_DSN=$(DATABASE_DSN) \
	SERVER_LISTEN_ADDR=$(SERVER_LISTEN_ADDR) \
	./bin/main


