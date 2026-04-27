package configurations

import (
	agent_identity_configurations "github.com/6022-labs/agentic-agent-node/src/agent_identity/configurations"
	agent_identity_blockchain_configurations "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/configurations"
	agent_knowledge_configurations "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/configurations"
	agent_knowledge_db_configurations "github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/configurations"
	agent_reasoning_configurations "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/configurations"
	common_vault_configurations "github.com/6022-labs/agentic-agent-node/src/common_vault/configurations"
	"github.com/6022-labs/agentic-agent-node/src/runtime_host/settings"
	runtime_mvc_configurations "github.com/6022-labs/agentic-agent-node/src/runtime_mvc/configurations"

	"github.com/knadh/koanf/v2"
	"go.uber.org/dig"
)

func ConfigureDI(config *koanf.Koanf) *dig.Container {
	container := dig.New()

	if err := container.Provide(settings.NewHostSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(func() *koanf.Koanf {
		return config
	}); err != nil {
		panic(err)
	}

	common_vault_configurations.AddCommonVaultConfiguration(container)
	agent_identity_blockchain_configurations.AddAgentIdentityBlockchainConfiguration(container)
	agent_identity_configurations.AddAgentIdentityConfiguration(container)
	agent_reasoning_configurations.AddAgentReasoningConfiguration(container)
	agent_knowledge_db_configurations.AddAgentKnowledgeDbConfiguration(container)
	agent_knowledge_configurations.AddAgentKnowledgeConfiguration(container)
	runtime_mvc_configurations.AddRuntimeMvcConfiguration(container)

	return container
}
