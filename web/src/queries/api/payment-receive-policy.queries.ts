import axios from "axios"

import type { PaymentReceivePolicy } from "@/models/payment-receive-policy"

// GET /config/payment-receive-policy — the policy head; the default policy is
// materialized on first access.
export const getPaymentReceivePolicy = async () => {
  return await axios.get<PaymentReceivePolicy>("/config/payment-receive-policy")
}

// DELETE /config/payment-receive-policy — 204, resets to the default policy.
export const deletePaymentReceivePolicy = async () => {
  return await axios.delete("/config/payment-receive-policy")
}
