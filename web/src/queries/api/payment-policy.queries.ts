import axios from "axios"

import type { PaymentPolicy } from "@/models/payment-policy"

// GET /config/payment-policy — the policy head; the default policy is
// materialized on first access.
export const getPaymentPolicy = async () => {
  return await axios.get<PaymentPolicy>("/config/payment-policy")
}

// DELETE /config/payment-policy — 204, resets to the default policy.
export const deletePaymentPolicy = async () => {
  return await axios.delete("/config/payment-policy")
}
