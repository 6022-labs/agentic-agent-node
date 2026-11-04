import React from "react"

import { fetchErc20Info, isLookupSupported } from "@/lib/erc20"
import { knownToken, type TokenUnits } from "@/lib/token-format"

// One lookup per token for the whole page, however many cells show it.
const lookups = new Map<string, Promise<TokenUnits | null>>()

const lookup = (network: string, asset: string): Promise<TokenUnits | null> => {
  const key = `${network}:${asset.toLowerCase()}`
  let pending = lookups.get(key)
  if (!pending) {
    pending = fetchErc20Info(network, asset).then((info) => {
      // A failed lookup is not cached, so a transient RPC error can be retried.
      if (!info) lookups.delete(key)
      return info ? { symbol: info.symbol, decimals: info.decimals } : null
    })
    lookups.set(key, pending)
  }
  return pending
}

// useTokenInfo resolves a rule token's symbol and decimals: the known list first, then the chain.
export function useTokenInfo(
  network: string,
  asset: string
): TokenUnits | undefined {
  const known = knownToken(network, asset)
  const key = `${network}:${asset.toLowerCase()}`
  // Keyed by token, so a row edited onto another token never keeps the old one's units.
  const [resolved, setResolved] = React.useState<{
    key: string
    token: TokenUnits
  }>()

  React.useEffect(() => {
    if (known || !isLookupSupported(network)) return
    let stale = false
    lookup(network, asset).then((info) => {
      if (!stale && info) setResolved({ key, token: info })
    })
    return () => {
      stale = true
    }
  }, [known, network, asset, key])

  return known ?? (resolved?.key === key ? resolved.token : undefined)
}
