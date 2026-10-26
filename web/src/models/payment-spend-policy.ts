import { X402_METHOD } from "@/models/payment-receive-policy"

// PaymentSpendRule is the flat view model the UI works with. The wire format is
// method-neutral (name/maxAmountPerPayment/method + a per-method config blob); the query
// layer maps between the two so the table and form keep flat fields.
export type PaymentSpendRule = {
  id?: string
  name: string
  // The payment method; only X402_METHOD exists today.
  method: string
  // The most this rule will ever sign for, in the asset's atomic units.
  // "0" denies every priced quote it matches; there is no unlimited sentinel.
  maxAmountPerPayment: string
  // A recurring budget: 0 seconds leaves the rule bounded per payment only.
  windowSeconds: number
  maxAmountPerWindow: string
  // Read-only: what the next payment is measured against. The runtime reports a lapsed
  // period as nothing spent, since it only rolls over on the next payment.
  windowSpent: string
  windowResetsAt?: string
  // A rule authorizes a quote on this CAIP-2 network and token, paid to a recipient
  // a payToPatterns regex full-matches; the smallest matching cap wins.
  network: string
  asset: string
  // Unset accepts either transfer method.
  assetTransferMethod?: string
  payToPatterns: string[]
}

export type ListPaymentSpendRulesResponse = {
  rules: PaymentSpendRule[]
}

// PaymentSpendRuleConfig is the x402 method's matcher, carried in the wire rule's
// method-specific config blob.
export type PaymentSpendRuleConfig = {
  network: string
  asset: string
  assetTransferMethod?: string
  payToPatterns: string[]
}

// PaymentSpendRuleWire mirrors the gateway's spend rule.
export type PaymentSpendRuleWire = {
  id?: string
  name: string
  maxAmountPerPayment: string
  windowSeconds?: number
  maxAmountPerWindow?: string
  windowSpent?: string
  windowResetsAt?: string
  method: string
  config: PaymentSpendRuleConfig
}

export type ListPaymentSpendRulesResponseWire = {
  rules: PaymentSpendRuleWire[]
}

// The spend policy head; rules are fetched separately.
export type PaymentSpendPolicy = {
  id: string
  agentId: string
  createdAt: string
  updatedAt: string
}

export { X402_METHOD }

// ANY_PATTERN opens one facet of the matcher to everything.
export const ANY_PATTERN = ".*"

// Max uint256: no cap at all.
export const UNBOUNDED_MAX_AMOUNT =
  "115792089237316195423570985008687907853269984665640564039457584007913129639935"

// isUnbounded reports whether a cap leaves the agent effectively unlimited.
export function isUnbounded(maxAmountPerPayment: string): boolean {
  return maxAmountPerPayment === UNBOUNDED_MAX_AMOUNT
}

// isUnrestricted reports whether a rule pays any amount of its token to anyone.
export function isUnrestricted(rule: PaymentSpendRule): boolean {
  return (
    isUnbounded(rule.maxAmountPerPayment) &&
    (rule.payToPatterns?.includes(ANY_PATTERN) ?? false)
  )
}
