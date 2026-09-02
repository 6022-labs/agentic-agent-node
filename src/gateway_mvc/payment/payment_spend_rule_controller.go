package payment

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const PaymentSpendRulesEndpoint = "/config/payment-spend-policy/rules"

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

func (c *PaymentSpendRuleController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *PaymentSpendRuleController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(PaymentSpendRulesEndpoint, c.List)
	app.Post(PaymentSpendRulesEndpoint, c.Create)
	app.Put(PaymentSpendRulesEndpoint+"/:id", c.Update)
	app.Delete(PaymentSpendRulesEndpoint+"/:id", c.Delete)
}

// List returns the spend rules in table order.
// @Summary List payment spend rules
// @Tags Config
// @Produce json
// @Success 200 {object} responses.ListPaymentSpendRulesResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy/rules [get]
func (c *PaymentSpendRuleController) List(ctx fiber.Ctx) error {
	resp, err := c.listPaymentSpendRules.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Create appends a spend rule.
// @Summary Create payment spend rule
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SavePaymentSpendRuleRequest true "spend rule"
// @Success 200 {object} responses.PaymentSpendRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy/rules [post]
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
// @Summary Update payment spend rule
// @Tags Config
// @Accept json
// @Produce json
// @Param id path string true "rule id"
// @Param request body requests.SavePaymentSpendRuleRequest true "spend rule"
// @Success 200 {object} responses.PaymentSpendRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy/rules/{id} [put]
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
// @Summary Delete payment spend rule
// @Tags Config
// @Param id path string true "rule id"
// @Success 204 ""
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} errors.NotFoundError
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-spend-policy/rules/{id} [delete]
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
