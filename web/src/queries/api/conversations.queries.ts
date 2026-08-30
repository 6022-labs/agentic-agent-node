import axios from "axios"

import type {
  AppendConversationMessageInput,
  AppendConversationMessagesResponse,
  Conversation,
  ListConversationMessagesResponse,
  ListConversationsResponse,
} from "@/models/conversation"

// GET /conversations — one page of memory sessions, most recent activity first.
export const listConversations = async (page = 1, elementsPerPage = 10) => {
  return await axios.get<ListConversationsResponse>("/conversations", {
    params: { page, elementsPerPage },
  })
}

// POST /conversations — open an empty named memory session.
export const createConversation = async (name: string) => {
  return await axios.post<Conversation>("/conversations", { name })
}

// PUT /conversations/{id} — rename a memory session.
export const updateConversation = async (
  conversationId: string,
  name: string
) => {
  return await axios.put<Conversation>(`/conversations/${conversationId}`, {
    name,
  })
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
