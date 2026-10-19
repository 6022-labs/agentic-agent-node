import { z } from "zod"

import { LLM_PROVIDERS } from "@/models/llm-config"

// The stored key belongs to currentProvider, so any other cloud provider needs a new one.
export const createLlmConfigFormSchema = (currentProvider?: string) =>
  z
    .object({
      provider: z.string().min(1, "Select a provider"),
      model: z.string().min(1, "Model is required"),
      apiKey: z.string().optional(),
      baseUrl: z.string().optional(),
    })
    .superRefine((data, ctx) => {
      const provider = LLM_PROVIDERS.find((p) => p.value === data.provider)
      const needsBaseUrl =
        provider?.kind === "self-hosted" || provider?.requiresBaseUrl
      if (needsBaseUrl && !data.baseUrl) {
        ctx.addIssue({
          code: "custom",
          message: "This provider needs a server base URL",
          path: ["baseUrl"],
        })
      }
      const needsNewKey =
        provider?.kind === "cloud" && data.provider !== currentProvider
      if (needsNewKey && !data.apiKey) {
        ctx.addIssue({
          code: "custom",
          message: "Enter an API key for this provider",
          path: ["apiKey"],
        })
      }
    })

export type LlmConfigFormSchemaType = z.infer<
  ReturnType<typeof createLlmConfigFormSchema>
>
