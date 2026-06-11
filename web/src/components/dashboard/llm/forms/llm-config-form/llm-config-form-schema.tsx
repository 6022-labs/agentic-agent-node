import { z } from "zod"

import { LLM_PROVIDERS } from "@/models/llm-config"

export const llmConfigFormSchema = z
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
  })

export type LlmConfigFormSchemaType = z.infer<typeof llmConfigFormSchema>
