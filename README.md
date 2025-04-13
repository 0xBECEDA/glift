# How to run
## Prerequisites
You have installed:
- Docker and Docker Compose

## Running app
To build docker image and run app using docker, use:
```shell
make run
```
By default, app will listen for requests on `8080` port and use **calibrationnet.** 

To stop app:
```shell
make stop
```
## API usage
### ⚠️ Security Notice

This implementation accepts a raw private key from the client to sign and send a transaction.  
This approach is **NOT SECURE** and should **never be used in production**.

In a real-world application, **I would not transmit private keys over the network**, especially not in plaintext.  
Instead, I would recommend one of the following secure alternatives:

1. ✅ Let the client **sign the transaction locally** and submit the signed transaction.
2. 🔐 Use a **Web3 wallet** (e.g., MetaMask) with **WalletConnect** or similar bridge protocol.
3. 💾 Integrate with a **hardware wallet** or signing service (e.g., Ledger, Fireblocks).
4. 🧰 For backend automation, use a **secure signer** stored in a Vault or HSM (e.g., AWS KMS, HashiCorp Vault).

> ⚠️ The current implementation was chosen **only due to time constraints** and the need to demonstrate a complete flow in a single backend service.

Send transaction:
```
curl -X POST http://localhost:8080/transaction/send \
  -H "Content-Type: application/json" \
  -d '{
    "private_key_hex": "your_private_key_hex",
    "receiver": "0xReceiverAddressHere",
    "amount": "1000000000000000000"
}'
```
In case of success you would get transaction hash.
⚠️ Transaction Status Tracking

At the moment, this application **does not scan the blockchain to track the status of submitted transactions**.

As a result:

- All submitted transactions are stored in the database with a `pending` status
- The status is **not updated** after submission, even if the transaction is later confirmed or failed on-chain

This is a known limitation and can be improved in the future by integrating a background job or blockchain event listener to update transaction statuses.
In case if transaction status check needed, use  [explorer](https://filfox.info/en). Do not forget, that transactions are sent in **calibrationnet**.


Get transactions from database:
```curl -X GET "http://localhost:8080/transactions/?sender=0xSenderAddressHere&receiver=0xReceiverAddressHere"```

- Both sender and receiver are optional
- Address matching is not case-sensitive
- If neither sender nor receiver is provided, the endpoint will return all transactions, up to a maximum of 100 entries

Check balance of wallet:
```curl -X GET "http://localhost:8080/balance/your_address_here" ```
In case of success you would get balances of FIL and IFIL of your address.

# Testing
To run all tests:

```shell
make test
```

# 🚀 What Would Be Improved in a Real-World Scenario

If this were a production-grade application, I would implement the following improvements:

- 🔐 **No raw private keys in API requests**  
  As mentioned above, I would never accept raw private keys from clients over the network. Instead, I would rely on client-side signing using wallets like MetaMask, WalletConnect, or hardware wallets, or securely integrate with custodial solutions (e.g., Vault, Fireblocks).

- 📡 **Transaction status tracking via event listener**  
  A blockchain event listener or polling mechanism would be implemented to monitor the status of submitted transactions and update them in the database accordingly (e.g., from `pending` to `confirmed` or `failed`).

- ⚙️ **Flexible configuration management**  
  The application would support loading configuration from a file (e.g., `config.yaml` or `config.json`) in addition to environment variables, allowing for better manageability and separation of concerns.

- 📚 **Auto-generated API documentation using Swagger**  
  I would document the API using [Swagger](https://swagger.io/) (e.g., with `swaggo/swag`), enabling users and developers to easily understand and interact with the available endpoints.

- ✅ **More tests with diverse scenarios**  
  I would expand test coverage to include a broader range of unit and integration tests, covering edge cases, invalid input, database errors, and blockchain interaction failure scenarios.

