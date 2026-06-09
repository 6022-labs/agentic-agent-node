package use_cases

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
)

// SaveAgentRegistration upserts the agent draft the connected wallet filled in,
// binding it to the node's bootstrap wallet, and seeds the registration state.
// On every save it (re)frames the uploaded base image into the collection's
// "default"/"icon" set, pins them to IPFS, and stores the resulting CIDs.
type SaveAgentRegistration struct {
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface
	agentCollectionRequester         identity_interfaces.AgentCollectionRequesterInterface
	walletCredentialsStorage         identity_interfaces.WalletCredentialsStorageInterface
	generateAgentImages              identity_interfaces.GenerateAgentImagesInterface
	ipfsFileStorage                  identity_interfaces.IPFSFileStorageInterface
	token6022Requester               identity_interfaces.Token6022RequesterInterface
	agentIdentitySettings            *settings.AgentIdentitySettings
}

func NewSaveAgentRegistration(
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
	generateAgentImages identity_interfaces.GenerateAgentImagesInterface,
	ipfsFileStorage identity_interfaces.IPFSFileStorageInterface,
	token6022Requester identity_interfaces.Token6022RequesterInterface,
	agentIdentitySettings *settings.AgentIdentitySettings,
) *SaveAgentRegistration {
	return &SaveAgentRegistration{
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		agentCollectionsManagerRequester: agentCollectionsManagerRequester,
		agentCollectionRequester:         agentCollectionRequester,
		walletCredentialsStorage:         walletCredentialsStorage,
		generateAgentImages:              generateAgentImages,
		ipfsFileStorage:                  ipfsFileStorage,
		token6022Requester:               token6022Requester,
		agentIdentitySettings:            agentIdentitySettings,
	}
}

func (u *SaveAgentRegistration) Execute(ctx context.Context, req *requests.SaveAgentRegistrationRequest) (*responses.AgentRegistrationResponse, error) {
	collection := common.HexToAddress(req.CollectionAddress)

	known, err := u.agentCollectionsManagerRequester.IsKnownCollection(collection)
	if err != nil {
		return nil, errors.NewUnavailableError("collection_check_failed", err.Error())
	}
	if !known {
		return nil, errors.NewValidationError("collectionAddress", "collection is not governed by the manager")
	}

	cred, found, err := u.walletCredentialsStorage.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("wallet_lookup_failed", err.Error())
	}
	if !found {
		return nil, errors.NewUnavailableError("wallet_not_ready", "node wallet has not been bootstrapped yet")
	}

	collectionName, err := u.agentCollectionRequester.CollectionName(collection)
	if err != nil {
		return nil, errors.NewUnavailableError("collection_check_failed", err.Error())
	}

	existing, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err.Error())
	}

	owner := req.Owner
	var role enums.AgentRole
	_ = role.Set(req.Role)

	agent := &models.Agent{
		CollectionAddress: collection.Hex(),
		Name:              req.Name,
		CloneOf:           req.CloneOf,
		EnsDomain:         u.ensDomain(req.Name, collectionName),
		Role:              role,
		Creator:           owner,
		WalletAddress:     cred.Address,
		Owner:             &owner,
		Attributes:        req.Attributes,
	}
	// Preserve an already-minted identity if the owner re-edits the form.
	if existing != nil {
		agent.CollectionTokenId = existing.CollectionTokenId
		agent.ProposalId = existing.ProposalId
	}

	images, err := u.buildImages(agent, common.HexToAddress(owner), cred.Address, req.Image)
	if err != nil {
		return nil, err
	}
	agent.Images = images

	saved, err := u.agentRepository.Upsert(ctx, agent)
	if err != nil {
		return nil, errors.NewInternalError("agent_save_failed", err.Error())
	}

	state, err := u.agentRegistrationStateRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("state_lookup_failed", err.Error())
	}
	if state == nil {
		state, err = u.agentRegistrationStateRepository.Upsert(ctx, &models.AgentRegistrationState{
			MintStatus:        enums.TransactionStatusPending,
			EnsRegistryStatus: enums.TransactionStatusPending,
		})
		if err != nil {
			return nil, errors.NewInternalError("state_save_failed", err.Error())
		}
	}

	resp := &responses.AgentRegistrationResponse{}
	resp.FromModel(saved, state)
	return resp, nil
}

// buildImages frames the uploaded base image, pins each rendered variant to
// IPFS, and returns the CID per image key (what the mint stores on-chain).
func (u *SaveAgentRegistration) buildImages(agent *models.Agent, creator common.Address, wallet string, baseImage *multipart.FileHeader) (map[string]string, error) {
	baseBytes, err := readMultipartFile(baseImage)
	if err != nil {
		return nil, errors.NewValidationError("image", "could not read uploaded image: "+err.Error())
	}
	mimeType := baseImage.Header.Get("Content-Type")

	stars, err := u.computeStars(creator)
	if err != nil {
		return nil, errors.NewUnavailableError("token_read_failed", err.Error())
	}

	generated, err := u.generateAgentImages.Generate(identity_interfaces.GenerateImageParams{
		Name:              agent.Name,
		EnsDomain:         agent.EnsDomain,
		Role:              agent.Role.String(),
		Creator:           creator.Hex(),
		BaseImageBytes:    baseBytes,
		BaseImageMimeType: mimeType,
		Wallets:           []string{wallet},
		CreatedAt:         time.Now(),
		Stars:             stars,
	})
	if err != nil {
		return nil, errors.NewInternalError("image_generation_failed", err.Error())
	}

	images := make(map[string]string, len(generated))
	for key, bytes := range generated {
		cid, err := u.ipfsFileStorage.UploadFile(bytes, wallet)
		if err != nil {
			return nil, errors.NewUnavailableError("image_upload_failed", err.Error())
		}
		images[key] = *cid
	}
	return images, nil
}

// computeStars maps the creator's $6022 balance to a 0-8 rating by the number
// of whole-token digits (matches the orchestrator's create-agent flow).
func (u *SaveAgentRegistration) computeStars(creator common.Address) (uint8, error) {
	balance, err := u.token6022Requester.BalanceOf(creator)
	if err != nil {
		return 0, err
	}

	decimals, err := u.token6022Requester.Decimals()
	if err != nil {
		return 0, err
	}

	whole := balance.String()
	if len(whole) <= int(decimals) {
		return 0, nil
	}

	return min(uint8(len(whole)-int(decimals)), 8), nil
}

// ensDomain composes the display domain: <name>.<collection>.<baseDomain>.
func (u *SaveAgentRegistration) ensDomain(name, collectionName string) string {
	return fmt.Sprintf("%s.%s.%s",
		strings.ToLower(name),
		strings.ToLower(collectionName),
		strings.ToLower(u.agentIdentitySettings.EnsBaseDomain),
	)
}

// readMultipartFile opens an uploaded file and reads its bytes.
func readMultipartFile(header *multipart.FileHeader) ([]byte, error) {
	file, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
