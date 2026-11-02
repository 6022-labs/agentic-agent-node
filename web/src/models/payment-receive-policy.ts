// PaymentReceiveRulePrice is what callers matching addressPatterns pay per call, in atomic units; "0" is free.
export type PaymentReceiveRulePrice = {
  price: string
  addressPatterns: string[]
}

// PaymentReceiveRule is the flat view model of one token's rule; the query layer maps the wire config blob.
export type PaymentReceiveRule = {
  id?: string
  name: string
  // The payment method; only X402_METHOD exists today. The form select is
  // locked to it, leaving room for future methods (e.g. Lightning).
  method: string
  network: string
  asset: string
  payTo: string
  // Read-only: detected from the token on-chain at save.
  assetTransferMethod?: string
  eip712Name?: string
  eip712Version?: string
  prices: PaymentReceiveRulePrice[]
}

export type ListPaymentReceiveRulesResponse = {
  rules: PaymentReceiveRule[]
}

// PaymentReceiveRuleConfig is the x402 token in the wire rule's config blob.
export type PaymentReceiveRuleConfig = {
  network: string
  asset: string
  payTo: string
  assetTransferMethod?: string
  eip712Name?: string
  eip712Version?: string
}

// PaymentReceiveRuleWire mirrors the gateway's rule: method-neutral fields plus a
// method-specific config blob.
export type PaymentReceiveRuleWire = {
  id?: string
  name: string
  method: string
  prices: PaymentReceiveRulePrice[]
  config: PaymentReceiveRuleConfig
}

export type ListPaymentReceiveRulesResponseWire = {
  rules: PaymentReceiveRuleWire[]
}

export const X402_METHOD = "x402"

// How an x402 "exact" payment moves the token (mirrors the gateway enum); the
// asset's method is detected on-chain at save, never operator input.
export const ASSET_TRANSFER_METHOD_EIP3009 = "eip3009"
export const ASSET_TRANSFER_METHOD_PERMIT2 = "permit2"

// The policy head; rules are fetched separately. agentWalletAddress is
// exposed to pre-fill rules paying the agent itself.
export type PaymentReceivePolicy = {
  id: string
  agentId: string
  agentWalletAddress: string
  createdAt: string
  updatedAt: string
}

export const EVERYONE_PATTERN = ".*"

// matchesAddressPatterns mirrors the runtime: full-match against the lowercased address.
export function matchesAddressPatterns(
  patterns: string[],
  address: string
): boolean {
  const lower = address.toLowerCase()
  return patterns.some((pattern) => {
    try {
      return new RegExp(`^(?:${pattern})$`).test(lower)
    } catch {
      return false
    }
  })
}

// isPublicPrice reports a price open to everyone, which the runtime advertises.
export function isPublicPrice(price: PaymentReceiveRulePrice): boolean {
  return price.addressPatterns.includes(EVERYONE_PATTERN)
}

// findFreeRule returns the rule granting the address free access, if any.
export function findFreeRule(
  rules: PaymentReceiveRule[],
  address: string
): PaymentReceiveRule | undefined {
  return rules.find((rule) =>
    rule.prices.some(
      (price) =>
        price.price === "0" &&
        matchesAddressPatterns(price.addressPatterns, address)
    )
  )
}

// Max uint256: the largest price the runtime stores.
export const MAX_UINT256 =
  115792089237316195423570985008687907853269984665640564039457584007913129639935n
