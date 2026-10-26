package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalPaymentReceiveRulesEndpoint = "/internal/config/payment-receive-policy/rules"

type PaymentReceiveRuleController struct {
	listPaymentReceiveRules  *use_cases.ListPaymentReceiveRules
	createPaymentReceiveRule *use_cases.CreatePaymentReceiveRule
	updatePaymentReceiveRule *use_cases.UpdatePaymentReceiveRule
	deletePaymentReceiveRule *use_cases.DeletePaymentReceiveRule
}

func NewPaymentReceiveRuleController(
	listPaymentReceiveRules *use_cases.ListPaymentReceiveRules,
	createPaymentReceiveRule *use_cases.CreatePaymentReceiveRule,
	updatePaymentReceiveRule *use_cases.UpdatePaymentReceiveRule,
	deletePaymentReceiveRule *use_cases.DeletePaymentReceiveRule,
) *PaymentReceiveRuleController {
	return &PaymentReceiveRuleController{
		listPaymentReceiveRules:  listPaymentReceiveRules,
		createPaymentReceiveRule: createPaymentReceiveRule,
		updatePaymentReceiveRule: updatePaymentReceiveRule,
		deletePaymentReceiveRule: deletePaymentReceiveRule,
	}
}

func (c *PaymentReceiveRuleController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentReceiveRulesEndpoint, c.List)
	app.Post(InternalPaymentReceiveRulesEndpoint, c.Create)
	app.Put(InternalPaymentReceiveRulesEndpoint+"/:id", c.Update)
	app.Delete(InternalPaymentReceiveRulesEndpoint+"/:id", c.Delete)
}

// List returns the pricing rules in table order.
// @Summary Internal list payment rules
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.ListPaymentReceiveRulesResponse
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "policy_failed or list_failed"
// @Router /internal/config/payment-receive-policy/rules [get]
func (c *PaymentReceiveRuleController) List(ctx fiber.Ctx) error {
	resp, err := c.listPaymentReceiveRules.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Create appends a pricing rule.
// @Summary Internal create payment rule
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.SavePaymentReceiveRuleRequest true "payment rule"
// @Success 200 {object} responses.PaymentReceiveRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "policy_failed or create_failed"
// @Router /internal/config/payment-receive-policy/rules [post]
func (c *PaymentReceiveRuleController) Create(ctx fiber.Ctx) error {
	body := new(requests.SavePaymentReceiveRuleRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.createPaymentReceiveRule.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update replaces a pricing rule's fields.
// @Summary Internal update payment rule
// @Tags Internal
// @Accept json
// @Produce json
// @Param id path string true "rule id"
// @Param request body requests.SavePaymentReceiveRuleRequest true "payment rule"
// @Success 200 {object} responses.PaymentReceiveRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError "rule_not_found"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "policy_failed or update_failed"
// @Router /internal/config/payment-receive-policy/rules/{id} [put]
func (c *PaymentReceiveRuleController) Update(ctx fiber.Ctx) error {
	request := new(requests.UpdatePaymentReceiveRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.Bind().Body(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updatePaymentReceiveRule.Execute(ctx.Context(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes a pricing rule.
// @Summary Internal delete payment rule
// @Tags Internal
// @Param id path string true "rule id"
// @Success 204
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError "rule_not_found"
// @Failure 500 {object} errors.InternalError "delete_failed"
// @Router /internal/config/payment-receive-policy/rules/{id} [delete]
func (c *PaymentReceiveRuleController) Delete(ctx fiber.Ctx) error {
	request := new(requests.DeletePaymentReceiveRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := c.deletePaymentReceiveRule.Execute(ctx.Context(), request); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
