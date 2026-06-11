import { BrowserProvider, ethers, type Eip1193Provider } from "ethers"

import { AgentCollectionABI } from "@/abi/agent-collection"
import { AgentEnsRegistryABI } from "@/abi/agent-ens-registry"
import { getAppConfig } from "@/configs/app-config"
import { resolveAgentNode } from "@/lib/ens-node"
import type { EnsInfo, KeyValue } from "@/models/agent-registration"

import { getScaledTxGasOverrides } from "./tx-gas-overrides"

const logAndRethrow = (error: unknown) => {
  try {
    const iface = new ethers.Interface(AgentEnsRegistryABI)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const parsed = iface.parseError((error as any)?.data)
    if (parsed?.name) {
      ;(error as { parsedRevertName?: string }).parsedRevertName = parsed.name
    }
    console.error("AgentEnsRegistry call failed", error, parsed)
  } catch {
    console.error("AgentEnsRegistry call failed", error)
  }
  throw error
}

// initAgentTexts(collection, tokenId, records[]) on the AgentEnsRegistry. The
// registry address comes from the gateway's /config at boot.
export const initAgentTexts = async (
  collectionAddress: string,
  tokenId: number,
  records: KeyValue[],
  walletProvider: Eip1193Provider
) => {
  const provider = new BrowserProvider(walletProvider)
  const signer = await provider.getSigner()
  const contract = new ethers.Contract(
    getAppConfig().ensRegistryAddress,
    AgentEnsRegistryABI,
    signer
  )

  try {
    const tx = await contract.initAgentTexts(
      collectionAddress,
      BigInt(tokenId),
      records.map((r) => ({ key: r.key, value: r.value })),
      await getScaledTxGasOverrides(signer)
    )
    await tx.wait()
    return tx
  } catch (error) {
    return logAndRethrow(error)
  }
}

export type EnsRecordDrift = { key: string; desired: string; current: string }

export type EnsRecordsState = {
  // False until initAgentTexts ran on this registry (e.g. after a redeploy);
  // repairs must then init instead of setAgentText, which reverts NotProvisioned.
  provisioned: boolean
  drift: EnsRecordDrift[]
}

// Reads the on-chain text records and returns the ones that differ from the
// node's desired values (mirrors the Go EnsPublisher reconciliation, which can
// silently fail at boot when the node wallet can't pay gas).
export const getEnsRecordsDrift = async (
  ensInfo: EnsInfo,
  agentLabel: string,
  walletProvider: Eip1193Provider
): Promise<EnsRecordsState> => {
  const provider = new BrowserProvider(walletProvider)
  const registry = new ethers.Contract(
    getAppConfig().ensRegistryAddress,
    AgentEnsRegistryABI,
    provider
  )
  const collection = new ethers.Contract(
    ensInfo.collectionAddress,
    AgentCollectionABI,
    provider
  )

  const [parentNode, collectionName] = await Promise.all([
    registry.parentNode(),
    collection.name(),
  ])
  const node = resolveAgentNode(
    parentNode,
    getAppConfig().chainId,
    collectionName,
    agentLabel
  )

  const [provisioned, ...current] = await Promise.all([
    registry.provisioned(node) as Promise<boolean>,
    ...ensInfo.records.map(
      (r) => registry.text(node, r.key) as Promise<string>
    ),
  ])
  const drift = ensInfo.records
    .map((r, i) => ({ key: r.key, desired: r.value, current: current[i] }))
    .filter((d) => d.desired !== d.current)
  return { provisioned, drift }
}

// setAgentText from the OWNER's browser wallet — repair path when the node
// wallet couldn't afford its boot-time reconciliation tx.
export const setAgentText = async (
  collectionAddress: string,
  tokenId: number,
  key: string,
  value: string,
  walletProvider: Eip1193Provider
) => {
  const provider = new BrowserProvider(walletProvider)
  const signer = await provider.getSigner()
  const contract = new ethers.Contract(
    getAppConfig().ensRegistryAddress,
    AgentEnsRegistryABI,
    signer
  )

  try {
    const tx = await contract.setAgentText(
      collectionAddress,
      BigInt(tokenId),
      key,
      value,
      await getScaledTxGasOverrides(signer)
    )
    await tx.wait()
    return tx
  } catch (error) {
    return logAndRethrow(error)
  }
}
