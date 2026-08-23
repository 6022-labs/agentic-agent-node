package payment

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const PaymentRulesEndpoint = "/config/payment-policy/rules"

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

func (c *PaymentRuleController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *PaymentRuleController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(PaymentRulesEndpoint, c.List)
	app.Post(PaymentRulesEndpoint, c.Create)
	app.Put(PaymentRulesEndpoint+"/:id", c.Update)
	app.Delete(PaymentRulesEndpoint+"/:id", c.Delete)
}

// List returns the pricing rules in table order.
// @Summary List payment rules
// @Tags Config
// @Produce json
// @Success 200 {object} responses.ListPaymentRulesResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-policy/rules [get]
func (c *PaymentRuleController) List(ctx fiber.Ctx) error {
	resp, err := c.listPaymentRules.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Create appends a pricing rule.
// @Summary Create payment rule
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SavePaymentRuleRequest true "payment rule"
// @Success 200 {object} models.ExternalPaymentRule
// @Failure 400 {object} errors.ValidationError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-policy/rules [post]
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
// @Summary Update payment rule
// @Tags Config
// @Accept json
// @Produce json
// @Param id path string true "rule id"
// @Param request body requests.SavePaymentRuleRequest true "payment rule"
// @Success 200 {object} models.ExternalPaymentRule
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-policy/rules/{id} [put]
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
// @Summary Delete payment rule
// @Tags Config
// @Param id path string true "rule id"
// @Success 204 ""
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-policy/rules/{id} [delete]
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
