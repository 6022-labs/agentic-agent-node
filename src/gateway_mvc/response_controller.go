package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const ResponsesEndpoint = "/responses"

type ResponseController struct {
	createResponse use_cases.CreateResponseInterface
}

func NewResponseController(createResponse use_cases.CreateResponseInterface) *ResponseController {
	return &ResponseController{createResponse: createResponse}
}

func (c *ResponseController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *ResponseController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Post(ResponsesEndpoint, c.CreateResponse)
}

// CreateResponse runs the agent's inference loop for one turn.
// @Summary Create response
// @Tags Responses
// @Accept json
// @Produce json
// @Param request body requests.CreateResponseRequest true "Conversation turn"
// @Success 200 {object} responses.Response
// @Failure 400 {object} errors.ValidationError
// @Failure 502 {object} fiber.Map "runtime_failed"
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

	result, err := c.createResponse.Execute(ctx.Context(), body)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"code": "runtime_failed", "message": err.Error()})
	}
	return ctx.JSON(result)
}
