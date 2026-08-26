import React from "react"
import { toast } from "sonner"
import { Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { Conversation } from "@/models/conversation"
import {
  createConversation,
  updateConversation,
} from "@/queries/api/conversations.queries"
import { extractErrorMessage } from "@/lib/api-errors"

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  // A conversation means rename; nothing means create.
  conversation?: Conversation
  onSaved: () => void
}

export default function SaveSessionDialog({
  open,
  onOpenChange,
  conversation,
  onSaved,
}: Props) {
  const [name, setName] = React.useState("")
  const [submitting, setSubmitting] = React.useState(false)

  React.useEffect(() => {
    if (open) setName(conversation?.name ?? "")
  }, [open, conversation])

  const submit = async () => {
    const trimmed = name.trim()
    if (!trimmed) {
      toast.error("A name is required")
      return
    }

    setSubmitting(true)
    try {
      if (conversation) {
        await updateConversation(conversation.conversationId, trimmed)
        toast.success("Session renamed")
      } else {
        await createConversation(trimmed)
        toast.success("Session created")
      }
      onOpenChange(false)
      onSaved()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to save the session"))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {conversation ? "Rename session" : "New memory session"}
          </DialogTitle>
          <DialogDescription>
            A short title the agent uses to decide which session a conversation
            belongs to.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-1.5">
          <Label htmlFor="session-name">Name</Label>
          <Input
            id="session-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="ERC20 mint audit"
          />
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={submitting}
          >
            Cancel
          </Button>
          <Button onClick={submit} disabled={submitting}>
            {submitting ? <Loader2 className="animate-spin" /> : null}
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
