package use_cases

import (
	"context"
	"io"
	"math/big"
	"mime/multipart"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
)

// SaveAgentRegistration upserts the agent draft the connected wallet filled in,
// binding it to the node's bootstrap wallet, and seeds the registration state.
// On every save it (re)frames the uploaded base image into the collection's
// "default"/"icon" set, pins them to IPFS, and stores the resulting CIDs.
type SaveAgentRegistration struct {
	agentRepository                  repositories.AgentRepositoryInterface
	agentAddressRepository           repositories.AgentAddressRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface
	agentCollectionRequester         identity_interfaces.AgentCollectionRequesterInterface
	walletCredentialsStorage         identity_interfaces.WalletCredentialsStorageInterface
	generateAgentImages              identity_interfaces.GenerateAgentImagesInterface
	ipfsFileStorage                  identity_interfaces.IPFSFileStorageInterface
	token6022Requester               identity_interfaces.Token6022RequesterInterface
	ensDomainBuilder                 identity_interfaces.EnsDomainBuilderInterface
	evmChainsSettings                *runtime_common_settings.EvmChainsSettings
	agentCollectionCreatorsSettings  *runtime_common_settings.AgentCollectionCreatorsSettings
}

func NewSaveAgentRegistration(
	agentRepository repositories.AgentRepositoryInterface,
	agentAddressRepository repositories.AgentAddressRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	agentCollectionsManagerRequester identity_interfaces.AgentCollectionsManagerRequesterInterface,
	agentCollectionRequester identity_interfaces.AgentCollectionRequesterInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
	generateAgentImages identity_interfaces.GenerateAgentImagesInterface,
	ipfsFileStorage identity_interfaces.IPFSFileStorageInterface,
	token6022Requester identity_interfaces.Token6022RequesterInterface,
	ensDomainBuilder identity_interfaces.EnsDomainBuilderInterface,
	evmChainsSettings *runtime_common_settings.EvmChainsSettings,
	agentCollectionCreatorsSettings *runtime_common_settings.AgentCollectionCreatorsSettings,
) *SaveAgentRegistration {
	return &SaveAgentRegistration{
		agentRepository:                  agentRepository,
		agentAddressRepository:           agentAddressRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		agentCollectionsManagerRequester: agentCollectionsManagerRequester,
		agentCollectionRequester:         agentCollectionRequester,
		walletCredentialsStorage:         walletCredentialsStorage,
		generateAgentImages:              generateAgentImages,
		ipfsFileStorage:                  ipfsFileStorage,
		token6022Requester:               token6022Requester,
		ensDomainBuilder:                 ensDomainBuilder,
		evmChainsSettings:                evmChainsSettings,
		agentCollectionCreatorsSettings:  agentCollectionCreatorsSettings,
	}
}

