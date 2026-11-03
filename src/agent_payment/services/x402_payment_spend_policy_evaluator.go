package services

import (
	"context"
	stderrors "errors"
	"math/big"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

// X402PaymentSpendPolicyEvaluator denies by default: an unlisted payee or a token without a rule is refused.
type X402PaymentSpendPolicyEvaluator struct {
	paymentSpendPolicyProvider           payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendRuleRepository           repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface
	paymentSpendRuleWindowRepository     repositories.PaymentSpendRuleWindowRepositoryInterface
	unitOfWorkFactory                    runtime_common_interfaces.UnitOfWorkFactoryInterface
}

func NewX402PaymentSpendPolicyEvaluator(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface,
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *X402PaymentSpendPolicyEvaluator {
	return &X402PaymentSpendPolicyEvaluator{
		paymentSpendPolicyProvider:           paymentSpendPolicyProvider,
		paymentSpendRuleRepository:           paymentSpendRuleRepository,
		paymentSpendRuleX402ConfigRepository: paymentSpendRuleX402ConfigRepository,
		paymentSpendRuleWindowRepository:     paymentSpendRuleWindowRepository,
		unitOfWorkFactory:                    unitOfWorkFactory,
	}
}

func (e *X402PaymentSpendPolicyEvaluator) Authorize(
	ctx context.Context,
	quote *payment_interfaces.X402PaymentQuote,
	amount *big.Int,
) (*payment_interfaces.X402PaymentSpendReservation, error) {
	policy, err := e.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		// An unminted agent has no policy anyone could have set, so it spends nothing.
		var notMinted *errors.ConflictError
		if stderrors.As(err, &notMinted) {
			return nil, errors.NewUnauthorizedError(
				"spend_agent_not_minted",
				"the agent is not minted, so it spends nothing",
			)
		}
		return nil, err
	}
	if !policy.AllowsPayee(quote.PayTo) {
		return nil, errors.NewUnauthorizedError("spend_payee_not_allowed", "the payee is not in the spend policy")
	}

	// Deciding and recording share one transaction, so concurrent signatures cannot overrun a window.
	unitOfWork, err := e.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	reservation, err := e.recordSpend(unitOfWork.Context(), policy, quote, amount)
	if err != nil {
		// A refusal wrote nothing, so it rolls back as is.
		var refusal *errors.UnauthorizedError
		if stderrors.As(err, &refusal) {
			return nil, err
		}
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	if err := unitOfWork.Commit(); err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	return reservation, nil
}

func (e *X402PaymentSpendPolicyEvaluator) Release(
	ctx context.Context,
	reservation *payment_interfaces.X402PaymentSpendReservation,
) error {
	// Runs after the caller's work failed, so the caller's cancellation must not skip the refund.
	ctx = context.WithoutCancel(ctx)
	if len(reservation.WindowStarts) == 0 {
		return nil
	}

	unitOfWork, err := e.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	windows, err := e.paymentSpendRuleWindowRepository.ListByRuleIdForUpdate(unitOfWork.Context(), reservation.RuleId)
	if err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}
	for i := range windows {
		startedAt, charged := reservation.WindowStarts[windows[i].Id]
		if !charged {
			continue
		}
		if err := windows[i].Release(reservation.Amount, startedAt); err != nil {
			return errors.NewInternalError("spend_policy_failed", err.Error())
		}
		if _, err := e.paymentSpendRuleWindowRepository.Update(unitOfWork.Context(), &windows[i]); err != nil {
			return errors.NewInternalError("spend_policy_failed", err.Error())
		}
	}
	if err := unitOfWork.Commit(); err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}
	return nil
}

// recordSpend records the amount on every window of the token's rule, all or none; a refusal is an UnauthorizedError.
func (e *X402PaymentSpendPolicyEvaluator) recordSpend(
	ctx context.Context,
	policy *models.PaymentSpendPolicy,
	quote *payment_interfaces.X402PaymentQuote,
	amount *big.Int,
) (*payment_interfaces.X402PaymentSpendReservation, error) {
	network := strings.TrimSpace(quote.Network)
	// Stored lowercased by the strategy, so the lookup must match that form.
	asset := strings.ToLower(strings.TrimSpace(quote.Asset))
	found, err := e.paymentSpendRuleX402ConfigRepository.GetByPolicyIdAndAsset(ctx, policy.Id, network, asset)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, e.noRule()
	}

	// Locked in the order rule updates take them (rule, token, windows), so an edit is read once committed.
	rule, err := e.paymentSpendRuleRepository.GetByIdForUpdate(ctx, found.PaymentSpendRuleId)
	if err != nil {
		return nil, err
	}
	if rule == nil || rule.Method != enums.PaymentMethodX402 {
		return nil, e.noRule()
	}
	config, err := e.paymentSpendRuleX402ConfigRepository.GetByRuleIdForUpdate(ctx, rule.Id)
	if err != nil {
		return nil, err
	}
	if config == nil || config.Network != found.Network || config.Asset != found.Asset {
		return nil, e.noRule()
	}
	if !config.AcceptsAssetTransferMethod(quote.AssetTransferMethod.String()) {
		return nil, errors.NewUnauthorizedError(
			"spend_transfer_method_not_allowed",
			"the token's spend rule does not allow this transfer method",
		)
	}
	covered, err := rule.CoversPayment(amount)
	if err != nil {
		return nil, err
	}
	if !covered {
		return nil, errors.NewUnauthorizedError("spend_cap_exceeded", "the amount exceeds the rule's cap per payment")
	}

	windows, err := e.paymentSpendRuleWindowRepository.ListByRuleIdForUpdate(ctx, rule.Id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	reservation := &payment_interfaces.X402PaymentSpendReservation{
		RuleId:       rule.Id,
		Amount:       new(big.Int).Set(amount),
		WindowStarts: make(map[uuid.UUID]time.Time, len(windows)),
	}
	for i := range windows {
		recorded, err := windows[i].RecordSpend(amount, now)
		if err != nil {
			return nil, err
		}
		if !recorded {
			return nil, errors.NewUnauthorizedError(
				"spend_window_exhausted",
				"the amount exceeds what is left of a budget this period",
			)
		}
		reservation.WindowStarts[windows[i].Id] = *windows[i].StartedAt
	}
	for i := range windows {
		if _, err := e.paymentSpendRuleWindowRepository.Update(ctx, &windows[i]); err != nil {
			return nil, err
		}
	}
	return reservation, nil
}

func (e *X402PaymentSpendPolicyEvaluator) noRule() error {
	return errors.NewUnauthorizedError("spend_token_has_no_rule", "no spend rule covers this token on this network")
}
