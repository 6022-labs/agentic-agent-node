package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SignTypedData struct {
	typedDataSigner identity_interfaces.TypedDataSignerInterface
}

func NewSignTypedData(typedDataSigner identity_interfaces.TypedDataSignerInterface) *SignTypedData {
	return &SignTypedData{typedDataSigner: typedDataSigner}
}

func (u *SignTypedData) Execute(
	ctx context.Context,
	request *requests.SignTypedDataRequest,
) (*responses.TypedDataSignatureResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	signature, err := u.typedDataSigner.SignTypedData(ctx, request.TypedData)
	if err != nil {
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
