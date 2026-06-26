package agent_payment_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
// @Failure 500 {object} fiber.Map "list_payment_rules_failed"
// @Router /internal/config/payment-policy/rules [get]
func (c *PaymentRuleController) List(ctx *fiber.Ctx) error {
	resp, err := c.listPaymentRules.Execute(ctx.UserContext())
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
// @Failure 500 {object} fiber.Map "create_payment_rule_failed"
// @Router /internal/config/payment-policy/rules [post]
func (c *PaymentRuleController) Create(ctx *fiber.Ctx) error {
	body := new(requests.SavePaymentRuleRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.createPaymentRule.Execute(ctx.UserContext(), body)
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
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "update_payment_rule_failed"
// @Router /internal/config/payment-policy/rules/{id} [put]
func (c *PaymentRuleController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_id", "message": err.Error()})
	}
	body := new(requests.SavePaymentRuleRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	resp, err := c.updatePaymentRule.Execute(ctx.UserContext(), id, body)
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
// @Failure 404 {object} errors.NotFoundError
// @Failure 500 {object} fiber.Map "delete_payment_rule_failed"
// @Router /internal/config/payment-policy/rules/{id} [delete]
func (c *PaymentRuleController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_id", "message": err.Error()})
	}
	if err := c.deletePaymentRule.Execute(ctx.UserContext(), id); err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
