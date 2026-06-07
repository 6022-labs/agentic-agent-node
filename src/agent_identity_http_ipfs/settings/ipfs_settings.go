package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const IPFSSettingsKey = "ipfs"

// IPFSSettings points the adapter at a Kubo (go-ipfs) node's HTTP API.
// Set ipfs.node_http_url (env: IPFS__NODE_HTTP_URL). BearerToken is optional —
// when set, requests carry `Authorization: Bearer <token>`, matching Kubo's
// API.Authorizations bearer scheme (server side: AuthSecret "bearer:<token>").
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
