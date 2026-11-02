package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

// StringArg requires the key to be present; an empty string is a valid value (e.g. clearing guidance).
func StringArg(args map[string]any, key string) (string, error) {
	value, found := args[key]
	if !found || value == nil {
		return "", fmt.Errorf("argument %q is required", key)
	}
	text, ok := scalarToString(value)
	if !ok {
		return "", fmt.Errorf("argument %q must be a string", key)
	}
	return text, nil
}

func OptionalStringArg(args map[string]any, key string) string {
	value, found := args[key]
	if !found || value == nil {
		return ""
	}
	text, _ := scalarToString(value)
	return text
}

// OptionalIntArg returns nil when the key is absent or unparseable, so an explicit 0 stays distinguishable.
func OptionalIntArg(args map[string]any, key string) *int {
	text := OptionalStringArg(args, key)
	if text == "" {
		return nil
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil
	}
	// Converting an out-of-range float is implementation-defined, so it is clamped rather than converted.
	if value > math.MaxInt32 {
		return lo.ToPtr(math.MaxInt32)
	}
	if value < math.MinInt32 {
		return lo.ToPtr(math.MinInt32)
	}
	return lo.ToPtr(int(value))
}

// OptionalBoolArg returns nil when the key is absent or unparseable, so an explicit false stays distinguishable.
func OptionalBoolArg(args map[string]any, key string) *bool {
	value, err := strconv.ParseBool(OptionalStringArg(args, key))
	if err != nil {
		return nil
	}
	return &value
}

func UuidArg(args map[string]any, key string) (uuid.UUID, error) {
	raw, err := StringArg(args, key)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("argument %q must be a uuid", key)
	}
	return id, nil
}

func RawJsonObjectArg(args map[string]any, key string) (json.RawMessage, error) {
	value, found := args[key]
	if !found || value == nil {
		return nil, fmt.Errorf("argument %q is required", key)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("argument %q must be a json object", key)
	}
	encoded, err := json.Marshal(object)
	if err != nil {
		return nil, fmt.Errorf("argument %q must be a json object: %w", key, err)
	}
	return encoded, nil
}

// scalarToString converts numbers/bools to a string, since models often send them for string-typed fields.
func scalarToString(value any) (string, bool) {
	switch typedValue := value.(type) {
	case string:
		return typedValue, true
	case json.Number:
		return typedValue.String(), true
	case float64:
		// Plain decimal, never fmt.Sprint's exponent form (1e+06), which breaks big.Int/chain parsing downstream.
		return strconv.FormatFloat(typedValue, 'f', -1, 64), true
	case float32:
		return strconv.FormatFloat(float64(typedValue), 'f', -1, 32), true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		return fmt.Sprint(typedValue), true
	default:
		return "", false
	}
}

func JsonObjectArg[T any](args map[string]any, key string) (*T, error) {
	raw, err := RawJsonObjectArg(args, key)
	if err != nil {
		return nil, err
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("argument %q has the wrong shape: %w", key, err)
	}
	return &value, nil
}

// JsonArrayArg decodes a required array argument into []T; an empty array is a valid value.
func JsonArrayArg[T any](args map[string]any, key string) ([]T, error) {
	value, found := args[key]
	if !found || value == nil {
		return nil, fmt.Errorf("argument %q is required", key)
	}
	return OptionalJsonArrayArg[T](args, key)
}

// OptionalJsonArrayArg decodes an array argument into []T, returning nil when it is absent.
func OptionalJsonArrayArg[T any](args map[string]any, key string) ([]T, error) {
	value, found := args[key]
	if !found || value == nil {
		return nil, nil
	}
	array, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("argument %q must be a json array", key)
	}
	encoded, err := json.Marshal(array)
	if err != nil {
		return nil, fmt.Errorf("argument %q must be a json array: %w", key, err)
	}
	var out []T
	if err := json.Unmarshal(encoded, &out); err != nil {
		return nil, fmt.Errorf("argument %q has the wrong shape: %w", key, err)
	}
	return out, nil
}
