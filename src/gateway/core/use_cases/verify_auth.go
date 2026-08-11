package use_cases

import (
	"context"
	"errors"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
	"github.com/ethereum/go-ethereum/common"
)

// Sentinel errors let tests assert on the domain reason; transport adapters
// only care about the outcome (Unauthorized, Unavailable, ...).
var (
	ErrChallengeNotFound    = errors.New("challenge not found or expired")
	ErrSignatureMismatch    = errors.New("signature does not match address")
	ErrAddressNotAuthorized = errors.New("address is not authorized")
)

type VerifyAuth struct {
	nonceStore         gateway_interfaces.NonceStoreInterface
	siweMessageBuilder gateway_interfaces.SiweMessageBuilderInterface
	signatureVerifier  gateway_interfaces.SignatureVerifierInterface
	authorizer         common_interfaces.AuthorizerInterface
	jwtIssuer          gateway_interfaces.JwtIssuerInterface
}

func NewVerifyAuth(
	nonceStore gateway_interfaces.NonceStoreInterface,
	siweMessageBuilder gateway_interfaces.SiweMessageBuilderInterface,
	signatureVerifier gateway_interfaces.SignatureVerifierInterface,
	authorizer common_interfaces.AuthorizerInterface,
	jwtIssuer gateway_interfaces.JwtIssuerInterface,
) *VerifyAuth {
	return &VerifyAuth{
		nonceStore:         nonceStore,
		siweMessageBuilder: siweMessageBuilder,
		signatureVerifier:  signatureVerifier,
		authorizer:         authorizer,
		jwtIssuer:          jwtIssuer,
	}
}

func (u *VerifyAuth) Execute(
	ctx context.Context,
	req *requests.VerifyAuthRequest,
) (*responses.VerifyAuthResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	claimed := common.HexToAddress(req.Address)
	entry := u.nonceStore.Consume(claimed)
	if entry == nil {
		return nil, apperrors.NewUnauthorizedError("challenge_not_found", ErrChallengeNotFound.Error())
	}

	message := u.siweMessageBuilder.Build(claimed, entry.ChainId, entry.Nonce, entry.IssuedAt, entry.ExpiresAt)
	recovered, err := u.signatureVerifier.Recover(message, req.Signature)
	if err != nil {
		return nil, apperrors.NewInternalError("verify_failed", err.Error())
	}
	if recovered != claimed {
		return nil, apperrors.NewUnauthorizedError("signature_mismatch", ErrSignatureMismatch.Error())
	}

	authorized, err := u.authorizer.IsAuthorized(ctx, claimed)
	if err != nil {
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return nil, apperrors.NewUnavailableError("identity_not_found", err.Error())
		}
		return nil, apperrors.NewInternalError("verify_failed", err.Error())
	}
	if !authorized {
		return nil, apperrors.NewUnauthorizedError("not_authorized", ErrAddressNotAuthorized.Error())
	}

	token, expiresAt, err := u.jwtIssuer.Issue(claimed)
	if err != nil {
		return nil, apperrors.NewInternalError("verify_failed", err.Error())
	}
	return &responses.VerifyAuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
