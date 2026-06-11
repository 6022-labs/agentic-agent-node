export type ConversationSummary = {
  conversationId: string
  messageCount: number
  firstMessageAt: string
  lastMessageAt: string
}

export type ListConversationsResponse = {
  conversations: ConversationSummary[]
}

export type ConversationMessage = {
  id: string
  conversationId: string
  role: string
  content: string
  timestamp: string
}

export type ListConversationMessagesResponse = {
  messages: ConversationMessage[]
}

export type ConversationMessageRole = "user" | "assistant" | "system"

export type AppendConversationMessageInput = {
  role: ConversationMessageRole
  content: string
  timestamp: string
}

export type AppendConversationMessagesResponse = {
  inserted: ConversationMessage[]
}
