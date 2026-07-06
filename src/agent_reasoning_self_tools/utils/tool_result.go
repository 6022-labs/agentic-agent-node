package utils

import (
	"encoding/json"
	"fmt"
)

// ToToolResult json-round-trips a use-case response into the generic tool result shape.
func ToToolResult(response any, err error) (map[string]any, error) {
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tool result: %w", err)
	}

	out := make(map[string]any)
	if err := json.Unmarshal(encoded, &out); err != nil {
		return nil, fmt.Errorf("failed to decode tool result: %w", err)
	}
	return out, nil
}

// SuccessResult is the shape returned by tools wrapping error-only use cases.
func SuccessResult() map[string]any {
	return map[string]any{"success": true}
}
