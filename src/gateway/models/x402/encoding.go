package x402

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const caip2Eip155Prefix = "eip155:"

// EncodeBase64Json renders v the way x402 headers carry JSON payloads.
func EncodeBase64Json(v any) (string, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encoded), nil
}

// DecodePaymentPayload parses a PAYMENT-SIGNATURE header value and rejects
// anything but x402 v2.
func DecodePaymentPayload(header string) (*PaymentPayload, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header))
	if err != nil {
		return nil, fmt.Errorf("payment header is not valid base64: %w", err)
	}
	var payload PaymentPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("payment header is not valid JSON: %w", err)
	}
	if payload.X402Version != Version {
		return nil, fmt.Errorf("unsupported x402 version %d (only %d is supported)", payload.X402Version, Version)
	}
	return &payload, nil
}

// ParseEip155ChainId extracts the chain id from a CAIP-2 eip155 network
// identifier ("eip155:80002" → 80002).
func ParseEip155ChainId(network string) (int64, error) {
	rest, found := strings.CutPrefix(network, caip2Eip155Prefix)
	if !found {
		return 0, fmt.Errorf("network %q is not a CAIP-2 eip155 identifier", network)
	}
	chainId, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || chainId <= 0 {
		return 0, fmt.Errorf("network %q has an invalid chain id", network)
	}
	return chainId, nil
}
