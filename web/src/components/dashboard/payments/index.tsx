import React from "react"
import axios from "axios"
import { toast } from "sonner"
import { AlertTriangle, Coins, Globe, Trash2 } from "lucide-react"

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
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import ConfirmDialog from "@/components/dashboard/confirm-dialog"
import EmptyState from "@/components/dashboard/empty-state"
import NotMintedState from "@/components/dashboard/not-minted-state"
import PaymentReceiveRuleDialog from "@/components/dashboard/payments/dialogs/payment-receive-rule-dialog"
import type {
  PaymentReceiveRuleFormField,
  PaymentReceiveRuleFormSchemaType,
} from "@/components/dashboard/payments/forms/payment-receive-rule-form/payment-receive-rule-form-schema"
import {
  EVERYONE_PATTERN,
  isPublicPrice,
  type PaymentReceivePolicy,
  type PaymentReceiveRule,
} from "@/models/payment-receive-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  deletePaymentReceivePolicy,
  getPaymentReceivePolicy,
} from "@/queries/api/payment-receive-policy.queries"
import {
  createPaymentReceiveRule,
  deletePaymentReceiveRule,
  listPaymentReceiveRules,
  updatePaymentReceiveRule,
} from "@/queries/api/payment-receive-rule.queries"
import { getFacilitatorNetworks } from "@/configs/app-config"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"
import {
  formatNetwork,
  formatPrice,
  isPlainAddressPattern,
} from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

const MAX_CALLER_CHIPS = 3

// The fields the rule form actually renders; runtime errors on any other
// (defaulted) config field are toasted instead of pinned to a control.
const FORM_FIELDS: Array<keyof PaymentReceiveRuleFormSchemaType> = [
  "name",
  "method",
  "network",
  "asset",
  "payTo",
  "prices",
]

// toFormField maps a runtime error path ("config.asset", "prices[0].price") to a form control.
const toFormField = (path: string): PaymentReceiveRuleFormField | null => {
  const field = path.replace(/^config\./, "").replace(/\[(\d+)\]/g, ".$1")
  const root = field.split(".")[0] as keyof PaymentReceiveRuleFormSchemaType
  return FORM_FIELDS.includes(root)
    ? (field as PaymentReceiveRuleFormField)
    : null
}

// isAssetTaken reports the runtime's refusal of a second rule on one token.
const isAssetTaken = (error: unknown): boolean =>
  axios.isAxiosError(error) &&
  error.response?.status === 409 &&
  (error.response.data as { code?: string } | undefined)?.code ===
    "asset_already_has_rule"

const isNotMintedError = (error: unknown) =>
  axios.isAxiosError(error) &&
  error.response?.status === 409 &&
  (error.response.data as { code?: string } | undefined)?.code ===
    "agent_not_minted"

