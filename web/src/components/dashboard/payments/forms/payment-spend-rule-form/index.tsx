import React from "react"
import { useFieldArray, useForm, useWatch } from "react-hook-form"
import { CircleCheck, CircleX, Loader2, Plus, Trash2 } from "lucide-react"
import { zodResolver } from "@hookform/resolvers/zod"

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  UNBOUNDED_MAX_AMOUNT,
  WINDOW_PRESETS,
  X402_METHOD,
  type PaymentSpendRule,
} from "@/models/payment-spend-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import type { Erc20Lookup } from "@/lib/erc20"
import { getAppConfig } from "@/configs/app-config"
import {
  formatAtomicAmount,
  formatNetwork,
  isZeroAmount,
  parseTokenAmount,
} from "@/lib/token-format"
import {
  ANY_TRANSFER_METHOD,
  paymentSpendRuleFormSchema,
  type PaymentSpendRuleFormField,
  type PaymentSpendRuleFormSchemaType,
} from "./payment-spend-rule-form-schema"

type Props = {
  // When set, the form edits this rule instead of creating one.
  initialRule?: PaymentSpendRule
  // Token metadata for the entered network + asset, resolved by the caller.
  tokenLookup: Erc20Lookup
  onValueChange: (values: PaymentSpendRuleFormSchemaType) => void
  onFormComplete: (
    values: PaymentSpendRuleFormSchemaType
  ) => Promise<FormCompletionResult<PaymentSpendRuleFormField>>
}

// CUSTOM_PERIOD is the select's value for a period no preset names.
const CUSTOM_PERIOD = "custom"

const presetFor = (windowSeconds: string): string =>
  WINDOW_PRESETS.some((preset) => String(preset.seconds) === windowSeconds)
    ? windowSeconds
    : CUSTOM_PERIOD

// Keystroke filter: digits and a single dot only.
const decimalOnly = (value: string): string =>
  value.replace(/[^0-9.]/g, "").replace(/(\..*)\./g, "$1")

// Non-zero amounts are atomic on the wire; they show once the token's decimals are known.
const tokenUnitsPlaceholder = (amount: string | undefined): string =>
  amount === undefined ? "" : amount === "0" ? "0" : ""

