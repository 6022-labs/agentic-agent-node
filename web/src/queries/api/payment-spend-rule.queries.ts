import axios from "axios"

import {
  X402_METHOD,
  type ListPaymentSpendRulesResponse,
  type ListPaymentSpendRulesResponseWire,
  type PaymentSpendRule,
  type PaymentSpendRuleWire,
} from "@/models/payment-spend-policy"

// toWire nests the flat view model's matcher into the method-specific config
// blob the gateway expects.
const toWire = (rule: PaymentSpendRule): PaymentSpendRuleWire => ({
  id: rule.id,
  name: rule.name,
  maxAmountPerPayment: rule.maxAmountPerPayment,
  method: rule.method || X402_METHOD,
  config: {
    networkPatterns: rule.networkPatterns,
    assetPatterns: rule.assetPatterns,
    payToPatterns: rule.payToPatterns,
    assetTransferMethodPatterns: rule.assetTransferMethodPatterns,
  },
})

// fromWire flattens a wire rule's config back into the view model.
const fromWire = (wire: PaymentSpendRuleWire): PaymentSpendRule => ({
  id: wire.id,
  name: wire.name,
  maxAmountPerPayment: wire.maxAmountPerPayment,
  method: wire.method,
  networkPatterns: wire.config?.networkPatterns ?? [],
  assetPatterns: wire.config?.assetPatterns ?? [],
  payToPatterns: wire.config?.payToPatterns ?? [],
  assetTransferMethodPatterns: wire.config?.assetTransferMethodPatterns ?? [],
})

// GET /config/payment-spend-policy/rules — the spend rules in table order.
export const listPaymentSpendRules = async (): Promise<ListPaymentSpendRulesResponse> => {
  const response = await axios.get<ListPaymentSpendRulesResponseWire>(
    "/config/payment-spend-policy/rules"
  )
  return { rules: (response.data.rules ?? []).map(fromWire) }
}

// POST /config/payment-spend-policy/rules — appends a rule.
export const createPaymentSpendRule = async (rule: PaymentSpendRule) => {
  return await axios.post<PaymentSpendRuleWire>("/config/payment-spend-policy/rules", toWire(rule))
}

// PUT /config/payment-spend-policy/rules/:id — replaces a rule's fields.
export const updatePaymentSpendRule = async (id: string, rule: PaymentSpendRule) => {
  return await axios.put<PaymentSpendRuleWire>(
    `/config/payment-spend-policy/rules/${id}`,
    toWire(rule)
  )
}

// DELETE /config/payment-spend-policy/rules/:id — 204.
export const deletePaymentSpendRule = async (id: string) => {
  return await axios.delete(`/config/payment-spend-policy/rules/${id}`)
}
