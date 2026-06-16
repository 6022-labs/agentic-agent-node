import type { MessageRole } from "@/models/enums/message-role"

export type InputMessageItem = {
  type?: "message"
  role: MessageRole
  content: string
}

// Mirrors the OpenAI Responses `create` body. Continuation is by
// previous_response_id (or conversation); only the new turn is sent.
export type CreateResponseRequest = {
  model?: string
  instructions?: string
  input: string | InputMessageItem[]
  conversation?: string | { id: string }
  previous_response_id?: string
  temperature?: number
  max_output_tokens?: number
  metadata?: Record<string, string>
}

export type OutputContent = {
  type: string
  text: string
  annotations: unknown[]
}

export type OutputItem = {
  type: string
  id: string
  status: string
  role: string
  content: OutputContent[]
}

export type Usage = {
  input_tokens: number
  output_tokens: number
  total_tokens: number
}

// Mirrors the OpenAI Responses object.
export type AgentResponse = {
  id: string
  object: string
  created_at: number
  status: string
  model?: string
  output: OutputItem[]
  usage: Usage | null
  previous_response_id: string | null
  conversation: { id: string } | null
  metadata?: Record<string, string>
  error: { code: string; message: string } | null
}
