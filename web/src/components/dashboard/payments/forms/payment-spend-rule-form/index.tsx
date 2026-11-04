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
import { useErc20Lookup } from "../../use-erc20-lookup"
import { getAppConfig } from "@/configs/app-config"
import {
  formatAtomicAmount,
  formatNetwork,
  isZeroAmount,
  parseTokenAmount,
} from "@/lib/token-format"
import {
  ANY_TRANSFER_METHOD,
  MAX_WINDOWS,
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

// A comma is read as the decimal point; anything else is refused, never silently dropped.
const TOKEN_AMOUNT = /^\d*\.?\d*$/
const ATOMIC_AMOUNT = /^\d*$/

// AmountMode is how amounts are typed: token units when the decimals are known, else atomic units.
type AmountMode = "token" | "atomic"

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
      // Stored atomic values; converted to token units once the decimals load.
      maxAmountPerPayment: initialRule?.maxAmountPerPayment ?? "",
      windows: (initialRule?.windows ?? []).map((window) => ({
        windowSeconds: String(window.windowSeconds),
        maxAmount: window.maxAmount,
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

  // Amounts wait for a settled lookup; the mode then stays fixed until every amount is cleared.
  const [lockedMode, setLockedMode] = React.useState<AmountMode | null>(null)
  // An edited rule holds raw atomic values until its stored amounts are converted.
  const [initDone, setInitDone] = React.useState(!initialRule)
  const mode: AmountMode = !initDone
    ? "atomic"
    : (lockedMode ?? (tokenInfo ? "token" : "atomic"))
  const lookupSettled =
    tokenLookup.status !== "idle" && tokenLookup.status !== "loading"

  // Stored amounts convert with the rule's own token, whatever is typed in the token field meanwhile.
  const storedLookup = useErc20Lookup(initialRule?.network, initialRule?.asset)
  const initialized = React.useRef(false)
  React.useEffect(() => {
    if (!initialRule || initialized.current) return
    if (storedLookup.status === "idle" || storedLookup.status === "loading") {
      return
    }
    initialized.current = true
    const storedInfo =
      storedLookup.status === "found" ? storedLookup.info : null
    if (storedInfo) {
      const cap = initialRule.maxAmountPerPayment
      if (cap !== undefined) {
        form.setValue(
          "maxAmountPerPayment",
          formatAtomicAmount(cap, storedInfo.decimals)
        )
      }
      initialRule.windows.forEach((window, index) => {
        form.setValue(
          `windows.${index}.maxAmount`,
          formatAtomicAmount(window.maxAmount, storedInfo.decimals)
        )
      })
    }
    setLockedMode(storedInfo ? "token" : "atomic")
    setInitDone(true)
  }, [initialRule, storedLookup, form])

  // Clearing every amount lets the mode follow the lookup again, so a token change never gets stuck.
  const amountsEmpty =
    (cap ?? "").trim().length === 0 &&
    (windows ?? []).every(
      (window) => (window?.maxAmount ?? "").trim().length === 0
    )
  React.useEffect(() => {
    if (amountsEmpty && lockedMode !== null) {
      setLockedMode(null)
    }
  }, [amountsEmpty, lockedMode])

  const onAmountChange = (
    field: PaymentSpendRuleFormField,
    onChange: (value: string) => void,
    value: string
  ) => {
    const normalized = mode === "token" ? value.replace(/,/g, ".") : value
    if (!(mode === "token" ? TOKEN_AMOUNT : ATOMIC_AMOUNT).test(normalized)) {
      form.setError(field, {
        type: "manual",
        message:
          mode === "token"
            ? "Digits and one decimal point only"
            : "Digits only",
      })
      return
    }
    form.clearErrors(field)
    if (lockedMode === null) setLockedMode(mode)
    onChange(normalized)
  }

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

  // An unchanged token falls back to its stored decimals, so a failed live lookup never blocks saving.
  const conversionInfo = (): typeof tokenInfo => {
    if (tokenInfo) return tokenInfo
    const { network, asset } = form.getValues()
    const sameToken =
      initialRule !== undefined &&
      network === initialRule.network &&
      asset.trim().toLowerCase() === initialRule.asset.toLowerCase()
    return sameToken && storedLookup.status === "found"
      ? storedLookup.info
      : null
  }

  // toAtomic converts one typed amount to atomic units, or sets the field's error and returns null.
  const toAtomic = (
    field: PaymentSpendRuleFormField,
    amount: string
  ): string | null => {
    const info = conversionInfo()
    let parsed: string | null
    if (mode === "atomic") {
      if (!/^\d+$/.test(amount)) {
        form.setError(field, {
          type: "manual",
          message: "Whole atomic units only",
        })
        return null
      }
      parsed = BigInt(amount).toString()
    } else {
      if (!info) {
        form.setError(field, {
          type: "manual",
          message:
            "The token's decimals are unknown; clear the amounts to enter atomic units",
        })
        return null
      }
      parsed = parseTokenAmount(amount, info.decimals)
      if (!parsed) {
        form.setError(field, {
          type: "manual",
          message: `Too many decimal places for $${info.symbol} (max ${info.decimals})`,
        })
        return null
      }
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
    if (!initDone) return
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

  // amountHint tells what a typed amount means once converted.
  const amountHint = (amount: string, empty: string): string => {
    if (!lookupSettled) {
      return tokenLookup.status === "loading"
        ? "Looking up the token…"
        : "Enter the token contract first."
    }
    if (amount.trim().length === 0) return empty
    if (isZeroAmount(amount)) return "0 refuses anything priced."
    if (mode === "atomic") return "Atomic units (token decimals unknown)"
    if (!tokenInfo) {
      return "The token's decimals are unknown; clear the amounts to enter atomic units."
    }
    const atomic = parseTokenAmount(amount, tokenInfo.decimals)
    return atomic
      ? `${atomic} atomic units`
      : `Up to ${tokenInfo.decimals} decimal places`
  }

  const unitsLabel =
    mode === "atomic"
      ? " (atomic units)"
      : tokenInfo
        ? ` ($${tokenInfo.symbol})`
        : ""

  return (
    <Form {...form}>
      <form
        id="payment-spend-rule-form"
        onSubmit={(event) => void form.handleSubmit(onSubmit)(event)}
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
                {tokenLookup.status === "timeout" ? (
                  <p className="flex items-center gap-1 text-sm text-destructive">
                    <CircleX className="size-3.5" />
                    Token lookup timed out.
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
                    No token lookup on this network; enter atomic units.
                  </FormDescription>
                ) : null}
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <div className="grid items-start gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
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
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="maxAmountPerPayment"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Cap per payment{unitsLabel}</FormLabel>
                <FormControl>
                  <Input
                    placeholder="Optional"
                    inputMode={mode === "token" ? "decimal" : "numeric"}
                    {...field}
                    disabled={!lookupSettled || !initDone}
                    onChange={(event) =>
                      onAmountChange(
                        "maxAmountPerPayment",
                        field.onChange,
                        event.target.value
                      )
                    }
                  />
                </FormControl>
                <FormDescription>
                  {amountHint(cap, "Empty = no cap.")}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium">Recurring budgets{unitsLabel}</p>
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
                    {/* Outside FormControl: the select trigger already owns the item's id. */}
                    {isCustomRow(row.id, field.value) ? (
                      <Input
                        id={`${row.id}-seconds`}
                        aria-label="Period in seconds"
                        placeholder="Seconds"
                        {...field}
                      />
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
                        inputMode={mode === "token" ? "decimal" : "numeric"}
                        {...field}
                        disabled={!lookupSettled || !initDone}
                        onChange={(event) =>
                          onAmountChange(
                            `windows.${index}.maxAmount`,
                            field.onChange,
                            event.target.value
                          )
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
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={windowRows.fields.length >= MAX_WINDOWS}
            onClick={addWindow}
          >
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
