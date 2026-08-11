export type SelfTool = {
  id: string
  name: string
  description: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export type ListSelfToolsResponse = {
  tools: SelfTool[]
}

export type UpdateSelfToolRequest = {
  enabled: boolean
}
