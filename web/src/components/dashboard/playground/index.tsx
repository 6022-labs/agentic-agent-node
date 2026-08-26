import React from "react"
import { ChevronRight, Loader2, RotateCcw, SendHorizontal } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import ChatTranscript from "@/components/dashboard/playground/chat-transcript"
import NotMintedState from "@/components/dashboard/not-minted-state"
import { usePlaygroundChat } from "@/components/dashboard/playground/use-playground-chat"
import { formatRelativeTime } from "@/lib/format"
import { conversationLabel, type Conversation } from "@/models/conversation"

const conversationOption = (conversation: Conversation) => {
  const count = conversation.messageCount
  const messages = `${count} message${count === 1 ? "" : "s"}`
  return `${conversationLabel(conversation)} · ${messages} · ${formatRelativeTime(conversation.updatedAt)}`
}

export default function PlaygroundSection() {
  const { loading: identityLoading, isMinted } = useAgentIdentity()
  const [prompt, setPrompt] = React.useState("You are a helpful assistant.")
  const [draft, setDraft] = React.useState("")
  const scrollRef = React.useRef<HTMLDivElement>(null)

  const {
    messages,
    sending,
    conversations,
    conversationId,
    loadingHistory,
    newChat,
    openConversation,
    send,
  } = usePlaygroundChat()

  React.useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [messages, sending])

  const onSend = async () => {
    const content = draft.trim()
    if (!content || sending) return

    setDraft("")
    const sent = await send(
      content,
      prompt.trim() || "You are a helpful assistant."
    )

    // Restore the draft so a failed message isn't lost.
    if (!sent) setDraft(content)
  }

  const selectedValue = conversations.some(
    (c) => c.conversationId === conversationId
  )
    ? conversationId
    : undefined

  // Pre-mint /responses is refused (402), so the chat interface is pointless.
  if (!identityLoading && !isMinted) {
    return (
      <Card className="flex min-h-0 flex-1 flex-col">
        <CardHeader>
          <CardTitle>Playground</CardTitle>
          <CardDescription>
            Talk to the agent through the gateway's reasoning loop.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <NotMintedState description="Mint the agent's on-chain identity to start chatting." />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="flex min-h-0 flex-1 flex-col">
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div>
            <CardTitle>Playground</CardTitle>
            <CardDescription>
              Talk to the agent through the gateway's reasoning loop.
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Select
              value={selectedValue}
              onValueChange={openConversation}
              disabled={loadingHistory || conversations.length === 0}
            >
              <SelectTrigger className="w-[260px]">
                <SelectValue
                  placeholder={
                    conversations.length === 0 ? "No history" : "History"
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {conversations.map((conversation) => (
                  <SelectItem
                    key={conversation.conversationId}
                    value={conversation.conversationId}
                  >
                    {conversationOption(conversation)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              variant="ghost"
              size="sm"
              onClick={newChat}
              disabled={sending}
            >
              <RotateCcw />
              New chat
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex min-h-0 flex-1 flex-col gap-4">
        <Collapsible className="group/prompt rounded-lg border px-3 py-2">
          <CollapsibleTrigger className="flex w-full cursor-pointer items-center gap-2 text-left">
            <ChevronRight className="size-4 shrink-0 text-muted-foreground transition-transform group-data-[state=open]/prompt:rotate-90" />
            <Label className="pointer-events-none shrink-0">
              System prompt
            </Label>
            <span className="min-w-0 flex-1 truncate text-sm text-muted-foreground group-data-[state=open]/prompt:hidden">
              {prompt}
            </span>
          </CollapsibleTrigger>
          <CollapsibleContent className="pt-2">
            <Textarea
              id="system-prompt"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              className="max-h-28"
            />
          </CollapsibleContent>
        </Collapsible>

        <ChatTranscript
          messages={messages}
          loadingHistory={loadingHistory}
          sending={sending}
          scrollRef={scrollRef}
        />

        <div className="flex items-end gap-2">
          <Textarea
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault()
                void onSend()
              }
            }}
            placeholder="Type a message… (Enter to send, Shift+Enter for newline)"
            className="max-h-32 min-h-11 flex-1"
          />
          <Button
            onClick={onSend}
            disabled={sending || !draft.trim()}
            size="icon-lg"
          >
            {sending ? (
              <Loader2 className="animate-spin" />
            ) : (
              <SendHorizontal />
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
