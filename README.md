# How to run

## API usage

```
curl -X POST http://localhost:8080/transaction/send \
  -H "Content-Type: application/json" \
  -d '{
    "private_key_hex": "your_private_key_hex_here_without_0x",
    "receiver": "0xReceiverAddressHere",
    "amount": "1000000000000000000"
}'
```

# Testing

The project includes unit and integration tests to ensure correctness of implementation. To run all tests:

```sh
go test ./...


## Run tests

Tests use Testcontainers to create isolated PostgreSQL containers. Make sure Docker is installed and running.

### Requirements
- Docker and Docker Compose
- Go 1.20 or higher (to run locally)

### Running Tests via Docker Compose
1. Create an `.env` file with environment variables (see `.env.example` for an example).
2. Run the command:
   ```sh
 docker-compose -f docker-compose.yml run --rm -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal app go test ./....

```
# What would be improved, if it would be real task
### ⚠️ Security Notice

This implementation accepts a raw private key from the client to sign and send a transaction.  
This approach is **NOT SECURE** and should **never be used in production**.

In a real-world application, **I would not transmit private keys over the network**, especially not in plaintext.  
Instead, I would recommend one of the following secure alternatives:

1. Let the client **sign the transaction locally** and submit the signed transaction.
2. 🔐Use a **Web3 wallet** (e.g., MetaMask) with **WalletConnect** or similar bridge protocol.
3. 💾Integrate with a **hardware wallet** or signing service (e.g., Ledger, Fireblocks).
4. 🧰For backend automation, use a **secure signer** stored in a Vault or HSM (e.g., AWS KMS, HashiCorp Vault).

> ⚠️ The current implementation was chosen **only due to time constraints** and the need to demonstrate a complete flow in a single backend service.


