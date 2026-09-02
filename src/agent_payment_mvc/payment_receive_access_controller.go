package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalPaymentReceiveAccessEndpoint = "/internal/payment/receive/access"

type PaymentReceiveAccessController struct {
	authorizePaymentReceiveAccess *use_cases.AuthorizePaymentReceiveAccess
}

func NewPaymentReceiveAccessController(
	authorizePaymentReceiveAccess *use_cases.AuthorizePaymentReceiveAccess,
) *PaymentReceiveAccessController {
	return &PaymentReceiveAccessController{authorizePaymentReceiveAccess: authorizePaymentReceiveAccess}
}

func (c *PaymentReceiveAccessController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalPaymentReceiveAccessEndpoint, c.Authorize)
}

// Authorize decides whether the caller may reach a gated route.
// @Summary Internal authorize payment receive access
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.AuthorizePaymentReceiveAccessRequest true "Caller credentials"
// @Success 200 {object} responses.AuthorizePaymentReceiveAccessResponse
// @Failure 400 {object} fiber.Map "validation_error"
// @Failure 500 {object} fiber.Map "authorize_payment_receive_access_failed"
// @Router /internal/payment/receive/access [post]
func (c *PaymentReceiveAccessController) Authorize(ctx fiber.Ctx) error {
	request := &requests.AuthorizePaymentReceiveAccessRequest{}
	if err := ctx.Bind().Body(request); err != nil {
		return mvc.WriteError(ctx, err)
	}

	resp, err := c.authorizePaymentReceiveAccess.Execute(ctx.Context(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
