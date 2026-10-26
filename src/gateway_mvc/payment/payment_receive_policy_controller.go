package payment

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/gofiber/fiber/v3"
)

const PaymentReceivePolicyEndpoint = "/config/payment-receive-policy"

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

func (c *PaymentReceivePolicyController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *PaymentReceivePolicyController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(PaymentReceivePolicyEndpoint, c.Get)
	app.Delete(PaymentReceivePolicyEndpoint, c.Delete)
}

// Get returns the policy head.
// @Summary Get payment policy
// @Tags Config
// @Produce json
// @Success 200 {object} responses.GetPaymentReceivePolicyResponse
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-receive-policy [get]
func (c *PaymentReceivePolicyController) Get(ctx fiber.Ctx) error {
	resp, err := c.getPaymentReceivePolicy.Execute(ctx.Context())
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
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-receive-policy [delete]
func (c *PaymentReceivePolicyController) Delete(ctx fiber.Ctx) error {
	if err := c.deletePaymentReceivePolicy.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
