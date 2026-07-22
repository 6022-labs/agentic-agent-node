import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import AppendMemoryDialog from "@/components/dashboard/memory/dialogs/append-memory-dialog"
import NotMintedState from "@/components/dashboard/not-minted-state"

export default function MemorySection() {
  const { loading: identityLoading, isMinted } = useAgentIdentity()

  return (
    <Card>
      <CardHeader>
        <CardTitle>Conversation memory</CardTitle>
        <CardDescription>
          Seed a conversation with messages out-of-band — useful for testing how
          the agent reacts to a given history ("false memory").
        </CardDescription>
      </CardHeader>
      <CardContent>
        {!identityLoading && !isMinted ? (
          <NotMintedState description="Mint the agent's on-chain identity before seeding conversation memory." />
        ) : (
          <AppendMemoryDialog />
        )}
      </CardContent>
    </Card>
  )
}