func (u *SaveAgentRegistration) Execute(
	ctx context.Context,
	req *requests.SaveAgentRegistrationRequest,
) (*responses.AgentRegistrationResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	// Identity-capable = reachable + a creator is deployed.
	if !u.evmChainsSettings.Has(req.ChainId) {
		return nil, errors.NewValidationError("chainId", "node has no RPC for this chain")
	}
	if _, ok := u.agentCollectionCreatorsSettings.AddressFor(req.ChainId); !ok {
		return nil, errors.NewValidationError("chainId", "no agent collection creator is deployed on this chain")
	}

	collection := common.HexToAddress(req.CollectionAddress)

	known, err := u.agentCollectionsManagerRequester.IsKnownCollection(req.ChainId, collection)
	if err != nil {
		return nil, errors.NewUnavailableError("collection_check_failed", err.Error())
	}
	if !known {
		return nil, errors.NewValidationError("collectionAddress", "collection is not governed by the manager")
	}

	existing, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err.Error())
	}
	if existing == nil {
		return nil, errors.NewUnavailableError("wallet_not_ready", "node wallet has not been bootstrapped yet")
	}

	collectionName, err := u.agentCollectionRequester.CollectionName(req.ChainId, collection)
	if err != nil {
		return nil, errors.NewUnavailableError("collection_check_failed", err.Error())
	}

	owner := req.Owner

	agent := &models.Agent{
		ChainId:           lo.ToPtr(req.ChainId),
		CollectionAddress: collection.Hex(),
		Name:              lo.ToPtr(req.Name),
		CloneOf:           req.CloneOf,
		EnsDomain:         lo.ToPtr(u.ensDomainBuilder.Build(req.Name, collectionName, req.ChainId)),
		Creator:           &owner,
		Owner:             &owner,
		Attributes:        req.Attributes,
	}
	// Preserve an already-minted identity if the owner re-edits the form.
	agent.CollectionTokenId = existing.CollectionTokenId
	agent.ProposalId = existing.ProposalId

	addresses, err := u.agentAddressRepository.GetAllByAgentId(ctx, existing.Id)
	if err != nil {
		return nil, errors.NewInternalError("agent_addresses_lookup_failed", err.Error())
	}

	agent.Images, err = u.resolveImages(agent, existing, addresses, common.HexToAddress(owner), req.Image)
	if err != nil {
		return nil, err
	}

	saved, err := u.agentRepository.Upsert(ctx, agent)
	if err != nil {
		return nil, errors.NewInternalError("agent_save_failed", err.Error())
	}

	state, err := u.agentRegistrationStateRepository.GetByAgentId(ctx, saved.Id)
	if err != nil {
		return nil, errors.NewInternalError("state_lookup_failed", err.Error())
	}
	if state == nil {
		state, err = u.agentRegistrationStateRepository.Upsert(ctx, &models.AgentRegistrationState{
			AgentId:           saved.Id,
			MintStatus:        enums.TransactionStatusPending,
			EnsRegistryStatus: enums.TransactionStatusPending,
		})
		if err != nil {
			return nil, errors.NewInternalError("state_save_failed", err.Error())
		}
	}

	resp := &responses.AgentRegistrationResponse{}
	resp.FromModel(saved, addresses, state)
	return resp, nil
}

func (u *SaveAgentRegistration) resolveImages(
	agent, existing *models.Agent,
	addresses map[enums.AddressType]string,
	creator common.Address,
	baseImage *multipart.FileHeader,
) (map[string]string, error) {
	if baseImage != nil {
		return u.buildImages(agent, addresses, creator, baseImage)
	}
	if len(existing.Images) > 0 {
		return existing.Images, nil
	}
	return nil, errors.NewValidationError("image", "is required")
}

// buildImages frames the uploaded base image, pins each rendered variant to
// IPFS, and returns the CID per image key (what the mint stores on-chain).
func (u *SaveAgentRegistration) buildImages(
	agent *models.Agent,
	addresses map[enums.AddressType]string,
	creator common.Address,
	baseImage *multipart.FileHeader,
) (map[string]string, error) {
	wallets := []string{}
	for _, addressType := range enums.GetAllAddressTypes() {
		if address := addresses[addressType]; address != "" {
			wallets = append(wallets, address)
		}
	}

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
		Name:              lo.FromPtr(agent.Name),
		EnsDomain:         lo.FromPtr(agent.EnsDomain),
		Role:              agent.Attributes["role"],
		Creator:           creator.Hex(),
		BaseImageBytes:    baseBytes,
		BaseImageMimeType: mimeType,
		Wallets:           wallets,
		CreatedAt:         time.Now(),
		Stars:             stars,
	})
	if err != nil {
		return nil, errors.NewInternalError("image_generation_failed", err.Error())
	}

	images := make(map[string]string, len(generated))
	for key, bytes := range generated {
		cid, err := u.ipfsFileStorage.UploadFile(bytes, lo.FromPtr(agent.EnsDomain))
		if err != nil {
			return nil, errors.NewUnavailableError("image_upload_failed", err.Error())
		}
		images[key] = *cid
	}
	return images, nil
}

// computeStars maps the creator's $6022 balance to a 0-8 rating by whole-token
// digit count. The balance is summed across every reachable chain, best-effort
// (chains that fail to read are skipped).
func (u *SaveAgentRegistration) computeStars(creator common.Address) (uint8, error) {
	total := new(big.Int)
	for _, chainId := range u.evmChainsSettings.ChainIds() {
		balance, err := u.token6022Requester.BalanceOf(chainId, creator)
		if err != nil {
			continue
		}
		total.Add(total, balance)
	}

	decimals, err := u.token6022Requester.Decimals()
	if err != nil {
		return 0, err
	}

	whole := total.String()
	if len(whole) <= int(decimals) {
		return 0, nil
	}

	return min(uint8(len(whole)-int(decimals)), 8), nil
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
