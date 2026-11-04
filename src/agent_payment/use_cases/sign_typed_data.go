package use_cases

import (
	"context"
	stderrors "errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"go.uber.org/zap"
)

// SignTypedData signs any EIP-712 payload, charging the ones that move tokens against the spend policy first.
type SignTypedData struct {
	typedDataSigner                 identity_interfaces.TypedDataSignerInterface
	typedDataSpendQuoter            payment_interfaces.TypedDataSpendQuoterInterface
	x402PaymentSpendPolicyEvaluator payment_interfaces.X402PaymentSpendPolicyEvaluatorInterface
	logger                          *zap.Logger
}

func NewSignTypedData(
	typedDataSigner identity_interfaces.TypedDataSignerInterface,
	typedDataSpendQuoter payment_interfaces.TypedDataSpendQuoterInterface,
	x402PaymentSpendPolicyEvaluator payment_interfaces.X402PaymentSpendPolicyEvaluatorInterface,
	logger *zap.Logger,
) *SignTypedData {
	return &SignTypedData{
		typedDataSigner:                 typedDataSigner,
		typedDataSpendQuoter:            typedDataSpendQuoter,
		x402PaymentSpendPolicyEvaluator: x402PaymentSpendPolicyEvaluator,
		logger:                          logger,
	}
}

func (u *SignTypedData) Execute(
	ctx context.Context,
	request *requests.SignTypedDataRequest,
) (*responses.TypedDataSignatureResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	quote, amount, err := u.typedDataSpendQuoter.Quote(request.TypedData)
	if err != nil {
		return nil, err
	}
	// Read before any spend is recorded, so this failure never leaves one to give back.
	address, err := u.typedDataSigner.WalletAddress(ctx)
	if err != nil {
		return nil, errors.NewInternalError("wallet_address_unavailable", err.Error())
	}
	var reservation *payment_interfaces.X402PaymentSpendReservation
	if quote != nil {
		reservation, err = u.x402PaymentSpendPolicyEvaluator.Authorize(ctx, quote, amount)
		if err != nil {
			var refusal *errors.UnauthorizedError
			if stderrors.As(err, &refusal) {
				u.logger.Info(
					"typed data payment refused",
					append(quote.AuditFields(), zap.String("code", refusal.Code))...)
			}
			return nil, err
		}
	}

	signature, err := u.typedDataSigner.SignTypedData(ctx, request.TypedData)
	if err != nil {
		// The caller's error matters more than a failed refund.
		if reservation != nil {
			if releaseErr := u.x402PaymentSpendPolicyEvaluator.Release(ctx, reservation); releaseErr != nil {
				u.logger.Error("failed to release an unsigned payload's spend", zap.Error(releaseErr))
			}
		}
		return nil, errors.NewInternalError("typed_data_signing_failed", err.Error())
	}

	if reservation != nil {
		u.logger.Info("typed data payment signed",
			append(quote.AuditFields(), zap.String("ruleId", reservation.RuleId.String()))...)
	}
	return &responses.TypedDataSignatureResponse{
		Signature: signature,
		Address:   address.Hex(),
	}, nil
}
