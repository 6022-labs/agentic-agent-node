package payment

import (
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const PaymentReceiveRulesEndpoint = "/config/payment-receive-policy/rules"

type PaymentReceiveRuleController struct {
	listPaymentReceiveRules  *use_cases.ListPaymentReceiveRules
	createPaymentReceiveRule *use_cases.CreatePaymentReceiveRule
	updatePaymentReceiveRule *use_cases.UpdatePaymentReceiveRule
	deletePaymentReceiveRule *use_cases.DeletePaymentReceiveRule
	errorWriter              *mvc.ErrorWriter
}

func NewPaymentReceiveRuleController(
	listPaymentReceiveRules *use_cases.ListPaymentReceiveRules,
	createPaymentReceiveRule *use_cases.CreatePaymentReceiveRule,
	updatePaymentReceiveRule *use_cases.UpdatePaymentReceiveRule,
	deletePaymentReceiveRule *use_cases.DeletePaymentReceiveRule,
	errorWriter *mvc.ErrorWriter,
) *PaymentReceiveRuleController {
	return &PaymentReceiveRuleController{
		listPaymentReceiveRules:  listPaymentReceiveRules,
		createPaymentReceiveRule: createPaymentReceiveRule,
		updatePaymentReceiveRule: updatePaymentReceiveRule,
		deletePaymentReceiveRule: deletePaymentReceiveRule,
		errorWriter:              errorWriter,
	}
}

func (c *PaymentReceiveRuleController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *PaymentReceiveRuleController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(PaymentReceiveRulesEndpoint, c.List)
	app.Post(PaymentReceiveRulesEndpoint, c.Create)
	app.Put(PaymentReceiveRulesEndpoint+"/:id", c.Update)
	app.Delete(PaymentReceiveRulesEndpoint+"/:id", c.Delete)
}

// List returns the pricing rules in table order.
// @Summary List payment rules
// @Tags Config
// @Produce json
// @Success 200 {object} responses.ListPaymentReceiveRulesResponse
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-receive-policy/rules [get]
func (c *PaymentReceiveRuleController) List(ctx fiber.Ctx) error {
	resp, err := c.listPaymentReceiveRules.Execute(ctx.Context())
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}

// Create appends a pricing rule.
// @Summary Create payment rule
// @Tags Config
// @Accept json
// @Produce json
// @Param request body requests.SavePaymentReceiveRuleRequest true "payment rule"
// @Success 200 {object} responses.PaymentReceiveRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 409 {object} errors.ConflictError "agent_not_minted or asset_already_has_rule"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-receive-policy/rules [post]
func (c *PaymentReceiveRuleController) Create(ctx fiber.Ctx) error {
	body := new(requests.SavePaymentReceiveRuleRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_body", err.Error()))
	}
	resp, err := c.createPaymentReceiveRule.Execute(ctx.Context(), body)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}

// Update replaces a pricing rule's fields.
// @Summary Update payment rule
// @Tags Config
// @Accept json
// @Produce json
// @Param id path string true "rule id"
// @Param request body requests.SavePaymentReceiveRuleRequest true "payment rule"
// @Success 200 {object} responses.PaymentReceiveRuleResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 404 {object} errors.NotFoundError "rule_not_found"
// @Failure 409 {object} errors.ConflictError "agent_not_minted or asset_already_has_rule"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-receive-policy/rules/{id} [put]
func (c *PaymentReceiveRuleController) Update(ctx fiber.Ctx) error {
	request := new(requests.UpdatePaymentReceiveRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_params", err.Error()))
	}
	if err := ctx.Bind().Body(request); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_body", err.Error()))
	}
	resp, err := c.updatePaymentReceiveRule.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.JSON(resp)
}

// Delete removes a pricing rule.
// @Summary Delete payment rule
// @Tags Config
// @Param id path string true "rule id"
// @Success 204 ""
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 404 {object} errors.NotFoundError "rule_not_found"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /config/payment-receive-policy/rules/{id} [delete]
func (c *PaymentReceiveRuleController) Delete(ctx fiber.Ctx) error {
	request := new(requests.DeletePaymentReceiveRuleRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return c.errorWriter.Write(ctx, apperrors.NewBadRequestError("invalid_params", err.Error()))
	}
	if err := c.deletePaymentReceiveRule.Execute(ctx.Context(), request); err != nil {
		return c.errorWriter.Write(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
