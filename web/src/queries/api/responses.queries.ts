import axios from "axios"

import type { AgentResponse, CreateResponseRequest } from "@/models/response"

// POST /responses — runs one inference turn and returns the full response.
// The route is x402-gated: paymentHeader carries the base64 PAYMENT-SIGNATURE
// (for the console, the free-auth signature of the connected wallet).
export const createResponse = async (
  request: CreateResponseRequest,
  paymentHeader: string
) => {
  return await axios.post<AgentResponse>("/responses", request, {
    headers: { "PAYMENT-SIGNATURE": paymentHeader },
  })
}
