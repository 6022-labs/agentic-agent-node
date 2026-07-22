export type KeyValue = {
  key: string
  value: string
}

export type AgentRegistration = {
  // Plain chain id the identity lives on (e.g. "80002").
  chainId: string
  collectionAddress: string
  tokenId: number | null
  proposalId: number | null
  name: string
  cloneOf: number | null
  owner: string
  // Wallet addresses per chain family, keyed by AddressType ("evm", "btc").
  addresses: Record<string, string>
  ensDomain: string
  images: Record<string, string>
  attributes: Record<string, string>
  mintStatus: string
  ensStatus: string
}

// A bootstrapped-but-empty agent carries no draft fields or status; a real draft
// or mint always sets at least one. Decides card vs. guided-empty-state.
export const hasStartedRegistration = (registration: AgentRegistration) =>
  !!registration.mintStatus || !!registration.ensStatus || !!registration.name

// Every field the mint needs is present. Gates advancing past the details step,
// from both the form's "Continue" button and the card's "Continue setup".
export const isDraftComplete = (registration: AgentRegistration | null) =>
  !!registration &&
  !!registration.chainId &&
  !!registration.collectionAddress &&
  !!registration.name &&
  !!registration.attributes?.["role"] &&
  !!registration.images &&
  Object.keys(registration.images).length > 0

export type SaveAgentRegistrationRequest = {
  chainId: string
  collectionAddress: string
  name: string
  cloneOf?: number | null
  // Single base image file the node frames into the NFT image set. Omitted on
  // re-edit to keep the stored image.
  image?: File
  // Must include the mandatory "role" key.
  attributes: Record<string, string>
}

export type MintInfo = {
  collectionAddress: string
  to: string
  name: string
  // Wallet addresses per chain family, keyed by AddressType ("evm", "btc").
  addresses: Record<string, string>
  images: KeyValue[]
  attributes: KeyValue[]
  cloneOf: number | null
  moderatorCount: number
  ownerIsModerator: boolean
}

// AgentAddress is one typed on-chain address entry ({addressType, value}).
export type AgentAddress = {
  addressType: string
  value: string
}

// ADDRESS_TYPE_ORDER fixes the on-chain addresses[] order across address maps.
export const ADDRESS_TYPE_ORDER = ["evm", "btc"] as const

// orderedAgentAddresses flattens an address map into the on-chain addresses[]
// order, dropping any chain family the agent doesn't hold yet.
export const orderedAgentAddresses = (
  addresses: Record<string, string>
): AgentAddress[] =>
  ADDRESS_TYPE_ORDER.filter((type) => !!addresses?.[type]).map((type) => ({
    addressType: type,
    value: addresses[type],
  }))

export type EnsInfo = {
  collectionAddress: string
  tokenId: number
  records: KeyValue[]
}

export type RegistrationStatus = {
  mintStatus: string
  ensStatus: string
  tokenId: number | null
  proposalId: number | null
  mintTxHash: string | null
  ensTxHash: string | null
}
