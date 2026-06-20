package x402

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

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
