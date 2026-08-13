package services

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/models"
	"go.uber.org/zap"
)

type IPFSFileStorage struct {
	logger     *zap.Logger
	ipfsClient clients.IPFSClientInterface
}

func NewIPFSFileStorage(logger *zap.Logger, ipfsClient clients.IPFSClientInterface) *IPFSFileStorage {
	return &IPFSFileStorage{logger: logger, ipfsClient: ipfsClient}
}

func (s *IPFSFileStorage) UploadFile(fileBytes []byte, filename string) (*string, error) {
	pin := true
	response, err := s.ipfsClient.Add(fileBytes, filename, &models.ExternalAddQueryRequest{Pin: &pin})
	if err != nil {
		return nil, err
	}

	// Announce the CID to the DHT in the background — best effort.
	go func(hash string) {
		if err := s.ipfsClient.RoutingProvideWithoutResponse(
			&models.ExternalRoutingProvideRequest{Arg: hash},
		); err != nil {
			s.logger.Error("failed to provide routing", zap.String("hash", hash), zap.Error(err))
		}
	}(response.Hash)

	return &response.Hash, nil
}
