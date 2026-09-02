package payment

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/gofiber/fiber/v3"
)

const PaymentSpendPolicyEndpoint = "/config/payment-spend-policy"

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

func (c *PaymentSpendPolicyController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *PaymentSpendPolicyController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(PaymentSpendPolicyEndpoint, c.Get)
	app.Delete(PaymentSpendPolicyEndpoint, c.Delete)
}

// Get returns the spend policy head.
// @Summary Get payment spend policy
// @Tags Config
// @Produce json
// @Success 200 {object} responses.GetPaymentSpendPolicyResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy [get]
func (c *PaymentSpendPolicyController) Get(ctx fiber.Ctx) error {
	resp, err := c.getPaymentSpendPolicy.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes every spend rule, leaving the agent unable to pay.
// @Summary Delete payment spend policy
// @Tags Config
// @Produce json
// @Success 204 ""
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy [delete]
func (c *PaymentSpendPolicyController) Delete(ctx fiber.Ctx) error {
	if err := c.deletePaymentSpendPolicy.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
