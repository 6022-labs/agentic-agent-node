import React from "react"
import { toast } from "sonner"
import { Coins, TriangleAlert } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useWallet } from "@/contexts/wallet-provider-context"
import type { AgentRegistration } from "@/models/agent-registration"
import { selfRegisterAgent } from "@/queries/api/registration.queries"
import { sendNative } from "@/queries/contracts/native-transfer"
import { extractErrorMessage } from "@/lib/api-errors"

type Props = {
  draft: AgentRegistration
  // True when the agent wallet is not a moderator of a moderated collection,
  // so the node will open a proposal instead of minting directly.
  willPropose?: boolean
  // Called once the node has accepted the self-registration request.
  onSubmitted: () => void
}

export default function SelfMintStep({
  draft,
  willPropose,
  onSubmitted,
}: Props) {
  const { wallet } = useWallet()
  const [amount, setAmount] = React.useState("0.05")
  const [busy, setBusy] = React.useState(false)

  const handleSelfMint = async () => {
    if (!wallet) {
      toast.error("Connect a wallet first")
      return
    }

    setBusy(true)
    try {
      // Fund the node wallet, then let the node mint + ENS-register itself.
      await sendNative(draft.walletAddress, amount, wallet)
      toast.success("Gas sent to the agent wallet")
      await selfRegisterAgent()
      toast.success("Agent self-registration submitted")
      onSubmitted()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Self-mint failed"))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Send a little gas to the agent's wallet, then the node mints and
        registers ENS on its own.
      </p>
      {willPropose ? (
        <div className="flex items-start gap-2 rounded-md bg-amber-500/10 p-3 text-sm text-amber-600 dark:text-amber-500">
          <TriangleAlert className="mt-0.5 size-4 shrink-0" />
          <span>
            This collection is moderated and the agent wallet is not a
            moderator: the node will open a mint proposal, and the agent stays
            unminted until a moderator approves it.
          </span>
        </div>
      ) : null}
      <div className="grid gap-1.5">
        <Label>Agent wallet</Label>
        <p className="rounded-md bg-muted px-3 py-2 font-mono text-xs break-all">
          {draft.walletAddress}
        </p>
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="self-mint-amount">Amount to send</Label>
        <Input
          id="self-mint-amount"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0.05"
          inputMode="decimal"
        />
      </div>
      <Button onClick={handleSelfMint} loading={busy} className="w-full">
        {!busy ? <Coins /> : null}
        Send gas &amp; self-register
      </Button>
    </div>
  )
}
