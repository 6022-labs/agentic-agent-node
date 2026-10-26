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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
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

// maxUint256 is the allowance the ensurer grants; a var since a const cannot hold a *big.Int.
var maxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

// approvalRpcServer answers the reads bind needs to send a tx and records every approve.
func approvalRpcServer(t *testing.T, allowances []*big.Int) (*httptest.Server, *[]approveCall) {
	t.Helper()

	parsed, err := abi.JSON(strings.NewReader(contract_abi.Erc20MetaData.ABI))
	assert.NoError(t, err)
	allowanceId := hexutil.Encode(parsed.Methods["allowance"].ID)
	approveId := hexutil.Encode(parsed.Methods["approve"].ID)

	submitted := &[]approveCall{}
	allowanceReads := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Id     json.RawMessage   `json:"id"`
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&request))

		switch request.Method {
		case "eth_chainId", "net_version":
			writeRpcResult(t, w, request.Id, "0x13882")
		case "eth_getCode":
			writeRpcResult(t, w, request.Id, "0x60")
		case "eth_getTransactionCount":
			writeRpcResult(t, w, request.Id, "0x0")
		case "eth_getBlockByNumber":
			// bind reads the head block's base fee to build a dynamic-fee tx.
			writeRpcRaw(t, w, request.Id, headerJson())
		case "eth_gasPrice", "eth_maxPriorityFeePerGas":
			writeRpcResult(t, w, request.Id, "0x1")
		case "eth_estimateGas":
			writeRpcResult(t, w, request.Id, "0x186a0")
		case "eth_sendRawTransaction":
			var raw string
			assert.NoError(t, json.Unmarshal(request.Params[0], &raw))
			*submitted = append(*submitted, decodeApprove(t, parsed, raw))
			writeRpcResult(t, w, request.Id, (*submitted)[len(*submitted)-1].hash)
		case "eth_getTransactionReceipt":
			var hash string
			assert.NoError(t, json.Unmarshal(request.Params[0], &hash))
			writeRpcRaw(t, w, request.Id, receiptJson(hash))
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
			switch data[:10] {
			case allowanceId:
				value := allowances[min(allowanceReads, len(allowances)-1)]
				allowanceReads++
				writeRpcResult(t, w, request.Id, hexutil.Encode(packErc20Output(t, "allowance", value)))
			case approveId:
				writeRpcResult(t, w, request.Id, hexutil.Encode(packErc20Output(t, "approve", true)))
			default:
				writeRpcRevert(t, w, request.Id)
			}
		default:
			writeRpcRevert(t, w, request.Id)
		}
	}))
	t.Cleanup(server.Close)
	return server, submitted
}

// approveCall is one submitted approve, decoded back.
type approveCall struct {
	hash    string
	spender common.Address
	amount  *big.Int
}

func decodeApprove(t *testing.T, parsed abi.ABI, rawTx string) approveCall {
	t.Helper()

	bytes, err := hexutil.Decode(rawTx)
	assert.NoError(t, err)
	tx := new(types.Transaction)
	assert.NoError(t, tx.UnmarshalBinary(bytes))

	args, err := parsed.Methods["approve"].Inputs.Unpack(tx.Data()[4:])
	assert.NoError(t, err)
	return approveCall{
		hash:    tx.Hash().Hex(),
		spender: args[0].(common.Address),
		amount:  args[1].(*big.Int),
	}
}

func receiptJson(hash string) string {
	return `{"transactionHash":"` + hash + `","status":"0x1","blockNumber":"0x1",` +
		`"blockHash":"0x` + strings.Repeat("11", 32) + `","transactionIndex":"0x0",` +
		`"cumulativeGasUsed":"0x1","gasUsed":"0x1","logs":[],"logsBloom":"0x` + strings.Repeat("00", 256) + `","type":"0x2"}`
}

func writeRpcRaw(t *testing.T, w http.ResponseWriter, id json.RawMessage, result string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(id) + `,"result":` + result + `}`))
	assert.NoError(t, err)
}

// headerJson is the minimum header go-ethereum unmarshals; bind reads baseFeePerGas from it.
func headerJson() string {
	zeroHash := "0x" + strings.Repeat("00", 32)
	zeroBloom := "0x" + strings.Repeat("00", 256)
	return `{"number":"0x1","baseFeePerGas":"0x1","gasLimit":"0x1c9c380","gasUsed":"0x0",` +
		`"parentHash":"` + zeroHash + `","sha3Uncles":"` + zeroHash + `","stateRoot":"` + zeroHash + `",` +
		`"transactionsRoot":"` + zeroHash + `","receiptsRoot":"` + zeroHash + `","logsBloom":"` + zeroBloom + `",` +
		`"difficulty":"0x0","timestamp":"0x1","extraData":"0x","mixHash":"` + zeroHash + `",` +
		`"nonce":"0x0000000000000000","miner":"0x0000000000000000000000000000000000000000"}`
}
