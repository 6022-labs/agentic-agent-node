import React from "react"
import { toast } from "sonner"
import { BrainCircuit, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import AppendMemoryForm from "@/components/dashboard/memory/forms/append-memory-form"
import {
  emptyMessage,
  type DraftMessage,
} from "@/components/dashboard/memory/forms/append-memory-form/draft-message"
import type { Conversation } from "@/models/conversation"
import {
  appendConversationMessages,
  listConversations,
} from "@/queries/api/conversations.queries"
import { extractErrorMessage } from "@/lib/api-errors"
import { newId } from "@/lib/utils"

type Props = {
  onAppended?: () => void
}

export default function AppendMemoryDialog({ onAppended }: Props) {
  const [open, setOpen] = React.useState(false)
  const [submitting, setSubmitting] = React.useState(false)

  // Draft state lives here, not in the form: Radix unmounts DialogContent on
  // close, and the draft should survive a close-and-reopen.
  const [conversationId, setConversationId] = React.useState(() => newId())
  const [conversations, setConversations] = React.useState<
    Conversation[]
  >([])
  const [messages, setMessages] = React.useState<DraftMessage[]>([
    emptyMessage(),
  ])

  React.useEffect(() => {
    if (!open) return
    listConversations(1, 100)
      .then(({ data }) => setConversations(data.conversations ?? []))
      .catch(() => setConversations([]))
  }, [open])

  const submit = async () => {
    const id = conversationId.trim()
    if (!id) {
      toast.error("A conversation id is required")
      return
    }

    const filled = messages.filter((m) => m.content.trim().length > 0)
    if (filled.length === 0) {
      toast.error("Add at least one non-empty message")
      return
    }

    setSubmitting(true)
    try {
      const now = Date.now()
      const { data } = await appendConversationMessages(
        id,
        filled.map((m, index) => ({
          role: m.role,
          content: m.content,
          // Stagger timestamps so the injected order is preserved on read.
          timestamp: new Date(now + index * 1000).toISOString(),
        }))
      )
      toast.success(`Injected ${data.inserted.length} message(s)`)
      setOpen(false)
      setMessages([emptyMessage()])
      setConversationId(newId())
      onAppended?.()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to inject messages"))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <BrainCircuit />
          Add memory
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Inject conversation memory</DialogTitle>
          <DialogDescription>
            Append messages directly into a conversation's transcript, bypassing
            the agent. Use an existing conversation or a fresh id.
          </DialogDescription>
        </DialogHeader>

        <AppendMemoryForm
          conversationId={conversationId}
          onConversationIdChange={setConversationId}
          conversations={conversations}
          messages={messages}
          onMessagesChange={setMessages}
        />

        <DialogFooter>
          <Button onClick={submit} disabled={submitting}>
            {submitting ? (
              <Loader2 className="animate-spin" />
            ) : (
              <BrainCircuit />
            )}
            Inject memory
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
