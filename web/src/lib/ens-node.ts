import { ethers } from "ethers"

const labelHash = (label: string) => ethers.keccak256(ethers.toUtf8Bytes(label))

const childNode = (parent: string, label: string) =>
  ethers.keccak256(ethers.concat([parent, label]))

// Mirrors AgentEnsRegistry._resolveAgentNode (and the Go helper.ResolveAgentNode):
// parentNode → chainId → collection name → agent label, each step
// keccak256(prev || keccak256(labelBytes)). Labels hash the raw strings —
// intentionally NOT ENS-normalized.
export function resolveAgentNode(
  parentNode: string,
  chainId: number,
  collectionName: string,
  agentLabel: string
): string {
  const chainNode = childNode(parentNode, labelHash(String(chainId)))
  const collectionNode = childNode(chainNode, labelHash(collectionName))
  return childNode(collectionNode, labelHash(agentLabel))
}
