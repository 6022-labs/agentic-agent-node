package use_cases

import (
	"context"
	stderrors "errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SignX402Payment struct {
	x402PaymentSigner interfaces.X402PaymentSignerInterface
}

func NewSignX402Payment(x402PaymentSigner interfaces.X402PaymentSignerInterface) *SignX402Payment {
	return &SignX402Payment{x402PaymentSigner: x402PaymentSigner}
}

func (u *SignX402Payment) Execute(
	ctx context.Context,
	request *requests.SignX402PaymentRequest,
) (*responses.X402PaymentSignatureResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	signed, err := u.x402PaymentSigner.Sign(ctx, &interfaces.X402PaymentQuote{
		Network:             request.Network,
		Asset:               request.Asset,
		PayTo:               request.PayTo,
		Amount:              request.Amount,
		AssetTransferMethod: request.ParsedAssetTransferMethod(),
		Eip712Name:          request.Eip712Name,
		Eip712Version:       request.Eip712Version,
		MaxTimeoutSeconds:   request.MaxTimeoutSeconds,
	})
	if err != nil {
		return nil, u.mapSignError(err)
	}

	response := &responses.X402PaymentSignatureResponse{}
	response.FromModel(signed)
	return response, nil
}

// mapSignError keeps a typed domain error typed; only an untyped signer failure reads as an upstream problem.
func (u *SignX402Payment) mapSignError(err error) error {
	var validationError *errors.ValidationError
	var unauthorizedError *errors.UnauthorizedError
	var internalError *errors.InternalError
	var conflictError *errors.ConflictError
	var notFoundError *errors.NotFoundError
	if stderrors.As(err, &validationError) ||
		stderrors.As(err, &unauthorizedError) ||
		stderrors.As(err, &internalError) ||
		stderrors.As(err, &conflictError) ||
		stderrors.As(err, &notFoundError) {
		return err
	}
	return errors.NewUnavailableError("x402_payment_signing_failed", err.Error())
}
