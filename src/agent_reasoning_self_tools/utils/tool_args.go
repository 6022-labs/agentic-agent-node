package utils

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
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

func RawJsonArg(args map[string]any, key string) (json.RawMessage, error) {
	value, found := args[key]
	if !found || value == nil {
		return nil, fmt.Errorf("argument %q is required", key)
	}
	encoded, err := json.Marshal(value)
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
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		return fmt.Sprint(typedValue), true
	default:
		return "", false
	}
}
