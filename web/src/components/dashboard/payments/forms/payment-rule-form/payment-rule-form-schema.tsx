import { z } from "zod"

import { getFacilitatorNetworks } from "@/configs/app-config"
import { isZeroAmount } from "@/lib/token-format"

const evmAddress = /^0x[0-9a-fA-F]{40}$/

export const paymentRuleFormSchema = z
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
    // Entered in token units; the form converts to atomic units on submit
    // using the resolved token's decimals.
    price: z
      .string()
      .regex(/^\d+(\.\d+)?$/, "Token units, e.g. 0.01 (0 = free)"),
    // The wire format is regex patterns, but the form only deals in plain
    // addresses — "allow everyone" maps to the ".*" pattern on submit.
    allowEveryone: z.boolean(),
    addresses: z.array(
      z.object({
        address: z.string().regex(evmAddress, "Must be a 0x address"),
      })
    ),
  })
  .superRefine((data, ctx) => {
    if (!data.allowEveryone && data.addresses.length === 0) {
      ctx.addIssue({
        code: "custom",
        message: "Add at least one address, or allow everyone",
        path: ["addresses"],
      })
    }
    // Paid rules settle through a facilitator, which only exists for some
    // networks; free rules never contact one.
    if (
      !isZeroAmount(data.price) &&
      !getFacilitatorNetworks().includes(data.network)
    ) {
      ctx.addIssue({
        code: "custom",
        message:
          "No payment facilitator is available for this network — only free rules can use it",
        path: ["network"],
      })
    }
  })

export type PaymentRuleFormSchemaType = z.infer<typeof paymentRuleFormSchema>
