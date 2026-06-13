import axios from "axios"

import type {
  ListPaymentRulesResponse,
  PaymentRule,
} from "@/models/payment-policy"

// GET /config/payment-policy/rules — the pricing rules in table order.
export const listPaymentRules = async () => {
  return await axios.get<ListPaymentRulesResponse>(
    "/config/payment-policy/rules"
  )
}

// POST /config/payment-policy/rules — appends a rule.
export const createPaymentRule = async (rule: PaymentRule) => {
  return await axios.post<PaymentRule>("/config/payment-policy/rules", rule)
}

// PUT /config/payment-policy/rules/:id — replaces a rule's fields.
export const updatePaymentRule = async (id: string, rule: PaymentRule) => {
  return await axios.put<PaymentRule>(
    `/config/payment-policy/rules/${id}`,
    rule
  )
}

// DELETE /config/payment-policy/rules/:id — 204.
export const deletePaymentRule = async (id: string) => {
  return await axios.delete(`/config/payment-policy/rules/${id}`)
}
