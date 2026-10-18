package agent_reasoning_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
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
// @Failure 401 {object} errors.UnauthorizedError
// @Failure 409 {object} errors.ConflictError
// @Failure 500 {object} errors.InternalError "create_response_failed"
// @Security CallerId
// @Router /internal/responses [post]
func (c *ResponseController) CreateResponse(ctx fiber.Ctx) error {
	body := new(requests.CreateResponseRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.createResponse.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
