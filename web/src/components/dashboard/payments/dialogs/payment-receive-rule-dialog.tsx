import React from "react"
import { Loader2, Pencil, Plus } from "lucide-react"

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
import PaymentReceiveRuleForm from "@/components/dashboard/payments/forms/payment-receive-rule-form"
import type { PaymentReceiveRuleFormSchemaType } from "@/components/dashboard/payments/forms/payment-receive-rule-form/payment-receive-rule-form-schema"
import type { PaymentReceiveRule } from "@/models/payment-receive-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { useErc20Lookup } from "@/components/dashboard/payments/use-erc20-lookup"

type Props = {
  agentWalletAddress: string
  // When set, the dialog edits this rule (pencil trigger) instead of adding
  // one (Add trigger).
  initialRule?: PaymentReceiveRule
  preservedPatternCount?: number
  // Persists the rule into the policy; returns a field error to keep the
  // dialog open for correction.
  onSave: (
    values: PaymentReceiveRuleFormSchemaType
  ) => Promise<FormCompletionResult<keyof PaymentReceiveRuleFormSchemaType>>
}

export default function PaymentReceiveRuleDialog({
  agentWalletAddress,
  initialRule,
  preservedPatternCount,
  onSave,
}: Props) {
  const [open, setOpen] = React.useState(false)
  const [saving, setSaving] = React.useState(false)
  const [values, setValues] =
    React.useState<PaymentReceiveRuleFormSchemaType | null>(null)
  const editing = initialRule !== undefined

  const tokenLookup = useErc20Lookup(values?.network, values?.asset)

  const onFormComplete = async (
    values: PaymentReceiveRuleFormSchemaType
  ): Promise<FormCompletionResult<keyof PaymentReceiveRuleFormSchemaType>> => {
    setSaving(true)
    try {
      const result = await onSave(values)
      if (!result.error) {
        setOpen(false)
      }
      return result
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {editing ? (
          <Button variant="ghost" size="icon-sm" aria-label="Edit rule">
            <Pencil />
          </Button>
        ) : (
          <Button size="sm">
            <Plus />
            Add rule
          </Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {editing ? "Edit pricing rule" : "Add pricing rule"}
          </DialogTitle>
          <DialogDescription>
            A network + token the matched callers pay with, per call.
            {editing
              ? ""
              : " Pre-filled with the node's chain and the $6022 token."}
          </DialogDescription>
        </DialogHeader>
        {/* Remount the form each time the dialog opens so an edited rule's
            values are re-derived fresh. */}
        {open ? (
          <PaymentReceiveRuleForm
            agentWalletAddress={agentWalletAddress}
            initialRule={initialRule}
            preservedPatternCount={preservedPatternCount}
            tokenLookup={tokenLookup}
            onValueChange={setValues}
            onFormComplete={onFormComplete}
          />
        ) : null}
        <DialogFooter>
          <Button
            type="submit"
            form="payment-receive-rule-form"
            disabled={saving}
          >
            {saving ? <Loader2 className="animate-spin" /> : null}
            {editing ? "Save" : "Add"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
