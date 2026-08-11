package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v2"
)

const InternalPaymentFacilitatorsEndpoint = "/internal/config/payment-facilitators"

type PaymentFacilitatorController struct {
	getPaymentFacilitators *use_cases.GetPaymentFacilitators
}

func NewPaymentFacilitatorController(
	getPaymentFacilitators *use_cases.GetPaymentFacilitators,
) *PaymentFacilitatorController {
	return &PaymentFacilitatorController{getPaymentFacilitators: getPaymentFacilitators}
}

func (c *PaymentFacilitatorController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentFacilitatorsEndpoint, c.Get)
}

// Get returns the network→facilitator mapping the gateway settles through.
// @Summary Internal get payment facilitators
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.PaymentFacilitatorsResponse
// @Router /internal/config/payment-facilitators [get]
func (c *PaymentFacilitatorController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getPaymentFacilitators.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
