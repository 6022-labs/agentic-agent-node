import axios from "axios"

import type { ListAgentCollectionsResponse } from "@/models/agent-collection"

// GET /agent/collections?chainId=… — collections the chain's on-chain manager governs.
export const listAgentCollections = async (chainId: string) => {
  return await axios.get<ListAgentCollectionsResponse>("/agent/collections", {
    params: { chainId },
  })
}
