import React from "react"
import { toast } from "sonner"
import axios from "axios"
import { AlertTriangle, Plus, Trash2, Wallet, X } from "lucide-react"

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
import { Input } from "@/components/ui/input"
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
import {
  ANY_TRANSFER_METHOD,
  type PaymentSpendRuleFormField,
  type PaymentSpendRuleFormSchemaType,
} from "@/components/dashboard/payments/forms/payment-spend-rule-form/payment-spend-rule-form-schema"
import {
  ANY_PATTERN,
  isUnbounded,
  isUnrestricted,
  type PaymentSpendPolicy,
  type PaymentSpendRule,
  type PaymentSpendRuleWindow,
} from "@/models/payment-spend-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  deletePaymentSpendPolicy,
  getPaymentSpendPolicy,
  updatePaymentSpendPolicy,
} from "@/queries/api/payment-spend-policy.queries"
import {
  createPaymentSpendRule,
  deletePaymentSpendRule,
  listPaymentSpendRules,
  updatePaymentSpendRule,
} from "@/queries/api/payment-spend-rule.queries"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"
import { formatNetwork, isPlainAddressPattern } from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

// formatResetsIn is coarse on purpose: the exact second the period rolls is not actionable.
const formatResetsIn = (resetsAt: string): string => {
  const seconds = Math.max(
    0,
    (new Date(resetsAt).getTime() - Date.now()) / 1000
  )
  if (seconds >= 86400) {
    return `${Math.round(seconds / 86400)}d`
  }
  if (seconds >= 3600) {
    return `${Math.round(seconds / 3600)}h`
  }
  return `${Math.max(1, Math.round(seconds / 60))}m`
}

// spentPercent stays in BigInt until the last step; atomic amounts overflow Number.
const spentPercent = (spent: string, budget: string): number => {
  try {
    const total = BigInt(budget)
    if (total === 0n) {
      return 100
    }
    return Math.min(100, Number((BigInt(spent) * 100n) / total))
  } catch {
    return 0
  }
}

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
      return "30 days"
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
  "windows",
  "network",
  "asset",
  "assetTransferMethod",
]

// isAssetTaken reports the runtime's refusal of a second rule on one token.
const isAssetTaken = (error: unknown): boolean =>
  axios.isAxiosError(error) && error.response?.status === 409

// toFormField maps a runtime error path ("config.network", "windows[0].maxAmount") to a form control.
const toFormField = (path: string): PaymentSpendRuleFormField | null => {
  const field = path.replace(/^config\./, "").replace(/\[(\d+)\]/g, ".$1")
  const root = field.split(".")[0] as keyof PaymentSpendRuleFormSchemaType
  return FORM_FIELDS.includes(root)
    ? (field as PaymentSpendRuleFormField)
    : null
}

// normalizePayee lowercases a plain address, as the runtime stores it, so duplicates are caught.
const normalizePayee = (pattern: string): string =>
  isPlainAddressPattern(pattern) ? pattern.toLowerCase() : pattern

