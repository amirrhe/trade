# Trade-Go

A small, pluggable crypto‑trading application in Go, built with a hexagonal (ports & adapters) architecture.  
It provides an HTTP API (via Fiber) and supports multiple exchanges through adapters (Bitpin, Wallex, etc.).

## Features

- **Hexagonal architecture**: clear separation of domain, application, adapters, and transport layers  
- **Multiple exchanges**: Bitpin and Wallex adapters out‑of‑the‑box, easy to add more  
- **Automatic token management**: background refresh + inline refresh within 30 s of expiry  
- **Structured JSON logging** via a `LoggerPort` interface  
- **HTTP API** (Fiber) exposing:
  - `POST /v1/orders` → create order  
  - `DELETE /v1/orders/:symbol/:id` → cancel order  
  - `GET /v1/balance` → get account balances  
  - `GET /v1/book/:symbol` → get orderbook  
- **Dockerized**, production‑ready static binary

---

## Prerequisites

- Go **1.22** (as declared in `go.mod`)  
- Git  
- (Optional) Docker & Docker Compose  
- API credentials for Bitpin or Wallex

---

## Setup

1. **Clone the repo**

   ```bash
   git clone https://github.com/you/Trade-Go.git
   cd Trade-Go
````

2. **Environment**

   Copy the sample env file and fill in your credentials:

   ```bash
   cp .env.sample .env
   ```

3. **Fetch dependencies**

   ```bash
   go mod download
   ```

---

## Build & Run

### Locally with `go run`

```bash
go run ./cmd/http
```

By default it listens on `:$HTTP_PORT` (8080 if unset).


### Docker

```bash
docker build -t trade-app .

docker run -d \
  --name trade \
  -p 8080:8080 \
  --env-file .env \
  trade-app
```

---

## API Documentation

Swagger specs live under `docs/`. Once your server is up, point any Swagger UI at:

```
http://localhost:8080/docs/swagger.yaml
```

(or serve `docs/swagger.yaml` through your preferred UI)

---

## Architecture Overview

1. **`internal/domain`**

   * `ExchangePort` & `LoggerPort`
   * Domain models (`OrderRequest`, `OrderResponse`, `Balance`, `OrderBook`)

2. **`internal/application`**

   * `TradingService` implements business use‑cases

3. **`internal/adapters`**

   * Bitpin & Wallex adapters (HTTP, JSON mapping, logging)

4. **`pkg/transport`**

   * Fiber router and handlers

5. **`internal/infrastructure`**

   * `.env` loader and DI wiring

---


