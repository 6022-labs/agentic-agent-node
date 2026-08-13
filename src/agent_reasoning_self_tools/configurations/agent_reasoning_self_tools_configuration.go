package configurations

import (
	identity_use_cases "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
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
	provideSelfTool(container, tools.NewGetPaymentPolicyTool)
	provideSelfTool(container, tools.NewListPaymentRulesTool)
	provideSelfTool(container, tools.NewGetPaymentFacilitatorsTool)
	provideSelfTool(container, tools.NewGetBlockchainConfigTool)
	provideSelfTool(container, tools.NewSetAgentGuidanceTool)
	provideSelfTool(container, tools.NewCreateMCPConnectionTool)
	provideSelfTool(container, tools.NewDeleteMCPConnectionTool)
	provideSelfTool(container, tools.NewCreatePaymentRuleTool)
	provideSelfTool(container, tools.NewUpdatePaymentRuleTool)
	provideSelfTool(container, tools.NewDeletePaymentRuleTool)
	provideSelfTool(container, tools.NewDeletePaymentPolicyTool)
	provideSelfTool(container, tools.NewCheckAgentMintTool)
	provideSelfTool(container, tools.NewCheckAgentEnsTool)
	provideSelfTool(container, tools.NewRunCommandTool)
	provideSelfTool(container, tools.NewListProcessesTool)
	provideSelfTool(container, tools.NewReadProcessOutputTool)
	provideSelfTool(container, tools.NewKillProcessTool)
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

	provideExecutor(container, asGetLLMConnection)
	provideExecutor(container, asListMCPConnections)
	provideExecutor(container, asSetAgentGuidance)
	provideExecutor(container, asCreateMCPConnection)
	provideExecutor(container, asDeleteMCPConnection)
}

func providePaymentExecutors(container *dig.Container) {
	asGetPaymentPolicy := func(u *payment_use_cases.GetPaymentPolicy) tool_interfaces.GetPaymentPolicyExecutorInterface {
		return u
	}
	asListPaymentRules := func(u *payment_use_cases.ListPaymentRules) tool_interfaces.ListPaymentRulesExecutorInterface {
		return u
	}
	asGetPaymentFacilitators := func(u *payment_use_cases.GetPaymentFacilitators) tool_interfaces.GetPaymentFacilitatorsExecutorInterface {
		return u
	}
	asCreatePaymentRule := func(u *payment_use_cases.CreatePaymentRule) tool_interfaces.CreatePaymentRuleExecutorInterface {
		return u
	}
	asUpdatePaymentRule := func(u *payment_use_cases.UpdatePaymentRule) tool_interfaces.UpdatePaymentRuleExecutorInterface {
		return u
	}
	asDeletePaymentRule := func(u *payment_use_cases.DeletePaymentRule) tool_interfaces.DeletePaymentRuleExecutorInterface {
		return u
	}
	asDeletePaymentPolicy := func(u *payment_use_cases.DeletePaymentPolicy) tool_interfaces.DeletePaymentPolicyExecutorInterface {
		return u
	}

	provideExecutor(container, asGetPaymentPolicy)
	provideExecutor(container, asListPaymentRules)
	provideExecutor(container, asGetPaymentFacilitators)
	provideExecutor(container, asCreatePaymentRule)
	provideExecutor(container, asUpdatePaymentRule)
	provideExecutor(container, asDeletePaymentRule)
	provideExecutor(container, asDeletePaymentPolicy)
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
