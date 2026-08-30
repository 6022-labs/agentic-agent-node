import React from "react"
import { toast } from "sonner"
import {
  CheckCircle2,
  Coins,
  ExternalLink,
  Loader2,
  RefreshCw,
  Wallet,
  type LucideIcon,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"
import { getEnsAppUrl } from "@/configs/networks"
import { useWallet } from "@/contexts/wallet-provider-context"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import type { AgentRegistration } from "@/models/agent-registration"
import { TRANSACTION_STATUS } from "@/models/enums/transaction-status"
import {
  checkAgentEns,
  checkAgentMint,
  getAgentEnsInfo,
  selfRegisterAgent,
} from "@/queries/api/registration.queries"
import { initAgentTexts } from "@/queries/contracts/agent-ens-registry"
import { ensureGas } from "@/lib/native-transfer"
import { extractErrorMessage } from "@/lib/api-errors"

type EnsMethod = "self" | "browser"

type Props = {
  draft: AgentRegistration
  onClose: () => void
}

export default function EnsStep({ draft, onClose }: Props) {
  const { wallet } = useWallet()
  const { refresh } = useAgentIdentity()
  const [method, setMethod] = React.useState<EnsMethod>("self")
  const [amount, setAmount] = React.useState("0.01")
  const [busy, setBusy] = React.useState(false)
  const [checking, setChecking] = React.useState(false)

  const isMinted = draft.mintStatus === TRANSACTION_STATUS.confirmed
  const isRegistered = draft.ensStatus === TRANSACTION_STATUS.confirmed
  const isRegistering = draft.ensStatus === TRANSACTION_STATUS.submitted

  const handleCheck = async () => {
    setChecking(true)
    try {
      const { data: status } = await checkAgentMint()
      if (status.mintStatus === TRANSACTION_STATUS.confirmed)
        await checkAgentEns()
      await refresh()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to refresh status"))
    } finally {
      setChecking(false)
    }
  }

  // Let the node publish ENS itself, funding its wallet only if short.
  const handleSelfEns = async () => {
    if (!wallet) {
      toast.error("Connect a wallet first")
      return
    }

    setBusy(true)
    try {
      if (await ensureGas(draft.addresses?.evm, amount, wallet)) {
        toast.success("Gas sent to the agent wallet")
      }
      await selfRegisterAgent()
      toast.success("ENS registration submitted")
      await refresh()
    } catch (error) {
      toast.error(extractErrorMessage(error, "ENS registration failed"))
    } finally {
      setBusy(false)
    }
  }

  // Sign the ENS records transaction from the connected (owner) wallet.
  const handleBrowserEns = async () => {
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

  if (!isMinted || isRegistering) {
    return (
      <div className="space-y-4">
        <div className="flex items-start gap-2 rounded-md bg-muted p-3 text-sm text-muted-foreground">
          <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin" />
          <span>
            {!isMinted
              ? "The node is completing the mint. ENS registration follows once it confirms."
              : "ENS registration is in progress."}
          </span>
        </div>
        <Button
          variant="outline"
          onClick={handleCheck}
          loading={checking}
          className="w-full"
        >
          {!checking ? <RefreshCw /> : null}
          Refresh status
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-2 sm:grid-cols-2">
        <MethodCard
          icon={Coins}
          title="Self-register"
          description="Fund the agent wallet if needed; the node publishes ENS itself."
          selected={method === "self"}
          onClick={() => setMethod("self")}
        />
        <MethodCard
          icon={Wallet}
          title="Browser wallet"
          description="Sign the ENS records transaction from your own wallet."
          selected={method === "browser"}
          onClick={() => setMethod("browser")}
        />
      </div>

      {method === "self" ? (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            The node publishes the agent's ENS records itself. Gas is only sent
            to the agent wallet if it doesn't already have enough.
          </p>
          <div className="grid gap-1.5">
            <Label htmlFor="self-ens-amount">Gas to ensure (if short)</Label>
            <Input
              id="self-ens-amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="0.01"
              inputMode="decimal"
            />
          </div>
          <div className="flex gap-2">
            <Button onClick={handleSelfEns} loading={busy} className="flex-1">
              {!busy ? <Coins /> : null}
              Self-register ENS
            </Button>
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
      ) : (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Publish the agent's records to the ENS registry by signing the
            transaction from your wallet.
          </p>
          <div className="flex gap-2">
            <Button
              onClick={handleBrowserEns}
              loading={busy}
              className="flex-1"
            >
              {!busy ? <Wallet /> : null}
              Register ENS
            </Button>
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
      )}
    </div>
  )
}

function MethodCard({
  icon: Icon,
  title,
  description,
  selected,
  onClick,
}: {
  icon: LucideIcon
  title: string
  description: string
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex cursor-pointer flex-col items-start gap-1 rounded-lg border p-3 text-left transition-colors",
        selected ? "border-primary bg-accent/50" : "hover:bg-accent/30"
      )}
    >
      <span className="flex items-center gap-2 text-sm font-medium">
        <Icon className="size-4" />
        {title}
      </span>
      <span className="text-xs text-muted-foreground">{description}</span>
    </button>
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
