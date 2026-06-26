package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v2"
)

const InternalPaymentPolicyEndpoint = "/internal/config/payment-policy"

type PaymentPolicyController struct {
	getPaymentPolicy    *use_cases.GetPaymentPolicy
	deletePaymentPolicy *use_cases.DeletePaymentPolicy
}

func NewPaymentPolicyController(
	getPaymentPolicy *use_cases.GetPaymentPolicy,
	deletePaymentPolicy *use_cases.DeletePaymentPolicy,
) *PaymentPolicyController {
	return &PaymentPolicyController{
		getPaymentPolicy:    getPaymentPolicy,
		deletePaymentPolicy: deletePaymentPolicy,
	}
}

func (c *PaymentPolicyController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentPolicyEndpoint, c.Get)
	app.Delete(InternalPaymentPolicyEndpoint, c.Delete)
}

// Get returns the policy head, materializing it on first access.
// @Summary Internal get payment policy
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.PaymentPolicyResponse
// @Failure 500 {object} fiber.Map "get_payment_policy_failed"
// @Router /internal/config/payment-policy [get]
func (c *PaymentPolicyController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getPaymentPolicy.Execute(ctx.UserContext())
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
// @Failure 500 {object} fiber.Map "delete_payment_policy_failed"
// @Router /internal/config/payment-policy [delete]
func (c *PaymentPolicyController) Delete(ctx *fiber.Ctx) error {
	if err := c.deletePaymentPolicy.Execute(ctx.UserContext()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
