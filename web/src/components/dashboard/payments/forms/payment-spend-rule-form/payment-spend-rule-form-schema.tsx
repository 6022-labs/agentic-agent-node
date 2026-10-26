import { z } from "zod"

// ANY_TRANSFER_METHOD is the select's stand-in for "either"; the wire omits the field.
export const ANY_TRANSFER_METHOD = "any"

// Recipients are entered one pattern per line; an empty list matches nobody, so
// the form requires at least one and offers ".*" as the open value. Patterns are
// compiled server-side by Go's RE2, which rejects lookahead and backreferences.
const patternList = (label: string) =>
  z
    .string()
    .trim()
    .min(1, `${label} needs at least one pattern (".*" matches everything)`)
    .superRefine((value, ctx) => {
      for (const pattern of value.split("\n").map((line) => line.trim())) {
        if (pattern.length === 0) {
          continue
        }
        try {
          new RegExp(`^(?:${pattern})$`)
        } catch {
          ctx.addIssue({
            code: "custom",
            message: `"${pattern}" is not a valid regex`,
          })
          return
        }
      }
    })

export const paymentSpendRuleFormSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, "Name is required")
      .max(64, "Max 64 characters"),
    method: z.string().min(1),
    // Atomic units, matching the runtime: "0" denies every priced quote in scope.
    maxAmountPerPayment: z
      .string()
      .trim()
      .regex(/^\d+$/, "Whole atomic units, e.g. 1000000 (0 = deny)"),
    // A period needs a budget and a budget needs a period; the runtime rejects either alone.
    windowSeconds: z
      .string()
      .trim()
      .regex(
        /^\d*$/,
        "Whole seconds, e.g. 86400 for a day (empty = no recurring budget)"
      ),
    maxAmountPerWindow: z
      .string()
      .trim()
      .regex(/^\d*$/, "Whole atomic units (empty = no recurring budget)"),
    network: z
      .string()
      .trim()
      .regex(/^eip155:\d+$/, "Pick a network"),
    asset: z
      .string()
      .trim()
      .regex(/^0x[0-9a-fA-F]{40}$/, "A 0x-prefixed token contract address"),
    assetTransferMethod: z.enum([ANY_TRANSFER_METHOD, "eip3009", "permit2"]),
    payToPatterns: patternList("Recipients"),
  })
  .superRefine((values, ctx) => {
    const period =
      values.windowSeconds.length === 0 ? 0 : Number(values.windowSeconds)
    const budget = values.maxAmountPerWindow

    if (period === 0) {
      if (budget.length !== 0) {
        ctx.addIssue({
          code: "custom",
          path: ["windowSeconds"],
          message: "A recurring budget needs a period to be spent over",
        })
      }
      return
    }
    if (budget.length === 0) {
      ctx.addIssue({
        code: "custom",
        path: ["maxAmountPerWindow"],
        message: "A period needs a budget",
      })
      return
    }
    // Atomic amounts run to 78 digits, past Number's exact range.
    if (BigInt(values.maxAmountPerPayment || "0") > BigInt(budget)) {
      ctx.addIssue({
        code: "custom",
        path: ["maxAmountPerPayment"],
        message: "Cannot exceed the budget for one period",
      })
    }
  })

export type PaymentSpendRuleFormSchemaType = z.infer<
  typeof paymentSpendRuleFormSchema
>
