package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const InternalPaymentSpendRulesEndpoint = "/internal/config/payment-spend-policy/rules"

type PaymentSpendRuleController struct {
	listPaymentSpendRules  *use_cases.ListPaymentSpendRules
	createPaymentSpendRule *use_cases.CreatePaymentSpendRule
	updatePaymentSpendRule *use_cases.UpdatePaymentSpendRule
	deletePaymentSpendRule *use_cases.DeletePaymentSpendRule
}

func NewPaymentSpendRuleController(
	listPaymentSpendRules *use_cases.ListPaymentSpendRules,
	createPaymentSpendRule *use_cases.CreatePaymentSpendRule,
	updatePaymentSpendRule *use_cases.UpdatePaymentSpendRule,
	deletePaymentSpendRule *use_cases.DeletePaymentSpendRule,
) *PaymentSpendRuleController {
	return &PaymentSpendRuleController{
		listPaymentSpendRules:  listPaymentSpendRules,
		createPaymentSpendRule: createPaymentSpendRule,
		updatePaymentSpendRule: updatePaymentSpendRule,
		deletePaymentSpendRule: deletePaymentSpendRule,
	}
}

func (c *PaymentSpendRuleController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalPaymentSpendRulesEndpoint, c.List)
	app.Post(InternalPaymentSpendRulesEndpoint, c.Create)
	app.Put(InternalPaymentSpendRulesEndpoint+"/:id", c.Update)
	app.Delete(InternalPaymentSpendRulesEndpoint+"/:id", c.Delete)
}

// List returns the spend rules in table order.
// @Summary Internal list payment spend rules
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.ListPaymentSpendRulesResponse
// @Failure 500 {object} fiber.Map "list_payment_spend_rules_failed"
// @Router /internal/config/payment-spend-policy/rules [get]
func (c *PaymentSpendRuleController) List(ctx fiber.Ctx) error {
	resp, err := c.listPaymentSpendRules.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Create appends a spend rule.
// @Summary Internal create payment spend rule
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.SavePaymentSpendRuleRequest true "payment spend rule"
// @Success 200 {object} responses.PaymentSpendRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 500 {object} fiber.Map "create_payment_spend_rule_failed"
// @Router /internal/config/payment-spend-policy/rules [post]
func (c *PaymentSpendRuleController) Create(ctx fiber.Ctx) error {
	body := new(requests.SavePaymentSpendRuleRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.createPaymentSpendRule.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update replaces a spend rule's fields.
// @Summary Internal update payment spend rule
// @Tags Internal
// @Accept json
// @Produce json
// @Param id path string true "rule id"
// @Param request body requests.SavePaymentSpendRuleRequest true "payment spend rule"
// @Success 200 {object} responses.PaymentSpendRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "update_payment_spend_rule_failed"
// @Router /internal/config/payment-spend-policy/rules/{id} [put]
func (c *PaymentSpendRuleController) Update(ctx fiber.Ctx) error {
	request := new(requests.UpdatePaymentSpendRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.Bind().Body(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updatePaymentSpendRule.Execute(ctx.Context(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes a spend rule.
// @Summary Internal delete payment spend rule
// @Tags Internal
// @Param id path string true "rule id"
// @Success 204
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "delete_payment_spend_rule_failed"
// @Router /internal/config/payment-spend-policy/rules/{id} [delete]
func (c *PaymentSpendRuleController) Delete(ctx fiber.Ctx) error {
	request := new(requests.DeletePaymentSpendRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := c.deletePaymentSpendRule.Execute(ctx.Context(), request); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
