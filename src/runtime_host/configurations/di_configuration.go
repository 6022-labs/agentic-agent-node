package configurations

import (
	agent_identity_configurations "github.com/6022-labs/agentic-agent-node/src/agent_identity/configurations"
	agent_identity_blockchain_configurations "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/configurations"
	agent_identity_ens_configurations "github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/configurations"
	agent_identity_http_ipfs_configurations "github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/configurations"
	agent_identity_mvc_configurations "github.com/6022-labs/agentic-agent-node/src/agent_identity_mvc/configurations"
	agent_knowledge_configurations "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/configurations"
	agent_knowledge_mvc_configurations "github.com/6022-labs/agentic-agent-node/src/agent_knowledge_mvc/configurations"
	agent_reasoning_configurations "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/configurations"
	agent_reasoning_http_inference_configurations "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/configurations"
	agent_reasoning_http_mcp_configurations "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/configurations"
	agent_reasoning_mvc_configurations "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_mvc/configurations"
	runtime_common_configurations "github.com/6022-labs/agentic-agent-node/src/runtime_common/configurations"
	runtime_common_mvc_configurations "github.com/6022-labs/agentic-agent-node/src/runtime_common_mvc/configurations"
	runtime_db_configurations "github.com/6022-labs/agentic-agent-node/src/runtime_db/configurations"
	"github.com/6022-labs/agentic-agent-node/src/runtime_host/settings"
	runtime_vault_configurations "github.com/6022-labs/agentic-agent-node/src/runtime_vault/configurations"

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

	runtime_vault_configurations.AddRuntimeVaultConfiguration(container)
	agent_identity_blockchain_configurations.AddAgentIdentityBlockchainConfiguration(container)
	agent_identity_configurations.AddAgentIdentityConfiguration(container)
	agent_identity_ens_configurations.AddAgentIdentityEnsConfiguration(container)
	agent_identity_http_ipfs_configurations.AddAgentIdentityHttpIpfsConfiguration(container)
	agent_reasoning_http_inference_configurations.AddAgentReasoningHttpInferenceConfiguration(container)
	agent_reasoning_http_mcp_configurations.AddAgentReasoningHttpMcpConfiguration(container)
	agent_reasoning_configurations.AddAgentReasoningConfiguration(container)
	runtime_db_configurations.AddRuntimeDbConfiguration(container)
	agent_knowledge_configurations.AddAgentKnowledgeConfiguration(container)
	agent_identity_mvc_configurations.AddAgentIdentityMvcConfiguration(container)
	runtime_common_configurations.AddRuntimeCommonConfiguration(container)
	runtime_common_mvc_configurations.AddRuntimeCommonMvcConfiguration(container)
	agent_knowledge_mvc_configurations.AddAgentKnowledgeMvcConfiguration(container)
	agent_reasoning_mvc_configurations.AddAgentReasoningMvcConfiguration(container)

	return container
}
