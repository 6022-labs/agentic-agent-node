import { z } from "zod"

import { ensLabelSchema } from "@/lib/ens-label"
import { AGENT_ROLES } from "@/models/enums/agent-role"

export const registrationDraftFormSchema = z.object({
  chainId: z.string().min(1, "Select a chain"),
  collectionAddress: z
    .string()
    .regex(/^0x[0-9a-fA-F]{40}$/, "Must be a 0x-prefixed address"),
  name: ensLabelSchema,
  // The base image file the node frames into the NFT set, uploaded as multipart.
  // Optional on re-edit (omitting it keeps the stored image); the backend
  // enforces it on first save.
  image: z
    .instanceof(File, { message: "Please choose an image file" })
    .optional(),
  attributes: z
    .array(
      z.object({
        key: z.string().min(1, "Key is required"),
        value: z.string().min(1, "Value is required"),
      })
    )
    .superRefine((attributes, ctx) => {
      // Row 0 is the pinned mandatory "role" attribute.
      const role = attributes[0]
      if (
        !role ||
        role.key !== "role" ||
        !AGENT_ROLES.includes(role.value as (typeof AGENT_ROLES)[number])
      ) {
        ctx.addIssue({
          code: "custom",
          message: "Select a role",
          path: [0, "value"],
        })
      }
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
