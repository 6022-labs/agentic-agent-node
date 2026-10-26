package use_cases

import (
	"context"

	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// AuthorizePaymentReceiveAccess is method-neutral; x402 is the only method with an evaluator today.
type AuthorizePaymentReceiveAccess struct {
	x402PaymentReceivePolicyEvaluator payment_interfaces.X402PaymentReceivePolicyEvaluatorInterface
}

func NewAuthorizePaymentReceiveAccess(
	x402PaymentReceivePolicyEvaluator payment_interfaces.X402PaymentReceivePolicyEvaluatorInterface,
) *AuthorizePaymentReceiveAccess {
	return &AuthorizePaymentReceiveAccess{x402PaymentReceivePolicyEvaluator: x402PaymentReceivePolicyEvaluator}
}

func (u *AuthorizePaymentReceiveAccess) Execute(
	ctx context.Context,
	request *requests.AuthorizePaymentReceiveAccessRequest,
) (*responses.AuthorizePaymentReceiveAccessResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	decision, err := u.x402PaymentReceivePolicyEvaluator.Evaluate(
		ctx,
		&payment_interfaces.PaymentReceiveAccessCredentials{
			Address:          request.Address,
			PaymentSignature: request.PaymentSignature,
		},
	)
	if err != nil {
		return nil, err
	}
	return responses.ToAuthorizePaymentReceiveAccessResponse(decision), nil
}
