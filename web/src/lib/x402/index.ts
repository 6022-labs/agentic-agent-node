import { BrowserProvider, type Eip1193Provider } from "ethers"

import {
  ASSET_TRANSFER_METHOD_EIP3009,
  ASSET_TRANSFER_METHOD_PERMIT2,
  type PaymentRule,
} from "@/models/payment-policy"

import { buildEip3009PaymentHeader } from "./eip3009"
import { buildPermit2PaymentHeader } from "./permit2"

// Signatures are cached and reused until shortly before expiry so the wallet
// isn't prompted on every message.
const VALIDITY_SECONDS = 300
const EXPIRY_MARGIN_SECONDS = 30

type CachedHeader = {
  key: string
  header: string
  validBefore: number
}

let cachedHeader: CachedHeader | null = null

function chainIdOf(network: string): number {
  const chainId = Number(network.replace("eip155:", ""))
  if (!Number.isInteger(chainId) || chainId <= 0) {
    throw new Error(`unsupported x402 network: ${network}`)
  }
  return chainId
}

// Signs an x402 "exact" payment for the rule (value "0" = free, never settled)
// and returns the PAYMENT-SIGNATURE header. assetTransferMethod selects EIP-3009
// vs Permit2.
export async function getX402PaymentHeader(
  wallet: Eip1193Provider,
  payerAddress: string,
  rule: PaymentRule,
  value: string = "0"
): Promise<string> {
  const cacheKey =
    `${payerAddress}:${rule.network}:${rule.payTo}:${value}:${rule.assetTransferMethod ?? ASSET_TRANSFER_METHOD_EIP3009}`.toLowerCase()
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
  const chainId = chainIdOf(rule.network)

  const provider = new BrowserProvider(wallet)
  const signer = await provider.getSigner()

  const header =
    rule.assetTransferMethod === ASSET_TRANSFER_METHOD_PERMIT2
      ? await buildPermit2PaymentHeader(
          signer,
          payerAddress,
          rule,
          value,
          chainId,
          validAfter,
          validBefore
        )
      : await buildEip3009PaymentHeader(
          signer,
          payerAddress,
          rule,
          value,
          chainId,
          validAfter,
          validBefore
        )

  cachedHeader = { key: cacheKey, header, validBefore }
  return header
}
