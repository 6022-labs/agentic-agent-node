package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalPaymentReceivePolicyEndpoint = "/internal/config/payment-receive-policy"

type PaymentReceivePolicyController struct {
	getPaymentReceivePolicy    *use_cases.GetPaymentReceivePolicy
	deletePaymentReceivePolicy *use_cases.DeletePaymentReceivePolicy
}

func NewPaymentReceivePolicyController(
	getPaymentReceivePolicy *use_cases.GetPaymentReceivePolicy,
	deletePaymentReceivePolicy *use_cases.DeletePaymentReceivePolicy,
) *PaymentReceivePolicyController {
	return &PaymentReceivePolicyController{
		getPaymentReceivePolicy:    getPaymentReceivePolicy,
		deletePaymentReceivePolicy: deletePaymentReceivePolicy,
	}
}

func (c *PaymentReceivePolicyController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentReceivePolicyEndpoint, c.Get)
	app.Delete(InternalPaymentReceivePolicyEndpoint, c.Delete)
}

// Get returns the policy head, materializing it on first access.
// @Summary Internal get payment policy
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.PaymentReceivePolicyResponse
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "policy_failed"
// @Router /internal/config/payment-receive-policy [get]
func (c *PaymentReceivePolicyController) Get(ctx fiber.Ctx) error {
	resp, err := c.getPaymentReceivePolicy.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete resets the policy: head and rules are removed, the default is
// re-seeded on next access.
// @Summary Internal delete payment policy
// @Tags Internal
// @Success 204
// @Failure 500 {object} errors.InternalError "delete_failed"
// @Router /internal/config/payment-receive-policy [delete]
func (c *PaymentReceivePolicyController) Delete(ctx fiber.Ctx) error {
	if err := c.deletePaymentReceivePolicy.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
