import { BrowserProvider, ethers, type Eip1193Provider } from "ethers"

import { AgentCollectionABI } from "@/abi/agent-collection"
import type { KeyValue } from "@/models/agent-registration"

import { getScaledTxGasOverrides } from "./tx-gas-overrides"

// The agent-node AgentCollectionV1.mint has NO signer/nonce/signature — it is the
// direct (moderator/open) mint. cloneOf is a NullableUint256 { isSet, value }.
const toCloneOfStruct = (cloneOf: number | null) =>
  cloneOf == null
    ? { isSet: false, value: 0n }
    : { isSet: true, value: BigInt(cloneOf) }

const toTuples = (pairs: KeyValue[]) =>
  pairs.map((p) => ({ key: p.key, value: p.value }))

const getContract = async (
  collectionAddress: string,
  walletProvider: Eip1193Provider
) => {
  const provider = new BrowserProvider(walletProvider)
  const signer = await provider.getSigner()
  return {
    contract: new ethers.Contract(
      collectionAddress,
      AgentCollectionABI,
      signer
    ),
    signer,
  }
}

const logAndRethrow = (error: unknown) => {
  try {
    const iface = new ethers.Interface(AgentCollectionABI)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const parsed = iface.parseError((error as any)?.data)
    if (parsed?.name) {
      ;(error as { parsedRevertName?: string }).parsedRevertName = parsed.name
    }
    console.error("AgentCollection call failed", error, parsed)
  } catch {
    console.error("AgentCollection call failed", error)
  }
  throw error
}

export type MintParams = {
  collectionAddress: string
  to: string
  role: string
  name: string
  wallets: string[]
  images: KeyValue[]
  attributes: KeyValue[]
  cloneOf: number | null
}

// Read-only probe mirroring the node's MintPublisher.canMintDirectly rule:
// direct mint iff the collection has no moderators or the sender is one.
export const canMintDirectly = async (
  collectionAddress: string,
  sender: string,
  walletProvider: Eip1193Provider
): Promise<boolean> => {
  const provider = new BrowserProvider(walletProvider)
  const contract = new ethers.Contract(
    collectionAddress,
    AgentCollectionABI,
    provider
  )
  const moderatorCount: bigint = await contract.moderatorCount()
  if (moderatorCount === 0n) return true
  return await contract.isModerator(sender)
}

// Direct mint — used when there are no moderators or the owner is a moderator.
export const mint = async (
  params: MintParams,
  walletProvider: Eip1193Provider
) => {
  const { contract, signer } = await getContract(
    params.collectionAddress,
    walletProvider
  )
  try {
    const tx = await contract.mint(
      params.to,
      params.role,
      params.name,
      params.wallets,
      toTuples(params.images),
      toTuples(params.attributes),
      toCloneOfStruct(params.cloneOf),
      await getScaledTxGasOverrides(signer)
    )
    await tx.wait()
    return tx
  } catch (error) {
    return logAndRethrow(error)
  }
}

// Proposal mint — used when the collection has moderators and the owner is not
// one (a moderator must approve via mintFromProposal afterwards).
export const createMintProposal = async (
  params: MintParams,
  walletProvider: Eip1193Provider
) => {
  const { contract, signer } = await getContract(
    params.collectionAddress,
    walletProvider
  )
  try {
    const tx = await contract.createMintProposal(
      params.to,
      params.role,
      params.name,
      params.wallets,
      toTuples(params.images),
      toTuples(params.attributes),
      toCloneOfStruct(params.cloneOf),
      await getScaledTxGasOverrides(signer)
    )
    await tx.wait()
    return tx
  } catch (error) {
    return logAndRethrow(error)
  }
}
