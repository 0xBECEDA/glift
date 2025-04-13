mockgen:
	mockgen -source=internal/blockchain/blockchain.go -destination=internal/blockchain/mock/mock_client.go -package=blockchain
	mockgen -source=internal/database/database.go -destination=internal/database/mock/mock_database.go -package=database

docker-build:
	docker build -t glif-app .

run: docker-build
	docker compose up -d

stop:
	docker compose down

test:
	go test ./...