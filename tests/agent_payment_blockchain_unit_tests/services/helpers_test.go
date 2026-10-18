package services_test

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contract_abi "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

// erc20Method is one canned eth_call answer; a method absent from the script reverts.
type erc20Method struct {
	data  []byte
	empty bool
}

// packErc20Output encodes a method return the way the contract would.
func packErc20Output(t *testing.T, method string, values ...any) []byte {
	t.Helper()

	parsed, err := abi.JSON(strings.NewReader(contract_abi.Erc20MetaData.ABI))
	assert.NoError(t, err)
	packed, err := parsed.Methods[method].Outputs.Pack(values...)
	assert.NoError(t, err)
	return packed
}

// newErc20RpcServer answers eth_call from a per-method script and eth_getCode from hasCode.
func newErc20RpcServer(t *testing.T, methods map[string]erc20Method, hasCode bool) *httptest.Server {
	t.Helper()

	parsed, err := abi.JSON(strings.NewReader(contract_abi.Erc20MetaData.ABI))
	assert.NoError(t, err)
	bySelector := map[string]string{}
	for name, method := range parsed.Methods {
		bySelector[hexutil.Encode(method.ID)] = name
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Id     json.RawMessage   `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&request))

		switch request.Method {
		case "eth_getCode":
			code := "0x"
			if hasCode {
				code = "0x60"
			}
			writeRpcResult(t, w, request.Id, code)
		case "eth_call":
			var call struct {
				Data  string `json:"data"`
				Input string `json:"input"`
			}
			assert.NoError(t, json.Unmarshal(request.Params[0], &call))
			data := call.Data
			if data == "" {
				data = call.Input
			}
			method, answer := methods[bySelector[data[:10]]]
			if !answer {
				writeRpcRevert(t, w, request.Id)
			} else if method.empty {
				writeRpcResult(t, w, request.Id, "0x")
			} else {
				writeRpcResult(t, w, request.Id, hexutil.Encode(method.data))
			}
		default:
			writeRpcRevert(t, w, request.Id)
		}
	}))
}

func writeRpcResult(t *testing.T, w http.ResponseWriter, id json.RawMessage, result string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": id, "result": result}))
}

func writeRpcRevert(t *testing.T, w http.ResponseWriter, id json.RawMessage) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0", "id": id,
		"error": map[string]any{"code": 3, "message": "execution reverted"},
	}))
}

// erc20ProbeMethods are the decimals/symbol/totalSupply reads every successful validation answers.
func erc20ProbeMethods(t *testing.T) map[string]erc20Method {
	t.Helper()
	return map[string]erc20Method{
		"decimals":    {data: packErc20Output(t, "decimals", uint8(6))},
		"symbol":      {data: packErc20Output(t, "symbol", "USDC")},
		"totalSupply": {data: packErc20Output(t, "totalSupply", big.NewInt(1_000_000))},
	}
}

// eip712DomainSeparator hashes the domain the way the token does, to verify the name()/version() fallback.
func eip712DomainSeparator(t *testing.T, name, version string) [32]byte {
	t.Helper()
	domain := apitypes.TypedDataDomain{
		Name:              name,
		Version:           version,
		ChainId:           math.NewHexOrDecimal256(erc20TestChainId),
		VerifyingContract: erc20TestAsset,
	}
	typedData := apitypes.TypedData{
		Types: apitypes.Types{"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		}},
		Domain: domain,
	}
	hash, err := typedData.HashStruct("EIP712Domain", domain.Map())
	assert.NoError(t, err)
	var out [32]byte
	copy(out[:], hash)
	return out
}
