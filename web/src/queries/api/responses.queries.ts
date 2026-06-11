import axios from "axios"

import type { AgentResponse, CreateResponseRequest } from "@/models/response"

// POST /responses — runs one inference turn and returns the full response.
export const createResponse = async (request: CreateResponseRequest) => {
  return await axios.post<AgentResponse>("/responses", request)
}
