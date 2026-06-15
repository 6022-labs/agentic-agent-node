import React from "react"
import { toast } from "sonner"
import {
  CheckCircle2,
  ExternalLink,
  Loader2,
  RefreshCw,
  Wallet,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { getEnsAppUrl } from "@/configs/networks"
import { useWallet } from "@/contexts/wallet-provider-context"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import {
  TRANSACTION_STATUS,
  type AgentRegistration,
} from "@/models/agent-registration"
import {
  checkAgentEns,
  checkAgentMint,
  getAgentEnsInfo,
} from "@/queries/api/registration.queries"
import { initAgentTexts } from "@/queries/contracts/agent-ens-registry"
import { extractErrorMessage } from "@/lib/api-errors"

type Props = {
  draft: AgentRegistration
  onClose: () => void
}

export default function EnsStep({ draft, onClose }: Props) {
  const { wallet } = useWallet()
  const { refresh } = useAgentIdentity()
  const [busy, setBusy] = React.useState(false)
  const [checking, setChecking] = React.useState(false)

  const isMinted = draft.mintStatus === TRANSACTION_STATUS.confirmed
  const isRegistered = draft.ensStatus === TRANSACTION_STATUS.confirmed
  const isRegistering = draft.ensStatus === TRANSACTION_STATUS.submitted

  const handleCheck = async () => {
    setChecking(true)
    try {
      await checkAgentMint()
      if (isMinted) await checkAgentEns()
      await refresh()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to refresh status"))
    } finally {
      setChecking(false)
    }
  }

  const handleEns = async () => {
    if (!wallet) {
      toast.error("Connect a wallet first")
      return
    }

    setBusy(true)
    try {
      const { data: info } = await getAgentEnsInfo()
      await initAgentTexts(
        draft.chainId,
        info.collectionAddress,
        info.tokenId,
        info.records,
        wallet
      )
      toast.success("ENS records transaction confirmed")
      await checkAgentEns()
      await refresh()
    } catch (error) {
      toast.error(extractErrorMessage(error, "ENS registration failed"))
    } finally {
      setBusy(false)
    }
  }

  if (isRegistered) {
    return (
      <div className="space-y-4">
        <div className="flex flex-col items-center gap-2 py-4 text-center">
          <CheckCircle2 className="size-10 text-primary" />
          <p className="font-medium">Identity complete</p>
          <p className="text-sm text-muted-foreground">
            The agent is minted and its ENS records are published
            {draft.ensDomain ? (
              <>
                {" "}
                at{" "}
                <EnsDomainLink
                  chainId={draft.chainId}
                  ensDomain={draft.ensDomain}
                />
              </>
            ) : null}
            .
          </p>
        </div>
        <Button onClick={onClose} className="w-full">
          Done
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {!isMinted || isRegistering ? (
        <div className="flex items-start gap-2 rounded-md bg-muted p-3 text-sm text-muted-foreground">
          <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin" />
          <span>
            {!isMinted
              ? "The node is completing the mint. ENS registration follows automatically."
              : "ENS registration is in progress."}
          </span>
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">
          Publish the agent's records to the ENS registry by signing the
          transaction from your wallet.
        </p>
      )}
      <div className="flex gap-2">
        {isMinted && !isRegistering ? (
          <Button onClick={handleEns} loading={busy} className="flex-1">
            {!busy ? <Wallet /> : null}
            Register ENS
          </Button>
        ) : null}
        <Button
          variant="outline"
          onClick={handleCheck}
          loading={checking}
          className="flex-1"
        >
          {!checking ? <RefreshCw /> : null}
          Refresh status
        </Button>
      </div>
    </div>
  )
}

function EnsDomainLink({
  chainId,
  ensDomain,
}: {
  chainId: string
  ensDomain: string
}) {
  const href = getEnsAppUrl(chainId, ensDomain)
  if (!href) return <span className="font-mono">{ensDomain}</span>
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className="inline-flex items-center gap-1 font-mono hover:underline"
    >
      {ensDomain}
      <ExternalLink className="size-3 shrink-0" />
    </a>
  )
}
