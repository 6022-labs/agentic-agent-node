export type McpConnection = {
  id: string
  name: string
  url: string
  // Regex patterns, full-matched against tool names; empty allows every tool.
  allowedToolPatterns: string[]
  headerAuthKeys: string[]
  createdAt: string
  updatedAt: string
}

export type ListMcpConnectionsResponse = {
  connections: McpConnection[]
}

export type CreateMcpConnectionRequest = {
  name: string
  url: string
  allowedToolPatterns: string[]
  headerAuth?: Record<string, string>
}

// headerAuth fully replaces the stored set: an omitted header is dropped, and one
// sent with a blank value keeps the value already in the vault.
export type UpdateMcpConnectionRequest = CreateMcpConnectionRequest

export type McpServerTool = {
  name: string
  description: string
}

// Every field is optional: connectionId names a stored connection each missing
// field falls back to (its url, its vault header values).
export type DiscoverMcpServerToolsRequest = {
  connectionId?: string
  url?: string
  headerAuth?: Record<string, string>
}

export type DiscoverMcpServerToolsResponse = {
  tools: McpServerTool[]
}
