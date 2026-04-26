package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const DbConfigEndpoint = "/config/db"

type DbConfigController struct{}

func NewDbConfigController() *DbConfigController {
	return &DbConfigController{}
}

func (c *DbConfigController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *DbConfigController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(DbConfigEndpoint, c.Get)
	app.Post(DbConfigEndpoint, c.Set)
}

func (c *DbConfigController) Get(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.db.get"))
}

func (c *DbConfigController) Set(ctx *fiber.Ctx) error {
	body := new(requests.SetDbConfigRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("config.db.set"))
}
