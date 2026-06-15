import { z } from "zod"

import { AGENT_ROLES } from "@/models/enums/agent-role"

export const registrationDraftFormSchema = z.object({
  chainId: z.string().min(1, "Select a chain"),
  collectionAddress: z
    .string()
    .regex(/^0x[0-9a-fA-F]{40}$/, "Must be a 0x-prefixed address"),
  name: z.string().min(1, "Name is required"),
  role: z.enum(AGENT_ROLES, { message: "Select a role" }),
  // The base image file the node frames into the NFT set, uploaded as multipart.
  image: z.instanceof(File, { message: "Please choose an image file" }),
  attributes: z
    .array(
      z.object({
        key: z.string().min(1, "Key is required"),
        value: z.string().min(1, "Value is required"),
      })
    )
    .superRefine((attributes, ctx) => {
      const seen = new Set<string>()
      attributes.forEach((attribute, index) => {
        const key = attribute.key.trim()
        if (seen.has(key)) {
          ctx.addIssue({
            code: "custom",
            message: "Duplicate key",
            path: [index, "key"],
          })
        }
        seen.add(key)
      })
    }),
})

export type RegistrationDraftFormSchemaType = z.infer<
  typeof registrationDraftFormSchema
>
