import { Plus, RefreshCw, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { conversationLabel, type Conversation } from "@/models/conversation"
import type { ConversationMessageRole } from "@/models/enums/conversation-message-role"
import { formatRelativeTime } from "@/lib/format"
import { newId } from "@/lib/utils"
import { emptyMessage, type DraftMessage } from "./draft-message"

// "system" is deliberately absent: the system slot belongs to the built prompt.
const ROLES: ConversationMessageRole[] = ["user", "assistant"]

type Props = {
  conversationId: string
  onConversationIdChange: (id: string) => void
  conversations: Conversation[]
  messages: DraftMessage[]
  onMessagesChange: (messages: DraftMessage[]) => void
}

export default function AppendMemoryForm({
  conversationId,
  onConversationIdChange,
  conversations,
  messages,
  onMessagesChange,
}: Props) {
  const updateMessage = (key: string, patch: Partial<DraftMessage>) =>
    onMessagesChange(
      messages.map((m) => (m.key === key ? { ...m, ...patch } : m))
    )

  return (
    <div className="space-y-4">
      <div className="grid gap-1.5">
        <Label htmlFor="memory-conversation-id">Conversation</Label>
        <div className="flex gap-2">
          <Input
            id="memory-conversation-id"
            value={conversationId}
            onChange={(e) => onConversationIdChange(e.target.value)}
            placeholder="conversation id"
            className="font-mono text-xs"
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            title="Generate a new conversation id"
            onClick={() => onConversationIdChange(newId())}
          >
            <RefreshCw />
          </Button>
        </div>
        {conversations.length > 0 ? (
          <Select value="" onValueChange={onConversationIdChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="…or pick an existing conversation" />
            </SelectTrigger>
            <SelectContent>
              {conversations.map((conversation) => (
                <SelectItem
                  key={conversation.conversationId}
                  value={conversation.conversationId}
                >
                  {conversationLabel(conversation)} · {conversation.messageCount}{" "}
                  message{conversation.messageCount === 1 ? "" : "s"} ·{" "}
                  {formatRelativeTime(conversation.updatedAt)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : null}
      </div>

      <div className="grid gap-2">
        <Label>Messages</Label>
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
              className="max-h-28 min-h-9 flex-1"
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
    </div>
  )
}
