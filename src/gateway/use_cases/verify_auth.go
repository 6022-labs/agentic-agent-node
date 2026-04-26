package use_cases

import (
	"context"
	"errors"

	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrNonceNotFound       = errors.New("nonce not found or expired")
	ErrSignatureMismatch   = errors.New("signature does not match address")
	ErrAddressNotAuthorized = errors.New("address is not authorized")
)

type VerifyAuth struct {
	nonceStore        gateway_interfaces.NonceStoreInterface
	signatureVerifier gateway_interfaces.SignatureVerifierInterface
	authorizer        common_interfaces.AuthorizerInterface
	jwtIssuer         gateway_interfaces.JwtIssuerInterface
}

func NewVerifyAuth(
	nonceStore gateway_interfaces.NonceStoreInterface,
	signatureVerifier gateway_interfaces.SignatureVerifierInterface,
	authorizer common_interfaces.AuthorizerInterface,
	jwtIssuer gateway_interfaces.JwtIssuerInterface,
) *VerifyAuth {
	return &VerifyAuth{
		nonceStore:        nonceStore,
		signatureVerifier: signatureVerifier,
		authorizer:        authorizer,
		jwtIssuer:         jwtIssuer,
	}
}

func (u *VerifyAuth) Execute(ctx context.Context, req *requests.VerifyAuthRequest) (*responses.VerifyAuthResponse, error) {
	if !u.nonceStore.Consume(req.Nonce) {
		return nil, ErrNonceNotFound
	}

	recovered, err := u.signatureVerifier.Recover(req.Nonce, req.Signature)
	if err != nil {
		return nil, err
	}
	claimed := common.HexToAddress(req.Address)
	if recovered != claimed {
		return nil, ErrSignatureMismatch
	}

	authorized, err := u.authorizer.IsAuthorized(ctx, claimed)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrAddressNotAuthorized
	}

	token, expiresAt, err := u.jwtIssuer.Issue(claimed)
	if err != nil {
		return nil, err
	}
	return &responses.VerifyAuthResponse{Token: token, ExpiresAt: expiresAt}, nil
}
