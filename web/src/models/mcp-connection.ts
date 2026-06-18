export type McpConnection = {
  id: string
  name: string
  url: string
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
  authHeaderName?: string
  authHeaderValue?: string
}
