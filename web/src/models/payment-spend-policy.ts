import { X402_METHOD } from "@/models/payment-receive-policy"

// PaymentSpendRuleWindow is one recurring budget of a rule, in the token's atomic units.
export type PaymentSpendRuleWindow = {
  windowSeconds: number
  maxAmount: string
  // Read-only: what the next payment is measured against. The runtime reports a lapsed
  // period as nothing spent, since it only rolls over on the next payment.
  spent?: string
  resetsAt?: string
}

// PaymentSpendRule is the flat view model the UI works with: one rule per token. The wire
// format is method-neutral (name/limits/method + a per-method config blob); the query
// layer maps between the two so the table and form keep flat fields.
export type PaymentSpendRule = {
  id?: string
  name: string
  // The payment method; only X402_METHOD exists today.
  method: string
  // Optional cap on one payment; "0" denies every priced quote. Unset leaves the windows
  // as the only bound, and a rule always has a cap, a window, or both.
  maxAmountPerPayment?: string
  windows: PaymentSpendRuleWindow[]
  // The token this rule limits: a CAIP-2 network and an ERC-20 contract.
  network: string
  asset: string
  // Unset accepts either transfer method.
  assetTransferMethod?: string
}

export type ListPaymentSpendRulesResponse = {
  rules: PaymentSpendRule[]
}

// PaymentSpendRuleConfig is the x402 method's token, carried in the wire rule's
// method-specific config blob.
type PaymentSpendRuleConfig = {
  network: string
  asset: string
  assetTransferMethod?: string
}

// PaymentSpendRuleWire mirrors the gateway's spend rule.
export type PaymentSpendRuleWire = {
  id?: string
  name: string
  maxAmountPerPayment?: string
  windows: PaymentSpendRuleWindow[]
  method: string
  config: PaymentSpendRuleConfig
}

export type ListPaymentSpendRulesResponseWire = {
  rules: PaymentSpendRuleWire[]
}

// The spend policy head: who the agent may pay, whatever the token. Rules are fetched separately.
export type PaymentSpendPolicy = {
  id: string
  agentId: string
  // Payee regexes, full-matched against the lowercase address; empty pays nobody.
  payToPatterns: string[]
  createdAt: string
  updatedAt: string
}

export { X402_METHOD }

// ANY_PATTERN allows every payee.
export const ANY_PATTERN = ".*"

// Max uint256: no cap at all.
export const UNBOUNDED_MAX_AMOUNT =
  "115792089237316195423570985008687907853269984665640564039457584007913129639935"

// WINDOW_PRESETS are the periods the form offers in one click.
export const WINDOW_PRESETS: Array<{ label: string; seconds: number }> = [
  { label: "Hour", seconds: 3600 },
  { label: "Day", seconds: 86400 },
  { label: "Week", seconds: 604800 },
  { label: "30 days", seconds: 2592000 },
]

// isUnbounded reports whether a cap leaves the agent effectively unlimited.
export function isUnbounded(maxAmount: string | undefined): boolean {
  return maxAmount === UNBOUNDED_MAX_AMOUNT
}

// isUnrestricted reports whether a rule pays any amount of its token, with no window to slow it.
export function isUnrestricted(rule: PaymentSpendRule): boolean {
  return (
    rule.windows.every((window) => isUnbounded(window.maxAmount)) &&
    (rule.maxAmountPerPayment === undefined ||
      isUnbounded(rule.maxAmountPerPayment))
  )
}
