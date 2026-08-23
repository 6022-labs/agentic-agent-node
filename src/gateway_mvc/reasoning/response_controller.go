package reasoning

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares"
	"github.com/gofiber/fiber/v3"
)

const ResponsesEndpoint = "/responses"

type ResponseController struct {
	createResponse               use_cases.CreateResponseInterface
	responseAccessGateMiddleware *middlewares.ResponseAccessGateMiddleware
}

func NewResponseController(
	createResponse use_cases.CreateResponseInterface,
	responseAccessGateMiddleware *middlewares.ResponseAccessGateMiddleware,
) *ResponseController {
	return &ResponseController{
		createResponse:               createResponse,
		responseAccessGateMiddleware: responseAccessGateMiddleware,
	}
}

// RegisterUnauthenticatedRoutes mounts /responses outside the JWT group, gated by the access chain.
func (c *ResponseController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Post(ResponsesEndpoint, c.responseAccessGateMiddleware.Handle, c.CreateResponse)
}

func (c *ResponseController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// CreateResponse runs the agent's inference loop for one turn.
// @Summary Create response
// @Description x402 v2 gated (PAYMENT-SIGNATURE header): callers pay the cheapest matching option; zero-price is free.
// @Tags Responses
// @Accept json
// @Produce json
// @Param PAYMENT-SIGNATURE header string true "x402 v2 payment payload (base64 JSON)"
// @Param request body requests.CreateResponseRequest true "Conversation turn"
// @Success 200 {object} responses.Response
// @Failure 400 {object} errors.ValidationError
// @Failure 402 {object} x402.PaymentRequired
// @Failure 502 {object} fiber.Map "runtime_failed"
// @Router /responses [post]
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
