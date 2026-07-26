import { z } from "zod"

// Keeps mcp_<name>_<tool> under the tightest provider function-name limit; mirrors the backend rule.
const MAX_NAME_LENGTH = 15

export const mcpConnectionFormSchema = z
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
    allowedTools: z.array(z.string()).min(1, "Allow at least one tool"),
    headerAuth: z.array(
      z.object({
        name: z.string(),
        // Blank is legal when editing a stored header: it keeps the vault value.
        value: z.string(),
      })
    ),
  })
  .superRefine((data, ctx) => {
    data.headerAuth.forEach((header, index) => {
      if (!header.name.trim()) {
        ctx.addIssue({
          code: "custom",
          message: "Header name is required",
          path: ["headerAuth", index, "name"],
        })
      }
    })

    const names = data.headerAuth.map((header) => header.name.trim())
    names.forEach((name, index) => {
      if (name && names.indexOf(name) !== index) {
        ctx.addIssue({
          code: "custom",
          message: "Duplicate header name",
          path: ["headerAuth", index, "name"],
        })
      }
    })
  })

export type McpConnectionFormSchemaType = z.infer<
  typeof mcpConnectionFormSchema
>
