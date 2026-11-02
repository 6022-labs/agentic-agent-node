import type { ReactNode } from "react"

import { formatNetwork } from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

import HoverHint from "./hover-hint"
import { useTokenInfo } from "./use-token-info"

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
  const symbol = useTokenInfo(network, asset)?.symbol
  return (
    <div className="flex min-w-0 items-start gap-1">
      <HoverHint
        className="block min-w-0"
        content={
          <>
            <p className="font-mono break-all">{asset}</p>
            <p className="text-muted-foreground">{network}</p>
          </>
        }
      >
        <span
          className={
            symbol
              ? "block truncate font-medium"
              : "block truncate font-mono text-xs"
          }
        >
          {symbol ? `$${symbol}` : truncateEthAddress(asset)}
        </span>
        <span className="block truncate text-xs text-muted-foreground">
          {formatNetwork(network)}
        </span>
      </HoverHint>
      {badge}
    </div>
  )
}
