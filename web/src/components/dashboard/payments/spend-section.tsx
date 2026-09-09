import React from "react"
import { toast } from "sonner"
import { AlertTriangle, Trash2, Wallet } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import ConfirmDialog from "@/components/dashboard/confirm-dialog"
import EmptyState from "@/components/dashboard/empty-state"
import PaymentSpendRuleDialog from "@/components/dashboard/payments/dialogs/payment-spend-rule-dialog"
import type { PaymentSpendRuleFormSchemaType } from "@/components/dashboard/payments/forms/payment-spend-rule-form/payment-spend-rule-form-schema"
import {
  ANY_PATTERN,
  isOpenToEverything,
  isUnbounded,
  type PaymentSpendPolicy,
  type PaymentSpendRule,
} from "@/models/payment-spend-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  deletePaymentSpendPolicy,
  getPaymentSpendPolicy,
} from "@/queries/api/payment-spend-policy.queries"
import {
  createPaymentSpendRule,
  deletePaymentSpendRule,
  listPaymentSpendRules,
  updatePaymentSpendRule,
} from "@/queries/api/payment-spend-rule.queries"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"
import { isPlainAddressPattern } from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

const MAX_PATTERN_CHIPS = 2

// formatPeriod names the common periods and falls back to raw seconds.
const formatPeriod = (seconds: number): string => {
  if (seconds % 86400 === 0) {
    const days = seconds / 86400
    if (days === 1) {
      return "day"
    }
    if (days === 7) {
      return "week"
    }
    if (days === 30) {
      return "month"
    }
    return `${days}d`
  }
  if (seconds % 3600 === 0) {
    return `${seconds / 3600}h`
  }
  return `${seconds}s`
}

// The fields the rule form renders; errors on anything else are toasted.
const FORM_FIELDS: Array<keyof PaymentSpendRuleFormSchemaType> = [
  "name",
  "method",
  "maxAmountPerPayment",
  "windowSeconds",
  "maxAmountPerWindow",
  "networkPatterns",
  "assetPatterns",
  "payToPatterns",
  "assetTransferMethodPatterns",
]

// splitPatterns turns the form's one-per-line text back into the wire array.
const splitPatterns = (value: string): string[] =>
  value
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line.length > 0)

