import axios from "axios"

import type {
  CreateMcpConnectionRequest,
  ListMcpConnectionsResponse,
  McpConnection,
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

// DELETE /config/mcp-connections/{id} — 204
export const deleteMcpConnection = async (id: string) => {
  return await axios.delete(`/config/mcp-connections/${id}`)
}
