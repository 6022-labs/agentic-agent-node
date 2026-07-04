import axios from "axios"

import type { AgentSoul } from "@/models/agent-soul"

// GET /agent/soul — 404 while the agent identity is not initialized. Read-only: the soul is immutable after mint.
export const getAgentSoul = async () => {
  return await axios.get<AgentSoul>("/agent/soul")
}
