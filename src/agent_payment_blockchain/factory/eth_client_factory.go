package factory

import (
	"fmt"
	"sync"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClientFactory dials and caches one RPC client per chain for this adapter.
type EthClientFactory struct {
	chainsSettings *settings.ChainsSettings
	mu             sync.Mutex
	clients        map[string]*ethclient.Client
}

func NewEthClientFactory(chainsSettings *settings.ChainsSettings) *EthClientFactory {
	return &EthClientFactory{chainsSettings: chainsSettings, clients: map[string]*ethclient.Client{}}
}

// For returns the client for a chain, dialing it on first use.
func (f *EthClientFactory) For(chainId string) (*ethclient.Client, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if client, ok := f.clients[chainId]; ok {
		return client, nil
	}
	httpUrl, ok := f.chainsSettings.HttpUrlFor(chainId)
	if !ok {
		return nil, fmt.Errorf("no RPC configured for chain %s", chainId)
	}
	client, err := ethclient.Dial(httpUrl)
	if err != nil {
		return nil, fmt.Errorf("dial chain %s: %w", chainId, err)
	}
	f.clients[chainId] = client
	return client, nil
}
