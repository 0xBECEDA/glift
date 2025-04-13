# 🧾 How to Run

## ✅ Prerequisites

Make sure you have the following installed:

- [Docker 1.26+](https://www.docker.com/)
- [Docker Compose v2.3+ or v1.19+](https://docs.docker.com/compose/)
- [GoLang 1.23.6+](https://go.dev/doc/install)

Following app is tested with setup:
- linux ubuntu 24.0
- go v1.23.6
- docker v1.28
- docker compose v2.31.0

## 🚀 Running the App
### Running with docker (preferred)
To build the docker image, run postgres:13 instance and run the application:
```bash
make docker-run
```
> ⚠️ If port 5433 is busy on your host, set up POSTGRES_PORT env variable to desired postgres port.
>   By default server listens 8080 port. It supports connections only to the **Calibration** testnet.

To stop the app:

```bash
make docker-stop
```

### Running locally
If you run application in docker before, please make down containers first by **make docker-stop** 

This command builds and runs the app locally with environment variables:
```bash
make local-run
```
- Starts the `postgres` service via Docker Compose
- Installs Go dependencies
- Builds the binary to `bin/main`
- Runs the app with environment variables

#### 🌱 Default environment variables:

| Variable              | Default Value                                                                             |
|-----------------------|--------------------------------------------------------------------------------------------|
| `CHAIN_ID`            | `testnet`                                                                                 |
| `POSTGRES_PORT`       | `5433`                                                                                     |
| `DATABASE_DSN`        | `postgres://postgres:password@localhost:$(POSTGRES_PORT)/postgres?sslmode=disable`        |
| `SERVER_LISTEN_ADDR`  | `:8080`                                                                                    |

Override like this:

```bash
POSTGRES_PORT=5432 make local-run
```
> ⚠️ Do not change CHAIN_ID: only calibrationnet is supported.


## 🛠️ API Usage

### ⚠️ Security Notice

This implementation accepts a raw private key from the client to sign and send a transaction.  
This approach is **NOT SECURE** and should **never be used in production**.

In a real-world application, **I would never transmit private keys over the network**, especially not in plaintext.  
Instead, I would recommend one of the following secure alternatives:

1. ✅ Let the client **sign the transaction locally** and submit the signed transaction.
2. 🔐 Use a **Web3 wallet** (e.g., MetaMask) with **WalletConnect** or a similar bridge protocol.
3. 💾 Integrate with a **hardware wallet** or signing service (e.g., Ledger, Fireblocks).
4. 🧰 For backend automation, use a **secure signer** stored in a Vault or HSM (e.g., AWS KMS, HashiCorp Vault).

> ⚠️ This insecure method was chosen **only due to time constraints** to demonstrate a complete backend flow.

### 📤 Send Transaction

```bash
curl -X POST http://localhost:8080/transaction/send \
  -H "Content-Type: application/json" \
  -d '{
    "private_key_hex": "your_private_key_hex",
    "receiver": "0xReceiverAddressHere",
    "amount": "1000000000000000000"
}'
```

Success response:
```json
{"hash":"0x15e54e2f6e60be523a4ef44e3dc3ab7245bdc98d8b007bfcf1628a320983384b"}
```

### 🕒 Transaction Status Tracking

Currently, this application **does not scan the blockchain to track the status of submitted transactions**.

As a result:

- All submitted transactions are stored in the database with a `pending` status.
- The status is **never updated**, even if the transaction is later confirmed or fails on-chain.

This limitation can be resolved in the future by adding a background worker or blockchain event listener to keep transaction statuses up to date.

> If transaction status tracking is needed, please check manually via [Filfox Explorer](https://filfox.info/en).  
> Remember that transactions are submitted on the **Calibration testnet**.

### 📄 Get Transactions from Database

```bash
curl -X GET "http://localhost:8080/transactions/?sender=0xSenderAddressHere&receiver=0xReceiverAddressHere"
```
Success response:
```json
[
  {
    "ID": 1,
    "Hash": "0xexamplehash000000000000000000000000000000000000000000000000000000",
    "Sender": "0xexampleSenderAddress0000000000000000000000000000",
    "Receiver": "0xexampleReceiverAddress000000000000000000000000",
    "Amount": "20000",
    "Timestamp": "2025-04-13T13:04:46.754419Z",
    "Status": "pending"
  }
]
```

- Both `sender` and `receiver` are optional. If no of them provided, endpoint returns **up to 100** transactions. 
- Address matching is not **case-insensitive**.

### 💰 Check Wallet Balance

```bash
curl -X GET "http://localhost:8080/balance/your_address_here"
```

Success response:
```json
{"fil":"1","ifil":"2"}
```

## 🧪 Testing
Tests are using **testcontainers**, make sure docker containers running by **make docker-run** is down.
To run all tests:

```bash
make test
```

## 🚀 What Would Be Improved in a Real-World Scenario

If this were a production-grade application, I would implement the following improvements:

- 🔐 **No raw private keys in API requests**  
  As mentioned above, raw private keys should never be transmitted over the network. Instead, signing should happen on the client side or via secure systems.

- 📡 **Transaction status tracking via event listener**  
  A background job or blockchain event listener would track the status of each transaction and update it in the database accordingly.

- ⚙️ **Flexible configuration management**  
  The app would support loading config values from a file (e.g., `config.yaml`) in addition to environment variables.

- 📚 **Auto-generated API documentation using Swagger**  
  API endpoints would be documented with [Swagger](https://swagger.io/) (e.g., using `swaggo/swag`) to provide an interactive developer experience.

- ✅ **More tests with diverse scenarios**  
  The test suite would be extended to include edge cases, failure simulations, invalid inputs, database errors, and blockchain response handling.
