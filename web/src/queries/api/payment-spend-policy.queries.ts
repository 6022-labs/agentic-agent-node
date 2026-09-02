import axios from "axios"

import type { PaymentSpendPolicy } from "@/models/payment-spend-policy"

// GET /config/payment-spend-policy — the policy head; the catch-all policy is
// materialized on first access.
export const getPaymentSpendPolicy = async () => {
  return await axios.get<PaymentSpendPolicy>("/config/payment-spend-policy")
}

// DELETE /config/payment-spend-policy — 204, resets to the seeded catch-all.
export const deletePaymentSpendPolicy = async () => {
  return await axios.delete("/config/payment-spend-policy")
}
