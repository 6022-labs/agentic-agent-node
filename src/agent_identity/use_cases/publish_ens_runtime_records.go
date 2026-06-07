package use_cases

import (
	"context"
	"errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
)

type PublishEnsRuntimeRecords struct {
	ensPublisher identity_interfaces.EnsPublisherInterface
}

func NewPublishEnsRuntimeRecords(ensPublisher identity_interfaces.EnsPublisherInterface) *PublishEnsRuntimeRecords {
	return &PublishEnsRuntimeRecords{ensPublisher: ensPublisher}
}

func (u *PublishEnsRuntimeRecords) Execute(ctx context.Context) error {
	if err := u.ensPublisher.PublishRuntimeRecords(ctx); err != nil {
		if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
			return apperrors.NewConflictError("identity_not_found", "agent NFT must be minted before ENS records can be published")
		}
		return apperrors.NewInternalError("ens_publish_failed", err.Error())
	}
	return nil
}
