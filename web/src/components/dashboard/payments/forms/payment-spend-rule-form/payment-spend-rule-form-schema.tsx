import { z } from "zod"

// A facet is entered one pattern per line; an empty facet matches nothing, so
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
          ctx.addIssue({ code: "custom", message: `"${pattern}" is not a valid regex` })
          return
        }
      }
    })

export const paymentSpendRuleFormSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(64, "Max 64 characters"),
  method: z.string().min(1),
  // Atomic units, matching the runtime: "0" denies every priced quote in scope.
  maxAmount: z
    .string()
    .trim()
    .regex(/^\d+$/, "Whole atomic units, e.g. 1000000 (0 = deny)"),
  networkPatterns: patternList("Networks"),
  assetPatterns: patternList("Assets"),
  payToPatterns: patternList("Recipients"),
  assetTransferMethodPatterns: patternList("Transfer methods"),
})

export type PaymentSpendRuleFormSchemaType = z.infer<typeof paymentSpendRuleFormSchema>
