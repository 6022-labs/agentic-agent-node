import type { FieldPath } from "react-hook-form"
import { z } from "zod"

import { UNBOUNDED_MAX_AMOUNT } from "@/models/payment-spend-policy"

// ANY_TRANSFER_METHOD is the select's stand-in for "either"; the wire omits the field.
export const ANY_TRANSFER_METHOD = "any"

// MAX_WINDOW_SECONDS mirrors the runtime's 10-year cap on a period.
export const MAX_WINDOW_SECONDS = 315360000

// fitsUint256 mirrors the runtime: an amount must fit an ERC-20 transfer.
const fitsUint256 = (value: string): boolean =>
  !/^\d+$/.test(value) || BigInt(value) <= BigInt(UNBOUNDED_MAX_AMOUNT)

const atomicAmount = z
  .string()
  .trim()
  .regex(/^\d+$/, "Whole atomic units, e.g. 1000000 (0 = deny)")
  .refine(fitsUint256, "At most 2^256-1 atomic units")

export const paymentSpendRuleFormSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, "Name is required")
      .max(64, "Max 64 characters"),
    method: z.string().min(1),
    // Atomic units, matching the runtime; empty leaves payments bounded by the windows only.
    maxAmountPerPayment: z
      .string()
      .trim()
      .regex(/^\d*$/, "Whole atomic units, e.g. 1000000 (empty = no cap)")
      .refine(fitsUint256, "At most 2^256-1 atomic units"),
    windows: z.array(
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
        maxAmount: atomicAmount,
      })
    ),
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

      // Atomic amounts run to 78 digits, past Number's exact range.
      if (
        values.maxAmountPerPayment.length > 0 &&
        /^\d+$/.test(window.maxAmount) &&
        BigInt(values.maxAmountPerPayment) > BigInt(window.maxAmount)
      ) {
        ctx.addIssue({
          code: "custom",
          path: ["maxAmountPerPayment"],
          message: "Cannot exceed any window's budget",
        })
      }
    })
  })

export type PaymentSpendRuleFormSchemaType = z.infer<
  typeof paymentSpendRuleFormSchema
>

// PaymentSpendRuleFormField names any form control, including one window's field.
export type PaymentSpendRuleFormField =
  FieldPath<PaymentSpendRuleFormSchemaType>
