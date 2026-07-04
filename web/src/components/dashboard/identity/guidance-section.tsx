import React from "react"
import { toast } from "sonner"
import { Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { getAgentGuidance, setAgentGuidance } from "@/queries/api/guidance.queries"
import { extractErrorMessage } from "@/lib/api-errors"

// Guidance = the owner-editable layer of the system prompt (soul stays immutable).
export default function GuidanceSection() {
  const [loading, setLoading] = React.useState(true)
  const [content, setContent] = React.useState("")
  const [savedContent, setSavedContent] = React.useState("")
  const [saving, setSaving] = React.useState(false)

  React.useEffect(() => {
    const load = async () => {
      try {
        const { data } = await getAgentGuidance()
        setContent(data.content)
        setSavedContent(data.content)
      } catch (error) {
        toast.error(extractErrorMessage(error, "Failed to load the guidance"))
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  const save = async () => {
    setSaving(true)
    try {
      const { data } = await setAgentGuidance({ content })
      setContent(data.content)
      setSavedContent(data.content)
      toast.success("Guidance updated")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to update the guidance"))
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Guidance</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-9 w-24" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Guidance</CardTitle>
        <CardDescription>
          Your standing instructions, applied to every conversation. Edit
          anytime — they steer behavior but never override the identity.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <Textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="e.g. Always answer concisely. Prefer French. Never discuss competitors."
          className="min-h-32 font-mono text-sm"
        />
        <div className="flex justify-end">
          <Button onClick={save} disabled={saving || content === savedContent}>
            {saving ? <Loader2 className="animate-spin" /> : null}
            Save
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
