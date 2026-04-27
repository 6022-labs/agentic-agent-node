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

// CreateResponse runs the agent's inner loop for one participant turn.
// @Summary Create response
// @Description Runs LLM ↔ MCP ↔ knowledge loop (P2.4 — currently 501).
// @Tags Responses
// @Accept json
// @Produce json
// @Param request body requests.CreateResponseRequest true "Conversation turn"
// @Success 501 {object} fiber.Map
// @Failure 400 {object} errors.ValidationError
// @Security BearerAuth
// @Router /responses [post]
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
