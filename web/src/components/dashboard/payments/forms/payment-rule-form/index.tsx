import React from "react"
import { useFieldArray, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { CircleCheck, CircleX, Loader2, Plus, Trash2 } from "lucide-react"

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
  X402_METHOD,
  type PaymentRule,
} from "@/models/payment-policy"
import {
  getAppConfig,
  getEvmChainConfig,
  getFacilitatorNetworks,
} from "@/configs/app-config"
import type { FormCompletionResult } from "@/models/api/validation-error"
import type { Erc20Lookup } from "@/lib/erc20"
import {
  formatAtomicAmount,
  formatNetwork,
  isPlainAddressPattern,
  isZeroAmount,
  parseTokenAmount,
} from "@/lib/token-format"
import {
  paymentRuleFormSchema,
  type PaymentRuleFormSchemaType,
} from "./payment-rule-form-schema"

type Props = {
  // Pre-fills new rules as paying to the agent's own wallet.
  agentWalletAddress: string
  // When set, the form edits this rule instead of creating one.
  initialRule?: PaymentRule
  // Non-address regex patterns of the edited rule (kept verbatim by the
  // caller on save); surfaced so the operator knows they exist.
  preservedPatternCount?: number
  // Token metadata for the entered network + asset, resolved by the caller.
  tokenLookup: Erc20Lookup
  onValueChange: (values: PaymentRuleFormSchemaType) => void
  onFormComplete: (
    values: PaymentRuleFormSchemaType
  ) => Promise<FormCompletionResult<keyof PaymentRuleFormSchemaType>>
}

