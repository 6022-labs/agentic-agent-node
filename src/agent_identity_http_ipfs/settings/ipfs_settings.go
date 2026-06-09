package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const IPFSSettingsKey = "ipfs"

// IPFSSettings targets a Kubo node's HTTP API. BearerToken (optional) is sent as
// `Authorization: Bearer <token>` for Kubo's API.Authorizations scheme.
type IPFSSettings struct {
	IpfsNodeHttpUrl string `koanf:"node_http_url"`
	BearerToken     string `koanf:"bearer_token"`
}

func NewIPFSSettings(logger *zap.Logger, k *koanf.Koanf) *IPFSSettings {
	var s IPFSSettings
	if err := k.Unmarshal(IPFSSettingsKey, &s); err != nil {
		logger.Fatal("failed to unmarshal ipfs settings", zap.Error(err))
	}

	s.IpfsNodeHttpUrl = strings.TrimSpace(s.IpfsNodeHttpUrl)
	if len(s.IpfsNodeHttpUrl) == 0 {
		logger.Fatal("please set your ipfs.node_http_url value in your configuration (IPFS__NODE_HTTP_URL)")
	}

	s.BearerToken = strings.TrimSpace(s.BearerToken)

	return &s
}