export default function PaymentReceivePolicySection() {
  const { loading: identityLoading, isMinted } = useAgentIdentity()
  const [policy, setPolicy] = React.useState<PaymentReceivePolicy | null>(null)
  const [rules, setRules] = React.useState<PaymentReceiveRule[]>([])
  const [loading, setLoading] = React.useState(true)
  const [mutating, setMutating] = React.useState(false)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      // Sequential on purpose: both calls lazily materialize the policy server-side.
      const policyResponse = await getPaymentReceivePolicy()
      const rulesResponse = await listPaymentReceiveRules()
      setPolicy(policyResponse.data)
      setRules(rulesResponse.rules)
    } catch (error) {
      // Pre-mint the runtime refuses the policy; the not-minted state already covers it.
      if (!isNotMintedError(error)) {
        toast.error(extractErrorMessage(error, "Failed to load payment policy"))
      }
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    if (isMinted) void load()
  }, [load, isMinted])

  const reloadRules = async () => {
    const { rules } = await listPaymentReceiveRules()
    setRules(rules)
  }

  const toRule = (
    values: PaymentReceiveRuleFormSchemaType
  ): PaymentReceiveRule => ({
    name: values.name,
    method: values.method,
    network: values.network,
    asset: values.asset,
    payTo: values.payTo,
    prices: values.prices,
  })

  const onSaveRule =
    (existing?: PaymentReceiveRule) =>
    async (
      values: PaymentReceiveRuleFormSchemaType
    ): Promise<FormCompletionResult<PaymentReceiveRuleFormField>> => {
      setMutating(true)
      try {
        const rule = toRule(values)
        if (existing?.id) {
          await updatePaymentReceiveRule(existing.id, rule)
        } else {
          await createPaymentReceiveRule(rule)
        }
        await reloadRules()
        toast.success(existing ? "Pricing rule updated" : "Pricing rule added")
        return {}
      } catch (error) {
        if (isAssetTaken(error)) {
          return {
            error: {
              field: "asset",
              message:
                "This token already has a pricing rule — edit that rule to add prices",
            },
          }
        }
        const validation = extractValidationError(error)
        if (validation) {
          const field = toFormField(validation.field)
          if (field) {
            return { error: { field, message: validation.message } }
          }
          // A config field with no form control (e.g. a defaulted one): toast
          // it rather than silently swallow the error.
          toast.error(validation.message)
          return {}
        }
        toast.error(extractErrorMessage(error, "Failed to save pricing rule"))
        return {}
      } finally {
        setMutating(false)
      }
    }

  const onRemoveRule = async (rule: PaymentReceiveRule) => {
    if (!rule.id) return
    setMutating(true)
    try {
      await deletePaymentReceiveRule(rule.id)
      await reloadRules()
      toast.success("Pricing rule removed")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to remove pricing rule"))
    } finally {
      setMutating(false)
    }
  }

  const onReset = async () => {
    try {
      await deletePaymentReceivePolicy()
      await load()
      toast.success("Payment policy reset")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to reset payment policy"))
    }
  }

  // Paid prices without a facilitator are never offered; free ones still work.
  const facilitatorNetworks = getFacilitatorNetworks()
  const hasFacilitator = (rule: PaymentReceiveRule) =>
    facilitatorNetworks.includes(rule.network)
  const hasUnsettleablePrice = (rule: PaymentReceiveRule) =>
    !hasFacilitator(rule) && rule.prices.some((row) => row.price !== "0")
  const acceptingPayments = rules.some(
    (rule) =>
      hasFacilitator(rule) &&
      rule.prices.some((row) => isPublicPrice(row) && row.price !== "0")
  )

  const notMinted = !identityLoading && !isMinted

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>Payment policy</CardTitle>
            <CardDescription>
              Every /responses call is priced by this policy: callers pay the
              cheapest price matching their address (via x402), and are refused
              when no price matches.
            </CardDescription>
          </div>
          <Badge variant={acceptingPayments ? "default" : "secondary"}>
            {acceptingPayments ? "accepting payments" : "private access only"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {notMinted ? (
          <NotMintedState description="The payment policy is created once the agent's on-chain identity is minted." />
        ) : loading || !policy ? (
          <div className="space-y-2">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm text-muted-foreground">
                One rule per token. When several of its prices match a caller,
                the cheapest wins — a free or discounted price always beats the
                public one for its callers. Prices open to everyone are
                advertised to anonymous callers; scoped ones are private offers.
              </p>
              <PaymentReceiveRuleDialog
                agentWalletAddress={policy.agentWalletAddress}
                onSave={onSaveRule()}
              />
            </div>
            {rules.length === 0 ? (
              <EmptyState
                icon={Coins}
                title="No pricing rules"
                description="Nobody can call this agent. Add a free rule for your wallet, or a public price for everyone."
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Rule</TableHead>
                    <TableHead>Prices per call</TableHead>
                    <TableHead>Network</TableHead>
                    <TableHead>Paid to</TableHead>
                    <TableHead className="w-20" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((rule) => (
                    <TableRow key={rule.id}>
                      <TableCell className="font-medium">{rule.name}</TableCell>
                      <TableCell>
                        <PricesCell
                          rule={rule}
                          settleable={hasFacilitator(rule)}
                        />
                      </TableCell>
                      <TableCell>
                        {hasUnsettleablePrice(rule) ? (
                          <Badge
                            variant="destructive"
                            title="No facilitator configured for this network — its paid prices can't be settled and aren't offered to callers; free prices still work."
                          >
                            <AlertTriangle />
                            {formatNetwork(rule.network)}
                          </Badge>
                        ) : (
                          <span title={rule.network}>
                            {formatNetwork(rule.network)}
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        {rule.payTo.toLowerCase() ===
                        policy.agentWalletAddress.toLowerCase() ? (
                          <Badge variant="outline" title={rule.payTo}>
                            agent wallet
                          </Badge>
                        ) : (
                          <span className="font-mono" title={rule.payTo}>
                            {truncateEthAddress(rule.payTo)}
                          </span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <PaymentReceiveRuleDialog
                          agentWalletAddress={policy.agentWalletAddress}
                          initialRule={rule}
                          onSave={onSaveRule(rule)}
                        />
                        <ConfirmDialog
                          title="Remove this pricing rule?"
                          description="Callers lose every price of this token — they fall back to other tokens or get refused."
                          confirmLabel="Remove"
                          onConfirm={() => onRemoveRule(rule)}
                        >
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label="Remove pricing rule"
                            disabled={mutating}
                          >
                            <Trash2 />
                          </Button>
                        </ConfirmDialog>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </>
        )}
      </CardContent>
      {notMinted ? null : (
        <CardFooter>
          <ConfirmDialog
            title="Reset the payment policy?"
            description="The policy goes back to the default: a single free price for the agent wallet and owner."
            confirmLabel="Reset"
            onConfirm={onReset}
          >
            <Button type="button" variant="destructive" size="sm">
              <Trash2 />
              Reset
            </Button>
          </ConfirmDialog>
        </CardFooter>
      )}
    </Card>
  )
}

// PricesCell lists a rule's prices, cheapest first, each with the callers who pay it.
function PricesCell({
  rule,
  settleable,
}: {
  rule: PaymentReceiveRule
  settleable: boolean
}) {
  const rows = [...rule.prices].sort((a, b) => {
    const left = BigInt(a.price)
    const right = BigInt(b.price)
    return left < right ? -1 : left > right ? 1 : 0
  })
  return (
    <div className="space-y-1">
      {rows.map((row) => (
        <div key={row.price} className="flex flex-wrap items-center gap-2">
          {row.price === "0" ? (
            <Badge variant="secondary">free</Badge>
          ) : settleable ? (
            <span
              className="font-medium"
              title={`${row.price} atomic units of ${rule.asset}`}
            >
              {formatPrice(row.price, rule.asset)}
            </span>
          ) : (
            <Badge
              variant="destructive"
              title={`${row.price} atomic units of ${rule.asset} — no facilitator on this network, so this price can't be settled and isn't offered to callers.`}
            >
              <AlertTriangle />
              {formatPrice(row.price, rule.asset)}
            </Badge>
          )}
          <CallersCell patterns={row.addressPatterns} />
        </div>
      ))}
    </div>
  )
}

// CallersCell renders "public" or MAX_CALLER_CHIPS chips plus a "+N" overflow chip.
function CallersCell({ patterns }: { patterns: string[] }) {
  if (patterns.includes(EVERYONE_PATTERN)) {
    return (
      <Badge>
        <Globe />
        public
      </Badge>
    )
  }

  const visible = patterns.slice(0, MAX_CALLER_CHIPS)
  const hidden = patterns.slice(MAX_CALLER_CHIPS)

  return (
    <div className="flex flex-wrap items-center gap-1">
      {visible.map((pattern, index) => (
        <Badge
          key={`${pattern}-${index}`}
          variant="outline"
          className="font-mono"
          title={pattern}
        >
          {isPlainAddressPattern(pattern)
            ? truncateEthAddress(pattern)
            : pattern.length > 14
              ? `${pattern.slice(0, 14)}…`
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
