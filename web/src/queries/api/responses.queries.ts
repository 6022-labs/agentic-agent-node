import axios from "axios"

import type { AgentResponse, CreateResponseRequest } from "@/models/response"

// POST /responses — one inference turn. x402-gated, but a JWT free-access
// caller needs no signature; paymentHeader is sent only on a 402 fallback.
export const createResponse = async (
  request: CreateResponseRequest,
  paymentHeader?: string
) => {
  return await axios.post<AgentResponse>("/responses", request, {
    headers: paymentHeader ? { "PAYMENT-SIGNATURE": paymentHeader } : undefined,
  })
}
