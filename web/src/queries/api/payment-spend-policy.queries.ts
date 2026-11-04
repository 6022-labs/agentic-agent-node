import axios from "axios"

import type { PaymentSpendPolicy } from "@/models/payment-spend-policy"

// GET /config/payment-spend-policy — the policy head and its payees; materialized on first access.
export const getPaymentSpendPolicy = async () => {
  return await axios.get<PaymentSpendPolicy>("/config/payment-spend-policy")
}

// PUT /config/payment-spend-policy — replaces the payees; an empty list pays nobody.
export const updatePaymentSpendPolicy = async (payToPatterns: string[]) => {
  return await axios.put<PaymentSpendPolicy>("/config/payment-spend-policy", {
    payToPatterns,
  })
}

// DELETE /config/payment-spend-policy — 204, removes every rule and payee.
export const deletePaymentSpendPolicy = async () => {
  return await axios.delete("/config/payment-spend-policy")
}
