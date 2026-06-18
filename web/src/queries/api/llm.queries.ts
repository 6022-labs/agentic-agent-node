import axios from "axios"

import type { LlmConfig, SetLlmConfigRequest } from "@/models/llm-config"

// GET /config/llm — 404 when not configured yet.
export const getLlmConfig = async () => {
  return await axios.get<LlmConfig>("/config/llm")
}

// POST /config/llm — upsert the single LLM connection.
export const setLlmConfig = async (request: SetLlmConfigRequest) => {
  return await axios.post<LlmConfig>("/config/llm", request)
}

// DELETE /config/llm — 204.
export const deleteLlmConfig = async () => {
  return await axios.delete("/config/llm")
}
