package configurations

import (
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	abi_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi/interfaces"
	blockchain_services "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/settings"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// ethClientName keeps this adapter's RPC client separate from other adapters'
// clients in the container.
const ethClientName = "agent-payment-blockchain"

func AddAgentPaymentBlockchainConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewRpcSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(newEthClient, dig.Name(ethClientName)); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewErc20Validator,
		dig.As(new(payment_interfaces.Erc20ValidatorInterface)),
	); err != nil {
		panic(err)
	}
}

func newEthClient(rpcSettings *settings.RpcSettings, logger *zap.Logger) abi_interfaces.Erc20CallerInterface {
	client, err := ethclient.Dial(rpcSettings.HttpUrl)
	if err != nil {
		logger.Fatal("failed to dial EVM RPC", zap.Error(err))
	}
	return client
}
