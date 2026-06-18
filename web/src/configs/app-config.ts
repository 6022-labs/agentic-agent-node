import axios from "axios"

// One chain's config. Empty addresses mean the contract isn't deployed there;
// facilitatorNetwork is the chain's CAIP-2 x402 network, null if none.
export type ChainConfig = {
  collectionsManagerAddress: string
  collectionCreatorAddress: string
  token6022Address: string
  ensRegistryAddress: string
  mintable: boolean
  facilitatorNetwork: string | null
}

// Config served by the gateway's GET /config at boot, chains keyed by chain id.
export type AppConfig = {
  chains: Record<string, ChainConfig>
}

let appConfig: AppConfig | null = null
let pending: Promise<AppConfig> | null = null

// Idempotent: concurrent/repeated calls (StrictMode double-mount) share one request.
export async function loadAppConfig(): Promise<AppConfig> {
  if (appConfig) return appConfig
  if (!pending) {
    pending = axios
      .get<AppConfig>("/config")
      .then(({ data }) => {
        appConfig = data
        return data
      })
      .finally(() => {
        pending = null
      })
  }
  return pending
}

export function getAppConfig(): AppConfig {
  if (!appConfig) {
    throw new Error("App config read before loadAppConfig() resolved")
  }
  return appConfig
}

// getChainConfig returns the config for a chain id, or undefined if the node
// doesn't serve it.
export function getChainConfig(chainId: string): ChainConfig | undefined {
  return getAppConfig().chains[chainId]
}

// getChainConfigOrThrow resolves a chain's contract addresses, failing loudly
// when the node doesn't serve that chain.
export function getChainConfigOrThrow(chainId: string): ChainConfig {
  const chainConfig = getChainConfig(chainId)
  if (!chainConfig) {
    throw new Error(`Chain ${chainId} is not configured on this node`)
  }
  return chainConfig
}

// getMintableChainIds lists the chains an agent identity can be created on.
export function getMintableChainIds(): string[] {
  const { chains } = getAppConfig()
  return Object.keys(chains).filter((id) => chains[id].mintable)
}

// getFacilitatorNetworks lists the CAIP-2 networks that can settle paid x402.
export function getFacilitatorNetworks(): string[] {
  return Object.values(getAppConfig().chains)
    .map((c) => c.facilitatorNetwork)
    .filter((network): network is string => network !== null)
}
