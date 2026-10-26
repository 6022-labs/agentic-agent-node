import React from "react"
import axios from "axios"
import { toast } from "sonner"

import type { ChatMessage } from "@/components/dashboard/playground/chat-transcript"
import type { MessageRole } from "@/models/enums/message-role"
import type { Conversation } from "@/models/conversation"
import { createResponse } from "@/queries/api/responses.queries"
import { listPaymentReceiveRules } from "@/queries/api/payment-receive-rule.queries"
import {
  getConversationMessages,
  listConversations,
} from "@/queries/api/conversations.queries"
import { extractErrorMessage } from "@/lib/api-errors"
import { getX402PaymentHeader } from "@/lib/x402"
import { findFreeRule } from "@/models/payment-receive-policy"
import { newId } from "@/lib/utils"
import { useWallet } from "@/contexts/wallet-provider-context"

export function usePlaygroundChat() {
  const { wallet, address } = useWallet()
  const [messages, setMessages] = React.useState<ChatMessage[]>([])
  const [sending, setSending] = React.useState(false)
  const [conversations, setConversations] = React.useState<Conversation[]>([])
  const [conversationId, setConversationId] = React.useState(() => newId())
  const [loadingHistory, setLoadingHistory] = React.useState(false)

  const loadConversations = React.useCallback(async () => {
    try {
      const { data } = await listConversations(1, 100)
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
      // Skip tool results and content-less tool-call turns.
      setMessages(
        data.messages
          .filter((m) => m.role !== "tool" && !!m.content)
          .map((m) => ({
            id: m.id,
            author: m.role,
            role: m.role as MessageRole,
            content: m.content ?? "",
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
      // OpenAI Responses contract: the server owns history so we send
      // only the new turn and continue the thread via previous_response_id.
      const body = {
        input: content,
        instructions: prompt,
        previous_response_id: conversationId,
      }

      // JWT free access needs no signature; sign only on a 402 fallback.
      let response: Awaited<ReturnType<typeof createResponse>>
      try {
        response = await createResponse(body)
      } catch (error) {
        if (!axios.isAxiosError(error) || error.response?.status !== 402) {
          throw error
        }
        if (!wallet || !address) {
          throw new Error("Connect a wallet to talk to the agent", {
            cause: error,
          })
        }
        const { rules } = await listPaymentReceiveRules()
        const freeRule = findFreeRule(rules, address)
        if (!freeRule) {
          throw new Error(
            "No free access for this wallet — add a zero-price rule for it in the Payment policy",
            { cause: error }
          )
        }
        const paymentHeader = await getX402PaymentHeader(
          wallet,
          address,
          freeRule
        )
        response = await createResponse(body, paymentHeader)
      }
      const { data } = response

      if (data.id && data.id !== conversationId) {
        setConversationId(data.id)
      }

      const output = data.output?.find((item) => item.type === "message")
      const text =
        output?.content
          ?.filter((part) => part.type === "output_text")
          .map((part) => part.text)
          .join("") || "(no output)"
      const assistantMessage: ChatMessage = {
        id: output?.id || newId(),
        author: "assistant",
        role: "assistant",
        content: text,
        timestamp: new Date().toISOString(),
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