// AmountLabel shows an atomic amount the way the cap column does: badges for the sentinels, else truncated.
function AmountLabel({ amount }: { amount: string }) {
  if (isUnbounded(amount)) {
    return <Badge variant="destructive">unbounded</Badge>
  }
  return (
    <span className="font-mono text-xs" title={amount}>
      {amount.length > 12 ? `${amount.slice(0, 12)}…` : amount}
    </span>
  )
}

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
    maxAmountPerPayment:
      values.maxAmountPerPayment.length === 0
        ? undefined
        : values.maxAmountPerPayment,
    windows: values.windows.map((window) => ({
      windowSeconds: Number(window.windowSeconds),
      maxAmount: window.maxAmount,
    })),
    network: values.network,
    asset: values.asset,
    assetTransferMethod:
      values.assetTransferMethod === ANY_TRANSFER_METHOD
        ? undefined
        : values.assetTransferMethod,
  })

  const onSaveRule =
    (existing?: PaymentSpendRule) =>
    async (
      values: PaymentSpendRuleFormSchemaType
    ): Promise<FormCompletionResult<PaymentSpendRuleFormField>> => {
      setMutating(true)
      try {
        const rule = toRule(values)
        if (existing?.id) {
          await updatePaymentSpendRule(existing.id, rule)
        } else {
          await createPaymentSpendRule(rule)
        }
        toast.success(existing ? "Spend rule updated" : "Spend rule added")
        // The rule is saved even if the refresh fails; reporting it as a save failure invites a 409 on retry.
        try {
          await reloadRules()
        } catch (error) {
          toast.error(extractErrorMessage(error, "Failed to refresh the rules"))
        }
        return {}
      } catch (error) {
        if (isAssetTaken(error)) {
          return {
            error: {
              field: "asset",
              message: "This token already has a rule — edit that one instead",
            },
          }
        }
        const validation = extractValidationError(error)
        if (validation) {
          const field = toFormField(validation.field)
          if (field) {
            return { error: { field, message: validation.message } }
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

  // savePayees returns the runtime's refusal so the editor can keep the draft and show why.
  const savePayees = async (
    payToPatterns: string[]
  ): Promise<string | null> => {
    setMutating(true)
    try {
      const { data } = await updatePaymentSpendPolicy(payToPatterns)
      setPolicy(data)
      return null
    } catch (error) {
      return (
        extractValidationError(error)?.message ??
        extractErrorMessage(error, "Failed to save the payees")
      )
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

  // With no payee or no rule the agent cannot pay anything: matching is deny-by-default.
  const payees = policy?.payToPatterns ?? []
  const cannotSpend = rules.length === 0 || payees.length === 0
  const unrestricted =
    payees.includes(ANY_PATTERN) && rules.some(isUnrestricted)

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>Spending policy</CardTitle>
            <CardDescription>
              What the agent may pay for. Every x402 quote it is asked to sign
              must go to an allowed payee, in a token that has a rule, and fit
              that rule's cap and every one of its windows.
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
                  Anyone can be paid, and a rule allows any amount of its token.
                  Narrow it before the agent holds funds worth taking.
                </p>
              </div>
            ) : null}
            <PayeesEditor
              payees={payees}
              disabled={mutating}
              onChange={savePayees}
            />
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm text-muted-foreground">
                One rule per token. A token with no rule is refused, so removing
                every rule stops the agent paying entirely.
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
                    <TableHead>Network</TableHead>
                    <TableHead>Token</TableHead>
                    <TableHead>Transfer method</TableHead>
                    <TableHead>Cap per payment</TableHead>
                    <TableHead>Spent this period</TableHead>
                    <TableHead className="w-20" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((rule) => (
                    <TableRow key={rule.id ?? rule.name}>
                      <TableCell className="font-medium">{rule.name}</TableCell>
                      <TableCell className="text-xs" title={rule.network}>
                        {formatNetwork(rule.network)}
                      </TableCell>
                      <TableCell
                        className="font-mono text-xs"
                        title={rule.asset}
                      >
                        {truncateEthAddress(rule.asset)}
                      </TableCell>
                      <TableCell className="text-xs">
                        {rule.assetTransferMethod ?? (
                          <span className="text-muted-foreground">either</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {rule.maxAmountPerPayment === undefined ? (
                          <span className="text-xs text-muted-foreground">
                            none
                          </span>
                        ) : rule.maxAmountPerPayment === "0" ? (
                          <Badge variant="secondary">denies</Badge>
                        ) : (
                          <AmountLabel amount={rule.maxAmountPerPayment} />
                        )}
                      </TableCell>
                      <TableCell>
                        {rule.windows.length === 0 ? (
                          <span className="text-xs text-muted-foreground">
                            per payment only
                          </span>
                        ) : (
                          <div className="min-w-36 space-y-2">
                            {rule.windows.map((window) => (
                              <WindowUsage
                                key={window.windowSeconds}
                                window={window}
                              />
                            ))}
                          </div>
                        )}
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
            title="Reset the spending policy?"
            description="Every rule and payee is removed; the agent will not be able to pay anything until both are set again."
            confirmLabel="Reset"
            onConfirm={onResetPolicy}
          >
            <Button variant="destructive" size="sm" disabled={mutating}>
              <Trash2 />
              Reset policy
            </Button>
          </ConfirmDialog>
        </CardFooter>
      ) : null}
    </Card>
  )
}

// WindowUsage shows one window's spend against its budget and when the period rolls over.
function WindowUsage({ window }: { window: PaymentSpendRuleWindow }) {
  const spent = window.spent ?? "0"
  return (
    <div className="space-y-1">
      <div
        className="flex flex-wrap items-center gap-1 text-xs"
        title={`${spent} of ${window.maxAmount} atomic units this period`}
      >
        <AmountLabel amount={spent} />
        <span>/</span>
        {window.maxAmount === "0" ? (
          <Badge variant="secondary">denies</Badge>
        ) : (
          <AmountLabel amount={window.maxAmount} />
        )}
        <span className="text-muted-foreground">
          per {formatPeriod(window.windowSeconds)}
        </span>
      </div>
      <div className="h-1 w-full overflow-hidden rounded-full bg-muted">
        <div
          className="h-full bg-primary"
          style={{ width: `${spentPercent(spent, window.maxAmount)}%` }}
        />
      </div>
      <div className="text-xs text-muted-foreground">
        {window.resetsAt
          ? `resets in ${formatResetsIn(window.resetsAt)}`
          : "nothing spent yet"}
      </div>
    </div>
  )
}

// PayeesEditor edits the policy's payee list; every change is saved at once.
function PayeesEditor({
  payees,
  disabled,
  onChange,
}: {
  payees: string[]
  disabled: boolean
  onChange: (payees: string[]) => Promise<string | null>
}) {
  const [draft, setDraft] = React.useState("")
  const [error, setError] = React.useState<string | null>(null)

  // The runtime judges the regex (Go RE2), so a refused pattern keeps the draft and shows its message.
  const save = async (next: string[]): Promise<boolean> => {
    const refusal = await onChange(next)
    setError(refusal)
    return refusal === null
  }

  const add = async () => {
    const pattern = normalizePayee(draft.trim())
    // Enter bypasses the button's disabled state, so a save in flight is checked here.
    if (pattern.length === 0 || disabled) {
      return
    }
    if (payees.includes(pattern) || (await save([...payees, pattern]))) {
      setError(null)
      setDraft("")
    }
  }

  return (
    <div className="space-y-2 rounded-md border p-3">
      <div className="flex items-center justify-between gap-2">
        <div>
          <p className="text-sm font-medium">Payees</p>
          <p className="text-xs text-muted-foreground">
            Who the agent may pay, whatever the token: an address in any case,
            or a regex written in lowercase.
          </p>
        </div>
        {payees.includes(ANY_PATTERN) ? null : (
          <Button
            variant="outline"
            size="sm"
            disabled={disabled}
            onClick={() => void save([ANY_PATTERN])}
          >
            Allow anyone (.*)
          </Button>
        )}
      </div>
      {payees.length === 0 ? (
        <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 p-2 text-sm">
          <AlertTriangle className="mt-0.5 size-4 shrink-0 text-destructive" />
          <p className="text-muted-foreground">
            Nobody can be paid yet: add a payee for the rules to take effect.
          </p>
        </div>
      ) : (
        <div className="flex flex-wrap items-center gap-1">
          {payees.map((pattern, index) => (
            <Badge
              key={`${index}:${pattern}`}
              variant={pattern === ANY_PATTERN ? "default" : "outline"}
              className="gap-1 font-mono text-xs"
              title={pattern}
            >
              {pattern === ANY_PATTERN
                ? "anyone"
                : isPlainAddressPattern(pattern)
                  ? truncateEthAddress(pattern)
                  : pattern}
              <button
                type="button"
                aria-label={`Remove ${pattern}`}
                disabled={disabled}
                onClick={() =>
                  void save(payees.filter((_, other) => other !== index))
                }
              >
                <X className="size-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
      <div className="flex items-center gap-2">
        <Input
          className="font-mono text-xs"
          placeholder="0x… or a regex"
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault()
              void add()
            }
          }}
        />
        <Button
          variant="outline"
          size="sm"
          disabled={disabled || draft.trim().length === 0}
          onClick={() => void add()}
        >
          <Plus />
          Add
        </Button>
      </div>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
    </div>
  )
}
