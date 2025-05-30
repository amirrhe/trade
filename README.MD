# Trade-Go

A pluggable crypto-trading application in Go with a hexagonal architecture.

## Features

- Supports multiple exchanges (Bitpin, Wallex, etc.)
- Provides an HTTP API (Fiber)

## Overview

This application provides an HTTP API for trading cryptocurrencies, built with a hexagonal (ports & adapters) architecture for flexibility and maintainability. It supports multiple exchanges through adapters.

## Features

- **Hexagonal architecture**: Ensures clear separation of concerns.
- **Multiple exchanges**: Supports Bitpin and Wallex out-of-the-box.
- **Automatic token management**: Refreshes tokens in the background.
- **Structured JSON logging**: Uses a `LoggerPort` interface for logging.
- **HTTP API**: Provides endpoints for creating, canceling, and retrieving orders and balances.
- **Dockerized**: Ready for production deployment.
## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

---

## Prerequisites

- Go 1.22 (as declared in `go.mod`)
- Git
- Docker (optional)
- API credentials for Bitpin or Wallex

---

## Setup

1.  **Clone the repository**

    ```bash
    git clone https://github.com/you/Trade-Go.git
    cd Trade-Go
    ```

2.  **Configure environment variables**

    Copy `.env.sample` to `.env` and update with your API keys and secrets.

    ```bash
    cp .env.sample .env
    ```

3.  **Download dependencies**

    ```bash
    go mod download
    ```

---

## Configuration

The application is configured using environment variables. Here's a list of the available options:

-   `EXCHANGE`: The exchange to use (`bitpin` or `wallex`). Default is `bitpin`.
-   `HTTP_PORT`: The port for the HTTP server to listen on. Default is `8080`.
-   `LOG_LEVEL`: The logging level (`debug`, `info`, `warn`, `error`, `fatal`, `panic`). Default is `info`.
-   `BITPIN_API_KEY`: The API key for Bitpin.
-   `BITPIN_API_SECRET`: The API secret for Bitpin.
-   `BITPIN_BASE_URL`: The base URL for Bitpin API. Default is `https://api.bitpin.ir`.
-   `WALLEX_API_KEY`: The API key for Wallex.
-   `WALLEX_BASE_URL`: The base URL for Wallex API. Default is `https://api.wallex.ir`.

---

## Build & Run

### Locally

```bash
go run ./cmd/http
```

The server listens on the port specified by the `$HTTP_PORT` environment variable (default: 8080).

### Docker

```bash
docker build -t trade-app .
docker run -d --name trade -p 8080:8080 --env-file .env trade-app
```

---

## API Documentation

The API documentation is available in Swagger format under the `docs/` directory. You can view it using a Swagger UI:

```
http://localhost:8080/docs/swagger.yaml
```

---

## Architecture

1.  **Domain Layer (`internal/domain`)**

    - Defines the core business entities and interfaces (`ExchangePort`, `LoggerPort`, `OrderRequest`, `OrderResponse`, `Balance`, `OrderBook`).
    - Contains the business logic and rules of the application.

2.  **Application Layer (`internal/application`)**

    - Implements the use cases of the application by orchestrating the domain layer.
    - Contains the `TradingService` which provides methods for creating orders, canceling orders, getting balances, and getting order books.

3.  **Adapters Layer (`internal/adapters`)**

    - Implements the interfaces defined in the domain layer for specific exchanges (Bitpin, Wallex).
    - Handles communication with external systems, such as APIs and databases.
    - Contains the `BitpinAdapter` and `WallexAdapter` which implement the `ExchangePort` interface.

4.  **Transport Layer (`pkg/transport`)**

    - Exposes the application's functionality through an HTTP API using Fiber.
    - Handles routing, request parsing, and response formatting.
    - Contains the `NewRouter` function which defines the API endpoints.

5.  **Infrastructure Layer (`internal/infrastructure`)**

    - Provides the infrastructure components required by the application, such as configuration loading and dependency injection.
    - Contains the `LoadConfig` function which loads the application configuration from environment variables.
    - Contains the `BuildApp` function which wires up the application dependencies using Wire.

---


