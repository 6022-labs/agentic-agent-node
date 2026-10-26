import axios from "axios"

import {
  X402_METHOD,
  type ListPaymentReceiveRulesResponse,
  type ListPaymentReceiveRulesResponseWire,
  type PaymentReceiveRule,
  type PaymentReceiveRuleWire,
} from "@/models/payment-receive-policy"

// toWire nests the flat view model's x402 settings into the method-specific
// config blob the gateway expects.
const toWire = (rule: PaymentReceiveRule): PaymentReceiveRuleWire => ({
  id: rule.id,
  name: rule.name,
  price: rule.price,
  method: rule.method || X402_METHOD,
  config: {
    network: rule.network,
    asset: rule.asset,
    payTo: rule.payTo,
    addressPatterns: rule.addressPatterns,
  },
})

// fromWire flattens a wire rule's config back into the view model.
const fromWire = (wire: PaymentReceiveRuleWire): PaymentReceiveRule => ({
  id: wire.id,
  name: wire.name,
  price: wire.price,
  method: wire.method,
  network: wire.config.network,
  asset: wire.config.asset,
  payTo: wire.config.payTo,
  assetTransferMethod: wire.config.assetTransferMethod,
  eip712Name: wire.config.eip712Name,
  eip712Version: wire.config.eip712Version,
  addressPatterns: wire.config.addressPatterns ?? [],
})

// GET /config/payment-receive-policy/rules — the pricing rules in table order. Unlike
// the passthrough queries, this maps the wire rules' config blob into the flat
// view model, so it returns the model directly rather than the axios response.
export const listPaymentReceiveRules = async (): Promise<ListPaymentReceiveRulesResponse> => {
  const response = await axios.get<ListPaymentReceiveRulesResponseWire>(
    "/config/payment-receive-policy/rules"
  )
  return { rules: (response.data.rules ?? []).map(fromWire) }
}

// POST /config/payment-receive-policy/rules — appends a rule.
export const createPaymentReceiveRule = async (rule: PaymentReceiveRule) => {
  return await axios.post<PaymentReceiveRuleWire>(
    "/config/payment-receive-policy/rules",
    toWire(rule)
  )
}

// PUT /config/payment-receive-policy/rules/:id — replaces a rule's fields.
export const updatePaymentReceiveRule = async (id: string, rule: PaymentReceiveRule) => {
  return await axios.put<PaymentReceiveRuleWire>(
    `/config/payment-receive-policy/rules/${id}`,
    toWire(rule)
  )
}

// DELETE /config/payment-receive-policy/rules/:id — 204.
export const deletePaymentReceiveRule = async (id: string) => {
  return await axios.delete(`/config/payment-receive-policy/rules/${id}`)
}
