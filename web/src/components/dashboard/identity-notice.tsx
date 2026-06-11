import { Fingerprint, ArrowRight } from "lucide-react"
import { Link } from "react-router-dom"

import { useAgentIdentity } from "@/contexts/agent-identity-context"

// Slim, non-blocking callout: the API works without an identity, but a missing
// one is the most likely reason the agent behaves unexpectedly.
export default function IdentityNotice() {
  const { loading, isMinted } = useAgentIdentity()

  if (loading || isMinted) return null

  return (
    <Link
      to="/dashboard/identity"
      className="mb-4 flex shrink-0 items-center gap-2 rounded-lg border bg-muted/50 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted"
    >
      <Fingerprint className="size-4 shrink-0" />
      <span className="flex-1">
        This agent has no on-chain identity yet — finish the setup.
      </span>
      <ArrowRight className="size-4 shrink-0" />
    </Link>
  )
}
