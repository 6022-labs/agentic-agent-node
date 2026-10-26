// Package schema lists the gorm models the migrations cover; the loader and the migration tests both read it.
package schema

import (
	agent_identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	agent_knowledge_models "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	agent_payment_models "github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
)

// Models lists every table-backed model, parents before children.
func Models() []any {
	return []any{
		&mcpConnectionWithCrossDomainFK{},
		&llmConnectionWithCrossDomainFK{},
		&agentGuidanceWithCrossDomainFK{},
		&selfToolWithCrossDomainFK{},

		&conversationWithCrossDomainFK{},
		&agent_knowledge_models.ConversationMessage{},

		&agent_identity_models.Agent{},
		&agent_identity_models.AgentRegistrationState{},
		&agent_identity_models.AgentAddress{},

		&paymentReceivePolicyWithCrossDomainFK{},
		&agent_payment_models.PaymentReceiveRule{},
		&agent_payment_models.PaymentReceiveRuleX402Config{},
		&agent_payment_models.PaymentReceiveRulePrice{},

		&paymentSpendPolicyWithCrossDomainFK{},
		&agent_payment_models.PaymentSpendRule{},
		&agent_payment_models.PaymentSpendRuleX402Config{},
	}
}
