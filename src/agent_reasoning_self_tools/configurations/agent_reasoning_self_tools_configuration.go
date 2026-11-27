package configurations

import (
	identity_use_cases "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	knowledge_use_cases "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	payment_use_cases "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	reasoning_use_cases "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	os_services "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	common_use_cases "github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases"
	"go.uber.org/dig"
)

func AddAgentReasoningSelfToolsConfiguration(container *dig.Container) {
	if err := container.Provide(
		metrics.NewSelfToolMetrics,
		dig.As(new(metrics_interfaces.SelfToolMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewSelfToolRegistry,
		dig.As(new(reasoning_interfaces.SelfToolRegistryInterface)),
	); err != nil {
		panic(err)
	}

	provideExecutors(container)

	provideSelfTool(container, tools.NewGetAgentRegistrationTool)
	provideSelfTool(container, tools.NewListAgentCollectionsTool)
	provideSelfTool(container, tools.NewGetLLMConnectionTool)
	provideSelfTool(container, tools.NewListMCPConnectionsTool)
	provideSelfTool(container, tools.NewGetPaymentReceivePolicyTool)
	provideSelfTool(container, tools.NewListPaymentReceiveRulesTool)
	provideSelfTool(container, tools.NewGetPaymentFacilitatorsTool)
	provideSelfTool(container, tools.NewGetBlockchainConfigTool)
	provideSelfTool(container, tools.NewSetAgentGuidanceTool)
	provideSelfTool(container, tools.NewCreateMCPConnectionTool)
	provideSelfTool(container, tools.NewDeleteMCPConnectionTool)
	provideSelfTool(container, tools.NewCreatePaymentReceiveRuleTool)
	provideSelfTool(container, tools.NewUpdatePaymentReceiveRuleTool)
	provideSelfTool(container, tools.NewDeletePaymentReceiveRuleTool)
	provideSelfTool(container, tools.NewDeletePaymentReceivePolicyTool)
	provideSelfTool(container, tools.NewGetPaymentSpendPolicyTool)
	provideSelfTool(container, tools.NewUpdatePaymentSpendPolicyTool)
	provideSelfTool(container, tools.NewListPaymentSpendRulesTool)
	provideSelfTool(container, tools.NewCreatePaymentSpendRuleTool)
	provideSelfTool(container, tools.NewUpdatePaymentSpendRuleTool)
	provideSelfTool(container, tools.NewDeletePaymentSpendRuleTool)
	provideSelfTool(container, tools.NewDeletePaymentSpendPolicyTool)
	provideSelfTool(container, tools.NewCheckAgentMintTool)
	provideSelfTool(container, tools.NewCheckAgentEnsTool)
	provideSelfTool(container, tools.NewSignX402PaymentTool)
	provideSelfTool(container, tools.NewSignTypedDataTool)
	provideSelfTool(container, tools.NewRunCommandTool)
	provideSelfTool(container, tools.NewListProcessesTool)
	provideSelfTool(container, tools.NewReadProcessOutputTool)
	provideSelfTool(container, tools.NewKillProcessTool)
	provideSelfTool(container, tools.NewListConversationsTool)
	provideSelfTool(container, tools.NewCreateConversationTool)
	provideSelfTool(container, tools.NewUpdateConversationTool)
	provideSelfTool(container, tools.NewCallAgentTool)
}

func provideSelfTool(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("self_tools"),
		dig.As(new(reasoning_interfaces.SelfToolInterface)),
	); err != nil {
		panic(err)
	}
}

// provideExecutors binds each concrete use case or service to the narrow port its tool depends on.
func provideExecutors(container *dig.Container) {
	provideIdentityExecutors(container)
	provideKnowledgeExecutors(container)
	provideReasoningExecutors(container)
	providePaymentExecutors(container)
	provideCommonExecutors(container)
	provideOsExecutors(container)
}

func provideIdentityExecutors(container *dig.Container) {
	asGetAgentRegistration := func(u *identity_use_cases.GetAgentRegistration) tool_interfaces.GetAgentRegistrationExecutorInterface {
		return u
	}
	asListAgentCollections := func(u *identity_use_cases.ListAgentCollections) tool_interfaces.ListAgentCollectionsExecutorInterface {
		return u
	}
	asCheckAgentMint := func(u *identity_use_cases.CheckAgentMint) tool_interfaces.CheckAgentMintExecutorInterface {
		return u
	}
	asCheckAgentEns := func(u *identity_use_cases.CheckAgentEns) tool_interfaces.CheckAgentEnsExecutorInterface {
		return u
	}

	provideExecutor(container, asGetAgentRegistration)
	provideExecutor(container, asListAgentCollections)
	provideExecutor(container, asCheckAgentMint)
	provideExecutor(container, asCheckAgentEns)
}

func provideKnowledgeExecutors(container *dig.Container) {
	asListConversations := func(u *knowledge_use_cases.ListConversations) tool_interfaces.ListConversationsExecutorInterface {
		return u
	}
	asCreateConversation := func(u *knowledge_use_cases.CreateConversation) tool_interfaces.CreateConversationExecutorInterface {
		return u
	}
	asUpdateConversation := func(u *knowledge_use_cases.UpdateConversation) tool_interfaces.UpdateConversationExecutorInterface {
		return u
	}

	provideExecutor(container, asListConversations)
	provideExecutor(container, asCreateConversation)
	provideExecutor(container, asUpdateConversation)
}

func provideReasoningExecutors(container *dig.Container) {
	asGetLLMConnection := func(u *reasoning_use_cases.GetLLMConnection) tool_interfaces.GetLLMConnectionExecutorInterface {
		return u
	}
	asListMCPConnections := func(u *reasoning_use_cases.ListMCPConnections) tool_interfaces.ListMCPConnectionsExecutorInterface {
		return u
	}
	asSetAgentGuidance := func(u *reasoning_use_cases.SetAgentGuidance) tool_interfaces.SetAgentGuidanceExecutorInterface {
		return u
	}
	asCreateMCPConnection := func(u *reasoning_use_cases.CreateMCPConnection) tool_interfaces.CreateMCPConnectionExecutorInterface {
		return u
	}
	asDeleteMCPConnection := func(u *reasoning_use_cases.DeleteMCPConnection) tool_interfaces.DeleteMCPConnectionExecutorInterface {
		return u
	}
	asCallAgent := func(u *reasoning_use_cases.CallAgent) tool_interfaces.CallAgentExecutorInterface {
		return u
	}

	provideExecutor(container, asGetLLMConnection)
	provideExecutor(container, asListMCPConnections)
	provideExecutor(container, asSetAgentGuidance)
	provideExecutor(container, asCreateMCPConnection)
	provideExecutor(container, asDeleteMCPConnection)
	provideExecutor(container, asCallAgent)
}

func providePaymentExecutors(container *dig.Container) {
	asGetPaymentReceivePolicy := func(u *payment_use_cases.GetPaymentReceivePolicy) tool_interfaces.GetPaymentReceivePolicyExecutorInterface {
		return u
	}
	asListPaymentReceiveRules := func(u *payment_use_cases.ListPaymentReceiveRules) tool_interfaces.ListPaymentReceiveRulesExecutorInterface {
		return u
	}
	asGetPaymentFacilitators := func(u *payment_use_cases.GetPaymentFacilitators) tool_interfaces.GetPaymentFacilitatorsExecutorInterface {
		return u
	}
	asCreatePaymentReceiveRule := func(u *payment_use_cases.CreatePaymentReceiveRule) tool_interfaces.CreatePaymentReceiveRuleExecutorInterface {
		return u
	}
	asUpdatePaymentReceiveRule := func(u *payment_use_cases.UpdatePaymentReceiveRule) tool_interfaces.UpdatePaymentReceiveRuleExecutorInterface {
		return u
	}
	asDeletePaymentReceiveRule := func(u *payment_use_cases.DeletePaymentReceiveRule) tool_interfaces.DeletePaymentReceiveRuleExecutorInterface {
		return u
	}
	asDeletePaymentReceivePolicy := func(u *payment_use_cases.DeletePaymentReceivePolicy) tool_interfaces.DeletePaymentReceivePolicyExecutorInterface {
		return u
	}
	asSignX402Payment := func(u *payment_use_cases.SignX402Payment) tool_interfaces.SignX402PaymentExecutorInterface {
		return u
	}
	asSignTypedData := func(u *payment_use_cases.SignTypedData) tool_interfaces.SignTypedDataExecutorInterface {
		return u
	}
	asGetPaymentSpendPolicy := func(u *payment_use_cases.GetPaymentSpendPolicy) tool_interfaces.GetPaymentSpendPolicyExecutorInterface {
		return u
	}
	asUpdatePaymentSpendPolicy := func(u *payment_use_cases.UpdatePaymentSpendPolicy) tool_interfaces.UpdatePaymentSpendPolicyExecutorInterface {
		return u
	}
	asListPaymentSpendRules := func(u *payment_use_cases.ListPaymentSpendRules) tool_interfaces.ListPaymentSpendRulesExecutorInterface {
		return u
	}
	asCreatePaymentSpendRule := func(u *payment_use_cases.CreatePaymentSpendRule) tool_interfaces.CreatePaymentSpendRuleExecutorInterface {
		return u
	}
	asUpdatePaymentSpendRule := func(u *payment_use_cases.UpdatePaymentSpendRule) tool_interfaces.UpdatePaymentSpendRuleExecutorInterface {
		return u
	}
	asDeletePaymentSpendRule := func(u *payment_use_cases.DeletePaymentSpendRule) tool_interfaces.DeletePaymentSpendRuleExecutorInterface {
		return u
	}
	asDeletePaymentSpendPolicy := func(u *payment_use_cases.DeletePaymentSpendPolicy) tool_interfaces.DeletePaymentSpendPolicyExecutorInterface {
		return u
	}

	provideExecutor(container, asGetPaymentReceivePolicy)
	provideExecutor(container, asListPaymentReceiveRules)
	provideExecutor(container, asGetPaymentFacilitators)
	provideExecutor(container, asCreatePaymentReceiveRule)
	provideExecutor(container, asUpdatePaymentReceiveRule)
	provideExecutor(container, asDeletePaymentReceiveRule)
	provideExecutor(container, asDeletePaymentReceivePolicy)
	provideExecutor(container, asGetPaymentSpendPolicy)
	provideExecutor(container, asUpdatePaymentSpendPolicy)
	provideExecutor(container, asListPaymentSpendRules)
	provideExecutor(container, asCreatePaymentSpendRule)
	provideExecutor(container, asUpdatePaymentSpendRule)
	provideExecutor(container, asDeletePaymentSpendRule)
	provideExecutor(container, asDeletePaymentSpendPolicy)
	provideExecutor(container, asSignX402Payment)
	provideExecutor(container, asSignTypedData)
}

func provideCommonExecutors(container *dig.Container) {
	asGetBlockchainConfig := func(u *common_use_cases.GetBlockchainConfig) tool_interfaces.GetBlockchainConfigExecutorInterface {
		return u
	}

	provideExecutor(container, asGetBlockchainConfig)
}

// provideOsExecutors binds the one command executor service to the four narrow ports its tools depend on.
func provideOsExecutors(container *dig.Container) {
	asRunCommand := func(s *os_services.CommandExecutor) tool_interfaces.RunCommandExecutorInterface { return s }
	asListProcesses := func(s *os_services.CommandExecutor) tool_interfaces.ListProcessesExecutorInterface { return s }
	asReadProcessOutput := func(s *os_services.CommandExecutor) tool_interfaces.ReadProcessOutputExecutorInterface {
		return s
	}
	asKillProcess := func(s *os_services.CommandExecutor) tool_interfaces.KillProcessExecutorInterface { return s }

	provideExecutor(container, asRunCommand)
	provideExecutor(container, asListProcesses)
	provideExecutor(container, asReadProcessOutput)
	provideExecutor(container, asKillProcess)
}

func provideExecutor(container *dig.Container, binder any) {
	if err := container.Provide(binder); err != nil {
		panic(err)
	}
}
