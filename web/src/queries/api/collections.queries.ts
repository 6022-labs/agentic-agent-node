import axios from "axios"

import type { ListAgentCollectionsResponse } from "@/models/agent-collection"

// GET /agent/collections — collections the on-chain manager governs.
export const listAgentCollections = async () => {
  return await axios.get<ListAgentCollectionsResponse>("/agent/collections")
}
