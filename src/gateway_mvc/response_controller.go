package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const ResponsesEndpoint = "/responses"

type ResponseController struct{}

func NewResponseController() *ResponseController {
	return &ResponseController{}
}

func (c *ResponseController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *ResponseController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Post(ResponsesEndpoint, c.CreateResponse)
}

func (c *ResponseController) CreateResponse(ctx *fiber.Ctx) error {
	body := new(requests.CreateResponseRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	return ctx.Status(fiber.StatusNotImplemented).JSON(notImplementedBody("responses.create"))
}
