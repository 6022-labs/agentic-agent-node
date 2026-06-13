import { Contract, JsonRpcProvider } from "ethers"

import { findEvmNetwork } from "@/models/networks"

const ERC20_ABI = [
  "function name() view returns (string)",
  "function symbol() view returns (string)",
  "function decimals() view returns (uint8)",
]

export type Erc20Info = {
  name: string
  symbol: string
  decimals: number
}

export function isLookupSupported(network: string): boolean {
  return findEvmNetwork(network) !== undefined
}

// fetchErc20Info resolves the token metadata via the network's public RPC,
// or null when nothing ERC-20-shaped answers at the address. The runtime
// re-validates the contract on save with the node's own RPC.
export async function fetchErc20Info(
  network: string,
  asset: string
): Promise<Erc20Info | null> {
  const rpcUrl = findEvmNetwork(network)?.rpcUrl
  if (!rpcUrl) {
    return null
  }

  const provider = new JsonRpcProvider(rpcUrl)
  try {
    const token = new Contract(asset, ERC20_ABI, provider)
    const [symbol, decimals, name] = await Promise.all([
      token.symbol() as Promise<string>,
      token.decimals() as Promise<bigint>,
      (token.name() as Promise<string>).catch(() => ""),
    ])
    return { name, symbol, decimals: Number(decimals) }
  } catch {
    return null
  } finally {
    provider.destroy()
  }
}
