package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/gofiber/fiber/v2"
)

const PaymentPolicyEndpoint = "/config/payment-policy"

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

func (c *PaymentPolicyController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *PaymentPolicyController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(PaymentPolicyEndpoint, c.Get)
	app.Delete(PaymentPolicyEndpoint, c.Delete)
}

// Get returns the policy head.
// @Summary Get payment policy
// @Tags Config
// @Produce json
// @Success 200 {object} responses.GetPaymentPolicyResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-policy [get]
func (c *PaymentPolicyController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getPaymentPolicy.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete resets the policy to its seeded default.
// @Summary Delete payment policy
// @Tags Config
// @Produce json
// @Success 204 ""
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-policy [delete]
func (c *PaymentPolicyController) Delete(ctx *fiber.Ctx) error {
	if err := c.deletePaymentPolicy.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
