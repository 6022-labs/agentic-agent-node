package main

import (
	agent_identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	agent_knowledge_models "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	agent_payment_models "github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	agent_reasoning_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// Loader-only wrappers that add the cross-domain agent_id -> agents.id FK without coupling domain models to agent_identity (which would risk an import cycle).
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

type paymentPolicyWithCrossDomainFK struct {
	agent_payment_models.PaymentPolicy
	Agent agent_identity_models.Agent `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`
}

func (model paymentPolicyWithCrossDomainFK) TableName() string {
	return model.PaymentPolicy.TableName()
}
