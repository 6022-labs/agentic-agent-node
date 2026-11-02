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

type Props = {
  // When set, the dialog edits this rule (pencil trigger) instead of adding one.
  initialRule?: PaymentSpendRule
  onSave: (
    values: PaymentSpendRuleFormSchemaType
  ) => Promise<FormCompletionResult<PaymentSpendRuleFormField>>
}

export default function PaymentSpendRuleDialog({ initialRule, onSave }: Props) {
  const [open, setOpen] = React.useState(false)
  const [saving, setSaving] = React.useState(false)
  const editing = initialRule !== undefined

  const onFormComplete = async (
    values: PaymentSpendRuleFormSchemaType
  ): Promise<FormCompletionResult<PaymentSpendRuleFormField>> => {
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
            {editing ? "Edit spend rule" : "Add spend rule"}
          </DialogTitle>
          <DialogDescription>
            How much the agent may pay in one token. A payment must fit the cap
            and every window; who may be paid is set on the policy.
          </DialogDescription>
        </DialogHeader>
        {/* Remount per open so an edited rule's values are re-derived fresh. */}
        {open ? (
          <PaymentSpendRuleForm
            initialRule={initialRule}
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
