package helper

import (
	"math/big"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
)

// ToKeyValues maps domain mint key/values to the AgentCollectionV1 abi type.
func ToKeyValues(in []identity_interfaces.MintKeyValue) []abi.KeyValue {
	out := make([]abi.KeyValue, 0, len(in))
	for _, kv := range in {
		out = append(out, abi.KeyValue{Key: kv.Key, Value: kv.Value})
	}
	return out
}

// ToAgentAddresses maps domain mint addresses to the AgentCollectionV1 abi type.
func ToAgentAddresses(in []identity_interfaces.MintAddress) []abi.AgentAddress {
	out := make([]abi.AgentAddress, 0, len(in))
	for _, agentAddress := range in {
		out = append(out, abi.AgentAddress{
			AddressType: agentAddress.AddressType.String(),
			Value:       agentAddress.Value,
		})
	}
	return out
}

// KeyValuesToMap collapses an AgentCollectionV1 key/value list into a map.
func KeyValuesToMap(in []abi.KeyValue) map[string]string {
	out := make(map[string]string, len(in))
	for _, kv := range in {
		out[kv.Key] = kv.Value
	}
	return out
}

// ToNullableUint256 maps an optional uint64 to the AgentCollectionV1 abi type.
func ToNullableUint256(v *uint64) abi.NullableUint256 {
	if v == nil {
		return abi.NullableUint256{IsSet: false, Value: big.NewInt(0)}
	}
	return abi.NullableUint256{IsSet: true, Value: new(big.Int).SetUint64(*v)}
}

// ToUint64 maps a contract uint256 to uint64, collapsing nil/overflow to 0.
func ToUint64(v *big.Int) uint64 {
	if v == nil || !v.IsUint64() {
		return 0
	}
	return v.Uint64()
}
