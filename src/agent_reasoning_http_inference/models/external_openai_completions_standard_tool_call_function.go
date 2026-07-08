package models

import (
	"encoding/json"
	"strings"
)

type ExternalOpenAICompletionsStandardToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (f *ExternalOpenAICompletionsStandardToolCallFunction) ParsedArguments() (map[string]any, error) {
	if strings.TrimSpace(f.Arguments) == "" {
		return map[string]any{}, nil
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(f.Arguments), &args); err != nil {
		return nil, err
	}

	return args, nil
}
