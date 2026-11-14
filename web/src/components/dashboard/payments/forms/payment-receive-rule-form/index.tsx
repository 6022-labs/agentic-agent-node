import React from "react"
import { useFieldArray, useForm, useWatch } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { CircleCheck, CircleX, Loader2, Plus, Trash2, X } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  EVERYONE_PATTERN,
  MAX_UINT256,
  X402_METHOD,
  type PaymentReceiveRule,
} from "@/models/payment-receive-policy"
import {
  getAppConfig,
  getEvmChainConfig,
  getFacilitatorNetworks,
} from "@/configs/app-config"
import type { FormCompletionResult } from "@/models/api/validation-error"
import type { Erc20Lookup } from "@/lib/erc20"
import { truncateEthAddress } from "@/lib/utils"
import {
  formatAtomicAmount,
  formatNetwork,
  isPlainAddressPattern,
  isZeroAmount,
  parseTokenAmount,
} from "@/lib/token-format"
import { normalizePattern } from "@/components/dashboard/payments/rule-form-errors"
import {
  paymentReceiveRuleFormSchema,
  type PaymentReceiveRuleFormField,
  type PaymentReceiveRuleFormSchemaType,
} from "./payment-receive-rule-form-schema"

type Props = {
  // Pre-fills new rules as paying to the agent's own wallet.
  agentWalletAddress: string
  // When set, the form edits this rule instead of creating one.
  initialRule?: PaymentReceiveRule
  // Token metadata for the entered network + asset, resolved by the caller.
  tokenLookup: Erc20Lookup
  onValueChange: (values: PaymentReceiveRuleFormSchemaType) => void
  onFormComplete: (
    values: PaymentReceiveRuleFormSchemaType
  ) => Promise<FormCompletionResult<PaymentReceiveRuleFormField>>
}

