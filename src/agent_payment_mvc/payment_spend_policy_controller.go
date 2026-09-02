package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalPaymentSpendPolicyEndpoint = "/internal/config/payment-spend-policy"

type PaymentSpendPolicyController struct {
	getPaymentSpendPolicy    *use_cases.GetPaymentSpendPolicy
	deletePaymentSpendPolicy *use_cases.DeletePaymentSpendPolicy
}

func NewPaymentSpendPolicyController(
	getPaymentSpendPolicy *use_cases.GetPaymentSpendPolicy,
	deletePaymentSpendPolicy *use_cases.DeletePaymentSpendPolicy,
) *PaymentSpendPolicyController {
	return &PaymentSpendPolicyController{
		getPaymentSpendPolicy:    getPaymentSpendPolicy,
		deletePaymentSpendPolicy: deletePaymentSpendPolicy,
	}
}

func (c *PaymentSpendPolicyController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentSpendPolicyEndpoint, c.Get)
	app.Delete(InternalPaymentSpendPolicyEndpoint, c.Delete)
}

// Get returns the spend policy head, materializing it on first access.
// @Summary Internal get payment spend policy
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.PaymentSpendPolicyResponse
// @Failure 500 {object} fiber.Map "get_payment_spend_policy_failed"
// @Router /internal/config/payment-spend-policy [get]
func (c *PaymentSpendPolicyController) Get(ctx fiber.Ctx) error {
	resp, err := c.getPaymentSpendPolicy.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes the policy head and its rules, leaving the agent unable to pay.
// @Summary Internal delete payment spend policy
// @Tags Internal
// @Success 204
// @Failure 500 {object} fiber.Map "delete_payment_spend_policy_failed"
// @Router /internal/config/payment-spend-policy [delete]
func (c *PaymentSpendPolicyController) Delete(ctx fiber.Ctx) error {
	if err := c.deletePaymentSpendPolicy.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
