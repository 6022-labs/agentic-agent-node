export type AgentCollection = {
  address: string
  name: string
  description: string
  moderatorCount: number
  nextTokenId: number
}

export type ListAgentCollectionsResponse = {
  collections: AgentCollection[]
}
