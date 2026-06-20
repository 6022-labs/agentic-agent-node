package use_cases

import (
	"context"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/ethereum/go-ethereum/common"
)

type Authorize struct {
	authorizer           common_interfaces.AuthorizerInterface
	authorizationMetrics metrics_interfaces.AuthorizationMetricsInterface
}

func NewAuthorize(
	authorizer common_interfaces.AuthorizerInterface,
	authorizationMetrics metrics_interfaces.AuthorizationMetricsInterface,
) *Authorize {
	return &Authorize{authorizer: authorizer, authorizationMetrics: authorizationMetrics}
}

func (u *Authorize) Execute(ctx context.Context, req *requests.AuthorizeRequest) (*responses.AuthorizeResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	authorized, err := u.authorizer.IsAuthorized(ctx, common.HexToAddress(req.Address))
	if err != nil {
		u.authorizationMetrics.RecordAuthorizationFailed(ctx)
		return nil, errors.NewInternalError("authorize_failed", err.Error())
	}
	if authorized {
		u.authorizationMetrics.RecordAuthorizationAuthorized(ctx)
	} else {
		u.authorizationMetrics.RecordAuthorizationDenied(ctx)
	}
	return &responses.AuthorizeResponse{Authorized: authorized}, nil
}
