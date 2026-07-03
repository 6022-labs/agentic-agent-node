import axios from "axios"

// One EVM chain's config. Empty addresses mean the contract isn't deployed there;
// facilitatorNetwork is the chain's CAIP-2 x402 network, null if none.
export type EvmChainConfig = {
  collectionsManagerAddress: string
  collectionCreatorAddress: string
  token6022Address: string
  ensRegistryAddress: string
  mintable: boolean
  facilitatorNetwork: string | null
}

// Config served by the gateway's GET /config at boot; evmChains keyed by chain id.
export type AppConfig = {
  evmChains: Record<string, EvmChainConfig>
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

// getEvmChainConfig returns the config for a chain id, or undefined if the node
// doesn't serve it.
export function getEvmChainConfig(chainId: string): EvmChainConfig | undefined {
  return getAppConfig().evmChains[chainId]
}

// getEvmChainConfigOrThrow resolves a chain's contract addresses, failing loudly
// when the node doesn't serve that chain.
export function getEvmChainConfigOrThrow(chainId: string): EvmChainConfig {
  const evmChainConfig = getEvmChainConfig(chainId)
  if (!evmChainConfig) {
    throw new Error(`Chain ${chainId} is not configured on this node`)
  }
  return evmChainConfig
}

// getMintableChainIds lists the chains an agent identity can be created on.
export function getMintableChainIds(): string[] {
  const { evmChains } = getAppConfig()
  return Object.keys(evmChains).filter((id) => evmChains[id].mintable)
}

// getFacilitatorNetworks lists the CAIP-2 networks that can settle paid x402.
export function getFacilitatorNetworks(): string[] {
  return Object.values(getAppConfig().evmChains)
    .map((c) => c.facilitatorNetwork)
    .filter((network): network is string => network !== null)
}
