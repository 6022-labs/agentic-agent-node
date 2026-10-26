package schema

import (
	agent_identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	agent_knowledge_models "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	agent_payment_models "github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	agent_reasoning_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// Loader-only wrappers adding the agent_id -> agents.id FK without coupling models to agent_identity (import cycle).
type conversationWithCrossDomainFK struct {
	agent_knowledge_models.Conversation
	Agent agent_identity_models.Agent `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`
}

func (model conversationWithCrossDomainFK) TableName() string { return model.Conversation.TableName() }

type llmConnectionWithCrossDomainFK struct {
	agent_reasoning_models.LLMConnection
	Agent agent_identity_models.Agent `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`
}

func (model llmConnectionWithCrossDomainFK) TableName() string {
	return model.LLMConnection.TableName()
}

type agentGuidanceWithCrossDomainFK struct {
	agent_reasoning_models.AgentGuidance
	Agent agent_identity_models.Agent `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`
}

func (model agentGuidanceWithCrossDomainFK) TableName() string {
	return model.AgentGuidance.TableName()
}

type mcpConnectionWithCrossDomainFK struct {
	agent_reasoning_models.MCPConnection
	Agent agent_identity_models.Agent `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`
}

func (model mcpConnectionWithCrossDomainFK) TableName() string {
	return model.MCPConnection.TableName()
}

type selfToolWithCrossDomainFK struct {
	agent_reasoning_models.SelfTool
	Agent agent_identity_models.Agent `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`
}

func (model selfToolWithCrossDomainFK) TableName() string {
	return model.SelfTool.TableName()
}

type paymentReceivePolicyWithCrossDomainFK struct {
	agent_payment_models.PaymentReceivePolicy
	Agent agent_identity_models.Agent `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`
}

func (model paymentReceivePolicyWithCrossDomainFK) TableName() string {
	return model.PaymentReceivePolicy.TableName()
}
