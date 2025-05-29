// pkg/transport/router.go
package transport

import (
	"trade/internal/application"
	"trade/internal/domain"
	"trade/internal/ports"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a JSON error response
// swagger:response ErrorResponse
// in: body
// schema:
//
//	type: object
//	properties:
//	  error:
//	    type: string
//	    example: "error message"
type ErrorResponse struct {
	Error string `json:"error"`
}

func NewRouter(svc *application.TradingService, log ports.LoggerPort) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
		},
	})

	api := app.Group("/v1")
	app.Use(func(c *fiber.Ctx) error {
		log.Info(c.Context(), "HTTP request", ports.Fields{
			"method": c.Method(),
			"path":   c.OriginalURL(),
		})
		return c.Next()
	})
	api.Post("/orders", createOrderHandler(svc))
	api.Delete("/orders/:symbol/:id", cancelOrderHandler(svc))
	api.Get("/balance", getBalanceHandler(svc))
	api.Get("/book/:symbol", getOrderBookHandler(svc))

	return app
}

// createOrderHandler parses a JSON body into OrderRequest and calls CreateOrder.
// @Summary Create a new order
// @Description Place a new order on the configured exchange
// @Tags orders
// @Accept application/json
// @Produce application/json
// @Param order body domain.OrderRequest true "Order payload"
// @Success 201 {object} domain.OrderResponse
// @Failure 400 {object} transport.ErrorResponse
// @Failure 500 {object} transport.ErrorResponse
// @Router /v1/orders [post]
func createOrderHandler(svc *application.TradingService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req domain.OrderRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		}

		resp, err := svc.CreateOrder(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(resp)
	}
}

// cancelOrderHandler reads symbol and id from the path and calls CancelOrder.
// @Summary Cancel an existing order
// @Description Cancel a placed order by symbol and ID
// @Tags orders
// @Param symbol path string true "Trading symbol"
// @Param id path string true "Order ID"
// @Success 204
// @Failure 500 {object} transport.ErrorResponse
// @Router /v1/orders/{symbol}/{id} [delete]
func cancelOrderHandler(svc *application.TradingService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		symbol := c.Params("symbol")
		id := c.Params("id")

		if err := svc.CancelOrder(c.Context(), symbol, id); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// getBalanceHandler calls GetBalance and returns account balances.
// @Summary Get account balances
// @Description Retrieve all asset balances of the account
// @Tags balance
// @Produce application/json
// @Success 200 {array} domain.Balance
// @Failure 500 {object} transport.ErrorResponse
// @Router /v1/balance [get]
func getBalanceHandler(svc *application.TradingService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		balances, err := svc.GetBalance(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.JSON(balances)
	}
}

// getOrderBookHandler fetches the order book for the given symbol.
// @Summary Get order book
// @Description Fetch the current order book for a trading symbol
// @Tags market
// @Param symbol path string true "Trading symbol"
// @Produce application/json
// @Success 200 {object} domain.OrderBook
// @Failure 500 {object} transport.ErrorResponse
// @Router /v1/book/{symbol} [get]
func getOrderBookHandler(svc *application.TradingService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		symbol := c.Params("symbol")
		book, err := svc.GetOrderBook(c.Context(), symbol)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.JSON(book)
	}
}
