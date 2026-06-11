export type MessageRole = "user" | "assistant" | "system"

export type InputMessage = {
  id: string
  author: string
  role: MessageRole
  content: string
  timestamp: string
}

export type CreateResponseRequest = {
  conversationId: string
  platform: "none" | "slack"
  prompt: string
  input: InputMessage[]
  forceConclusion?: boolean
  options?: {
    temperature?: number
    maxTokens?: number
    stop?: string[]
  }
}

export type OutputMessage = {
  id: string
  content: string
  timestamp: string
}

export type ResponseStats = {
  iterations: number
  toolCalls: number
  tokens: number
  wallClockMs: number
}

export type AgentResponse = {
  id: string
  isFinal: boolean
  output: OutputMessage[]
  stats?: ResponseStats
}
