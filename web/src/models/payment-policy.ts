export type PaymentRule = {
  id?: string
  name: string
  network: string
  asset: string
  payTo: string
  price: string
  maxTimeoutSeconds?: number
  eip712Name?: string
  eip712Version?: string
  // Regexes full-matched against the caller's lowercase 0x address;
  // EVERYONE_PATTERN opens the rule to everyone. Price "0" makes the rule a
  // free pass for the matched addresses.
  addressPatterns: string[]
}

export type ListPaymentRulesResponse = {
  rules: PaymentRule[]
}

// The policy head; rules are fetched separately. agentWalletAddress is
// exposed to pre-fill rules paying the agent itself.
export type PaymentPolicy = {
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
  rules: PaymentRule[],
  address: string
): PaymentRule | undefined {
  return rules.find(
    (rule) =>
      rule.price === "0" &&
      matchesAddressPatterns(rule.addressPatterns ?? [], address)
  )
}
