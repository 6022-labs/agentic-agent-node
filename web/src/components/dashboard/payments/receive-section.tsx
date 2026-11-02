import React from "react"
import axios from "axios"
import { toast } from "sonner"
import { AlertTriangle, Coins, Trash2 } from "lucide-react"

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
import TokenCell from "@/components/dashboard/payments/token-cell"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  AudienceBadge,
  DocsLink,
} from "@/components/dashboard/payments/address-chips"
import type {
  PaymentReceiveRuleFormField,
  PaymentReceiveRuleFormSchemaType,
} from "@/components/dashboard/payments/forms/payment-receive-rule-form/payment-receive-rule-form-schema"
import {
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
import { audienceLabel, formatPrice } from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

// MAX_PRICE_LINES keeps a row short; further prices fold into a "+N more prices" badge.
const MAX_PRICE_LINES = 2

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
    <Card className="lg:h-full">
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>Receive policy</CardTitle>
            <CardDescription>
              What callers pay the agent per call. <DocsLink />
            </CardDescription>
          </div>
          <Badge variant={acceptingPayments ? "default" : "secondary"}>
            {acceptingPayments ? "accepting payments" : "private access only"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="flex-1 space-y-3">
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
            <div className="flex items-center justify-end gap-2">
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
                    <TableHead>Token</TableHead>
                    <TableHead>Prices per call</TableHead>
                    <TableHead>Paid to</TableHead>
                    <TableHead className="w-20" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((rule) => (
                    <TableRow key={rule.id}>
                      <TableCell className="font-medium">
                        {/* A td ignores max-width in auto table layout, so the inner block truncates. */}
                        <div className="max-w-40 truncate" title={rule.name}>
                          {rule.name}
                        </div>
                      </TableCell>
                      <TableCell className="max-w-40">
                        <TokenCell
                          network={rule.network}
                          asset={rule.asset}
                          badge={
                            hasUnsettleablePrice(rule) ? (
                              <Badge
                                variant="destructive"
                                title="No facilitator configured for this network — its paid prices can't be settled and aren't offered to callers; free prices still work."
                              >
                                <AlertTriangle />
                              </Badge>
                            ) : null
                          }
                        />
                      </TableCell>
                      <TableCell>
                        <PricesCell
                          rule={rule}
                          settleable={hasFacilitator(rule)}
                        />
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
        <CardFooter className="justify-end">
          <ConfirmDialog
            title="Reset the receive policy?"
            description="The policy goes back to the default: a single free price for the agent wallet and owner."
            confirmLabel="Reset"
            onConfirm={onReset}
          >
            <Button type="button" variant="destructive" size="sm">
              <Trash2 />
              Reset policy
            </Button>
          </ConfirmDialog>
        </CardFooter>
      )}
    </Card>
  )
}

// PricesCell lists a rule's cheapest prices, each with one badge for who pays it.
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
  const visible = rows.slice(0, MAX_PRICE_LINES)
  const hidden = rows.slice(MAX_PRICE_LINES)
  return (
    <div className="space-y-1">
      {visible.map((row) => (
        <div key={row.price} className="flex min-w-0 items-center gap-2">
          <PriceLabel
            price={row.price}
            asset={rule.asset}
            settleable={settleable}
          />
          <AudienceBadge patterns={row.addressPatterns} anyoneLabel="public" />
        </div>
      ))}
      {hidden.length > 0 ? (
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              aria-label={`Show ${hidden.length} more prices`}
            >
              <Badge variant="secondary">+{hidden.length} more prices</Badge>
            </button>
          </TooltipTrigger>
          <TooltipContent className="block max-w-sm p-2">
            <ul className="max-h-60 space-y-0.5 overflow-y-auto text-xs select-text">
              {hidden.map((row) => (
                <li key={row.price}>
                  {row.price === "0"
                    ? "free"
                    : formatPrice(row.price, rule.asset)}
                  {" — "}
                  {audienceLabel(row.addressPatterns, "public")}
                </li>
              ))}
            </ul>
          </TooltipContent>
        </Tooltip>
      ) : null}
    </div>
  )
}

function PriceLabel({
  price,
  asset,
  settleable,
}: {
  price: string
  asset: string
  settleable: boolean
}) {
  if (price === "0") {
    return <Badge variant="secondary">free</Badge>
  }
  if (settleable) {
    return (
      <span className="font-medium" title={`${price} atomic units of ${asset}`}>
        {formatPrice(price, asset)}
      </span>
    )
  }
  return (
    <Badge
      variant="destructive"
      title={`${price} atomic units of ${asset} — no facilitator on this network, so this price can't be settled and isn't offered to callers.`}
    >
      <AlertTriangle />
      {formatPrice(price, asset)}
    </Badge>
  )
}
