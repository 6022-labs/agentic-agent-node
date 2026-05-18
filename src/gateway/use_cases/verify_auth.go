package use_cases

import (
	"context"
	"errors"

	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	"github.com/ethereum/go-ethereum/common"
)

// Sentinel errors are kept for tests that assert on the domain reason; the
// outcome (Unauthorized, Unavailable, ...) is what the transport adapter cares
// about, but having a named error helps unit tests be explicit.
var (
	ErrNonceNotFound        = errors.New("nonce not found or expired")
	ErrSignatureMismatch    = errors.New("signature does not match address")
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

func (u *VerifyAuth) Execute(ctx context.Context, req *requests.VerifyAuthRequest) *usecase_result.Result {
	if !u.nonceStore.Consume(req.Nonce) {
		return usecase_result.Unauthorized("nonce_not_found", ErrNonceNotFound.Error())
	}

	recovered, err := u.signatureVerifier.Recover(req.Nonce, req.Signature)
	if err != nil {
		return usecase_result.InternalError("verify_failed", err.Error())
	}
	claimed := common.HexToAddress(req.Address)
	if recovered != claimed {
		return usecase_result.Unauthorized("signature_mismatch", ErrSignatureMismatch.Error())
	}

	authorized, err := u.authorizer.IsAuthorized(ctx, claimed)
	if err != nil {
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return usecase_result.Unavailable("identity_not_found", err.Error())
		}
		return usecase_result.InternalError("verify_failed", err.Error())
	}
	if !authorized {
		return usecase_result.Unauthorized("not_authorized", ErrAddressNotAuthorized.Error())
	}

	token, expiresAt, err := u.jwtIssuer.Issue(claimed)
	if err != nil {
		return usecase_result.InternalError("verify_failed", err.Error())
	}
	return usecase_result.Ok(&responses.VerifyAuthResponse{Token: token, ExpiresAt: expiresAt})
}
