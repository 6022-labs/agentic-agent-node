import React from "react"

import {
  fetchErc20Info,
  isLookupSupported,
  type Erc20Lookup,
} from "@/lib/erc20"
import { knownToken } from "@/lib/token-format"

const evmAddress = /^0x[0-9a-fA-F]{40}$/
// ethers waits up to 5 minutes by default, which would keep the amount inputs disabled.
const LOOKUP_TIMEOUT_MS = 10_000
const caip2Network = /^eip155:\d+$/

// useErc20Lookup resolves token metadata while the operator types — informative
// only; the runtime re-validates the contract on save with the node's own RPC.
export function useErc20Lookup(
  rawNetwork: string | undefined,
  rawAsset: string | undefined
): Erc20Lookup {
  const [lookup, setLookup] = React.useState<Erc20Lookup>({ status: "idle" })
  // A pasted address with surrounding spaces must still resolve.
  const network = rawNetwork?.trim()
  const asset = rawAsset?.trim()

  React.useEffect(() => {
    if (!network || !asset) {
      setLookup({ status: "idle" })
      return
    }
    if (!evmAddress.test(asset) || !caip2Network.test(network)) {
      setLookup({ status: "idle" })
      return
    }
    const known = knownToken(network, asset)
    if (known) {
      setLookup({ status: "found", info: { name: "", ...known } })
      return
    }
    if (!isLookupSupported(network)) {
      setLookup({ status: "unsupported" })
      return
    }

    setLookup({ status: "loading" })
    let stale = false
    const timer = setTimeout(async () => {
      const info = await Promise.race([
        fetchErc20Info(network, asset),
        new Promise<"timeout">((resolve) =>
          setTimeout(() => resolve("timeout"), LOOKUP_TIMEOUT_MS)
        ),
      ])
      if (stale) return
      if (info === "timeout") {
        setLookup({ status: "timeout" })
        return
      }
      setLookup(info ? { status: "found", info } : { status: "not-found" })
    }, 400)

    return () => {
      stale = true
      clearTimeout(timer)
    }
  }, [network, asset])

  return lookup
}
