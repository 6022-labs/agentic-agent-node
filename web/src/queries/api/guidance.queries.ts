import axios from "axios"

import type { AgentGuidance } from "@/models/agent-guidance"

// GET /agent/guidance — empty content until the owner sets one.
export const getAgentGuidance = async () => {
  return await axios.get<AgentGuidance>("/agent/guidance")
}

// PUT /agent/guidance — replaces the owner-editable guidance layer.
export const setAgentGuidance = async (request: AgentGuidance) => {
  return await axios.put<AgentGuidance>("/agent/guidance", request)
}
