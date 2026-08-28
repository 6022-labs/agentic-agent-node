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
import SessionMessagesForm from "@/components/dashboard/memory/forms/session-messages-form"
import {
  emptyMessage,
  type DraftMessage,
} from "@/components/dashboard/memory/forms/session-messages-form/draft-message"
import type { Conversation } from "@/models/conversation"
import {
  appendConversationMessages,
  createConversation,
  updateConversation,
} from "@/queries/api/conversations.queries"
import { extractErrorMessage } from "@/lib/api-errors"

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  // A conversation means edit; nothing means create.
  conversation?: Conversation
  onSaved: () => void
}

export default function SessionDialog({
  open,
  onOpenChange,
  conversation,
  onSaved,
}: Props) {
  const [name, setName] = React.useState("")
  const [messages, setMessages] = React.useState<DraftMessage[]>([
    emptyMessage(),
  ])
  const [submitting, setSubmitting] = React.useState(false)

  React.useEffect(() => {
    if (!open) return
    setName(conversation?.name ?? "")
    setMessages([emptyMessage()])
  }, [open, conversation])

  const seed = async (conversationId: string, drafts: DraftMessage[]) => {
    const now = Date.now()
    const { data } = await appendConversationMessages(
      conversationId,
      drafts.map((m, index) => ({
        role: m.role,
        content: m.content,
        // Stagger timestamps so the injected order is preserved on read.
        timestamp: new Date(now + index * 1000).toISOString(),
      }))
    )
    return data.inserted.length
  }

  const submit = async () => {
    const trimmed = name.trim()
    if (!trimmed) {
      toast.error("A name is required")
      return
    }

    const drafts = messages.filter((m) => m.content.trim().length > 0)
    const renamed = trimmed !== (conversation?.name ?? "")
    if (conversation && !renamed && drafts.length === 0) {
      toast.error("Nothing to save")
      return
    }

    setSubmitting(true)
    try {
      let conversationId = conversation?.conversationId
      if (!conversation) {
        const { data } = await createConversation(trimmed)
        conversationId = data.conversationId
        toast.success("Session created")
      } else if (renamed) {
        await updateConversation(conversation.conversationId, trimmed)
        toast.success("Session renamed")
      }

      if (drafts.length > 0 && conversationId) {
        try {
          const inserted = await seed(conversationId, drafts)
          toast.success(`Injected ${inserted} message(s)`)
        } catch (error) {
          // The session itself is saved, so the dialog closes rather than
          // risking a second create on retry; seeding is reachable by editing.
          toast.error(extractErrorMessage(error, "Failed to inject messages"))
        }
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
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {conversation ? "Edit memory session" : "New memory session"}
          </DialogTitle>
          <DialogDescription>
            A short title the agent uses to decide which session a conversation
            belongs to. You can seed the transcript at the same time.
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

        <SessionMessagesForm
          messages={messages}
          onMessagesChange={setMessages}
        />

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
