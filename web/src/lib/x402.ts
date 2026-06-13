import { BrowserProvider, type Eip1193Provider } from "ethers"

import type { PaymentRule } from "@/models/payment-policy"

// Validity window of one signature; signatures are cached and reused until
// shortly before expiry so the wallet isn't prompted on every message.
const VALIDITY_SECONDS = 300
const EXPIRY_MARGIN_SECONDS = 30

const TRANSFER_WITH_AUTHORIZATION_TYPES = {
  TransferWithAuthorization: [
    { name: "from", type: "address" },
    { name: "to", type: "address" },
    { name: "value", type: "uint256" },
    { name: "validAfter", type: "uint256" },
    { name: "validBefore", type: "uint256" },
    { name: "nonce", type: "bytes32" },
  ],
}

type CachedHeader = {
  key: string
  header: string
  validBefore: number
}

let cachedHeader: CachedHeader | null = null

function randomNonce(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(32))
  return (
    "0x" +
    Array.from(bytes)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("")
  )
}

function chainIdOf(network: string): number {
  const chainId = Number(network.replace("eip155:", ""))
  if (!Number.isInteger(chainId) || chainId <= 0) {
    throw new Error(`unsupported x402 network: ${network}`)
  }
  return chainId
}

// getX402PaymentHeader signs an EIP-3009 authorization for the given payment
// rule (value "0" for free rules — never settled on-chain) and returns the
// PAYMENT-SIGNATURE header value.
export async function getX402PaymentHeader(
  wallet: Eip1193Provider,
  payerAddress: string,
  rule: PaymentRule,
  value: string = "0"
): Promise<string> {
  const cacheKey =
    `${payerAddress}:${rule.network}:${rule.payTo}:${value}`.toLowerCase()
  const nowSeconds = Math.floor(Date.now() / 1000)
  if (
    cachedHeader &&
    cachedHeader.key === cacheKey &&
    nowSeconds < cachedHeader.validBefore - EXPIRY_MARGIN_SECONDS
  ) {
    return cachedHeader.header
  }

  const validAfter = nowSeconds - 60
  const validBefore = nowSeconds + VALIDITY_SECONDS

  const authorization = {
    from: payerAddress,
    to: rule.payTo,
    value,
    validAfter: String(validAfter),
    validBefore: String(validBefore),
    nonce: randomNonce(),
  }

  const provider = new BrowserProvider(wallet)
  const signer = await provider.getSigner()
  const signature = await signer.signTypedData(
    {
      name: rule.eip712Name ?? "USDC",
      version: rule.eip712Version ?? "2",
      chainId: chainIdOf(rule.network),
      verifyingContract: rule.asset,
    },
    TRANSFER_WITH_AUTHORIZATION_TYPES,
    authorization
  )

  const payload = {
    x402Version: 2,
    scheme: "exact",
    network: rule.network,
    payload: { signature, authorization },
  }
  const header = btoa(JSON.stringify(payload))

  cachedHeader = { key: cacheKey, header, validBefore }
  return header
}
