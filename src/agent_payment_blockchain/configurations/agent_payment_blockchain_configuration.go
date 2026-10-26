package configurations

import (
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/factory"
	factory_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/factory/interfaces"
	blockchain_services "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/services"
	"go.uber.org/dig"
)

func AddAgentPaymentBlockchainConfiguration(container *dig.Container) {
	if err := container.Provide(
		factory.NewEthClientFactory,
		dig.As(new(factory_interfaces.EthClientFactoryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		factory.NewTransactorFactory,
		dig.As(new(factory_interfaces.TransactorFactoryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewPermit2ApprovalEnsurer,
		dig.As(new(payment_interfaces.Permit2ApprovalEnsurerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewErc20Validator,
		dig.As(new(payment_interfaces.Erc20ValidatorInterface)),
	); err != nil {
		panic(err)
	}
}
