package factory

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// EthClientFactory dials and caches one RPC client per chain for this adapter.
type EthClientFactory struct {
	evmChainsSettings *settings.EvmChainsSettings
	rpcMetrics        metrics_interfaces.BlockchainRpcMetricsInterface
	mu                sync.Mutex
	clients           map[string]*ethclient.Client
}

func NewEthClientFactory(
	evmChainsSettings *settings.EvmChainsSettings,
	rpcMetrics metrics_interfaces.BlockchainRpcMetricsInterface,
) *EthClientFactory {
	return &EthClientFactory{
		evmChainsSettings: evmChainsSettings,
		rpcMetrics:        rpcMetrics,
		clients:           map[string]*ethclient.Client{},
	}
}

// For returns the client for a chain, dialing it on first use.
func (f *EthClientFactory) For(chainId string) (*ethclient.Client, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if client, ok := f.clients[chainId]; ok {
		return client, nil
	}
	httpUrl, ok := f.evmChainsSettings.HttpUrlFor(chainId)
	if !ok {
		return nil, fmt.Errorf("no RPC configured for chain %s", chainId)
	}
	httpClient := &http.Client{
		Transport: metrics.NewRpcRoundTripper(http.DefaultTransport, f.rpcMetrics, chainId),
	}
	rpcClient, err := rpc.DialOptions(context.Background(), httpUrl, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("dial chain %s: %w", chainId, err)
	}
	client := ethclient.NewClient(rpcClient)
	f.clients[chainId] = client
	return client, nil
}
