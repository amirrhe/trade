Here’s the cleaned‑up README with proper code fences and formatting:

````markdown
# Trade-Go

A small, pluggable crypto‐trading application in Go, built with a hexagonal (ports & adapters) architecture.  
It provides both an HTTP API (via Fiber) and a CLI entrypoint, and supports multiple exchanges through adapters (Bitpin, Wallex, etc.), complete with automatic token refresh, structured logging (Logrus), and full test coverage.

---

## 📂 Project Structure

```
.
├── cmd
│   └── http
│       └── main.go           # HTTP server entrypoint
├── internal
│   ├── adapters              # Exchange adapters (Bitpin, Wallex…)
│   ├── application           # Core “use cases” / service layer
│   ├── domain                # Business models & port interfaces
│   ├── infrastructure
│   │   ├── config            # Env/.env loading
│   │   └── di                # Dependency‐injection (Wire)
│   └── ports                 # LoggerPort, etc.
├── pkg
│   └── transport             # HTTP routing & handlers
├── docs                      # Swagger & design docs
├── go.mod
├── go.sum
└── Dockerfile
````

---

## 🚀 Features

* **Hexagonal architecture**: clear separation of domain, application, adapters, and transport layers
* **Multiple exchanges**: Bitpin and Wallex adapters out‑of‑the‑box, easy to add more
* **Automatic token management**: background refresh + inline refresh within 30s of expiry
* **Structured JSON logging** with Logrus, injected via a `LoggerPort` interface
* **HTTP API** (Fiber) exposing:

  * `POST /v1/orders` → create order
  * `DELETE /v1/orders/:symbol/:id` → cancel order
  * `GET /v1/balance` → get account balances
  * `GET /v1/book/:symbol` → get orderbook
* **Dockerized**, production‑ready static binary

---

## 🎲 Prerequisites

* Go **1.22** (as declared in `go.mod`)
* Git
* (Optional) Docker & Docker Compose
* Exchange API credentials for Bitpin or Wallex

---

## 🛠️ Setup

1. **Clone the repo**

   ```bash
   git clone https://github.com/you/Trade-Go.git
   cd Trade-Go
   ```

2. **Environment**

   Create a `.env` file in the project root:

   ```dotenv
   EXCHANGE=bitpin           # or "wallex"
   HTTP_PORT=8080
   LOG_LEVEL=info

   # Bitpin (if using Bitpin adapter)
   BITPIN_API_KEY=your_key
   BITPIN_API_SECRET=your_secret
   BITPIN_BASE_URL=https://api.bitpin.ir

   # Wallex (if using Wallex adapter)
   WALLEX_API_KEY=your_key
   WALLEX_BASE_URL=https://api.wallex.ir
   ```

3. **Fetch modules**

   ```bash
   go mod download
   ```

---

## ▶️ Build & Run

### Locally

```bash
# Build
go build -ldflags="-s -w" -o trade ./cmd/http

# Run
./trade
```

The server will start on `http://localhost:8080` (or `$HTTP_PORT`).

### Docker

```bash
# Build
docker build -t trade-app .

# Run
docker run -d \
  --name trade \
  -p 8080:8080 \
  --env-file .env \
  trade-app
```

---

## 📑 API Documentation

We’ve included Swagger docs under `docs/swagger.yaml`. After running the server, you can serve them with any Swagger UI, or point your browser at:

```
http://localhost:8080/docs/index.html
```

## 🏗️ Architecture Overview

1. **`internal/domain`**

   * `ExchangePort` & `LoggerPort` interfaces
   * Domain models (`OrderRequest`, `OrderResponse`, `Balance`, `OrderBook`)

2. **`internal/application`**

   * `TradingService` implements main use‑cases by calling `ExchangePort`

3. **`internal/adapters`**

   * Exchange‑specific logic (HTTP client, JSON mapping, logging)
   * Bitpin adapter with background token refresh
   * Wallex adapter with robust parsing of mixed JSON types

4. **`pkg/transport`**

   * Fiber router, request validation, and JSON responses

5. **`internal/infrastructure`**

   * Configuration loader (`.env`)
   * DI wiring to assemble the service, adapter, and router

This separation keeps your core business logic **exchange‑agnostic** and makes it trivial to plug in new adapters (e.g. KuCoin, Binance) with minimal code changes.

```
```
