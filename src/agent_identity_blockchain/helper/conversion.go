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

// ToNullableUint256 maps an optional uint64 to the AgentCollectionV1 abi type.
func ToNullableUint256(v *uint64) abi.NullableUint256 {
	if v == nil {
		return abi.NullableUint256{IsSet: false, Value: big.NewInt(0)}
	}
	return abi.NullableUint256{IsSet: true, Value: new(big.Int).SetUint64(*v)}
}
