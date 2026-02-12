# E-Wallet API (Echo + Gorm)

REST API for simple e-wallet management with JWT auth, wallet topup, transfer, and transaction history.

## Prerequisites
- Go 1.24+
- PostgreSQL
- Migration tool: `migrate` (https://github.com/golang-migrate/migrate)

## Environment Variables
Use `.env.example` as template.

```env
APP_ENV=development
PORT=8080
JWT_SECRET=change_this_secret

# DB configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ewallet
DB_SSLMODE=disable

# Test DB configuration
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD=postgres
TEST_DB_NAME=ewallet_test
TEST_DB_SSLMODE=disable
```

## Setup
1. Copy `.env.example` to `.env` and set values.
2. Create databases (example: `ewallet` and `ewallet_test`).
3. Load environment variables:

```bash
set -a
source .env
set +a
```

4. Run migrations:

```bash
migrate -path ./migrations -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" up
```

## Run
```bash
go run ./cmd/server
```

## Caching
- Transaction history (`GET /api/transactions/history`) uses in-memory cache.
- Cache implementation uses `github.com/AnggaBS86/gocachemem`.
- For more information please see the documentation https://github.com/AnggaBS86/gocachemem?tab=readme-ov-file#gocachemem 
- Cache is invalidated automatically after successful transfer (`POST /api/transactions/transfer`) for both sender and receiver.

## Transaction Handling
- Transfer is executed in a single database transaction for atomic debit-credit behavior.
- Wallet rows are locked using `SELECT ... FOR UPDATE` during transfer execution.
- Locking order is deterministic (sorted user IDs) to avoid deadlock in concurrent transfers.
- Validations enforced before commit:
  - Amount must be greater than zero.
  - Sender cannot transfer to self.
  - Sender balance must be sufficient.
- Transfer record is written only after both wallet updates succeed.
- Any failure triggers rollback, so no partial update is persisted.

## Author Note
- Wallet implementation approach in this project is based on practical backend experience while working at `cicil.co.id`

## Postman Collection
- You can import ready-to-use collection from `docs/postman_collection.json`.
- In Postman: `Import` -> `Upload Files` -> choose `docs/postman_collection.json`.
- Collection includes auth, wallet, and transaction requests with sample payloads.

## Response Format
Success:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {}
}
```

Error:

```json
{
  "status": "ERROR",
  "message": "error message",
  "data": null
}
```

## API Documentation
Base URL:

```bash
export BASE_URL="http://localhost:8080"
```

### 1) Register
- Method: `POST`
- Path: `/api/auth/register`
- Auth: `No`

Body:

```json
{
  "name": "Alice",
  "email": "alice@example.com",
  "password": "password123"
}
```

Curl:

```bash
curl -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

### 2) Login
- Method: `POST`
- Path: `/api/auth/login`
- Auth: `No`

Body:

```json
{
  "email": "alice@example.com",
  "password": "password123"
}
```

Curl:

```bash
curl -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123"
  }'
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "token": "<jwt_token>"
  }
}
```

Set token manually from login response:

```bash
export TOKEN="<jwt_token_from_login_response>"
```

### 3) Logout
- Method: `POST`
- Path: `/api/auth/logout`
- Auth: `Yes`

Curl:

```bash
curl -X POST "$BASE_URL/api/auth/logout" \
  -H "Authorization: Bearer $TOKEN"
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "message": "logged out"
  }
}
```

### 4) Get Profile
- Method: `GET`
- Path: `/api/users/profile`
- Auth: `Yes`

Curl:

```bash
curl -X GET "$BASE_URL/api/users/profile" \
  -H "Authorization: Bearer $TOKEN"
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

### 5) Top Up Wallet
- Method: `POST`
- Path: `/api/wallets/topup`
- Auth: `Yes`

Body:

```json
{
  "amount": 50000
}
```

Curl:

```bash
curl -X POST "$BASE_URL/api/wallets/topup" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 50000
  }'
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "balance": 50000
  }
}
```

### 6) Get Wallet Balance
- Method: `GET`
- Path: `/api/wallets/balance`
- Auth: `Yes`

Curl:

```bash
curl -X GET "$BASE_URL/api/wallets/balance" \
  -H "Authorization: Bearer $TOKEN"
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "balance": 50000
  }
}
```

### 7) Transfer
- Method: `POST`
- Path: `/api/transactions/transfer`
- Auth: `Yes`

Body:

```json
{
  "receiver_email": "bob@example.com",
  "amount": 10000
}
```

Curl:

```bash
curl -X POST "$BASE_URL/api/transactions/transfer" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver_email": "bob@example.com",
    "amount": 10000
  }'
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "transaction_id": 1,
    "status": "completed"
  }
}
```

### 8) Transaction History
- Method: `GET`
- Path: `/api/transactions/history`
- Auth: `Yes`
- Query Params:
  - `limit` optional, default `50`, max `100`

Curl:

```bash
curl -X GET "$BASE_URL/api/transactions/history?limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

Response example:

```json
{
  "status": "OK",
  "message": "Response result successfully",
  "data": {
    "transactions": [
      {
        "id": 1,
        "sender": {
          "id": 1,
          "name": "Alice",
          "email": "alice@example.com"
        },
        "receiver": {
          "id": 2,
          "name": "Bob",
          "email": "bob@example.com"
        },
        "amount": 10000,
        "status": "completed",
        "created_at": "2026-02-12T10:00:00Z"
      }
    ]
  }
}
```

## Testing
Run all tests:

```bash
go test ./...
```

### Integration Test
- Integration test file: `tests/integration_test.go`
- Scope covered:
  - Register user
  - Login user
  - Top up wallet
  - Transfer balance
  - Fetch transaction history (with sender/receiver user info)
- Test uses real PostgreSQL and runs migration reset (`down` then `up`) before execution.

Required test DB env (already available in `.env.example`):
- `TEST_DB_HOST`
- `TEST_DB_PORT`
- `TEST_DB_USER`
- `TEST_DB_PASSWORD`
- `TEST_DB_NAME`
- `TEST_DB_SSLMODE`

Run integration test only:

```bash
set -a
source .env
set +a
go test ./tests -v
```
