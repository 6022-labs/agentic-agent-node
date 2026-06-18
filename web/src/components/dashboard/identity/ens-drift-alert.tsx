import React from "react"
import { toast } from "sonner"
import { Send, TriangleAlert } from "lucide-react"

import { Button } from "@/components/ui/button"
import { useWallet } from "@/contexts/wallet-provider-context"
import { useEnsRecordsDrift } from "./use-ens-records-drift"
import { extractErrorMessage } from "@/lib/api-errors"
import {
  initAgentTexts,
  setAgentText,
} from "@/queries/contracts/agent-ens-registry"

// Shown when on-chain ENS text records lag the node config (e.g. the node
// wallet couldn't pay gas for its boot-time reconciliation). The owner
// republishes from their own wallet — setAgentText allows owner or agent wallet.
export default function EnsDriftAlert() {
  const { wallet } = useWallet()
  const { chainId, drift, provisioned, ensInfo, recheck } = useEnsRecordsDrift()
  const [publishing, setPublishing] = React.useState(false)

  if (drift.length === 0 || !ensInfo || !wallet || !chainId) return null

  const republish = async () => {
    setPublishing(true)
    try {
      if (!provisioned) {
        // Fresh registry (e.g. after a redeploy): setAgentText reverts with
        // NotProvisioned until the full record set is initialized.
        await initAgentTexts(
          chainId,
          ensInfo.collectionAddress,
          ensInfo.tokenId,
          ensInfo.records,
          wallet
        )
      } else {
        for (const record of drift) {
          await setAgentText(
            chainId,
            ensInfo.collectionAddress,
            ensInfo.tokenId,
            record.key,
            record.desired,
            wallet
          )
        }
      }
      await recheck()
      toast.success("ENS records republished")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to republish ENS records"))
    } finally {
      setPublishing(false)
    }
  }

  return (
    <div className="rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm">
      <div className="flex items-center gap-2">
        <TriangleAlert className="size-4 shrink-0 text-amber-600 dark:text-amber-400" />
        <span className="flex-1">
          {provisioned
            ? "On-chain ENS records are out of date — the node could not update them itself."
            : "ENS records are not published on the current registry yet."}{" "}
          Republish from your wallet (you pay the gas).
        </span>
        <Button
          size="sm"
          variant="outline"
          onClick={republish}
          disabled={publishing}
        >
          <Send className={publishing ? "animate-pulse" : ""} />
          {publishing ? "Republishing…" : "Republish"}
        </Button>
      </div>
      <div className="mt-1.5 space-y-0.5 pl-6 font-mono text-xs text-muted-foreground">
        {drift.map((d) => (
          <div key={d.key} className="truncate">
            <span className="font-semibold">{d.key}</span>:{" "}
            {d.current || "(empty)"} → {d.desired}
          </div>
        ))}
      </div>
    </div>
  )
}
