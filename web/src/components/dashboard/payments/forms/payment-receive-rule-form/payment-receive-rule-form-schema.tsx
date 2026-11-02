import type { FieldPath } from "react-hook-form"
import { z } from "zod"

import { getFacilitatorNetworks } from "@/configs/app-config"
import { isZeroAmount } from "@/lib/token-format"

const evmAddress = /^0x[0-9a-fA-F]{40}$/

// isValidPattern mirrors the runtime's check that a pattern compiles; RE2 may still refuse rare syntax.
const isValidPattern = (pattern: string): boolean => {
  try {
    new RegExp(pattern)
    return true
  } catch {
    return false
  }
}

// canonicalUnits drops leading and trailing zeros so "1", "01" and "1.0" read as one price.
export const canonicalUnits = (amount: string): string => {
  const [whole, fraction = ""] = amount.trim().split(".")
  const trimmedWhole = whole.replace(/^0+(?=\d)/, "")
  const trimmedFraction = fraction.replace(/0+$/, "")
  return trimmedFraction.length > 0
    ? `${trimmedWhole}.${trimmedFraction}`
    : trimmedWhole
}

export const paymentReceiveRuleFormSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, "Name is required")
      .max(64, "Max 64 characters"),
    method: z.string().min(1),
    network: z
      .string()
      .regex(/^eip155:\d+$/, "Must be a CAIP-2 id, e.g. eip155:80002"),
    asset: z.string().regex(evmAddress, "Must be the token contract address"),
    payTo: z.string().regex(evmAddress, "Must be the receiving wallet address"),
    prices: z
      .array(
        z.object({
          // Token units; converted to atomic units on submit.
          price: z
            .string()
            .trim()
            .regex(/^\d+(\.\d+)?$/, "Token units, e.g. 0.01 (0 = free)"),
          addressPatterns: z
            .array(z.string().refine(isValidPattern, "Not a valid regex"))
            .min(1, "Add at least one caller, or allow anyone"),
        })
      )
      .min(1, "Add at least one price"),
  })
  .superRefine((data, ctx) => {
    const seen = new Set<string>()
    data.prices.forEach((row, index) => {
      const price = canonicalUnits(row.price)
      if (seen.has(price)) {
        ctx.addIssue({
          code: "custom",
          message: "This price already has a row — add its callers there",
          path: ["prices", index, "price"],
        })
      }
      seen.add(price)
    })
    // Only paid prices need a facilitator network.
    if (
      data.prices.some((row) => !isZeroAmount(row.price)) &&
      !getFacilitatorNetworks().includes(data.network)
    ) {
      ctx.addIssue({
        code: "custom",
        message:
          "No payment facilitator is available for this network — only free prices can use it",
        path: ["network"],
      })
    }
  })

export type PaymentReceiveRuleFormSchemaType = z.infer<
  typeof paymentReceiveRuleFormSchema
>

// PaymentReceiveRuleFormField names any form control, including one price row's field.
export type PaymentReceiveRuleFormField =
  FieldPath<PaymentReceiveRuleFormSchemaType>
