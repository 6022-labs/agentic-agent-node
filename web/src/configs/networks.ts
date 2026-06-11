import {
  hardhat,
  polygon,
  polygonAmoy,
  type AppKitNetwork,
} from "@reown/appkit/networks"

export const networkByChainId: Record<number, AppKitNetwork> = {
  137: polygon,
  80002: polygonAmoy,
  31337: hardhat,
}

// ENS lives on Ethereum L1: test chains resolve via the Sepolia ENS app,
// production chains via the mainnet one. Local chains have no public ENS app.
export const getEnsAppUrl = (
  chainId: number,
  ensDomain: string
): string | null => {
  const network = networkByChainId[chainId]
  if (!network || network.id === hardhat.id) return null
  const base =
    network.testnet === true
      ? "https://sepolia.app.ens.domains"
      : "https://app.ens.domains"
  return `${base}/${ensDomain}`
}
