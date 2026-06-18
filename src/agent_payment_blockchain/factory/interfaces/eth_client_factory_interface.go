package interfaces

import "github.com/ethereum/go-ethereum/ethclient"

// EthClientFactoryInterface resolves a cached RPC client per chain so the
// validator can probe assets on any reachable chain.
type EthClientFactoryInterface interface {
	For(chainId string) (*ethclient.Client, error)
}
