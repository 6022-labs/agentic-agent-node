export type LlmConfig = {
  id: string
  provider: string
  model: string
  baseUrl?: string
  setupStatus: string
  createdAt: string
  updatedAt: string
}

export type SetLlmConfigRequest = {
  provider: string
  model: string
  apiKey: string
  baseUrl?: string
  extensions?: Record<string, unknown>
}

export type LlmProvider = {
  value: string
  label: string
  kind: "cloud" | "self-hosted"
  // Cloud providers that also need a per-connection endpoint (e.g. the
  // regional Bedrock runtime URL the backend requires alongside the key).
  requiresBaseUrl?: boolean
}

// Providers the runtime supports (mirrors the orchestrator platform list).
export const LLM_PROVIDERS: LlmProvider[] = [
  { value: "anthropic", label: "Anthropic", kind: "cloud" },
  { value: "openai", label: "OpenAI", kind: "cloud" },
  { value: "mistral_ai", label: "Mistral", kind: "cloud" },
  { value: "deepseek", label: "Deepseek", kind: "cloud" },
  { value: "libert_ai", label: "LibertAI", kind: "cloud" },
  { value: "x_ai", label: "X", kind: "cloud" },
  {
    value: "aws.bedrock",
    label: "Amazon Bedrock",
    kind: "cloud",
    requiresBaseUrl: true,
  },
  { value: "ollama", label: "Ollama", kind: "self-hosted" },
]
