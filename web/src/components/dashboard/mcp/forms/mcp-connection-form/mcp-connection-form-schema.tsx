import { z } from "zod"

// Keeps mcp_<name>_<tool> under the tightest provider function-name limit; mirrors the backend rule.
const MAX_NAME_LENGTH = 15

export const MCP_HEADER_AUTH_TYPES = ["bearer", "custom"] as const

export type McpHeaderAuthType = (typeof MCP_HEADER_AUTH_TYPES)[number]

export const MCP_HEADER_AUTH_TYPE_LABELS: Record<McpHeaderAuthType, string> = {
  bearer: "API key",
  custom: "Custom header",
}

export const BEARER_HEADER_NAME = "Authorization"
export const BEARER_PREFIX = "Bearer "

// A bearer row carries no name of its own: the scheme dictates it.
export const headerAuthName = (row: {
  type: McpHeaderAuthType
  name: string
}) => (row.type === "bearer" ? BEARER_HEADER_NAME : row.name.trim())

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
        type: z.enum(MCP_HEADER_AUTH_TYPES),
        name: z.string(),
        // Blank is legal when editing a stored header: it keeps the vault value.
        value: z.string(),
      })
    ),
  })
  .superRefine((data, ctx) => {
    data.headerAuth.forEach((header, index) => {
      if (header.type === "custom" && !header.name.trim()) {
        ctx.addIssue({
          code: "custom",
          message: "Header name is required",
          path: ["headerAuth", index, "name"],
        })
      }
    })

    const names = data.headerAuth.map(headerAuthName)
    names.forEach((name, index) => {
      if (name && names.indexOf(name) !== index) {
        ctx.addIssue({
          code: "custom",
          message: "Duplicate header name",
          path: [
            "headerAuth",
            index,
            data.headerAuth[index].type === "bearer" ? "type" : "name",
          ],
        })
      }
    })
  })

export type McpConnectionFormSchemaType = z.infer<
  typeof mcpConnectionFormSchema
>
