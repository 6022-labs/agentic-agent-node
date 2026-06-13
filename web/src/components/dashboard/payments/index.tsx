import React from "react"
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
import ConfirmDialog from "@/components/dashboard/confirm-dialog"
import EmptyState from "@/components/dashboard/empty-state"
import PaymentRuleDialog from "@/components/dashboard/payments/dialogs/payment-rule-dialog"
import type { PaymentRuleFormSchemaType } from "@/components/dashboard/payments/forms/payment-rule-form/payment-rule-form-schema"
import {
  EVERYONE_PATTERN,
  type PaymentPolicy,
  type PaymentRule,
} from "@/models/payment-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  deletePaymentPolicy,
  getPaymentPolicy,
} from "@/queries/api/payment-policy.queries"
import {
  createPaymentRule,
  deletePaymentRule,
  listPaymentRules,
  updatePaymentRule,
} from "@/queries/api/payment-rule.queries"
import { getAppConfig } from "@/configs/app-config"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"
import {
  formatNetwork,
  formatPrice,
  isPlainAddressPattern,
} from "@/lib/token-format"
import { truncateEthAddress } from "@/lib/utils"

const MAX_CALLER_CHIPS = 3

export default function PaymentPolicySection() {
  const [policy, setPolicy] = React.useState<PaymentPolicy | null>(null)
  const [rules, setRules] = React.useState<PaymentRule[]>([])
  const [loading, setLoading] = React.useState(true)
  const [mutating, setMutating] = React.useState(false)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const [policyResponse, rulesResponse] = await Promise.all([
        getPaymentPolicy(),
        listPaymentRules(),
      ])
      setPolicy(policyResponse.data)
      setRules(rulesResponse.data.rules ?? [])
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to load payment policy"))
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void load()
  }, [load])

  const reloadRules = async () => {
    const { data } = await listPaymentRules()
    setRules(data.rules ?? [])
  }

  // The form deals in plain addresses; custom regex patterns set via the API
  // on an edited rule are preserved verbatim.
  const toRule = (
    values: PaymentRuleFormSchemaType,
    existing?: PaymentRule
  ): PaymentRule => ({
    name: values.name,
    network: values.network,
    asset: values.asset,
    payTo: values.payTo,
    price: values.price,
    addressPatterns: values.allowEveryone
      ? [EVERYONE_PATTERN]
      : [
          ...values.addresses.map((entry) => entry.address.toLowerCase()),
          ...(existing ? customPatternsOf(existing) : []),
        ],
  })

  const customPatternsOf = (rule: PaymentRule): string[] =>
    (rule.addressPatterns ?? []).filter(
      (pattern) =>
        pattern !== EVERYONE_PATTERN && !isPlainAddressPattern(pattern)
    )

  const onSaveRule =
    (existing?: PaymentRule) =>
    async (
      values: PaymentRuleFormSchemaType
    ): Promise<FormCompletionResult<keyof PaymentRuleFormSchemaType>> => {
      setMutating(true)
      try {
        const rule = toRule(values, existing)
        if (existing?.id) {
          await updatePaymentRule(existing.id, rule)
        } else {
          await createPaymentRule(rule)
        }
        await reloadRules()
        toast.success(existing ? "Pricing rule updated" : "Pricing rule added")
        return {}
      } catch (error) {
        const validation = extractValidationError(error)
        if (validation) {
          // Re-home errors on wire-only fields onto the form field they
          // originate from.
          const field = (
            validation.field === "addressPatterns"
              ? "addresses"
              : validation.field === "facilitatorBaseUrl"
                ? "network"
                : validation.field
          ) as keyof PaymentRuleFormSchemaType
          return { error: { field, message: validation.message } }
        }
        toast.error(extractErrorMessage(error, "Failed to save pricing rule"))
        return {}
      } finally {
        setMutating(false)
      }
    }

  const onRemoveRule = async (rule: PaymentRule) => {
    if (!rule.id) return
    setMutating(true)
    try {
      await deletePaymentRule(rule.id)
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
      await deletePaymentPolicy()
      await load()
      toast.success("Payment policy reset")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to reset payment policy"))
    }
  }

  const isPublic = (rule: PaymentRule) =>
    rule.addressPatterns?.includes(EVERYONE_PATTERN)
  // A paid rule whose network has no configured facilitator can't be settled,
  // so callers can never pay it — flag it.
  const facilitatorNetworks = getAppConfig().facilitatorNetworks
  const isUnsettleable = (rule: PaymentRule) =>
    rule.price !== "0" && !facilitatorNetworks.includes(rule.network)
  const acceptingPayments = rules.some(
    (rule) => isPublic(rule) && rule.price !== "0"
  )

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>Payment policy</CardTitle>
            <CardDescription>
              Every /responses call is priced by this policy: callers pay the
              cheapest rule matching their address (via x402), and are refused
              when no rule matches.
            </CardDescription>
          </div>
          <Badge variant={acceptingPayments ? "default" : "secondary"}>
            {acceptingPayments ? "accepting payments" : "private access only"}
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
            <div className="flex items-center justify-between gap-2">
              <p className="text-sm text-muted-foreground">
                When several rules match the same caller for the same token and
                network, the cheapest price wins — a free or discounted rule
                always beats the public price for its callers. Rules open to
                everyone are advertised to anonymous callers; scoped rules are
                private offers.
              </p>
              <PaymentRuleDialog
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
                    <TableHead>Callers</TableHead>
                    <TableHead>Price per call</TableHead>
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
                        <CallersCell patterns={rule.addressPatterns ?? []} />
                      </TableCell>
                      <TableCell>
                        {rule.price === "0" ? (
                          <Badge variant="secondary">Free</Badge>
                        ) : (
                          <span
                            className="font-medium"
                            title={`${rule.price} atomic units of ${rule.asset}`}
                          >
                            {formatPrice(rule.price, rule.asset)}
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        {isUnsettleable(rule) ? (
                          <Badge
                            variant="destructive"
                            title="No facilitator configured for this network — callers can't pay this rule, and it isn't offered to them."
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
                        <PaymentRuleDialog
                          agentWalletAddress={policy.agentWalletAddress}
                          initialRule={rule}
                          preservedPatternCount={customPatternsOf(rule).length}
                          onSave={onSaveRule(rule)}
                        />
                        <ConfirmDialog
                          title="Remove this pricing rule?"
                          description="The matched callers lose this price — they fall back to other rules or get refused."
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
      <CardFooter>
        <ConfirmDialog
          title="Reset the payment policy?"
          description="The policy goes back to the default: a single free rule for the agent wallet and owner."
          confirmLabel="Reset"
          onConfirm={onReset}
        >
          <Button type="button" variant="destructive" size="sm">
            <Trash2 />
            Reset
          </Button>
        </ConfirmDialog>
      </CardFooter>
    </Card>
  )
}

// CallersCell renders "Everyone" or up to MAX_CALLER_CHIPS address/pattern
// chips, collapsing the rest into a "+N" chip whose tooltip lists them all.
function CallersCell({ patterns }: { patterns: string[] }) {
  if (patterns.includes(EVERYONE_PATTERN)) {
    return (
      <Badge>
        <Globe />
        Everyone
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
