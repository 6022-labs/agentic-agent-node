package x402

import "encoding/json"

// X402Config is the x402 method's per-rule settings, carried in a rule's
// method-specific Config blob.
type X402Config struct {
	Network         string   `json:"network"`
	Asset           string   `json:"asset"`
	PayTo           string   `json:"payTo"`
	Eip712Name      string   `json:"eip712Name"`
	Eip712Version   string   `json:"eip712Version"`
	AddressPatterns []string `json:"addressPatterns"`
}

func ParseConfig(raw []byte) (*X402Config, error) {
	var config X402Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
