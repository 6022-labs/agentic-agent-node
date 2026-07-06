import { z } from "zod"

// Keeps mcp_<name>_<tool> under the tightest provider function-name limit; mirrors the backend rule.
const MAX_NAME_LENGTH = 15

export const addMcpConnectionFormSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, "Name is required")
      .max(MAX_NAME_LENGTH, `Name cannot exceed ${MAX_NAME_LENGTH} characters`)
      .regex(
        /^[a-zA-Z][a-zA-Z0-9-]*$/,
        "Must start with a letter and contain only letters, digits, and hyphens"
      )
      .refine((name) => !name.toLowerCase().startsWith("mcp"), {
        message: 'Name cannot start with "mcp"',
      })
      .refine((name) => name.toLowerCase() !== "self", {
        message: '"self" prefix is reserved for built-in tools',
      }),
    url: z.string().min(1, "URL is required"),
    authHeaderName: z.string().optional(),
    authHeaderValue: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    if (data.authHeaderName && !data.authHeaderValue) {
      ctx.addIssue({
        code: "custom",
        message: "Required when a header name is set",
        path: ["authHeaderValue"],
      })
    }
  })

export type AddMcpConnectionFormSchemaType = z.infer<
  typeof addMcpConnectionFormSchema
>
