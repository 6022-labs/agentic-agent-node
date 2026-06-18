import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import AppendMemoryDialog from "@/components/dashboard/memory/dialogs/append-memory-dialog"

export default function MemorySection() {
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
        <AppendMemoryDialog />
      </CardContent>
    </Card>
  )
}
