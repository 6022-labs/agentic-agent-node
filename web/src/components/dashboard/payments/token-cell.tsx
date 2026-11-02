import type { ReactNode } from "react"

import { formatNetwork, knownTokenSymbol } from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

// TokenCell shows a rule's token over its chain, shared by the receive and spend tables.
export default function TokenCell({
  network,
  asset,
  badge,
}: {
  network: string
  asset: string
  badge?: ReactNode
}) {
  const symbol = knownTokenSymbol(asset)
  return (
    <div className="min-w-0">
      <div className="flex min-w-0 items-center gap-1">
        <span
          className={
            symbol ? "truncate font-medium" : "truncate font-mono text-xs"
          }
          title={asset}
        >
          {symbol ? `$${symbol}` : truncateEthAddress(asset)}
        </span>
        {badge}
      </div>
      <div className="truncate text-xs text-muted-foreground" title={network}>
        {formatNetwork(network)}
      </div>
    </div>
  )
}
