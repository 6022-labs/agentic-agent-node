import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import SessionsTable from "@/components/dashboard/memory/sessions-table"
import NotMintedState from "@/components/dashboard/not-minted-state"

export default function MemorySection() {
  const { loading: identityLoading, isMinted } = useAgentIdentity()

  return (
    <Card>
      <CardHeader>
        <CardTitle>Memory sessions</CardTitle>
        <CardDescription>
          Named buckets the agent keeps its history in. It picks one per
          conversation and titles new ones itself; you can add, rename and seed
          them here.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {!identityLoading && !isMinted ? (
          <NotMintedState description="Mint the agent's on-chain identity before managing conversation memory." />
        ) : (
          <SessionsTable />
        )}
      </CardContent>
    </Card>
  )
}
