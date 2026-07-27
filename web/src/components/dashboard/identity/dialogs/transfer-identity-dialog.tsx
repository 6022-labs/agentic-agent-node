import React from "react"
import { toast } from "sonner"
import { ArrowRightLeft, TriangleAlert } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import TransferIdentityForm from "@/components/dashboard/identity/forms/transfer-identity-form"
import type { TransferIdentityFormSchemaType } from "@/components/dashboard/identity/forms/transfer-identity-form/transfer-identity-form-schema"
import { useAuth } from "@/contexts/auth-context"
import { useWallet } from "@/contexts/wallet-provider-context"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { transferFrom } from "@/queries/contracts/agent-collection"
import { extractErrorMessage } from "@/lib/api-errors"

type Props = {
  collectionAddress: string
  collectionTokenId: string
  agentName?: string
}

export default function TransferIdentityDialog({
  collectionAddress,
  collectionTokenId,
  agentName,
}: Props) {
  const { signOut } = useAuth()
  const { wallet } = useWallet()

  const [open, setOpen] = React.useState(false)
  const [saving, setSaving] = React.useState(false)

  const onFormComplete = async (
    values: TransferIdentityFormSchemaType
  ): Promise<FormCompletionResult<keyof TransferIdentityFormSchemaType>> => {
    if (!wallet) {
      toast.error("Connect a wallet first")
      return {}
    }

    setSaving(true)
    try {
      await transferFrom(
        collectionAddress,
        collectionTokenId,
        values.to,
        wallet
      )
      setOpen(false)
      toast.success("Ownership transferred — this wallet no longer has access.")
      // The gateway only issues a JWT to the current owner, so this session is
      // dead: drop it rather than leave a console open on a foreign agent.
      signOut()
      return {}
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to transfer the identity"))
      return {}
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button" variant="ghost" size="sm">
          <ArrowRightLeft />
          Transfer
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Transfer identity</DialogTitle>
          <DialogDescription>
            {agentName ? `Transfers "${agentName}"` : "Transfers this agent"} to
            another wallet. The transaction is signed by the current owner.
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-start gap-2 rounded-md bg-amber-500/10 p-3 text-sm text-amber-600 dark:text-amber-500">
          <TriangleAlert className="mt-0.5 size-4 shrink-0" />
          <span>
            Sign-in to this console is restricted to the identity NFT owner. Once
            the transfer confirms, this wallet is signed out and can no longer
            sign in — only the new owner can configure this agent, and only they
            can transfer it back.
          </span>
        </div>

        <TransferIdentityForm onFormComplete={onFormComplete} />

        <DialogFooter>
          <Button
            type="submit"
            form="transfer-identity-form"
            variant="destructive"
            loading={saving}
          >
            Transfer
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
