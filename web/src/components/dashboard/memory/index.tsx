import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import AppendMemoryDialog from "@/components/dashboard/memory/dialogs/append-memory-dialog"
import SessionsTable from "@/components/dashboard/memory/sessions-table"
import NotMintedState from "@/components/dashboard/not-minted-state"

export default function MemorySection() {
  const { loading: identityLoading, isMinted } = useAgentIdentity()

  if (!identityLoading && !isMinted) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Conversation memory</CardTitle>
        </CardHeader>
        <CardContent>
          <NotMintedState description="Mint the agent's on-chain identity before seeding conversation memory." />
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Memory sessions</CardTitle>
          <CardDescription>
            Named buckets the agent keeps its history in. The agent picks one per
            conversation and titles new ones itself.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <SessionsTable />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Seed a conversation</CardTitle>
          <CardDescription>
            Inject messages out-of-band — useful for testing how the agent reacts
            to a given history ("false memory").
          </CardDescription>
        </CardHeader>
        <CardContent>
          <AppendMemoryDialog />
        </CardContent>
      </Card>
    </div>
  )
}
