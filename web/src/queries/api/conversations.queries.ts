import axios from "axios"

import type {
  AppendConversationMessageInput,
  AppendConversationMessagesResponse,
  ListConversationMessagesResponse,
  ListConversationsResponse,
} from "@/models/conversation"

// GET /conversations — past conversations, most recent first.
export const listConversations = async () => {
  return await axios.get<ListConversationsResponse>("/conversations")
}

// GET /conversations/{id}/messages — one conversation's transcript.
export const getConversationMessages = async (conversationId: string) => {
  return await axios.get<ListConversationMessagesResponse>(
    `/conversations/${conversationId}/messages`
  )
}

// POST /conversations/{id}/messages — inject messages ("false memory").
export const appendConversationMessages = async (
  conversationId: string,
  messages: AppendConversationMessageInput[]
) => {
  return await axios.post<AppendConversationMessagesResponse>(
    `/conversations/${conversationId}/messages`,
    { messages }
  )
}
