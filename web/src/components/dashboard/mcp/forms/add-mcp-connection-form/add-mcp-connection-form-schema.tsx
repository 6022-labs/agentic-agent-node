import { z } from "zod"

export const addMcpConnectionFormSchema = z
  .object({
    name: z.string().min(1, "Name is required"),
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
