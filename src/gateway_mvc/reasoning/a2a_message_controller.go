package reasoning

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares"
	"github.com/gofiber/fiber/v3"
)

const A2AMessageEndpoint = "/a2a"

type A2AMessageController struct {
	sendA2AMessage               use_cases.SendA2AMessageInterface
	responseAccessGateMiddleware *middlewares.ResponseAccessGateMiddleware
}

func NewA2AMessageController(
	sendA2AMessage use_cases.SendA2AMessageInterface,
	responseAccessGateMiddleware *middlewares.ResponseAccessGateMiddleware,
) *A2AMessageController {
	return &A2AMessageController{
		sendA2AMessage:               sendA2AMessage,
		responseAccessGateMiddleware: responseAccessGateMiddleware,
	}
}

// RegisterUnauthenticatedRoutes mounts /a2a behind the same x402 gate as /responses.
func (c *A2AMessageController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Post(A2AMessageEndpoint, c.responseAccessGateMiddleware.Handle, c.Handle)
}

func (c *A2AMessageController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// Handle runs one A2A SendMessage turn; failures return HTTP error status so the x402 gate never settles a failed turn.
// @Summary A2A SendMessage (JSON-RPC)
// @Description JSON-RPC 2.0 SendMessage (A2A v1.0): one inference turn, immediate reply. x402-gated like /responses.
// @Tags A2A
// @Accept json
// @Produce json
// @Param PAYMENT-SIGNATURE header string false "x402 v2 payment payload (base64 JSON)"
// @Param request body requests.A2ARequest true "JSON-RPC SendMessage request"
// @Success 200 {object} responses.A2AResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError
// @Failure 402 {object} x402.PaymentRequired
// @Failure 409 {object} errors.ConflictError "conversation_unavailable or agent_not_minted"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Router /a2a [post]
func (c *A2AMessageController) Handle(ctx fiber.Ctx) error {
	body := new(requests.A2ARequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.sendA2AMessage.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
