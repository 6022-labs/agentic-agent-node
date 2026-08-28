import type { ConversationMessageRole } from "@/models/enums/conversation-message-role"

export type Conversation = {
  conversationId: string
  name?: string
  messageCount: number
  createdAt: string
  updatedAt: string
}

export type ListConversationsResponse = {
  page: number
  elementsPerPage: number
  totalElements: number
  conversations: Conversation[]
}

// conversationLabel prefers the session name; the id is the fallback the wire actually keys on.
export const conversationLabel = (conversation: Conversation) =>
  conversation.name ?? conversation.conversationId

export type ConversationMessage = {
  id: string
  conversationId: string
  role: string
  content?: string
  toolCalls?: string
  toolCallId?: string
  timestamp: string
}

export type ListConversationMessagesResponse = {
  messages: ConversationMessage[]
}

export type AppendConversationMessageInput = {
  role: ConversationMessageRole
  content: string
  timestamp: string
}

export type AppendConversationMessagesResponse = {
  inserted: ConversationMessage[]
}