export default function PaymentRuleForm({
  agentWalletAddress,
  initialRule,
  preservedPatternCount = 0,
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

  const form = useForm<PaymentRuleFormSchemaType>({
    resolver: zodResolver(paymentRuleFormSchema),
    defaultValues: initialRule
      ? {
          name: initialRule.name,
          method: initialRule.method || X402_METHOD,
          network: initialRule.network,
          asset: initialRule.asset,
          payTo: initialRule.payTo,
          // Non-zero prices are atomic on the wire; they are converted to
          // token units below once the token's decimals are known.
          price: initialRule.price === "0" ? "0" : "",
          allowEveryone: initialRule.addressPatterns.includes(EVERYONE_PATTERN),
          addresses: initialRule.addressPatterns
            .filter(isPlainAddressPattern)
            .map((address) => ({ address })),
        }
      : // New rules default to a facilitator network and the $6022 token.
        {
          name: "",
          method: X402_METHOD,
          network: defaultNewRuleNetwork,
          asset: defaultToken,
          payTo: agentWalletAddress,
          price: "0.01",
          allowEveryone: true,
          addresses: [],
        },
  })

  const addressRows = useFieldArray({
    control: form.control,
    name: "addresses",
  })
  const allowEveryone = form.watch("allowEveryone")

  const price = form.watch("price")

  React.useEffect(() => {
    onValueChange(form.getValues())

    const subscription = form.watch(() => onValueChange(form.getValues()))
    return () => subscription.unsubscribe()
  }, [form, onValueChange])

  // Paid rules can only target a configured-facilitator network; free rules
  // never settle, so they offer every reachable chain. The edited rule's
  // network is always kept so editing never drops it.
  const networkOptions = React.useMemo(() => {
    const ids = new Set<string>(facilitatorNetworks)
    if (isZeroAmount(price)) reachableNetworks.forEach((n) => ids.add(n))
    if (initialRule) ids.add(initialRule.network)
    return [...ids]
  }, [price, initialRule, facilitatorNetworks, reachableNetworks])

  const tokenInfo = tokenLookup.status === "found" ? tokenLookup.info : null

  // When editing a paid rule, the stored atomic price becomes token units as
  // soon as the token's decimals are known (one-shot).
  const priceInitialized = React.useRef(false)
  React.useEffect(() => {
    if (
      initialRule &&
      initialRule.price !== "0" &&
      tokenInfo &&
      !priceInitialized.current
    ) {
      priceInitialized.current = true
      form.setValue(
        "price",
        formatAtomicAmount(initialRule.price, tokenInfo.decimals)
      )
    }
  }, [initialRule, tokenInfo, form])

  // No reset on success: the dialog closes and unmounts the form. Values are
  // kept on failure so the user can correct and retry. The price is entered
  // in token units and converted to the atomic units the gateway stores.
  const onSubmit = async (values: PaymentRuleFormSchemaType) => {
    let atomicPrice: string
    if (isZeroAmount(values.price)) {
      atomicPrice = "0"
    } else if (!tokenInfo) {
      form.setError("price", {
        type: "manual",
        message:
          "Resolve the token contract first — its decimals are needed to convert the price",
      })
      return
    } else {
      const parsed = parseTokenAmount(values.price, tokenInfo.decimals)
      if (!parsed) {
        form.setError("price", {
          type: "manual",
          message: `Too many decimal places for $${tokenInfo.symbol} (max ${tokenInfo.decimals})`,
        })
        return
      }
      atomicPrice = parsed
    }

    const { error } = await onFormComplete({ ...values, price: atomicPrice })

    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  return (
    <Form {...form}>
      <form
        id="payment-rule-form"
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-4"
      >
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
              <FormDescription>Only x402 is supported today.</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="network"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Network</FormLabel>
              <Select value={field.value} onValueChange={field.onChange}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a network" />
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
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <CircleX className="size-3.5" />
                  No ERC-20 token answers at this address on this network.
                </p>
              ) : null}
              {tokenLookup.status === "unsupported" ? (
                <FormDescription>
                  Token lookup isn't available for this network — the node
                  validates the contract on save when it can.
                </FormDescription>
              ) : null}
              <FormMessage />
            </FormItem>
          )}
        />
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
        <FormField
          control={form.control}
          name="price"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                Price per call{tokenInfo ? ` ($${tokenInfo.symbol})` : ""}
              </FormLabel>
              <FormControl>
                {/* Keystroke filter: digits and a single dot only. */}
                <Input
                  placeholder="0.01"
                  inputMode="decimal"
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
                {tokenInfo
                  ? `${
                      parseTokenAmount(price, tokenInfo.decimals)
                        ? `${price} $${tokenInfo.symbol} = ${parseTokenAmount(price, tokenInfo.decimals)} atomic units (${tokenInfo.decimals} decimals)`
                        : `Up to ${tokenInfo.decimals} decimal places`
                    }. 0 = free for the matched addresses.`
                  : "In token units, converted on save. 0 = free for the matched addresses."}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="space-y-2">
          <FormLabel>Allowed callers</FormLabel>
          <FormField
            control={form.control}
            name="allowEveryone"
            render={({ field }) => (
              <FormItem>
                <label className="flex items-center gap-2 text-sm font-normal">
                  <FormControl>
                    <input
                      type="checkbox"
                      className="size-4 accent-primary"
                      checked={field.value}
                      onChange={(event) => {
                        field.onChange(event.target.checked)
                        // Switching modes resets the list: a fresh empty row
                        // when scoping, nothing when opening to everyone.
                        addressRows.replace(
                          event.target.checked ? [] : [{ address: "" }]
                        )
                      }}
                    />
                  </FormControl>
                  Allow everyone
                </label>
                <FormDescription>
                  {field.value
                    ? "Anyone can pay this price; the rule is advertised to anonymous callers."
                    : "Only the addresses below get this price — a private offer, never advertised."}
                </FormDescription>
              </FormItem>
            )}
          />
          {allowEveryone ? null : (
            <>
              {addressRows.fields.map((row, index) => (
                <div key={row.id} className="flex items-start gap-2">
                  <FormField
                    control={form.control}
                    name={`addresses.${index}.address`}
                    render={({ field }) => (
                      <FormItem className="flex-1">
                        <FormControl>
                          <Input placeholder="0x…" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    aria-label="Remove address"
                    onClick={() => addressRows.remove(index)}
                  >
                    <Trash2 />
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => addressRows.append({ address: "" })}
              >
                <Plus />
                Add address
              </Button>
              {preservedPatternCount > 0 ? (
                <p className="text-sm text-muted-foreground">
                  {preservedPatternCount} custom regex pattern
                  {preservedPatternCount === 1 ? "" : "s"} set via the API are
                  kept as-is.
                </p>
              ) : null}
            </>
          )}
          {form.formState.errors.addresses?.root?.message ||
          form.formState.errors.addresses?.message ? (
            <p className="text-sm text-destructive">
              {form.formState.errors.addresses.root?.message ??
                form.formState.errors.addresses.message}
            </p>
          ) : null}
        </div>
      </form>
    </Form>
  )
}
