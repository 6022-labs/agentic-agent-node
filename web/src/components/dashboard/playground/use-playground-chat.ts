import React from "react"
import { toast } from "sonner"

import type { ChatMessage } from "@/components/dashboard/playground/chat-transcript"
import type { MessageRole } from "@/models/response"
import type { ConversationSummary } from "@/models/conversation"
import { createResponse } from "@/queries/api/responses.queries"
import { listPaymentRules } from "@/queries/api/payment-rule.queries"
import {
  getConversationMessages,
  listConversations,
} from "@/queries/api/conversations.queries"
import { extractErrorMessage } from "@/lib/api-errors"
import { getX402PaymentHeader } from "@/lib/x402"
import { findFreeRule } from "@/models/payment-policy"
import { newId } from "@/lib/utils"
import { useWallet } from "@/contexts/wallet-provider-context"

export function usePlaygroundChat() {
  const { wallet, address } = useWallet()
  const [messages, setMessages] = React.useState<ChatMessage[]>([])
  const [sending, setSending] = React.useState(false)
  const [conversations, setConversations] = React.useState<
    ConversationSummary[]
  >([])
  const [conversationId, setConversationId] = React.useState(() => newId())
  const [loadingHistory, setLoadingHistory] = React.useState(false)

  const loadConversations = React.useCallback(async () => {
    try {
      const { data } = await listConversations()
      setConversations(data.conversations ?? [])
    } catch {
      // History is best-effort; the playground still works without it.
    }
  }, [])

  React.useEffect(() => {
    void loadConversations()
  }, [loadConversations])

  const newChat = () => {
    setMessages([])
    setConversationId(newId())
  }

  const openConversation = async (id: string) => {
    if (id === conversationId) return

    setLoadingHistory(true)
    try {
      const { data } = await getConversationMessages(id)
      setConversationId(id)
      setMessages(
        data.messages.map((m) => ({
          id: m.id,
          author: m.role,
          role: m.role as MessageRole,
          content: m.content,
          timestamp: m.timestamp,
        }))
      )
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to load conversation"))
    } finally {
      setLoadingHistory(false)
    }
  }

  // Optimistically appends the user message; returns false (and rolls the
  // message back) when the agent fails to respond.
  const send = async (content: string, prompt: string): Promise<boolean> => {
    const userMessage: ChatMessage = {
      id: newId(),
      author: "user",
      role: "user",
      content,
      timestamp: new Date().toISOString(),
    }
    const nextMessages = [...messages, userMessage]
    setMessages(nextMessages)
    setSending(true)

    try {
      if (!wallet || !address) {
        throw new Error("Connect a wallet to talk to the agent")
      }
      // /responses is x402-gated; the console signs a zero-value
      // authorization against a free rule the connected wallet is
      // entitled to (the agent wallet + owner have one by default).
      const { data: rulesData } = await listPaymentRules()
      const freeRule = findFreeRule(rulesData.rules ?? [], address)
      if (!freeRule) {
        throw new Error(
          "No free access for this wallet — add a zero-price rule for it in the Payment policy"
        )
      }
      const paymentHeader = await getX402PaymentHeader(wallet, address, freeRule)

      const { data } = await createResponse(
        {
          conversationId,
          platform: "none",
          prompt,
          input: nextMessages.map(
            ({ id, author, role, content, timestamp }) => ({
              id,
              author,
              role,
              content,
              timestamp,
            })
          ),
        },
        paymentHeader
      )

      const output = data.output?.[0]
      const assistantMessage: ChatMessage = {
        id: output?.id ?? newId(),
        author: "assistant",
        role: "assistant",
        content: output?.content ?? "(no output)",
        timestamp: output?.timestamp ?? new Date().toISOString(),
        stats: data.stats
          ? `${data.stats.iterations} iters · ${data.stats.toolCalls} tools · ${data.stats.tokens} tokens · ${data.stats.wallClockMs}ms`
          : undefined,
      }
      setMessages((prev) => [...prev, assistantMessage])
      void loadConversations()
      return true
    } catch (error) {
      toast.error(extractErrorMessage(error, "The agent failed to respond"))
      setMessages((prev) => prev.filter((m) => m.id !== userMessage.id))
      return false
    } finally {
      setSending(false)
    }
  }

  return {
    messages,
    sending,
    conversations,
    conversationId,
    loadingHistory,
    newChat,
    openConversation,
    send,
  }
}
