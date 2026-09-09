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
import {
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
const toLines = (patterns: string[] | undefined) => (patterns ?? [ANY_PATTERN]).join("\n")

export default function PaymentSpendRuleForm({ initialRule, onFormComplete }: Props) {
  const form = useForm<PaymentSpendRuleFormSchemaType>({
    resolver: zodResolver(paymentSpendRuleFormSchema),
    defaultValues: {
      name: initialRule?.name ?? "",
      method: initialRule?.method || X402_METHOD,
      maxAmountPerPayment: initialRule?.maxAmountPerPayment ?? "",
      networkPatterns: toLines(initialRule?.networkPatterns),
      assetPatterns: toLines(initialRule?.assetPatterns),
      payToPatterns: toLines(initialRule?.payToPatterns),
      assetTransferMethodPatterns: toLines(initialRule?.assetTransferMethodPatterns),
    },
  })

  const onSubmit = async (values: PaymentSpendRuleFormSchemaType) => {
    const { error } = await onFormComplete(values)
    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  const facet = (
    name: "networkPatterns" | "assetPatterns" | "payToPatterns" | "assetTransferMethodPatterns",
    label: string,
    description: string
  ) => (
    <FormField
      control={form.control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Textarea rows={3} className="font-mono text-xs" {...field} />
          </FormControl>
          <FormDescription>{description}</FormDescription>
          <FormMessage />
        </FormItem>
      )}
    />
  )

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
              <Select value={field.value} onValueChange={field.onChange} disabled>
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
                Atomic units of the quoted asset. When several rules match a quote, the
                smallest cap wins — 0 refuses anything priced.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {facet(
          "networkPatterns",
          "Networks",
          'One regex per line, matched against the quote\'s CAIP-2 network, e.g. eip155:8453 — ".*" allows any.'
        )}
        {facet(
          "assetPatterns",
          "Assets",
          "One regex per line, matched against the token contract address."
        )}
        {facet(
          "payToPatterns",
          "Recipients",
          "One regex per line, matched against the address being paid."
        )}
        {facet(
          "assetTransferMethodPatterns",
          "Transfer methods",
          "eip3009 authorizes one transfer; permit2 grants a standing allowance."
        )}
      </form>
    </Form>
  )
}
