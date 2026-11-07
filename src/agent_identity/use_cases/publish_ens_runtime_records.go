package use_cases

import (
	"context"
	"errors"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/metrics/interfaces"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type PublishEnsRuntimeRecords struct {
	ensPublisher          identity_interfaces.EnsPublisherInterface
	ensPublicationMetrics metrics_interfaces.EnsPublicationMetricsInterface
}

func NewPublishEnsRuntimeRecords(
	ensPublisher identity_interfaces.EnsPublisherInterface,
	ensPublicationMetrics metrics_interfaces.EnsPublicationMetricsInterface,
) *PublishEnsRuntimeRecords {
	return &PublishEnsRuntimeRecords{ensPublisher: ensPublisher, ensPublicationMetrics: ensPublicationMetrics}
}

func (u *PublishEnsRuntimeRecords) Execute(ctx context.Context) error {
	if err := u.ensPublisher.PublishRuntimeRecords(ctx); err != nil {
		if errors.Is(err, apperrors.ErrIdentityNotFound) {
			u.ensPublicationMetrics.RecordEnsPublishedNotMinted(ctx)
			return apperrors.NewConflictError(
				"identity_not_found",
				"agent NFT must be minted before ENS records can be published",
			)
		}
		u.ensPublicationMetrics.RecordEnsPublishedFailed(ctx)
		return apperrors.NewInternalError("ens_publish_failed", err)
	}
	u.ensPublicationMetrics.RecordEnsPublishedSucceeded(ctx)
	return nil
}
