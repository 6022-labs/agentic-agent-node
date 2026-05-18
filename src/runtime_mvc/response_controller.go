package runtime_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc_result"
	"github.com/gofiber/fiber/v2"
)

const InternalResponsesEndpoint = "/internal/responses"

type ResponseController struct {
	createResponse use_cases.CreateResponseInterface
}

func NewResponseController(createResponse use_cases.CreateResponseInterface) *ResponseController {
	return &ResponseController{createResponse: createResponse}
}

func (c *ResponseController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalResponsesEndpoint, c.CreateResponse)
}

// CreateResponse runs the agent's inference loop and persists the resulting
// assistant turn alongside any new inputs.
// @Summary Internal create response
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.CreateResponseRequest true "Inputs"
// @Success 200 {object} responses.CreateResponseResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "create_response_failed"
// @Router /internal/responses [post]
func (c *ResponseController) CreateResponse(ctx *fiber.Ctx) error {
	body := new(requests.CreateResponseRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if err := body.ValidateAndSanitize(); err != nil {
		if vErr, ok := err.(*errors.ValidationError); ok {
			return ctx.Status(fiber.StatusBadRequest).JSON(vErr)
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	return mvc_result.Write(ctx, c.createResponse.Execute(ctx.Context(), body))
}
