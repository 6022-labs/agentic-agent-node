package use_cases

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
)

// wellKnownCardSigner is embedded by every well-known use case so they share one signing path.
type wellKnownCardSigner struct {
	agentCardSignatureClient identity_interfaces.AgentCardSignatureClientInterface
}

// signCard must run while the card's Signatures is still nil — omitempty is what excludes it from the signed bytes.
func (s *wellKnownCardSigner) signCard(ctx context.Context, card any) ([]responses.AgentCardSignature, error) {
	body, err := json.Marshal(card)
	if err != nil {
		return nil, err
	}

	signature, err := s.agentCardSignatureClient.Sign(ctx, body)
	if err != nil {
		return nil, err
	}
	return []responses.AgentCardSignature{{Protected: signature.Protected, Signature: signature.Signature}}, nil
}
