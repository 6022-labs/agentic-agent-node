import React from "react"

import {
  fetchErc20Info,
  isLookupSupported,
  type Erc20Lookup,
} from "@/lib/erc20"

const evmAddress = /^0x[0-9a-fA-F]{40}$/
const caip2Network = /^eip155:\d+$/

// useErc20Lookup resolves token metadata while the operator types — informative
// only; the runtime re-validates the contract on save with the node's own RPC.
export function useErc20Lookup(
  network: string | undefined,
  asset: string | undefined
): Erc20Lookup {
  const [lookup, setLookup] = React.useState<Erc20Lookup>({ status: "idle" })

  React.useEffect(() => {
    if (!network || !asset) {
      setLookup({ status: "idle" })
      return
    }
    if (!evmAddress.test(asset) || !caip2Network.test(network)) {
      setLookup({ status: "idle" })
      return
    }
    if (!isLookupSupported(network)) {
      setLookup({ status: "unsupported" })
      return
    }

    setLookup({ status: "loading" })
    let stale = false
    const timer = setTimeout(async () => {
      const info = await fetchErc20Info(network, asset)
      if (stale) return
      setLookup(info ? { status: "found", info } : { status: "not-found" })
    }, 400)

    return () => {
      stale = true
      clearTimeout(timer)
    }
  }, [network, asset])

  return lookup
}