export default function PaymentSpendPolicySection() {
  const [policy, setPolicy] = React.useState<PaymentSpendPolicy | null>(null)
  const [rules, setRules] = React.useState<PaymentSpendRule[]>([])
  const [loading, setLoading] = React.useState(true)
  const [mutating, setMutating] = React.useState(false)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      // Sequential on purpose: both calls lazily materialize the policy server-side.
      const policyResponse = await getPaymentSpendPolicy()
      const rulesResponse = await listPaymentSpendRules()
      setPolicy(policyResponse.data)
      setRules(rulesResponse.rules)
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to load spend policy"))
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void load()
  }, [load])

  const reloadRules = async () => {
    const { rules } = await listPaymentSpendRules()
    setRules(rules)
  }

  const toRule = (
    values: PaymentSpendRuleFormSchemaType
  ): PaymentSpendRule => ({
    name: values.name,
    method: values.method,
    maxAmountPerPayment: values.maxAmountPerPayment,
    windowSeconds:
      values.windowSeconds.length === 0 ? 0 : Number(values.windowSeconds),
    maxAmountPerWindow:
      values.maxAmountPerWindow.length === 0 ? "0" : values.maxAmountPerWindow,
    networkPatterns: splitPatterns(values.networkPatterns),
    assetPatterns: splitPatterns(values.assetPatterns),
    payToPatterns: splitPatterns(values.payToPatterns),
    assetTransferMethodPatterns: splitPatterns(
      values.assetTransferMethodPatterns
    ),
  })

  const onSaveRule =
    (existing?: PaymentSpendRule) =>
    async (
      values: PaymentSpendRuleFormSchemaType
    ): Promise<FormCompletionResult<keyof PaymentSpendRuleFormSchemaType>> => {
      setMutating(true)
      try {
        const rule = toRule(values)
        if (existing?.id) {
          await updatePaymentSpendRule(existing.id, rule)
        } else {
          await createPaymentSpendRule(rule)
        }
        await reloadRules()
        toast.success(existing ? "Spend rule updated" : "Spend rule added")
        return {}
      } catch (error) {
        const validation = extractValidationError(error)
        if (validation) {
          // Runtime errors on the x402 matcher come back prefixed (e.g.
          // "config.networkPatterns"); strip it to reach the form control.
          const field = validation.field.replace(/^config\./, "")
          if (
            FORM_FIELDS.includes(field as keyof PaymentSpendRuleFormSchemaType)
          ) {
            return {
              error: {
                field: field as keyof PaymentSpendRuleFormSchemaType,
                message: validation.message,
              },
            }
          }
          // A field with no form control: toast it rather than swallow it.
          toast.error(validation.message)
          return {}
        }
        toast.error(extractErrorMessage(error, "Failed to save the spend rule"))
        return {}
      } finally {
        setMutating(false)
      }
    }

  const onDeleteRule = async (rule: PaymentSpendRule) => {
    if (!rule.id) {
      return
    }
    setMutating(true)
    try {
      await deletePaymentSpendRule(rule.id)
      await reloadRules()
      toast.success("Spend rule removed")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to remove the spend rule"))
    } finally {
      setMutating(false)
    }
  }

  const onResetPolicy = async () => {
    setMutating(true)
    try {
      await deletePaymentSpendPolicy()
      await load()
      toast.success("Spend policy reset")
    } catch (error) {
      toast.error(
        extractErrorMessage(error, "Failed to reset the spend policy")
      )
    } finally {
      setMutating(false)
    }
  }

  // With no rule at all the agent cannot pay anything: matching is deny-by-default.
  const cannotSpend = rules.length === 0
  const unrestricted = rules.some(isOpenToEverything)

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>Spending policy</CardTitle>
            <CardDescription>
              What the agent may pay for. Every x402 quote it is asked to sign
              must match a rule on all four facets, and is capped by the
              smallest matching rule.
            </CardDescription>
          </div>
          <Badge
            variant={
              unrestricted
                ? "destructive"
                : cannotSpend
                  ? "secondary"
                  : "default"
            }
          >
            {unrestricted
              ? "unrestricted"
              : cannotSpend
                ? "cannot spend"
                : "restricted"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {loading || !policy ? (
          <div className="space-y-2">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : (
          <>
            {unrestricted ? (
              <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm">
                <AlertTriangle className="mt-0.5 size-4 shrink-0 text-destructive" />
                <p className="text-muted-foreground">
                  A seeded rule allows any amount, to any recipient, on any
                  network. Narrow or remove it before the agent holds funds
                  worth taking.
                </p>
              </div>
            ) : null}
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm text-muted-foreground">
                A quote no rule matches is refused, so removing every rule stops
                the agent paying entirely.
              </p>
              <PaymentSpendRuleDialog onSave={onSaveRule()} />
            </div>
            {rules.length === 0 ? (
              <EmptyState
                icon={Wallet}
                title="No spend rules"
                description="The agent cannot sign any payment until a rule allows one."
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Cap (atomic units)</TableHead>
                    <TableHead>Budget per period</TableHead>
                    <TableHead>Networks</TableHead>
                    <TableHead>Assets</TableHead>
                    <TableHead>Recipients</TableHead>
                    <TableHead>Transfer methods</TableHead>
                    <TableHead className="w-20" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((rule) => (
                    <TableRow key={rule.id ?? rule.name}>
                      <TableCell className="font-medium">{rule.name}</TableCell>
                      <TableCell>
                        {isUnbounded(rule.maxAmountPerPayment) ? (
                          <Badge variant="destructive">unbounded</Badge>
                        ) : rule.maxAmountPerPayment === "0" ? (
                          <Badge variant="secondary">denies</Badge>
                        ) : (
                          <span
                            className="font-mono text-xs"
                            title={rule.maxAmountPerPayment}
                          >
                            {rule.maxAmountPerPayment.length > 12
                              ? `${rule.maxAmountPerPayment.slice(0, 12)}…`
                              : rule.maxAmountPerPayment}
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        {rule.windowSeconds > 0 ? (
                          <span
                            className="font-mono text-xs"
                            title={`${rule.maxAmountPerWindow} per ${rule.windowSeconds}s`}
                          >
                            {rule.maxAmountPerWindow} /{" "}
                            {formatPeriod(rule.windowSeconds)}
                          </span>
                        ) : (
                          <span className="text-xs text-muted-foreground">
                            per payment only
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <PatternsCell patterns={rule.networkPatterns} />
                      </TableCell>
                      <TableCell>
                        <PatternsCell patterns={rule.assetPatterns} />
                      </TableCell>
                      <TableCell>
                        <PatternsCell patterns={rule.payToPatterns} />
                      </TableCell>
                      <TableCell>
                        <PatternsCell
                          patterns={rule.assetTransferMethodPatterns}
                        />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center justify-end gap-1">
                          <PaymentSpendRuleDialog
                            initialRule={rule}
                            onSave={onSaveRule(rule)}
                          />
                          <ConfirmDialog
                            title="Remove this spend rule?"
                            description="The agent will stop being able to pay quotes only this rule allowed."
                            confirmLabel="Remove"
                            onConfirm={() => onDeleteRule(rule)}
                          >
                            <Button
                              variant="ghost"
                              size="icon-sm"
                              aria-label="Remove rule"
                              disabled={mutating}
                            >
                              <Trash2 />
                            </Button>
                          </ConfirmDialog>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </>
        )}
      </CardContent>
      {policy ? (
        <CardFooter className="justify-end">
          <ConfirmDialog
            title="Reset to the unrestricted default?"
            description="Every spend rule is removed, and the agent is seeded again with a rule allowing any amount, to any recipient, on any network."
            confirmLabel="Reset"
            onConfirm={onResetPolicy}
          >
            <Button variant="destructive" size="sm" disabled={mutating}>
              <Trash2 />
              Reset to unrestricted
            </Button>
          </ConfirmDialog>
        </CardFooter>
      ) : null}
    </Card>
  )
}

// PatternsCell shows the open marker, or up to MAX_PATTERN_CHIPS patterns with
// the rest collapsed into a "+N" chip.
function PatternsCell({ patterns }: { patterns: string[] }) {
  if (patterns.includes(ANY_PATTERN)) {
    return <Badge>any</Badge>
  }

  // An empty facet matches nothing, so the rule can never authorize a quote.
  if (patterns.length === 0) {
    return <Badge variant="destructive">none — never matches</Badge>
  }

  const visible = patterns.slice(0, MAX_PATTERN_CHIPS)
  const hidden = patterns.slice(MAX_PATTERN_CHIPS)

  return (
    <div className="flex flex-wrap items-center gap-1">
      {visible.map((pattern, index) => (
        <Badge
          key={`${pattern}-${index}`}
          variant="outline"
          className="font-mono text-xs"
          title={pattern}
        >
          {isPlainAddressPattern(pattern)
            ? truncateEthAddress(pattern)
            : pattern.length > 16
              ? `${pattern.slice(0, 16)}…`
              : pattern}
        </Badge>
      ))}
      {hidden.length > 0 ? (
        <Badge variant="outline" title={hidden.join("\n")}>
          +{hidden.length}
        </Badge>
      ) : null}
    </div>
  )
}
