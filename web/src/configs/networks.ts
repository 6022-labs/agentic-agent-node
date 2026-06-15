import {
  hardhat,
  polygon,
  polygonAmoy,
  type AppKitNetwork,
} from "@reown/appkit/networks"

import { findEvmNetwork } from "@/models/networks"

export const networkByChainId: Record<number, AppKitNetwork> = {
  137: polygon,
  80002: polygonAmoy,
  31337: hardhat,
}

// ENS lives on Ethereum L1: test chains resolve via the Sepolia ENS app,
// production chains via the mainnet one. Chains the console doesn't know (e.g.
// local) have no public ENS app. chainId is the plain id the identity lives on.
export const getEnsAppUrl = (
  chainId: string,
  ensDomain: string
): string | null => {
  const network = findEvmNetwork(`eip155:${chainId}`)
  if (!network) return null
  const base = network.testnet
    ? "https://sepolia.app.ens.domains"
    : "https://app.ens.domains"
  return `${base}/${ensDomain}`
}
