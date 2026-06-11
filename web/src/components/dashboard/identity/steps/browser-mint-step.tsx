import React from "react"
import { toast } from "sonner"
import { Hourglass, RefreshCw, Wallet } from "lucide-react"

import { Button } from "@/components/ui/button"
import { useWallet } from "@/contexts/wallet-provider-context"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import {
  TRANSACTION_STATUS,
  type AgentRegistration,
} from "@/models/agent-registration"
import {
  checkAgentMint,
  getAgentMintInfo,
} from "@/queries/api/registration.queries"
import { createMintProposal, mint } from "@/queries/contracts/agent-collection"
import { extractErrorMessage } from "@/lib/api-errors"

type Props = {
  draft: AgentRegistration
  // True when the connected wallet is not a moderator of a moderated
  // collection, so signing submits a proposal instead of minting directly.
  willPropose?: boolean
  onMinted: () => void
}

export default function BrowserMintStep({
  draft,
  willPropose,
  onMinted,
}: Props) {
  const { wallet } = useWallet()
  const { refresh } = useAgentIdentity()
  const [busy, setBusy] = React.useState(false)
  const [checking, setChecking] = React.useState(false)

  const awaitingModerator = draft.proposalId != null

  const handleMint = async () => {
    if (!wallet) {
      toast.error("Connect a wallet first")
      return
    }

    setBusy(true)
    try {
      const { data: info } = await getAgentMintInfo()
      const params = {
        collectionAddress: info.collectionAddress,
        to: info.to,
        role: info.role,
        name: info.name,
        wallets: info.wallets,
        images: info.images,
        attributes: info.attributes,
        cloneOf: info.cloneOf,
      }
      const needsProposal = info.moderatorCount !== 0 && !info.ownerIsModerator
      if (needsProposal) {
        await createMintProposal(params, wallet)
        toast.success("Mint proposal submitted — a moderator must approve it")
      } else {
        await mint(params, wallet)
        toast.success("Mint transaction confirmed")
      }
      const { data: status } = await checkAgentMint()
      await refresh()
      if (status.mintStatus === TRANSACTION_STATUS.confirmed) onMinted()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Mint failed"))
    } finally {
      setBusy(false)
    }
  }

  const handleCheck = async () => {
    setChecking(true)
    try {
      const { data: status } = await checkAgentMint()
      await refresh()
      if (status.mintStatus === TRANSACTION_STATUS.confirmed) onMinted()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to refresh status"))
    } finally {
      setChecking(false)
    }
  }

  if (awaitingModerator) {
    return (
      <div className="space-y-4">
        <div className="flex items-start gap-2 rounded-md bg-muted p-3 text-sm text-muted-foreground">
          <Hourglass className="mt-0.5 size-4 shrink-0" />
          <span>
            Mint proposal #{draft.proposalId} is waiting for a moderator's
            approval. Check again once it has been approved.
          </span>
        </div>
        <Button
          variant="outline"
          onClick={handleCheck}
          loading={checking}
          className="w-full"
        >
          {!checking ? <RefreshCw /> : null}
          Check approval status
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Sign the mint transaction from your own wallet. If the collection is
        moderated, this submits a proposal instead.
      </p>
      {willPropose ? (
        <div className="flex items-start gap-2 rounded-md bg-amber-500/10 p-3 text-sm text-amber-600 dark:text-amber-500">
          <Hourglass className="mt-0.5 size-4 shrink-0" />
          <span>
            This collection is moderated and your wallet is not a moderator:
            signing will submit a mint proposal that a moderator must approve.
          </span>
        </div>
      ) : null}
      <Button onClick={handleMint} loading={busy} className="w-full">
        {!busy ? <Wallet /> : null}
        {willPropose ? "Submit mint proposal" : "Mint agent"}
      </Button>
    </div>
  )
}
