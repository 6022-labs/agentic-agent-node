import React from "react"

import { useAgentIdentity } from "@/contexts/agent-identity-context"
import { useWallet } from "@/contexts/wallet-provider-context"
import type { EnsInfo } from "@/models/agent-registration"
import { getAgentEnsInfo } from "@/queries/api/registration.queries"
import {
  getEnsRecordsDrift,
  type EnsRecordDrift,
} from "@/queries/contracts/agent-ens-registry"

// Advisory drift check: compares the node's desired ENS records against the
// chain once per mount and per identity refresh. Failures stay silent — a
// flaky RPC must never raise a false alarm.
export function useEnsRecordsDrift() {
  const { registration, isMinted } = useAgentIdentity()
  const { wallet } = useWallet()
  const [drift, setDrift] = React.useState<EnsRecordDrift[]>([])
  const [provisioned, setProvisioned] = React.useState(true)
  const [ensInfo, setEnsInfo] = React.useState<EnsInfo | null>(null)

  const recheck = React.useCallback(async () => {
    // Pre-confirmation states are the wizard's job; no wallet means no RPC.
    if (!wallet || !isMinted || registration?.ensStatus !== "confirmed") {
      setDrift([])
      return
    }
    try {
      const { data: info } = await getAgentEnsInfo()
      const state = await getEnsRecordsDrift(info, registration.name, wallet)
      setEnsInfo(info)
      setProvisioned(state.provisioned)
      setDrift(state.drift)
    } catch (error) {
      console.warn("ENS records drift check skipped", error)
      setDrift([])
    }
  }, [wallet, isMinted, registration])

  React.useEffect(() => {
    void recheck()
  }, [recheck])

  return { drift, provisioned, ensInfo, recheck }
}
