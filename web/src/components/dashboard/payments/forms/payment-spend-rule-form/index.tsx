import React from "react"
import { useForm } from "react-hook-form"
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
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  ANY_PATTERN,
  X402_METHOD,
  type PaymentSpendRule,
} from "@/models/payment-spend-policy"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { getAppConfig } from "@/configs/app-config"
import { formatNetwork } from "@/lib/token-format"
import {
  ANY_TRANSFER_METHOD,
  paymentSpendRuleFormSchema,
  type PaymentSpendRuleFormSchemaType,
} from "./payment-spend-rule-form-schema"

type Props = {
  // When set, the form edits this rule instead of creating one.
  initialRule?: PaymentSpendRule
  onFormComplete: (
    values: PaymentSpendRuleFormSchemaType
  ) => Promise<FormCompletionResult<keyof PaymentSpendRuleFormSchemaType>>
}

// toLines renders a stored facet for editing, one pattern per line.
const toLines = (patterns: string[] | undefined) =>
  (patterns ?? [ANY_PATTERN]).join("\n")

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
      windowSeconds: initialRule?.windowSeconds
        ? String(initialRule.windowSeconds)
        : "",
      maxAmountPerWindow:
        initialRule?.maxAmountPerWindow &&
        initialRule.maxAmountPerWindow !== "0"
          ? initialRule.maxAmountPerWindow
          : "",
      network: initialRule?.network ?? networkOptions[0] ?? "",
      asset: initialRule?.asset ?? "",
      assetTransferMethod: (initialRule?.assetTransferMethod ||
        ANY_TRANSFER_METHOD) as PaymentSpendRuleFormSchemaType["assetTransferMethod"],
      payToPatterns: toLines(initialRule?.payToPatterns),
    },
  })

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
                <Input className="font-mono text-xs" {...field} />
              </FormControl>
              <FormDescription>
                Atomic units of the token. When several rules match a quote, the
                smallest cap wins — 0 refuses anything priced.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="windowSeconds"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Recurring period (seconds)</FormLabel>
              <FormControl>
                <Input
                  className="font-mono text-xs"
                  placeholder="86400"
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Leave empty to bound this rule per payment only. 86400 is a day,
                604800 a week.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="maxAmountPerWindow"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Budget per period</FormLabel>
              <FormControl>
                <Input className="font-mono text-xs" {...field} />
              </FormControl>
              <FormDescription>
                Atomic units the rule may spend before the period rolls over.
                Changing it restarts the period.
              </FormDescription>
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
                <Input
                  placeholder="0x…"
                  className="font-mono text-xs"
                  {...field}
                />
              </FormControl>
              <FormDescription>
                The caps are counted in this token's atomic units.
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

        <FormField
          control={form.control}
          name="payToPatterns"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Recipients</FormLabel>
              <FormControl>
                <Textarea rows={3} className="font-mono text-xs" {...field} />
              </FormControl>
              <FormDescription>
                One regex per line, matched against the address being paid —
                ".*" allows anyone.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}
