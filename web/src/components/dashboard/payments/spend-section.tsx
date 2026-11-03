import React from "react"
import { toast } from "sonner"
import { AlertTriangle, Plus, RefreshCw, Trash2, Wallet, X } from "lucide-react"

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
import NotMintedState from "@/components/dashboard/not-minted-state"
import EmptyState from "@/components/dashboard/empty-state"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import PaymentSpendRuleDialog from "@/components/dashboard/payments/dialogs/payment-spend-rule-dialog"
import TokenCell from "@/components/dashboard/payments/token-cell"
import {
  isAssetTaken,
  normalizePattern,
  toFormField,
} from "@/components/dashboard/payments/rule-form-errors"
import { DocsLink } from "@/components/dashboard/payments/address-chips"
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
  WINDOW_PRESETS,
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
import {
  formatAtomicAmount,
  formatTokenAmount,
  isPlainAddressPattern,
} from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

import HoverHint from "./hover-hint"
import { useTokenInfo } from "./use-token-info"

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
    return Math.min(100, Number((BigInt(spent) * 100n) / BigInt(budget)))
  } catch {
    return 0
  }
}

// formatPeriod reuses the form's preset names and falls back to raw seconds.
const formatPeriod = (seconds: number): string => {
  const preset = WINDOW_PRESETS.find((p) => p.seconds === seconds)
  if (preset) {
    return preset.label.toLowerCase()
  }
  if (seconds % 86400 === 0) {
    return `${seconds / 86400} days`
  }
  if (seconds % 3600 === 0) {
    return `${seconds / 3600} hours`
  }
  return `${seconds} seconds`
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

// AmountLabel shows an atomic amount in token units once the token resolves, else truncated atomic.
function AmountLabel({
  amount,
  network,
  asset,
  bare,
}: {
  amount: string
  network: string
  asset: string
  // bare drops the symbol, for a spent amount shown right before its budget.
  bare?: boolean
}) {
  const token = useTokenInfo(network, asset)
  if (isUnbounded(amount)) {
    return <Badge variant="destructive">unbounded</Badge>
  }
  if (token) {
    return (
      <span className="font-mono text-xs">
        {bare
          ? formatAtomicAmount(amount, token.decimals)
          : formatTokenAmount(amount, token)}
      </span>
    )
  }
  if (amount.length <= 12) {
    return (
      <span className="font-mono text-xs">{formatTokenAmount(amount)}</span>
    )
  }
  return (
    <HoverHint
      className="font-mono text-xs"
      content={formatTokenAmount(amount)}
    >
      {`${amount.slice(0, 12)}… units`}
    </HoverHint>
  )
}

export default function PaymentSpendPolicySection() {
  const { loading: identityLoading, isMinted } = useAgentIdentity()
  const [policy, setPolicy] = React.useState<PaymentSpendPolicy | null>(null)
  const [rules, setRules] = React.useState<PaymentSpendRule[]>([])
  const [loading, setLoading] = React.useState(true)
  const [loadFailed, setLoadFailed] = React.useState(false)
  const [mutating, setMutating] = React.useState(false)

  // load reports whether it succeeded; on failure it already shows the retry state.
  const load = React.useCallback(async (): Promise<boolean> => {
    setLoading(true)
    setLoadFailed(false)
    try {
      // Sequential on purpose: both calls lazily materialize the policy server-side.
      const policyResponse = await getPaymentSpendPolicy()
      const rulesResponse = await listPaymentSpendRules()
      setPolicy(policyResponse.data)
      setRules(rulesResponse.rules)
      return true
    } catch (error) {
      setLoadFailed(true)
      setPolicy(null)
      toast.error(extractErrorMessage(error, "Failed to load spend policy"))
      return false
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    if (isMinted) void load()
  }, [load, isMinted])

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
    ): Promise<FormCompletionResult<PaymentSpendRuleFormField, "saved">> => {
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
        return { result: "saved" }
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
          const field = toFormField<PaymentSpendRuleFormField>(
            validation.field,
            FORM_FIELDS
          )
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
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to remove the spend rule"))
      setMutating(false)
      return
    }
    toast.success("Spend rule removed")
    // The rule is gone even if the refresh fails; reporting it as a removal failure invites a 404 on retry.
    try {
      await reloadRules()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to refresh the rules"))
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
      toast.success("Spend policy reset")
      await load()
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
    <Card className="lg:h-full lg:min-h-0 lg:overflow-hidden">
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>Spend policy</CardTitle>
            <CardDescription>
              What the agent may pay, and to whom. <DocsLink />
            </CardDescription>
          </div>
          {policy ? (
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
          ) : null}
        </div>
      </CardHeader>
      <CardContent className="space-y-3 lg:flex lg:min-h-0 lg:flex-1 lg:flex-col">
        {!identityLoading && !isMinted ? (
          <NotMintedState description="The spend policy is set once the agent's on-chain identity is minted." />
        ) : loadFailed && !loading ? (
          <EmptyState
            icon={AlertTriangle}
            title="Couldn't load the spend policy"
            action={
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => void load()}
              >
                <RefreshCw />
                Retry
              </Button>
            }
          />
        ) : loading || !policy ? (
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
                  Anyone can be paid any amount of a token. Narrow it first.
                </p>
              </div>
            ) : null}
            <PayeesEditor
              payees={payees}
              disabled={mutating}
              onChange={savePayees}
            />
            <div className="flex items-center justify-end gap-2">
              <PaymentSpendRuleDialog onSave={onSaveRule()} />
            </div>
            {rules.length === 0 ? (
              <EmptyState
                icon={Wallet}
                title="No spend rules"
                description="The agent cannot sign any payment until a rule allows one."
              />
            ) : (
              <div className="max-h-[28rem] overflow-y-auto lg:max-h-none lg:min-h-0 lg:flex-1">
                <Table>
                  <TableHeader className="sticky top-0 z-10 bg-card">
                    <TableRow>
                      <TableHead>Rule</TableHead>
                      <TableHead>Token</TableHead>
                      <TableHead>Cap per payment</TableHead>
                      <TableHead>Spent this period</TableHead>
                      <TableHead className="w-20" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {rules.map((rule) => (
                      <TableRow key={rule.id ?? rule.name}>
                        <TableCell className="font-medium">
                          {/* A td ignores max-width in auto table layout, so the inner block truncates. */}
                          <HoverHint
                            className="block max-w-40 truncate"
                            content={rule.name}
                          >
                            {rule.name}
                          </HoverHint>
                        </TableCell>
                        <TableCell className="max-w-40">
                          <TokenCell
                            network={rule.network}
                            asset={rule.asset}
                            badge={
                              rule.assetTransferMethod ? (
                                <Badge variant="outline" className="text-xs">
                                  {rule.assetTransferMethod}
                                </Badge>
                              ) : null
                            }
                          />
                        </TableCell>
                        <TableCell>
                          {rule.maxAmountPerPayment === undefined ? (
                            <span className="text-xs text-muted-foreground">
                              none
                            </span>
                          ) : rule.maxAmountPerPayment === "0" ? (
                            <Badge variant="secondary">denies</Badge>
                          ) : (
                            <AmountLabel
                              amount={rule.maxAmountPerPayment}
                              network={rule.network}
                              asset={rule.asset}
                            />
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
                                  network={rule.network}
                                  asset={rule.asset}
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
                                aria-label={`Remove rule ${rule.name}`}
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
              </div>
            )}
          </>
        )}
      </CardContent>
      {policy && isMinted ? (
        <CardFooter className="justify-end">
          <ConfirmDialog
            title="Reset the spend policy?"
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
function WindowUsage({
  window,
  network,
  asset,
}: {
  window: PaymentSpendRuleWindow
  network: string
  asset: string
}) {
  const spent = window.spent ?? "0"
  return (
    <div className="flex min-w-0 items-center gap-2 text-xs">
      <span className="flex min-w-0 items-center gap-1">
        <AmountLabel bare amount={spent} network={network} asset={asset} />
        <span>/</span>
        {window.maxAmount === "0" ? (
          <Badge variant="secondary">denies</Badge>
        ) : (
          <AmountLabel
            amount={window.maxAmount}
            network={network}
            asset={asset}
          />
        )}
        <span className="text-muted-foreground">
          · {formatPeriod(window.windowSeconds)}
        </span>
      </span>
      {window.maxAmount === "0" ? null : (
        <div className="h-1 w-12 shrink-0 overflow-hidden rounded-full bg-muted">
          <div
            className="h-full bg-primary"
            style={{ width: `${spentPercent(spent, window.maxAmount)}%` }}
          />
        </div>
      )}
      {window.resetsAt ? (
        <span className="shrink-0 text-muted-foreground">
          resets in {formatResetsIn(window.resetsAt)}
        </span>
      ) : null}
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
    const pattern = normalizePattern(draft.trim())
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
            Addresses in any case, or regexes written in lowercase.
          </p>
        </div>
        {payees.includes(ANY_PATTERN) ? null : payees.length === 0 ? (
          <Button
            variant="outline"
            size="sm"
            disabled={disabled}
            onClick={() => void save([ANY_PATTERN])}
          >
            Allow anyone
          </Button>
        ) : (
          <ConfirmDialog
            title={`Replace ${payees.length} ${payees.length === 1 ? "payee" : "payees"} with anyone?`}
            description="The agent will be able to pay any address its rules allow."
            confirmLabel="Allow anyone"
            onConfirm={() => save([ANY_PATTERN])}
          >
            <Button variant="outline" size="sm" disabled={disabled}>
              Allow anyone
            </Button>
          </ConfirmDialog>
        )}
      </div>
      {payees.length === 0 ? (
        <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 p-2 text-sm">
          <AlertTriangle className="mt-0.5 size-4 shrink-0 text-destructive" />
          <p className="text-muted-foreground">Nobody can be paid yet.</p>
        </div>
      ) : (
        <div className="flex flex-wrap items-center gap-1">
          {payees.map((pattern, index) => (
            <HoverHint
              key={`${index}:${pattern}`}
              className="inline-flex max-w-full min-w-0"
              content={<span className="font-mono break-all">{pattern}</span>}
            >
              <Badge
                variant={pattern === ANY_PATTERN ? "default" : "outline"}
                className="max-w-full gap-1 font-mono text-xs"
              >
                <span className="truncate">
                  {pattern === ANY_PATTERN
                    ? "anyone"
                    : isPlainAddressPattern(pattern)
                      ? truncateEthAddress(pattern)
                      : pattern}
                </span>
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
            </HoverHint>
          ))}
        </div>
      )}
      <div className="flex items-center gap-2">
        <Input
          className="font-mono text-xs"
          aria-label="Payee address or regex"
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
