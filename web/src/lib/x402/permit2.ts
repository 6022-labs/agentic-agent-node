import { Contract, MaxUint256, type Signer } from "ethers"

import type { PaymentRule } from "@/models/payment-policy"

import { randomBytes32Hex } from "./nonce"

// Uniswap's canonical Permit2, deployed at the same address on every EVM chain
// via CREATE2. Permit2's EIP-712 domain has NO version field.
const PERMIT2_ADDRESS = "0x000000000022D473030F116dDEE9F6B43aC78BA3"
const PERMIT2_DOMAIN_NAME = "Permit2"
// The x402 exact-permit2 proxy the signature authorizes to pull the tokens.
const PERMIT2_SPENDER = "0x402085c248EeA27D92E8b30b2C58ed07f9E20001"

const PERMIT_WITNESS_TRANSFER_FROM_TYPES = {
  PermitWitnessTransferFrom: [
    { name: "permitted", type: "TokenPermissions" },
    { name: "spender", type: "address" },
    { name: "nonce", type: "uint256" },
    { name: "deadline", type: "uint256" },
    { name: "witness", type: "Witness" },
  ],
  TokenPermissions: [
    { name: "token", type: "address" },
    { name: "amount", type: "uint256" },
  ],
  Witness: [
    { name: "to", type: "address" },
    { name: "validAfter", type: "uint256" },
  ],
}

const ERC20_ALLOWANCE_ABI = [
  "function allowance(address owner, address spender) view returns (uint256)",
  "function approve(address spender, uint256 amount) returns (bool)",
]

// A full-uint256 unordered nonce, decimal-encoded.
function randomPermit2Nonce(): string {
  return BigInt(randomBytes32Hex()).toString()
}

// One-time max Permit2 allowance, only for paid settlement (free rules never
// settle). Sends a tx only when the current allowance falls short.
async function ensurePermit2Approval(
  signer: Signer,
  payerAddress: string,
  token: string,
  amount: bigint
): Promise<void> {
  const erc20 = new Contract(token, ERC20_ALLOWANCE_ABI, signer)
  const allowance = (await erc20.allowance(
    payerAddress,
    PERMIT2_ADDRESS
  )) as bigint
  if (allowance >= amount) {
    return
  }
  const tx = await erc20.approve(PERMIT2_ADDRESS, MaxUint256)
  await tx.wait()
}

// Signs a Permit2 PermitWitnessTransferFrom over Permit2's fixed domain (the
// recipient and earliest-spend time go in the witness). Paid amounts approve first.
export async function buildPermit2PaymentHeader(
  signer: Signer,
  payerAddress: string,
  rule: PaymentRule,
  value: string,
  chainId: number,
  validAfter: number,
  validBefore: number
): Promise<string> {
  if (value !== "0") {
    await ensurePermit2Approval(signer, payerAddress, rule.asset, BigInt(value))
  }

  const permit2Authorization = {
    permitted: { token: rule.asset, amount: value },
    from: payerAddress,
    spender: PERMIT2_SPENDER,
    nonce: randomPermit2Nonce(),
    deadline: String(validBefore),
    witness: { to: rule.payTo, validAfter: String(validAfter) },
  }

  const signature = await signer.signTypedData(
    {
      name: PERMIT2_DOMAIN_NAME,
      chainId,
      verifyingContract: PERMIT2_ADDRESS,
    },
    PERMIT_WITNESS_TRANSFER_FROM_TYPES,
    permit2Authorization
  )

  const payload = {
    x402Version: 2,
    scheme: "exact",
    network: rule.network,
    payload: { signature, permit2Authorization },
  }
  return btoa(JSON.stringify(payload))
}
