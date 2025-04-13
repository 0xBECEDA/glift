DOCKER_COMPOSE = $(shell command -v docker-compose >/dev/null 2>&1 && echo docker-compose || echo docker compose)

mockgen:
	mockgen -source=internal/blockchain/blockchain.go -destination=internal/blockchain/mock/mock_client.go -package=blockchain
	mockgen -source=internal/database/database.go -destination=internal/database/mock/mock_database.go -package=database

docker-build:
	docker build -t glif-app .

run: docker-build
	$(DOCKER_COMPOSE) up -d

stop:
	$(DOCKER_COMPOSE) down

test:
	go test ./... -v