package use_cases

import (
	"context"

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
	var reservation *payment_interfaces.X402PaymentSpendReservation
	if quote != nil {
		reservation, err = u.x402PaymentSpendPolicyEvaluator.Authorize(ctx, quote, amount)
		if err != nil {
			return nil, err
		}
		if reservation == nil {
			return nil, errors.NewUnauthorizedError(
				"spend_refused",
				"the agent's spend policy does not cover this token payload",
			)
		}
	}

	signature, err := u.typedDataSigner.SignTypedData(ctx, request.TypedData)
	if err != nil {
		u.release(ctx, reservation)
		return nil, errors.NewInternalError("typed_data_signing_failed", err.Error())
	}

	address, err := u.typedDataSigner.WalletAddress(ctx)
	if err != nil {
		return nil, errors.NewInternalError("wallet_address_unavailable", err.Error())
	}

	return &responses.TypedDataSignatureResponse{
		Signature: signature,
		Address:   address.Hex(),
	}, nil
}

// release gives the spend back when no signature came out; the caller's error matters more than a failed refund.
func (u *SignTypedData) release(
	ctx context.Context,
	reservation *payment_interfaces.X402PaymentSpendReservation,
) {
	if reservation == nil {
		return
	}
	if err := u.x402PaymentSpendPolicyEvaluator.Release(context.WithoutCancel(ctx), reservation); err != nil {
		u.logger.Error("failed to release an unsigned payload's spend", zap.Error(err))
	}
}
