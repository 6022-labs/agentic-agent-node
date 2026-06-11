/* eslint-disable react-refresh/only-export-components */
import React from "react"
import axios from "axios"

import { useAuth } from "@/contexts/auth-context"
import type { AgentRegistration } from "@/models/agent-registration"
import { getAgentRegistration } from "@/queries/api/registration.queries"

const DEFAULT_TITLE = "Agent Node Dapp"

type AgentIdentityContextType = {
  registration: AgentRegistration | null
  loading: boolean
  isMinted: boolean
  ensDomain: string | null
  refresh: () => Promise<void>
  setRegistration: (registration: AgentRegistration | null) => void
}

const AgentIdentityContext = React.createContext<AgentIdentityContextType>({
  registration: null,
  loading: false,
  isMinted: false,
  ensDomain: null,
  refresh: async () => {},
  setRegistration: () => {},
})

export const AgentIdentityProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const { jwt } = useAuth()
  const [registration, setRegistration] =
    React.useState<AgentRegistration | null>(null)
  const [loading, setLoading] = React.useState(false)

  const refresh = React.useCallback(async () => {
    if (!jwt) {
      setRegistration(null)
      return
    }
    setLoading(true)
    try {
      const { data } = await getAgentRegistration()
      setRegistration(data)
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        setRegistration(null)
      }
    } finally {
      setLoading(false)
    }
  }, [jwt])

  React.useEffect(() => {
    void refresh()
  }, [refresh])

  const isMinted = registration?.mintStatus === "confirmed"
  const ensDomain =
    isMinted && registration?.ensDomain ? registration.ensDomain : null

  // Reflect the agent's ENS domain in the browser tab title once minted.
  React.useEffect(() => {
    document.title = ensDomain ?? DEFAULT_TITLE
  }, [ensDomain])

  return (
    <AgentIdentityContext.Provider
      value={{
        registration,
        loading,
        isMinted: !!isMinted,
        ensDomain,
        refresh,
        setRegistration,
      }}
    >
      {children}
    </AgentIdentityContext.Provider>
  )
}

export const useAgentIdentity = () => {
  const context = React.useContext(AgentIdentityContext)
  if (!context)
    throw new Error(
      "useAgentIdentity must be used within an AgentIdentityProvider"
    )
  return context
}
