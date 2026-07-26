import axios from "axios"

import type {
  CreateMcpConnectionRequest,
  ListMcpConnectionsResponse,
  DiscoverMcpServerToolsRequest,
  DiscoverMcpServerToolsResponse,
  McpConnection,
  UpdateMcpConnectionRequest,
} from "@/models/mcp-connection"

// GET /config/mcp-connections
export const listMcpConnections = async () => {
  return await axios.get<ListMcpConnectionsResponse>("/config/mcp-connections")
}

// POST /config/mcp-connections — 201
export const createMcpConnection = async (
  request: CreateMcpConnectionRequest
) => {
  return await axios.post<McpConnection>("/config/mcp-connections", request)
}

// PUT /config/mcp-connections/{id}
export const updateMcpConnection = async (
  id: string,
  request: UpdateMcpConnectionRequest
) => {
  return await axios.put<McpConnection>(
    `/config/mcp-connections/${id}`,
    request
  )
}

// POST /config/mcp-connections/tools — asks the server which tools it exposes.
export const discoverMcpServerTools = async (
  request: DiscoverMcpServerToolsRequest
) => {
  return await axios.post<DiscoverMcpServerToolsResponse>(
    "/config/mcp-connections/tools",
    request
  )
}

// DELETE /config/mcp-connections/{id} — 204
export const deleteMcpConnection = async (id: string) => {
  return await axios.delete(`/config/mcp-connections/${id}`)
}
