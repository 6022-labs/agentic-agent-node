import { X402_METHOD } from "@/models/payment-receive-policy"

// PaymentSpendRule is the flat view model the UI works with. The wire format is
// method-neutral (name/maxAmount/method + a per-method config blob); the query
// layer maps between the two so the table and form keep flat fields.
export type PaymentSpendRule = {
  id?: string
  name: string
  // The payment method; only X402_METHOD exists today.
  method: string
  // The most this rule will ever sign for, in the asset's atomic units.
  // "0" denies every priced quote it matches; there is no unlimited sentinel.
  maxAmount: string
  // Regexes full-matched against the quote's facets. A rule authorizes a quote
  // only when all four match, and the effective cap is the smallest maxAmount
  // among the rules that do.
  networkPatterns: string[]
  assetPatterns: string[]
  payToPatterns: string[]
  assetTransferMethodPatterns: string[]
}

export type ListPaymentSpendRulesResponse = {
  rules: PaymentSpendRule[]
}

// PaymentSpendRuleConfig is the x402 method's matcher, carried in the wire rule's
// method-specific config blob.
export type PaymentSpendRuleConfig = {
  networkPatterns: string[]
  assetPatterns: string[]
  payToPatterns: string[]
  assetTransferMethodPatterns: string[]
}

// PaymentSpendRuleWire mirrors the gateway's spend rule.
export type PaymentSpendRuleWire = {
  id?: string
  name: string
  maxAmount: string
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

// Max uint256: what the seeded catch-all rule carries, i.e. no cap at all.
export const UNBOUNDED_MAX_AMOUNT =
  "115792089237316195423570985008687907853269984665640564039457584007913129639935"

// isUnbounded reports whether a cap leaves the agent effectively unlimited.
export function isUnbounded(maxAmount: string): boolean {
  return maxAmount === UNBOUNDED_MAX_AMOUNT
}

// isOpenToEverything reports whether a rule constrains nothing at all — the
// state a freshly seeded agent starts in.
export function isOpenToEverything(rule: PaymentSpendRule): boolean {
  return (
    isUnbounded(rule.maxAmount) &&
    [
      rule.networkPatterns,
      rule.assetPatterns,
      rule.payToPatterns,
      rule.assetTransferMethodPatterns,
    ].every((patterns) => patterns?.includes(ANY_PATTERN) ?? false)
  )
}
