import React from "react"

import type { AgentCollection } from "@/models/agent-collection"
import { listAgentCollections } from "@/queries/api/collections.queries"

// Collections are per-chain; the hook (re)loads whenever the selected chain id
// changes. An empty chainId leaves the list null (nothing selected yet).
export function useAgentCollections(chainId: string | undefined) {
  // null = still loading; [] = unavailable/empty -> fall back to manual input.
  const [collections, setCollections] = React.useState<
    AgentCollection[] | null
  >(null)

  const loadCollections = React.useCallback(async (): Promise<
    AgentCollection[]
  > => {
    if (!chainId) {
      setCollections(null)
      return []
    }
    try {
      const { data } = await listAgentCollections(chainId)
      const loaded = data.collections ?? []
      setCollections(loaded)
      return loaded
    } catch {
      setCollections([])
      return []
    }
  }, [chainId])

  React.useEffect(() => {
    setCollections(null)
    void loadCollections()
  }, [loadCollections])

  return { collections, loadCollections }
}
