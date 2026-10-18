package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalPaymentRulesEndpoint = "/internal/config/payment-policy/rules"

type PaymentRuleController struct {
	listPaymentRules  *use_cases.ListPaymentRules
	createPaymentRule *use_cases.CreatePaymentRule
	updatePaymentRule *use_cases.UpdatePaymentRule
	deletePaymentRule *use_cases.DeletePaymentRule
}

func NewPaymentRuleController(
	listPaymentRules *use_cases.ListPaymentRules,
	createPaymentRule *use_cases.CreatePaymentRule,
	updatePaymentRule *use_cases.UpdatePaymentRule,
	deletePaymentRule *use_cases.DeletePaymentRule,
) *PaymentRuleController {
	return &PaymentRuleController{
		listPaymentRules:  listPaymentRules,
		createPaymentRule: createPaymentRule,
		updatePaymentRule: updatePaymentRule,
		deletePaymentRule: deletePaymentRule,
	}
}

func (c *PaymentRuleController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentRulesEndpoint, c.List)
	app.Post(InternalPaymentRulesEndpoint, c.Create)
	app.Put(InternalPaymentRulesEndpoint+"/:id", c.Update)
	app.Delete(InternalPaymentRulesEndpoint+"/:id", c.Delete)
}

// List returns the pricing rules in table order.
// @Summary Internal list payment rules
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.ListPaymentRulesResponse
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "policy_failed or list_failed"
// @Router /internal/config/payment-policy/rules [get]
func (c *PaymentRuleController) List(ctx fiber.Ctx) error {
	resp, err := c.listPaymentRules.Execute(ctx.Context())
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
// @Param request body requests.SavePaymentRuleRequest true "payment rule"
// @Success 200 {object} responses.PaymentRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "policy_failed or create_failed"
// @Router /internal/config/payment-policy/rules [post]
func (c *PaymentRuleController) Create(ctx fiber.Ctx) error {
	body := new(requests.SavePaymentRuleRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.createPaymentRule.Execute(ctx.Context(), body)
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
// @Param request body requests.SavePaymentRuleRequest true "payment rule"
// @Success 200 {object} responses.PaymentRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError "rule_not_found"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "policy_failed or update_failed"
// @Router /internal/config/payment-policy/rules/{id} [put]
func (c *PaymentRuleController) Update(ctx fiber.Ctx) error {
	request := new(requests.UpdatePaymentRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.Bind().Body(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updatePaymentRule.Execute(ctx.Context(), request)
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
// @Router /internal/config/payment-policy/rules/{id} [delete]
func (c *PaymentRuleController) Delete(ctx fiber.Ctx) error {
	request := new(requests.DeletePaymentRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := c.deletePaymentRule.Execute(ctx.Context(), request); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
