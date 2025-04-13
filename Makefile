DOCKER_COMPOSE = $(shell command -v docker-compose >/dev/null 2>&1 && echo docker-compose || echo docker compose)

check-compose-version:
	@version_str="$$( $(DOCKER_COMPOSE) version --short 2>/dev/null )"; \
	if [ -z "$$version_str" ]; then \
		echo "❌ Could not determine Docker Compose version"; exit 1; \
	fi; \
	major=$$(echo $$version_str | cut -d. -f1); \
	minor=$$(echo $$version_str | cut -d. -f2); \
	if [ "$$major" -lt 1 ] || { [ "$$major" -eq 1 ] && [ "$$minor" -lt 28 ]; }; then \
		echo "❌ Docker Compose version $$version_str is too old. Please update to 1.28 or higher."; \
		exit 1; \
	else \
		echo "✅ Docker Compose version $$version_str is OK."; \
	fi

mockgen:
	mockgen -source=internal/blockchain/blockchain.go -destination=internal/blockchain/mock/mock_client.go -package=blockchain
	mockgen -source=internal/database/database.go -destination=internal/database/mock/mock_database.go -package=database

docker-build: check-compose-version
	docker build -t glif-app .

run: docker-build
	$(DOCKER_COMPOSE) up -d

stop:
	$(DOCKER_COMPOSE) down

test:
	go test ./... -v