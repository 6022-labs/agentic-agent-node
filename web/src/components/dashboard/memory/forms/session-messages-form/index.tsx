import { Plus, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { ConversationMessageRole } from "@/models/enums/conversation-message-role"
import { emptyMessage, type DraftMessage } from "./draft-message"

// "system" is deliberately absent: the system slot belongs to the built prompt.
const ROLES: ConversationMessageRole[] = ["user", "assistant"]

type Props = {
  messages: DraftMessage[]
  onMessagesChange: (messages: DraftMessage[]) => void
}

export default function SessionMessagesForm({
  messages,
  onMessagesChange,
}: Props) {
  const updateMessage = (key: string, patch: Partial<DraftMessage>) =>
    onMessagesChange(
      messages.map((m) => (m.key === key ? { ...m, ...patch } : m))
    )

  return (
    <div className="grid gap-2">
      <Label>Messages</Label>
      <p className="text-sm text-muted-foreground">
        Optional. Messages go straight into the transcript, bypassing the agent
        — useful for testing how it reacts to a given history.
      </p>
      {messages.map((message) => (
        <div key={message.key} className="flex items-start gap-2">
          <Select
            value={message.role}
            onValueChange={(value) =>
              updateMessage(message.key, {
                role: value as ConversationMessageRole,
              })
            }
          >
            <SelectTrigger className="w-32 shrink-0">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {ROLES.map((role) => (
                <SelectItem key={role} value={role}>
                  {role}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Textarea
            value={message.content}
            onChange={(e) =>
              updateMessage(message.key, { content: e.target.value })
            }
            placeholder="Message content…"
            className="max-h-48 min-h-20 flex-1"
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            disabled={messages.length === 1}
            onClick={() =>
              onMessagesChange(messages.filter((m) => m.key !== message.key))
            }
            aria-label="Remove message"
          >
            <Trash2 />
          </Button>
        </div>
      ))}
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="w-fit"
        onClick={() => onMessagesChange([...messages, emptyMessage()])}
      >
        <Plus />
        Add message
      </Button>
    </div>
  )
}
