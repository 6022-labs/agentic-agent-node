package payment

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const PaymentSpendPolicyEndpoint = "/config/payment-spend-policy"

type PaymentSpendPolicyController struct {
	getPaymentSpendPolicy    *use_cases.GetPaymentSpendPolicy
	updatePaymentSpendPolicy *use_cases.UpdatePaymentSpendPolicy
	deletePaymentSpendPolicy *use_cases.DeletePaymentSpendPolicy
}

func NewPaymentSpendPolicyController(
	getPaymentSpendPolicy *use_cases.GetPaymentSpendPolicy,
	updatePaymentSpendPolicy *use_cases.UpdatePaymentSpendPolicy,
	deletePaymentSpendPolicy *use_cases.DeletePaymentSpendPolicy,
) *PaymentSpendPolicyController {
	return &PaymentSpendPolicyController{
		getPaymentSpendPolicy:    getPaymentSpendPolicy,
		updatePaymentSpendPolicy: updatePaymentSpendPolicy,
		deletePaymentSpendPolicy: deletePaymentSpendPolicy,
	}
}

func (c *PaymentSpendPolicyController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *PaymentSpendPolicyController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(PaymentSpendPolicyEndpoint, c.Get)
	app.Put(PaymentSpendPolicyEndpoint, c.Update)
	app.Delete(PaymentSpendPolicyEndpoint, c.Delete)
}

// Get returns the spend policy head and the payees the agent may pay.
// @Summary Get payment spend policy
// @Tags Config
// @Produce json
// @Success 200 {object} responses.GetPaymentSpendPolicyResponse
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy [get]
func (c *PaymentSpendPolicyController) Get(ctx fiber.Ctx) error {
	resp, err := c.getPaymentSpendPolicy.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update replaces the payees the agent may pay.
// @Summary Update payment spend policy
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.UpdatePaymentSpendPolicyRequest true "spend policy"
// @Success 200 {object} responses.GetPaymentSpendPolicyResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy [put]
func (c *PaymentSpendPolicyController) Update(ctx fiber.Ctx) error {
	body := new(requests.UpdatePaymentSpendPolicyRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updatePaymentSpendPolicy.Execute(ctx.Context(), body)
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
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy [delete]
func (c *PaymentSpendPolicyController) Delete(ctx fiber.Ctx) error {
	if err := c.deletePaymentSpendPolicy.Execute(ctx.Context()); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
