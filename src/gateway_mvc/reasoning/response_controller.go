package reasoning

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/payment/middlewares"
	"github.com/gofiber/fiber/v2"
)

const ResponsesEndpoint = "/responses"

type ResponseController struct {
	createResponse        use_cases.CreateResponseInterface
	paymentGateMiddleware *middlewares.PaymentGateMiddleware
}

func NewResponseController(
	createResponse use_cases.CreateResponseInterface,
	paymentGateMiddleware *middlewares.PaymentGateMiddleware,
) *ResponseController {
	return &ResponseController{
		createResponse:        createResponse,
		paymentGateMiddleware: paymentGateMiddleware,
	}
}

// RegisterUnauthenticatedRoutes mounts /responses outside the JWT group: the
// payment gate dispatches to the method matching the caller's proof (today
// only x402), pricing each caller via the rules whose patterns match them
// (price 0 = free).
func (c *ResponseController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Post(ResponsesEndpoint, c.paymentGateMiddleware.Handle, c.CreateResponse)
}

func (c *ResponseController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// CreateResponse runs the agent's inference loop for one turn.
// @Summary Create response
// @Description x402 v2 gated: send a PAYMENT-SIGNATURE header. Each caller pays the cheapest accept option whose address patterns match them — zero-price options (agent wallet and NFT owner by default) are served free without settlement. Anonymous 402 responses advertise only the options open to everyone.
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
func (c *ResponseController) CreateResponse(ctx *fiber.Ctx) error {
	body := new(requests.CreateResponseRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.createResponse.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
