import React from "react"
import axios from "axios"
import { toast } from "sonner"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { getAgentSoul } from "@/queries/api/soul.queries"
import { extractErrorMessage } from "@/lib/api-errors"

// Soul = the immutable persona seeded at mint; read-only by design.
export default function SoulSection() {
  const [loading, setLoading] = React.useState(true)
  const [soul, setSoul] = React.useState<string | null>(null)

  React.useEffect(() => {
    const load = async () => {
      try {
        const { data } = await getAgentSoul()
        setSoul(data.soul)
      } catch (error) {
        if (!(axios.isAxiosError(error) && error.response?.status === 404)) {
          toast.error(extractErrorMessage(error, "Failed to load the soul"))
        }
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Soul</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-24 w-full" />
        </CardContent>
      </Card>
    )
  }

  // No identity yet, or not minted — the soul is born at mint.
  if (!soul) return null

  return (
    <Card>
      <CardHeader>
        <CardTitle>Soul</CardTitle>
        <CardDescription>
          The agent's persona, generated from its on-chain identity at mint.
          Read-only — use Guidance below to steer behavior.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <pre className="rounded-lg border bg-muted/40 p-4 font-mono text-sm whitespace-pre-wrap">
          {soul}
        </pre>
      </CardContent>
    </Card>
  )
}
