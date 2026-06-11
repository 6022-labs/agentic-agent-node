import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  addMcpConnectionFormSchema,
  type AddMcpConnectionFormSchemaType,
} from "./add-mcp-connection-form-schema"

type Props = {
  onFormComplete: (
    values: AddMcpConnectionFormSchemaType
  ) => Promise<FormCompletionResult<keyof AddMcpConnectionFormSchemaType>>
}

export default function AddMcpConnectionForm({ onFormComplete }: Props) {
  const form = useForm<AddMcpConnectionFormSchemaType>({
    resolver: zodResolver(addMcpConnectionFormSchema),
    defaultValues: {
      name: "",
      url: "",
      authHeaderName: "",
      authHeaderValue: "",
    },
  })

  // No reset on success: the dialog closes and unmounts the form. Values are
  // kept on failure so the user can correct and retry.
  const onSubmit = async (values: AddMcpConnectionFormSchemaType) => {
    const { error } = await onFormComplete(values)

    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  return (
    <Form {...form}>
      <form
        id="mcp-create-form"
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-4"
      >
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="my-tools" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>URL</FormLabel>
              <FormControl>
                <Input placeholder="https://my.mcp.com/v1/…" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="authHeaderName"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Auth header name (optional)</FormLabel>
              <FormControl>
                <Input placeholder="Authorization" {...field} />
              </FormControl>
              <FormDescription>
                Leave empty for an unauthenticated server.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="authHeaderValue"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Auth header value (optional)</FormLabel>
              <FormControl>
                <Input type="password" placeholder="Bearer …" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}
