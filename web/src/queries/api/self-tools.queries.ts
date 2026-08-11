import axios from "axios"

import type {
  ListSelfToolsResponse,
  SelfTool,
  UpdateSelfToolRequest,
} from "@/models/self-tool"

// GET /config/self-tools
export const listSelfTools = async () => {
  return await axios.get<ListSelfToolsResponse>("/config/self-tools")
}

// PUT /config/self-tools/{id}
export const updateSelfTool = async (
  id: string,
  request: UpdateSelfToolRequest
) => {
  return await axios.put<SelfTool>(`/config/self-tools/${id}`, request)
}
