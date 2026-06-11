import React from "react"

import type { AgentCollection } from "@/models/agent-collection"
import { listAgentCollections } from "@/queries/api/collections.queries"

export function useAgentCollections() {
  // null = still loading; [] = unavailable/empty -> fall back to manual input.
  const [collections, setCollections] = React.useState<
    AgentCollection[] | null
  >(null)

  const loadCollections = React.useCallback(async (): Promise<
    AgentCollection[]
  > => {
    try {
      const { data } = await listAgentCollections()
      const loaded = data.collections ?? []
      setCollections(loaded)
      return loaded
    } catch {
      setCollections([])
      return []
    }
  }, [])

  React.useEffect(() => {
    void loadCollections()
  }, [loadCollections])

  return { collections, loadCollections }
}