export default function PaymentSpendRuleForm({
  initialRule,
  tokenLookup,
  onValueChange,
  onFormComplete,
}: Props) {
  // Reachable chains as CAIP-2 networks; the edited rule's network is always kept.
  const networkOptions = React.useMemo(() => {
    const ids = new Set(
      Object.keys(getAppConfig().evmChains).map((id) => `eip155:${id}`)
    )
    if (initialRule?.network) ids.add(initialRule.network)
    return [...ids]
  }, [initialRule])

  const form = useForm<PaymentSpendRuleFormSchemaType>({
    resolver: zodResolver(paymentSpendRuleFormSchema),
    defaultValues: {
      name: initialRule?.name ?? "",
      method: initialRule?.method || X402_METHOD,
      maxAmountPerPayment: tokenUnitsPlaceholder(
        initialRule?.maxAmountPerPayment
      ),
      windows: (initialRule?.windows ?? []).map((window) => ({
        windowSeconds: String(window.windowSeconds),
        maxAmount: tokenUnitsPlaceholder(window.maxAmount),
      })),
      network: initialRule?.network ?? networkOptions[0] ?? "",
      asset: initialRule?.asset ?? "",
      assetTransferMethod: (initialRule?.assetTransferMethod ||
        ANY_TRANSFER_METHOD) as PaymentSpendRuleFormSchemaType["assetTransferMethod"],
    },
  })

  const windowRows = useFieldArray({ control: form.control, name: "windows" })
  const cap = useWatch({ control: form.control, name: "maxAmountPerPayment" })
  const windows = useWatch({ control: form.control, name: "windows" })

  const watched = useWatch({ control: form.control })
  React.useEffect(() => {
    onValueChange(form.getValues())
  }, [watched, form, onValueChange])

  const tokenInfo = tokenLookup.status === "found" ? tokenLookup.info : null

  // Editing: atomic amounts become token units once the token's decimals load.
  const amountsInitialized = React.useRef(false)
  React.useEffect(() => {
    if (initialRule && tokenInfo && !amountsInitialized.current) {
      amountsInitialized.current = true
      const cap = initialRule.maxAmountPerPayment
      if (cap !== undefined && cap !== "0") {
        form.setValue(
          "maxAmountPerPayment",
          formatAtomicAmount(cap, tokenInfo.decimals)
        )
      }
      initialRule.windows.forEach((window, index) => {
        if (window.maxAmount !== "0") {
          form.setValue(
            `windows.${index}.maxAmount`,
            formatAtomicAmount(window.maxAmount, tokenInfo.decimals)
          )
        }
      })
    }
  }, [initialRule, tokenInfo, form])

  // Custom mode is per row, not derived from the value, so typing 3600 does not snap to "Hour".
  const [customRows, setCustomRows] = React.useState<Record<string, boolean>>(
    {}
  )
  const isCustomRow = (rowId: string, windowSeconds: string): boolean =>
    customRows[rowId] ?? presetFor(windowSeconds) === CUSTOM_PERIOD

  // The next preset not used yet, so "Add window" rarely needs a second click.
  const addWindow = () => {
    const used = new Set(form.getValues("windows").map((w) => w.windowSeconds))
    const next = WINDOW_PRESETS.find((p) => !used.has(String(p.seconds)))
    windowRows.append({
      windowSeconds: next ? String(next.seconds) : "",
      maxAmount: "",
    })
  }

  // toAtomic converts one token-unit amount, or sets the field's error and returns null.
  const toAtomic = (
    field: PaymentSpendRuleFormField,
    amount: string
  ): string | null => {
    if (isZeroAmount(amount)) return "0"
    if (!tokenInfo) {
      form.setError(field, {
        type: "manual",
        message:
          "Resolve the token contract first — its decimals are needed to convert the amount",
      })
      return null
    }
    const parsed = parseTokenAmount(amount, tokenInfo.decimals)
    if (!parsed) {
      form.setError(field, {
        type: "manual",
        message: `Too many decimal places for $${tokenInfo.symbol} (max ${tokenInfo.decimals})`,
      })
      return null
    }
    if (BigInt(parsed) > BigInt(UNBOUNDED_MAX_AMOUNT)) {
      form.setError(field, {
        type: "manual",
        message: "Too large for a token amount (max 2^256-1 atomic units)",
      })
      return null
    }
    return parsed
  }

  // No reset: success unmounts the form, failure keeps values for a retry. Amounts go out in atomic units.
  const onSubmit = async (values: PaymentSpendRuleFormSchemaType) => {
    // An unconverted edit would save the shown placeholders over the stored amounts.
    if (initialRule && !tokenInfo && !amountsInitialized.current) {
      const storedAmounts = [
        initialRule.maxAmountPerPayment,
        ...initialRule.windows.map((window) => window.maxAmount),
      ]
      if (storedAmounts.some((amount) => amount && amount !== "0")) {
        form.setError("asset", {
          type: "manual",
          message:
            "The token's decimals are needed to show this rule's amounts — check the contract",
        })
        return
      }
    }

    let atomicCap = ""
    if (values.maxAmountPerPayment.length > 0) {
      const converted = toAtomic(
        "maxAmountPerPayment",
        values.maxAmountPerPayment
      )
      if (converted === null) return
      atomicCap = converted
    }
    const atomicWindows: string[] = []
    for (const [index, window] of values.windows.entries()) {
      const converted = toAtomic(`windows.${index}.maxAmount`, window.maxAmount)
      if (converted === null) return
      atomicWindows.push(converted)
    }
    if (
      atomicCap.length > 0 &&
      atomicWindows.some((max) => BigInt(atomicCap) > BigInt(max))
    ) {
      form.setError("maxAmountPerPayment", {
        type: "manual",
        message: "Cannot exceed any window's budget",
      })
      return
    }

    const { error } = await onFormComplete({
      ...values,
      maxAmountPerPayment: atomicCap,
      windows: values.windows.map((window, index) => ({
        windowSeconds: window.windowSeconds,
        maxAmount: atomicWindows[index],
      })),
    })
    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  // amountHint tells what a token-unit amount means once converted.
  const amountHint = (amount: string, empty: string): string => {
    if (amount.trim().length === 0) return empty
    if (isZeroAmount(amount)) return "0 refuses anything priced."
    if (!tokenInfo) return "In token units, converted on save."
    const atomic = parseTokenAmount(amount, tokenInfo.decimals)
    return atomic
      ? `${atomic} atomic units`
      : `Up to ${tokenInfo.decimals} decimal places`
  }

  return (
    <Form {...form}>
      <form
        id="payment-spend-rule-form"
        onSubmit={(event) => void form.handleSubmit(onSubmit)(event)}
        className="space-y-4"
      >
        <div className="grid items-start gap-4 sm:grid-cols-2">
          <FormField
            control={form.control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input placeholder="Only Base, small amounts" {...field} />
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
        <div className="grid items-start gap-4 sm:grid-cols-2">
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
                    Token lookup isn't available for this network — only 0
                    amounts can be converted without the token's decimals.
                  </FormDescription>
                ) : null}
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <div className="grid items-start gap-4 sm:grid-cols-2">
          <FormField
            control={form.control}
            name="assetTransferMethod"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Transfer method</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value={ANY_TRANSFER_METHOD}>Either</SelectItem>
                    <SelectItem value="eip3009">eip3009</SelectItem>
                    <SelectItem value="permit2">permit2</SelectItem>
                  </SelectContent>
                </Select>
                <FormDescription>
                  eip3009 authorizes one transfer; permit2 grants a standing
                  allowance.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="maxAmountPerPayment"
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  Cap per payment{tokenInfo ? ` ($${tokenInfo.symbol})` : ""}
                </FormLabel>
                <FormControl>
                  <Input
                    placeholder="Optional"
                    inputMode="decimal"
                    {...field}
                    onChange={(event) =>
                      field.onChange(decimalOnly(event.target.value))
                    }
                  />
                </FormControl>
                <FormDescription>
                  {amountHint(cap, "Empty: only the budgets bound a payment.")}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <div className="space-y-2">
          <FormLabel>
            Recurring budgets{tokenInfo ? ` ($${tokenInfo.symbol})` : ""}
          </FormLabel>
          <p className="text-sm text-muted-foreground">
            Each budget caps the total spent over its period; a kept one keeps
            what it has spent.
          </p>
          {windowRows.fields.map((row, index) => (
            <div
              key={row.id}
              className="flex items-start gap-2 rounded-md border p-3"
            >
              <FormField
                control={form.control}
                name={`windows.${index}.windowSeconds`}
                render={({ field }) => (
                  <FormItem className="w-44 shrink-0">
                    <FormLabel>Period</FormLabel>
                    <Select
                      value={
                        isCustomRow(row.id, field.value)
                          ? CUSTOM_PERIOD
                          : field.value
                      }
                      onValueChange={(value) => {
                        const custom = value === CUSTOM_PERIOD
                        setCustomRows((rows) => ({ ...rows, [row.id]: custom }))
                        field.onChange(custom ? "" : value)
                      }}
                    >
                      <FormControl>
                        <SelectTrigger className="w-full">
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {WINDOW_PRESETS.map((preset) => (
                          <SelectItem
                            key={preset.seconds}
                            value={String(preset.seconds)}
                          >
                            {preset.label}
                          </SelectItem>
                        ))}
                        <SelectItem value={CUSTOM_PERIOD}>Custom</SelectItem>
                      </SelectContent>
                    </Select>
                    {isCustomRow(row.id, field.value) ? (
                      <FormControl>
                        <Input placeholder="Seconds" {...field} />
                      </FormControl>
                    ) : null}
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name={`windows.${index}.maxAmount`}
                render={({ field }) => (
                  <FormItem className="flex-1">
                    <FormLabel>Max amount</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="10"
                        inputMode="decimal"
                        {...field}
                        onChange={(event) =>
                          field.onChange(decimalOnly(event.target.value))
                        }
                      />
                    </FormControl>
                    <FormDescription>
                      {amountHint(windows[index]?.maxAmount ?? "", "")}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="mt-7"
                aria-label="Remove window"
                onClick={() => windowRows.remove(index)}
              >
                <Trash2 />
              </Button>
            </div>
          ))}
          <Button type="button" variant="outline" size="sm" onClick={addWindow}>
            <Plus />
            Add window
          </Button>
          {form.formState.errors.windows?.root?.message ||
          form.formState.errors.windows?.message ? (
            <p className="text-sm text-destructive">
              {form.formState.errors.windows.root?.message ??
                form.formState.errors.windows.message}
            </p>
          ) : null}
        </div>
      </form>
    </Form>
  )
}