export default function PaymentReceiveRuleForm({
  agentWalletAddress,
  initialRule,
  tokenLookup,
  onValueChange,
  onFormComplete,
}: Props) {
  const facilitatorNetworks = getFacilitatorNetworks()
  // Reachable chains as CAIP-2 networks (the x402 rule shape): free rules may
  // target any reachable chain, paid rules only facilitator networks.
  const reachableNetworks = React.useMemo(
    () => Object.keys(getAppConfig().evmChains).map((id) => `eip155:${id}`),
    []
  )
  const defaultNewRuleNetwork =
    facilitatorNetworks[0] ?? reachableNetworks[0] ?? ""
  // Default the asset to the $6022 token on the default network's chain.
  const defaultToken =
    getEvmChainConfig(defaultNewRuleNetwork.replace("eip155:", ""))
      ?.token6022Address ?? ""

  const form = useForm<PaymentReceiveRuleFormSchemaType>({
    resolver: zodResolver(paymentReceiveRuleFormSchema),
    defaultValues: initialRule
      ? {
          name: initialRule.name,
          method: initialRule.method || X402_METHOD,
          network: initialRule.network,
          asset: initialRule.asset,
          payTo: initialRule.payTo,
          // Non-zero prices are atomic on the wire; they are converted to
          // token units below once the token's decimals are known.
          prices: initialRule.prices.map((row) => ({
            price: row.price === "0" ? "0" : "",
            addressPatterns: row.addressPatterns,
          })),
        }
      : // New rules default to a facilitator network and the $6022 token.
        {
          name: "",
          method: X402_METHOD,
          network: defaultNewRuleNetwork,
          asset: defaultToken,
          payTo: agentWalletAddress,
          prices: [{ price: "0.01", addressPatterns: [] }],
        },
  })

  const priceRows = useFieldArray({ control: form.control, name: "prices" })
  const prices = useWatch({ control: form.control, name: "prices" })
  const allFree = prices.every((row) => isZeroAmount(row.price))

  const watched = useWatch({ control: form.control })
  React.useEffect(() => {
    onValueChange(form.getValues())
  }, [watched, form, onValueChange])

  // Paid prices need a facilitator network; all-free rules may use any reachable chain; the edited one is kept.
  const networkOptions = React.useMemo(() => {
    const ids = new Set<string>(facilitatorNetworks)
    if (allFree) reachableNetworks.forEach((n) => ids.add(n))
    if (initialRule) ids.add(initialRule.network)
    return [...ids]
  }, [allFree, initialRule, facilitatorNetworks, reachableNetworks])

  const tokenInfo = tokenLookup.status === "found" ? tokenLookup.info : null

  // Editing: atomic prices become token units once the token's decimals load.
  const pricesInitialized = React.useRef(false)
  React.useEffect(() => {
    if (initialRule && tokenInfo && !pricesInitialized.current) {
      pricesInitialized.current = true
      initialRule.prices.forEach((row, index) => {
        if (row.price !== "0") {
          form.setValue(
            `prices.${index}.price`,
            formatAtomicAmount(row.price, tokenInfo.decimals)
          )
        }
      })
    }
  }, [initialRule, tokenInfo, form])

  // No reset: success unmounts the form, failure keeps values for a retry. Prices go out in atomic units.
  const onSubmit = async (values: PaymentReceiveRuleFormSchemaType) => {
    const atomicPrices: string[] = []
    for (const [index, row] of values.prices.entries()) {
      const field = `prices.${index}.price` as const
      if (isZeroAmount(row.price)) {
        atomicPrices.push("0")
        continue
      }
      if (!tokenInfo) {
        form.setError(field, {
          type: "manual",
          message:
            "Resolve the token contract first — its decimals are needed to convert the price",
        })
        return
      }
      const parsed = parseTokenAmount(row.price, tokenInfo.decimals)
      if (!parsed) {
        form.setError(field, {
          type: "manual",
          message: `Too many decimal places for $${tokenInfo.symbol} (max ${tokenInfo.decimals})`,
        })
        return
      }
      if (BigInt(parsed) > MAX_UINT256) {
        form.setError(field, {
          type: "manual",
          message: "Too large for a token amount (max 2^256-1 atomic units)",
        })
        return
      }
      atomicPrices.push(parsed)
    }

    const { error } = await onFormComplete({
      ...values,
      prices: values.prices.map((row, index) => ({
        price: atomicPrices[index],
        addressPatterns: row.addressPatterns,
      })),
    })

    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  const pricesError =
    form.formState.errors.prices?.root?.message ??
    form.formState.errors.prices?.message

  // A row's pattern error is either the list's own (none given) or the first invalid pattern's.
  const patternsError = (index: number): string | undefined => {
    const errors = form.formState.errors.prices?.[index]?.addressPatterns
    if (!errors) return undefined
    if (errors.message) return errors.message
    return Object.values(errors).find(
      (entry): entry is { message: string } =>
        typeof entry === "object" &&
        entry !== null &&
        typeof (entry as { message?: unknown }).message === "string"
    )?.message
  }

  return (
    <Form {...form}>
      <form
        id="payment-receive-rule-form"
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-4"
      >
        <div className="grid items-start gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
          <FormField
            control={form.control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input placeholder="Public price" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="method"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Payment method</FormLabel>
                {/* Locked to x402 — the seam is in place for future methods. */}
                <Select
                  value={field.value}
                  onValueChange={field.onChange}
                  disabled
                >
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value={X402_METHOD}>x402</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <div className="grid items-start gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
          <FormField
            control={form.control}
            name="network"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Network</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger className="w-full min-w-0">
                      {/* The trigger's flex layout defeats line-clamp, so a long chain name would overflow. */}
                      <SelectValue
                        className="!block min-w-0 truncate"
                        placeholder="Select a network"
                      />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {networkOptions.map((id) => (
                      <SelectItem key={id} value={id}>
                        {formatNetwork(id) === id
                          ? id
                          : `${formatNetwork(id)} (${id})`}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="asset"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Token contract</FormLabel>
                <FormControl>
                  <Input placeholder="0x…" {...field} />
                </FormControl>
                {tokenLookup.status === "loading" ? (
                  <p className="flex items-center gap-1 text-sm text-muted-foreground">
                    <Loader2 className="size-3.5 animate-spin" />
                    Looking up the token…
                  </p>
                ) : null}
                {tokenInfo ? (
                  <p className="flex items-center gap-1 text-sm text-emerald-600 dark:text-emerald-500">
                    <CircleCheck className="size-3.5" />
                    {tokenInfo.name || tokenInfo.symbol} (${tokenInfo.symbol}) —{" "}
                    {tokenInfo.decimals} decimals
                  </p>
                ) : null}
                {tokenLookup.status === "not-found" ? (
                  <p className="flex items-center gap-1 text-sm text-destructive dark:text-red-400">
                    <CircleX className="size-3.5" />
                    No ERC-20 token answers at this address on this network.
                  </p>
                ) : null}
                {tokenLookup.status === "unsupported" ? (
                  <FormDescription>
                    No token lookup on this network.
                  </FormDescription>
                ) : null}
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <FormField
          control={form.control}
          name="payTo"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Receiving wallet</FormLabel>
              <FormControl>
                <Input placeholder="0x…" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="space-y-2">
          <FormLabel>
            Prices per call{tokenInfo ? ` ($${tokenInfo.symbol})` : ""}
          </FormLabel>
          <p className="text-sm text-muted-foreground">
            Cheapest matching price wins. 0 = free.
          </p>
          {priceRows.fields.map((row, index) => (
            <div key={row.id} className="space-y-3 rounded-md border p-3">
              <div className="flex items-start gap-2">
                <FormField
                  control={form.control}
                  name={`prices.${index}.price`}
                  render={({ field }) => (
                    <FormItem className="flex-1">
                      <FormLabel>Price</FormLabel>
                      <div className="flex items-center gap-3">
                        <FormControl>
                          {/* Keystroke filter: digits and a single dot only. */}
                          <Input
                            className="w-44"
                            placeholder="0.01"
                            inputMode="decimal"
                            aria-label="Price per call"
                            {...field}
                            onChange={(event) =>
                              field.onChange(
                                event.target.value
                                  .replace(/[^0-9.]/g, "")
                                  .replace(/(\..*)\./g, "$1")
                              )
                            }
                          />
                        </FormControl>
                        <FormDescription>
                          {tokenInfo &&
                          !isZeroAmount(prices[index]?.price ?? "")
                            ? parseTokenAmount(
                                prices[index]?.price ?? "",
                                tokenInfo.decimals
                              )
                              ? `${parseTokenAmount(prices[index]?.price ?? "", tokenInfo.decimals)} atomic units`
                              : `Up to ${tokenInfo.decimals} decimal places`
                            : isZeroAmount(prices[index]?.price ?? "")
                              ? "Free for these callers."
                              : "In token units, converted on save."}
                        </FormDescription>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="mt-7"
                  aria-label="Remove price"
                  disabled={priceRows.fields.length === 1}
                  onClick={() => priceRows.remove(index)}
                >
                  <Trash2 />
                </Button>
              </div>
              <FormField
                control={form.control}
                name={`prices.${index}.addressPatterns`}
                render={({ field }) => (
                  <FormItem>
                    <AddressPatternsEditor
                      patterns={field.value}
                      onChange={field.onChange}
                    />
                    {patternsError(index) ? (
                      <p className="text-sm text-destructive dark:text-red-400">
                        {patternsError(index)}
                      </p>
                    ) : null}
                  </FormItem>
                )}
              />
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => priceRows.append({ price: "", addressPatterns: [] })}
          >
            <Plus />
            Add price
          </Button>
          {pricesError ? (
            <p className="text-sm text-destructive dark:text-red-400">
              {pricesError}
            </p>
          ) : null}
        </div>
      </form>
    </Form>
  )
}

// AddressPatternsEditor edits who pays one price: addresses or full-match regexes.
function AddressPatternsEditor({
  patterns,
  onChange,
}: {
  patterns: string[]
  onChange: (patterns: string[]) => void
}) {
  const [draft, setDraft] = React.useState("")

  const add = () => {
    const pattern = normalizePattern(draft.trim())
    if (pattern.length === 0) return
    if (!patterns.includes(pattern)) onChange([...patterns, pattern])
    setDraft("")
  }

  const isAnyone = patterns.includes(EVERYONE_PATTERN)

  return (
    <div className="space-y-2">
      <p className="text-sm font-medium">
        Who pays{patterns.length > 5 ? ` (${patterns.length})` : ""}
      </p>
      {/* ".*" already covers every caller, so nothing else can be added beside it. */}
      {isAnyone ? null : (
        <div className="flex items-center gap-2">
          <Input
            className="font-mono text-xs"
            aria-label="Payer address or regex"
            placeholder="0x… or a regex"
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault()
                add()
              }
            }}
          />
          <Button
            type="button"
            variant="outline"
            disabled={draft.trim().length === 0}
            onClick={add}
          >
            <Plus />
            Add
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => onChange([EVERYONE_PATTERN])}
          >
            Anyone
          </Button>
        </div>
      )}
      {patterns.length > 0 ? (
        <div className="flex max-h-24 flex-wrap items-center gap-1 overflow-y-auto">
          {patterns.map((pattern, index) => (
            <Badge
              key={`${index}:${pattern}`}
              variant={pattern === EVERYONE_PATTERN ? "default" : "outline"}
              className="gap-1 font-mono text-xs"
              title={pattern}
            >
              {pattern === EVERYONE_PATTERN
                ? "anyone"
                : isPlainAddressPattern(pattern)
                  ? truncateEthAddress(pattern)
                  : pattern}
              <button
                type="button"
                aria-label={`Remove ${pattern}`}
                onClick={() =>
                  onChange(patterns.filter((_, other) => other !== index))
                }
              >
                <X className="size-3" />
              </button>
            </Badge>
          ))}
        </div>
      ) : null}
    </div>
  )
}
