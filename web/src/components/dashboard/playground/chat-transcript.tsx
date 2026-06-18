import React from "react"
import { Loader2, MessagesSquare } from "lucide-react"

import { Skeleton } from "@/components/ui/skeleton"
import EmptyState from "@/components/dashboard/empty-state"
import { cn } from "@/lib/utils"
import type { InputMessage } from "@/models/response"

export type ChatMessage = InputMessage & { stats?: string }

type Props = {
  messages: ChatMessage[]
  loadingHistory: boolean
  sending: boolean
  scrollRef: React.RefObject<HTMLDivElement | null>
}

export default function ChatTranscript({
  messages,
  loadingHistory,
  sending,
  scrollRef,
}: Props) {
  return (
    <div
      ref={scrollRef}
      className="flex-1 space-y-3 overflow-y-auto rounded-xl border bg-muted/30 p-4"
    >
      {loadingHistory ? (
        <div className="space-y-3">
          <Skeleton className="ml-auto h-9 w-2/5 rounded-2xl" />
          <Skeleton className="h-14 w-3/5 rounded-2xl" />
          <Skeleton className="ml-auto h-9 w-1/3 rounded-2xl" />
          <Skeleton className="h-20 w-3/4 rounded-2xl" />
        </div>
      ) : messages.length === 0 ? (
        <EmptyState
          icon={MessagesSquare}
          title="No messages yet"
          description="Say hello to your agent."
        />
      ) : (
        messages.map((message) => (
          <div
            key={message.id}
            className={cn(
              "flex flex-col gap-1",
              message.role === "user" ? "items-end" : "items-start"
            )}
          >
            <div
              className={cn(
                "max-w-[85%] rounded-2xl px-3 py-2 text-sm whitespace-pre-wrap",
                message.role === "user"
                  ? "bg-primary text-primary-foreground"
                  : "border bg-background"
              )}
            >
              {message.content}
            </div>
            {message.stats ? (
              <span className="text-xs text-muted-foreground">
                {message.stats}
              </span>
            ) : null}
          </div>
        ))
      )}
      {sending ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" />
          Thinking…
        </div>
      ) : null}
    </div>
  )
}
