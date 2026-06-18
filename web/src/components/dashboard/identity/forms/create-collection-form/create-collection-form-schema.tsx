import { z } from "zod"

const addressRegex = /^0x[0-9a-fA-F]{40}$/

export const createCollectionFormSchema = z.object({
  // The contract requires a valid ENS label.
  name: z
    .string()
    .regex(/^[a-z0-9-]+$/, "Lowercase letters, digits and hyphens only"),
  symbol: z.string().min(1, "Symbol is required"),
  description: z.string().optional(),
  admin: z.string().regex(addressRegex, "Must be a 0x-prefixed address"),
  moderator: z.string().regex(addressRegex, "Must be a 0x-prefixed address"),
})

export type CreateCollectionFormSchemaType = z.infer<
  typeof createCollectionFormSchema
>
