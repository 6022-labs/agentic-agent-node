// PaymentReceiveRule is the flat view model the UI works with. The wire format is
// method-neutral (name/price/method + a per-method config blob); the query
// layer maps between the two so the table and form keep flat fields.
export type PaymentReceiveRule = {
  id?: string
  name: string
  // The payment method; only X402_METHOD exists today. The form select is
  // locked to it, leaving room for future methods (e.g. Lightning).
  method: string
  network: string
  asset: string
  payTo: string
  price: string
  assetTransferMethod?: string
  eip712Name?: string
  eip712Version?: string
  // Regexes full-matched against the caller's lowercase 0x address;
  // EVERYONE_PATTERN opens the rule to everyone. Price "0" makes the rule a
  // free pass for the matched addresses.
  addressPatterns: string[]
}

export type ListPaymentReceiveRulesResponse = {
  rules: PaymentReceiveRule[]
}

// PaymentReceiveRuleConfig is the x402 method's per-rule settings, carried in the
// wire rule's method-specific config blob.
export type PaymentReceiveRuleConfig = {
  network: string
  asset: string
  payTo: string
  assetTransferMethod?: string
  eip712Name?: string
  eip712Version?: string
  addressPatterns: string[]
}

// PaymentReceiveRuleWire mirrors the gateway's rule: method-neutral fields plus a
// method-specific config blob.
export type PaymentReceiveRuleWire = {
  id?: string
  name: string
  price: string
  method: string
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

// matchesAddressPatterns mirrors the gateway's rule: full-match regexes
// against the lowercase address; invalid patterns never match.
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

// findFreeRule returns the rule granting the address free access, if any.
export function findFreeRule(
  rules: PaymentReceiveRule[],
  address: string
): PaymentReceiveRule | undefined {
  return rules.find(
    (rule) =>
      rule.price === "0" &&
      matchesAddressPatterns(rule.addressPatterns ?? [], address)
  )
}
