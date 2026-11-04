import type { FieldPath } from "react-hook-form"
import { z } from "zod"

// ANY_TRANSFER_METHOD is the select's stand-in for "either"; the wire omits the field.
export const ANY_TRANSFER_METHOD = "any"

// MAX_WINDOW_SECONDS mirrors the runtime's 10-year cap on a period.
const MAX_WINDOW_SECONDS = 315360000

// MAX_WINDOWS mirrors the runtime's MaxPaymentSpendRuleWindows.
export const MAX_WINDOWS = 10

// Token units; the form converts to atomic units on submit, where the uint256 bound is checked.
const tokenAmount = z
  .string()
  .trim()
  .regex(/^\d+(\.\d+)?$/, "Token units, e.g. 10 (0 = deny)")

export const paymentSpendRuleFormSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, "Name is required")
      .max(64, "Max 64 characters"),
    method: z.string().min(1),
    // Empty leaves payments bounded by the windows only.
    maxAmountPerPayment: z
      .string()
      .trim()
      .regex(/^(\d+(\.\d+)?)?$/, "Token units, e.g. 0.5 (empty = no cap)"),
    windows: z
      .array(
        z.object({
          windowSeconds: z
            .string()
            .trim()
            .regex(/^[1-9]\d*$/, "Whole seconds, e.g. 86400 for a day")
            .refine(
              (value) =>
                !/^\d+$/.test(value) || Number(value) <= MAX_WINDOW_SECONDS,
              "At most 10 years (315360000 seconds)"
            ),
          maxAmount: tokenAmount,
        })
      )
      .max(MAX_WINDOWS, `At most ${MAX_WINDOWS} budgets`),
    network: z
      .string()
      .trim()
      .regex(/^eip155:\d+$/, "Pick a network"),
    asset: z
      .string()
      .trim()
      .regex(/^0x[0-9a-fA-F]{40}$/, "A 0x-prefixed token contract address"),
    assetTransferMethod: z.enum([ANY_TRANSFER_METHOD, "eip3009", "permit2"]),
  })
  .superRefine((values, ctx) => {
    if (
      values.maxAmountPerPayment.length === 0 &&
      values.windows.length === 0
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["windows"],
        message: "Set a cap per payment, a window, or both",
      })
      return
    }

    const seen = new Set<string>()
    values.windows.forEach((window, index) => {
      if (seen.has(window.windowSeconds)) {
        ctx.addIssue({
          code: "custom",
          path: ["windows", index, "windowSeconds"],
          message: "This period already has a window",
        })
      }
      seen.add(window.windowSeconds)
    })
  })

export type PaymentSpendRuleFormSchemaType = z.infer<
  typeof paymentSpendRuleFormSchema
>

// PaymentSpendRuleFormField names any form control, including one window's field.
export type PaymentSpendRuleFormField =
  FieldPath<PaymentSpendRuleFormSchemaType>
