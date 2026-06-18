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
  role: string
  cloneOf: number | null
  owner: string
  walletAddress: string
  ensDomain: string
  images: Record<string, string>
  attributes: Record<string, string>
  mintStatus: string
  ensStatus: string
}

export type SaveAgentRegistrationRequest = {
  chainId: string
  collectionAddress: string
  name: string
  role: string
  cloneOf?: number | null
  // Single base image file the node frames into the NFT image set.
  image: File
  attributes?: Record<string, string>
}

export type MintInfo = {
  collectionAddress: string
  to: string
  role: string
  name: string
  wallets: string[]
  images: KeyValue[]
  attributes: KeyValue[]
  cloneOf: number | null
  moderatorCount: number
  ownerIsModerator: boolean
}

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
