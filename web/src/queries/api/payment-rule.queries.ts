import axios from "axios"

import {
  X402_METHOD,
  type ListPaymentRulesResponse,
  type ListPaymentRulesResponseWire,
  type PaymentRule,
  type PaymentRuleWire,
} from "@/models/payment-policy"

// toWire nests the flat view model's x402 settings into the method-specific
// config blob the gateway expects.
const toWire = (rule: PaymentRule): PaymentRuleWire => ({
  id: rule.id,
  name: rule.name,
  price: rule.price,
  method: rule.method || X402_METHOD,
  config: {
    network: rule.network,
    asset: rule.asset,
    payTo: rule.payTo,
    eip712Name: rule.eip712Name,
    eip712Version: rule.eip712Version,
    addressPatterns: rule.addressPatterns,
  },
})

// fromWire flattens a wire rule's config back into the view model.
const fromWire = (wire: PaymentRuleWire): PaymentRule => ({
  id: wire.id,
  name: wire.name,
  price: wire.price,
  method: wire.method,
  network: wire.config.network,
  asset: wire.config.asset,
  payTo: wire.config.payTo,
  eip712Name: wire.config.eip712Name,
  eip712Version: wire.config.eip712Version,
  addressPatterns: wire.config.addressPatterns ?? [],
})

// GET /config/payment-policy/rules — the pricing rules in table order. Unlike
// the passthrough queries, this maps the wire rules' config blob into the flat
// view model, so it returns the model directly rather than the axios response.
export const listPaymentRules = async (): Promise<ListPaymentRulesResponse> => {
  const response = await axios.get<ListPaymentRulesResponseWire>(
    "/config/payment-policy/rules"
  )
  return { rules: (response.data.rules ?? []).map(fromWire) }
}

// POST /config/payment-policy/rules — appends a rule.
export const createPaymentRule = async (rule: PaymentRule) => {
  return await axios.post<PaymentRuleWire>(
    "/config/payment-policy/rules",
    toWire(rule)
  )
}

// PUT /config/payment-policy/rules/:id — replaces a rule's fields.
export const updatePaymentRule = async (id: string, rule: PaymentRule) => {
  return await axios.put<PaymentRuleWire>(
    `/config/payment-policy/rules/${id}`,
    toWire(rule)
  )
}

// DELETE /config/payment-policy/rules/:id — 204.
export const deletePaymentRule = async (id: string) => {
  return await axios.delete(`/config/payment-policy/rules/${id}`)
}
