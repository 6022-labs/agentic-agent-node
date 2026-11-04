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
import PaymentSpendRuleForm from "@/components/dashboard/payments/forms/payment-spend-rule-form"
import type {
  PaymentSpendRuleFormField,
  PaymentSpendRuleFormSchemaType,
} from "@/components/dashboard/payments/forms/payment-spend-rule-form/payment-spend-rule-form-schema"
import type { PaymentSpendRule } from "@/models/payment-spend-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { useErc20Lookup } from "@/components/dashboard/payments/use-erc20-lookup"

type Props = {
  // When set, the dialog edits this rule (pencil trigger) instead of adding one.
  initialRule?: PaymentSpendRule
  // result "saved" closes the dialog; anything else keeps the typed values for a retry.
  onSave: (
    values: PaymentSpendRuleFormSchemaType
  ) => Promise<FormCompletionResult<PaymentSpendRuleFormField, "saved">>
}

export default function PaymentSpendRuleDialog({ initialRule, onSave }: Props) {
  const [open, setOpen] = React.useState(false)
  const [saving, setSaving] = React.useState(false)
  const [values, setValues] =
    React.useState<PaymentSpendRuleFormSchemaType | null>(null)
  const editing = initialRule !== undefined

  const tokenLookup = useErc20Lookup(values?.network, values?.asset)

  const onFormComplete = async (
    values: PaymentSpendRuleFormSchemaType
  ): Promise<FormCompletionResult<PaymentSpendRuleFormField>> => {
    setSaving(true)
    try {
      const { error, result } = await onSave(values)
      if (result === "saved") {
        setValues(null)
        setOpen(false)
      }
      return { error }
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        // Closing mid-save would unmount the form and lose the save's error.
        if (saving) return
        // A previous session's token must never drive the next one's lookup.
        setValues(null)
        setOpen(next)
      }}
    >
      <DialogTrigger asChild>
        {editing ? (
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Edit rule ${initialRule?.name ?? ""}`}
          >
            <Pencil />
          </Button>
        ) : (
          <Button size="sm">
            <Plus />
            Add rule
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {editing ? "Edit spend rule" : "Add spend rule"}
          </DialogTitle>
          <DialogDescription>
            How much the agent may pay in one token.
          </DialogDescription>
        </DialogHeader>
        {/* Remount per open so an edited rule's values are re-derived fresh. */}
        {open ? (
          <PaymentSpendRuleForm
            initialRule={initialRule}
            tokenLookup={tokenLookup}
            onValueChange={setValues}
            onFormComplete={onFormComplete}
          />
        ) : null}
        <DialogFooter>
          <Button
            type="submit"
            form="payment-spend-rule-form"
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
