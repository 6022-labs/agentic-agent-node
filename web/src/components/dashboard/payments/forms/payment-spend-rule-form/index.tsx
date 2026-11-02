import React from "react"
import { useFieldArray, useForm } from "react-hook-form"
import { Plus, Trash2 } from "lucide-react"
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
  WINDOW_PRESETS,
  X402_METHOD,
  type PaymentSpendRule,
} from "@/models/payment-spend-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { getAppConfig } from "@/configs/app-config"
import { formatNetwork } from "@/lib/token-format"
import {
  ANY_TRANSFER_METHOD,
  paymentSpendRuleFormSchema,
  type PaymentSpendRuleFormField,
  type PaymentSpendRuleFormSchemaType,
} from "./payment-spend-rule-form-schema"

type Props = {
  // When set, the form edits this rule instead of creating one.
  initialRule?: PaymentSpendRule
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

export default function PaymentSpendRuleForm({
  initialRule,
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

  const onSubmit = async (values: PaymentSpendRuleFormSchemaType) => {
    const { error } = await onFormComplete(values)
    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  return (
    <Form {...form}>
      <form
        id="payment-spend-rule-form"
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
              <FormLabel>Method</FormLabel>
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
                  <SelectItem value={X402_METHOD}>{X402_METHOD}</SelectItem>
                </SelectContent>
              </Select>
              <FormDescription>Only x402 is supported today.</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="maxAmountPerPayment"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Cap per payment</FormLabel>
              <FormControl>
                <Input
                  className="font-mono text-xs"
                  placeholder="Optional"
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Atomic units of the token. Empty leaves the windows as the only
                bound — 0 refuses anything priced.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="space-y-2">
          <FormLabel>Recurring budgets</FormLabel>
          {windowRows.fields.map((row, index) => (
            <div key={row.id} className="flex items-start gap-2">
              <FormField
                control={form.control}
                name={`windows.${index}.windowSeconds`}
                render={({ field }) => (
                  <FormItem className="w-32 shrink-0">
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
                        <Input
                          className="font-mono text-xs"
                          placeholder="Seconds"
                          {...field}
                        />
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
                    <FormControl>
                      <Input
                        className="font-mono text-xs"
                        placeholder="Budget (atomic units)"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
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
          <FormDescription>
            Each window caps the total spent over its period. A window you keep
            keeps what it has spent so far.
          </FormDescription>
          {form.formState.errors.windows?.root?.message ||
          form.formState.errors.windows?.message ? (
            <p className="text-sm text-destructive">
              {form.formState.errors.windows.root?.message ??
                form.formState.errors.windows.message}
            </p>
          ) : null}
        </div>

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
                <Input
                  placeholder="0x…"
                  className="font-mono text-xs"
                  {...field}
                />
              </FormControl>
              <FormDescription>
                One rule per token; the caps are counted in its atomic units.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

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
      </form>
    </Form>
  )
}
