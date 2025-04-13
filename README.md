# 🧾 How to Run

## ✅ Prerequisites

Make sure you have the following installed:

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [GoLang 1.23+](https://go.dev/doc/install)
---

## 🚀 Running the App

To build the Docker image and run the application:

```bash
make run
```

By default, the app listens on port `8080` and connects to the **Calibration** testnet.

To stop the app:

```bash
make stop
```

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

If successful, the response will include the transaction hash.

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

- Both `sender` and `receiver` are optional.
- Address matching is **case-insensitive**.
- If neither is provided, the endpoint returns **up to 100** recent transactions.

### 💰 Check Wallet Balance

```bash
curl -X GET "http://localhost:8080/balance/your_address_here"
```

If successful, the response will include the **FIL** and **iFIL** balances of the specified address.

## 🧪 Testing

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
