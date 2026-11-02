// Display helpers for the x402 pricing table. Raw values stay canonical
// (CAIP-2 ids, atomic units, full addresses) — these only shape what the
// operator reads.

import { truncateEthAddress } from "@/lib/utils"
import { findEvmNetwork } from "@/models/networks"

// Known tokens by lowercase contract address — enough to render the prices
// the console itself sets up; anything else falls back to atomic units.
const KNOWN_TOKENS: Record<string, { symbol: string; decimals: number }> = {
  // Circle USDC: Polygon Amoy, Polygon, Base, Ethereum.
  "0x41e94eb019c0762f9bfcf9fb1e58725bfb0e7582": { symbol: "USDC", decimals: 6 },
  "0x3c499c542cef5e3811e1192ce70d8cc03d5c3359": { symbol: "USDC", decimals: 6 },
  "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913": { symbol: "USDC", decimals: 6 },
  "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": { symbol: "USDC", decimals: 6 },
}

// knownTokenSymbol is the symbol of a token the console knows, else undefined.
export function knownTokenSymbol(asset: string): string | undefined {
  return KNOWN_TOKENS[asset.toLowerCase()]?.symbol
}

export function formatNetwork(network: string): string {
  return findEvmNetwork(network)?.name ?? network
}

// formatPrice renders "0.01 $USDC" for known tokens and "10000 units"
// otherwise. Zero is the caller's "Free" — handled by the view, not here.
export function formatPrice(price: string, asset: string): string {
  const token = KNOWN_TOKENS[asset.toLowerCase()]
  if (!token) {
    return `${price} units`
  }
  return `${formatAtomicAmount(price, token.decimals)} $${token.symbol}`
}

export function formatAtomicAmount(amount: string, decimals: number): string {
  let value: bigint
  try {
    value = BigInt(amount)
  } catch {
    return amount
  }
  const base = 10n ** BigInt(decimals)
  const whole = value / base
  const fraction = (value % base).toString().padStart(decimals, "0")
  const trimmed = fraction.replace(/0+$/, "")
  return trimmed.length ? `${whole}.${trimmed}` : whole.toString()
}

// parseTokenAmount converts a human amount ("0.01") into atomic units
// ("10000" for 6 decimals). Returns null for malformed input or more
// fraction digits than the token carries.
export function parseTokenAmount(
  amount: string,
  decimals: number
): string | null {
  const match = /^(\d+)(?:\.(\d+))?$/.exec(amount.trim())
  if (!match) {
    return null
  }
  const fraction = match[2] ?? ""
  if (fraction.length > decimals) {
    return null
  }
  const atomic =
    BigInt(match[1]) * 10n ** BigInt(decimals) +
    BigInt(fraction.padEnd(decimals, "0") || "0")
  return atomic.toString()
}

// isZeroAmount accepts "0", "0.0", "000" — a free row needs no decimals.
export function isZeroAmount(amount: string): boolean {
  return /^0+(\.0+)?$/.test(amount.trim())
}

const evmAddress = /^0x[0-9a-fA-F]{40}$/

export function isPlainAddressPattern(pattern: string): boolean {
  return evmAddress.test(pattern)
}

const shortPattern = (pattern: string): string =>
  isPlainAddressPattern(pattern)
    ? truncateEthAddress(pattern)
    : pattern.length > 14
      ? `${pattern.slice(0, 14)}…`
      : pattern

// audienceLabel names who a pattern list covers in a few characters.
export const audienceLabel = (
  patterns: string[],
  anyoneLabel: string
): string =>
  patterns.includes(".*")
    ? anyoneLabel
    : patterns.length === 1
      ? shortPattern(patterns[0])
      : `${patterns.length} callers`
