import React from "react"
import { toast } from "sonner"
import {
  Coins,
  Loader2,
  RefreshCw,
  TriangleAlert,
  Wallet,
  type LucideIcon,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { useAuth } from "@/contexts/auth-context"
import { useWallet } from "@/contexts/wallet-provider-context"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import {
  TRANSACTION_STATUS,
  type AgentRegistration,
} from "@/models/agent-registration"
import { checkAgentMint } from "@/queries/api/registration.queries"
import { canMintDirectly } from "@/queries/contracts/agent-collection"
import { extractErrorMessage } from "@/lib/api-errors"
import SelfMintStep from "./self-mint-step"
import BrowserMintStep from "./browser-mint-step"

type MintMethod = "self" | "browser"

// Whether each sender wallet can mint directly; null while probing/unknown.
type DirectMintProbe = {
  self: boolean
  browser: boolean
}

type Props = {
  draft: AgentRegistration
  // Advance the wizard to the ENS step.
  onAdvance: () => void
}

export default function MintStep({ draft, onAdvance }: Props) {
  const { user } = useAuth()
  const { wallet } = useWallet()
  const { refresh } = useAgentIdentity()
  const [method, setMethod] = React.useState<MintMethod>("self")
  const [checking, setChecking] = React.useState(false)
  const [probe, setProbe] = React.useState<DirectMintProbe | null>(null)

  const isMinting = draft.mintStatus === TRANSACTION_STATUS.submitted
  const mintFailed = draft.mintStatus === TRANSACTION_STATUS.failed

  // The proposal rule depends on the tx sender: the agent's node wallet for
  // self-mint, the connected wallet for a browser mint.
  React.useEffect(() => {
    if (!wallet || !user?.address || !draft.collectionAddress) return
    let active = true
    Promise.all([
      canMintDirectly(draft.collectionAddress, draft.walletAddress, wallet),
      canMintDirectly(draft.collectionAddress, user.address, wallet),
    ])
      .then(([self, browser]) => {
        if (active) setProbe({ self, browser })
      })
      .catch(() => {
        // Probe is best-effort advice; minting still works without it.
        if (active) setProbe(null)
      })
    return () => {
      active = false
    }
  }, [wallet, user?.address, draft.collectionAddress, draft.walletAddress])

  const handleCheckMint = async () => {
    setChecking(true)
    try {
      const { data: status } = await checkAgentMint()
      await refresh()
      if (status.mintStatus === TRANSACTION_STATUS.confirmed) onAdvance()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to refresh status"))
    } finally {
      setChecking(false)
    }
  }

  if (isMinting) {
    return (
      <div className="space-y-4">
        <div className="flex items-start gap-2 rounded-md bg-muted p-3 text-sm text-muted-foreground">
          <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin" />
          <span>The mint transaction is being confirmed.</span>
        </div>
        <Button
          variant="outline"
          onClick={handleCheckMint}
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
      {mintFailed ? (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          The last mint attempt failed. You can try again below.
        </div>
      ) : null}

      <div className="grid gap-2 sm:grid-cols-2">
        <MethodCard
          icon={Coins}
          title="Self-mint"
          description="Fund the agent wallet; the node mints & registers ENS itself."
          selected={method === "self"}
          onClick={() => setMethod("self")}
          directMint={probe?.self}
        />
        <MethodCard
          icon={Wallet}
          title="Browser wallet"
          description="Sign the mint and ENS transactions from your own wallet."
          selected={method === "browser"}
          onClick={() => setMethod("browser")}
          directMint={probe?.browser}
        />
      </div>

      {method === "self" ? (
        <SelfMintStep
          draft={draft}
          willPropose={probe ? !probe.self : false}
          onSubmitted={() => {
            void refresh()
            onAdvance()
          }}
        />
      ) : (
        <BrowserMintStep
          draft={draft}
          willPropose={probe ? !probe.browser : false}
          onMinted={onAdvance}
        />
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
  directMint,
}: {
  icon: LucideIcon
  title: string
  description: string
  selected: boolean
  onClick: () => void
  // undefined = probe unavailable; show no hint rather than guessing.
  directMint?: boolean
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
      {directMint === true ? (
        <span className="text-xs text-muted-foreground">Mints directly.</span>
      ) : null}
      {directMint === false ? (
        <span className="flex items-center gap-1 text-xs text-amber-600 dark:text-amber-500">
          <TriangleAlert className="size-3 shrink-0" />
          Creates a proposal — needs moderator approval.
        </span>
      ) : null}
    </button>
  )
}
