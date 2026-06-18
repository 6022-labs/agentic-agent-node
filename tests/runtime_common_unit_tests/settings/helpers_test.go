package settings_test

import "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"

// reachableChains is a ChainsSettings covering the test facilitator networks, so
// the facilitator constructor's "needs a matching RPC" guard passes.
func reachableChains() *settings.ChainsSettings {
	chains := settings.ChainsSettings{
		"80002": {ChainId: "80002", HttpUrl: "http://amoy"},
		"84532": {ChainId: "84532", HttpUrl: "http://base-sepolia"},
	}
	return &chains
}
