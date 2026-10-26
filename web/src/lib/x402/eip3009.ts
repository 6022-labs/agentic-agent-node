import { type Signer } from "ethers"

import type { PaymentReceiveRule } from "@/models/payment-receive-policy"

import { randomBytes32Hex } from "./nonce"

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

// Signs an EIP-3009 transferWithAuthorization under the token's own EIP-712 domain.
export async function buildEip3009PaymentHeader(
  signer: Signer,
  payerAddress: string,
  rule: PaymentReceiveRule,
  value: string,
  chainId: number,
  validAfter: number,
  validBefore: number
): Promise<string> {
  const authorization = {
    from: payerAddress,
    to: rule.payTo,
    value,
    validAfter: String(validAfter),
    validBefore: String(validBefore),
    nonce: randomBytes32Hex(),
  }

  const signature = await signer.signTypedData(
    {
      name: rule.eip712Name ?? "USDC",
      version: rule.eip712Version ?? "2",
      chainId,
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
  return btoa(JSON.stringify(payload))
}
