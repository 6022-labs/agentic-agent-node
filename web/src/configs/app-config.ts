import axios from "axios"

// Chain config served by the gateway's unauthenticated GET /config — fetched
// once at boot so the frontend always matches the backend's chain binding.
export type AppConfig = {
  chainId: number
  collectionsManagerAddress: string
  collectionCreatorAddress: string
  token6022Address: string
  ensRegistryAddress: string
  // CAIP-2 networks the runtime can settle x402 payments on; paid rules are
  // restricted to these.
  facilitatorNetworks: string[]
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
