package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalPaymentSpendPolicyEndpoint = "/internal/config/payment-spend-policy"

type PaymentSpendPolicyController struct {
	getPaymentSpendPolicy    *use_cases.GetPaymentSpendPolicy
	updatePaymentSpendPolicy *use_cases.UpdatePaymentSpendPolicy
	deletePaymentSpendPolicy *use_cases.DeletePaymentSpendPolicy
	errorWriter              *mvc.ErrorWriter
}

func NewPaymentSpendPolicyController(
	getPaymentSpendPolicy *use_cases.GetPaymentSpendPolicy,
	updatePaymentSpendPolicy *use_cases.UpdatePaymentSpendPolicy,
	deletePaymentSpendPolicy *use_cases.DeletePaymentSpendPolicy,
	errorWriter *mvc.ErrorWriter,
) *PaymentSpendPolicyController {
	return &PaymentSpendPolicyController{
		getPaymentSpendPolicy:    getPaymentSpendPolicy,
		updatePaymentSpendPolicy: updatePaymentSpendPolicy,
		deletePaymentSpendPolicy: deletePaymentSpendPolicy,
		errorWriter:              errorWriter,
	}
}

func (c *PaymentSpendPolicyController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentSpendPolicyEndpoint, c.Get)
	app.Put(InternalPaymentSpendPolicyEndpoint, c.Update)
	app.Delete(InternalPaymentSpendPolicyEndpoint, c.Delete)
}

// Get returns the spend policy head, materializing it on first access.
// @Summary Internal get payment spend policy
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.PaymentSpendPolicyResponse
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "spend_policy_failed"
// @Router /internal/config/payment-spend-policy [get]
func (c *PaymentSpendPolicyController) Get(ctx fiber.Ctx) error {
	resp, err := c.getPaymentSpendPolicy.Execute(ctx.Context())
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update replaces the payees the agent may pay.
// @Summary Internal update payment spend policy
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.UpdatePaymentSpendPolicyRequest true "payment spend policy"
// @Success 200 {object} responses.PaymentSpendPolicyResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "spend_policy_failed or update_failed"
// @Router /internal/config/payment-spend-policy [put]
func (c *PaymentSpendPolicyController) Update(ctx fiber.Ctx) error {
	body := new(requests.UpdatePaymentSpendPolicyRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_body", err.Error()))
	}
	resp, err := c.updatePaymentSpendPolicy.Execute(ctx.Context(), body)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes the policy head and its rules, leaving the agent unable to pay.
// @Summary Internal delete payment spend policy
// @Tags Internal
// @Success 204
// @Failure 500 {object} errors.InternalError "delete_failed"
// @Router /internal/config/payment-spend-policy [delete]
func (c *PaymentSpendPolicyController) Delete(ctx fiber.Ctx) error {
	if err := c.deletePaymentSpendPolicy.Execute(ctx.Context()); err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
