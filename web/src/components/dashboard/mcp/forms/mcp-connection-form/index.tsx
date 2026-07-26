import React from "react"
import { useFieldArray, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Plus, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
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
import type { McpConnection, McpServerTool } from "@/models/mcp-connection"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { matchesToolPatterns } from "@/lib/tool-patterns"
import AllowedToolsField from "./allowed-tools-field"
import {
  mcpConnectionFormSchema,
  type McpConnectionFormSchemaType,
} from "./mcp-connection-form-schema"

// The submit button lives in the dialog footer and submits by this id.
export const MCP_CONNECTION_FORM_ID = "mcp-connection-form"

type Props = {
  // When set, the form edits this connection instead of creating one.
  connection?: McpConnection
  // Tools the server exposes, resolved by the caller; null until loaded.
  tools: McpServerTool[] | null
  loadingTools: boolean
  onValueChange: (values: McpConnectionFormSchemaType) => void
  onFormComplete: (
    values: McpConnectionFormSchemaType
  ) => Promise<FormCompletionResult<keyof McpConnectionFormSchemaType>>
}

export default function McpConnectionForm({
  connection,
  tools,
  loadingTools,
  onValueChange,
  onFormComplete,
}: Props) {
  const form = useForm<McpConnectionFormSchemaType>({
    resolver: zodResolver(mcpConnectionFormSchema),
    defaultValues: {
      name: connection?.name ?? "",
      url: connection?.url ?? "",
      allowedTools: [],
      // A stored header keeps its name; its value stays in the vault, so the input starts blank.
      headerAuth: (connection?.headerAuthKeys ?? []).map((name) => ({
        name,
        value: "",
      })),
    },
  })

  const headerRows = useFieldArray({ control: form.control, name: "headerAuth" })

  React.useEffect(() => {
    onValueChange(form.getValues())

    const subscription = form.watch(() => onValueChange(form.getValues()))
    return () => subscription.unsubscribe()
  }, [form, onValueChange])

  // A freshly listed server starts fully allowed; an edited one keeps whatever
  // its stored patterns still match.
  React.useEffect(() => {
    if (!tools) return

    const allowed = tools
      .filter(
        (tool) =>
          !connection ||
          matchesToolPatterns(tool.name, connection.allowedToolPatterns)
      )
      .map((tool) => tool.name)
    form.setValue("allowedTools", allowed)
  }, [tools, connection, form])

  const onSubmit = async (values: McpConnectionFormSchemaType) => {
    const { error } = await onFormComplete(values)

    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  return (
    <Form {...form}>
      <form
        id={MCP_CONNECTION_FORM_ID}
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
        <FormItem>
          <FormLabel>Auth headers</FormLabel>
          {headerRows.fields.map((row, index) => (
            <div key={row.id} className="flex items-start gap-2">
              <FormField
                control={form.control}
                name={`headerAuth.${index}.name`}
                render={({ field }) => (
                  <FormItem className="flex-1">
                    <FormControl>
                      <Input placeholder="Authorization" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name={`headerAuth.${index}.value`}
                render={({ field }) => (
                  <FormItem className="flex-1">
                    <FormControl>
                      <Input
                        type="password"
                        placeholder={connection ? "•••••• (unchanged)" : "Bearer …"}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label="Remove header"
                onClick={() => headerRows.remove(index)}
              >
                <Trash2 />
              </Button>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => headerRows.append({ name: "", value: "" })}
          >
            <Plus />
            Add header
          </Button>
          <FormDescription>
            Values are stored in the node vault and never returned.
            {connection
              ? " Leave a value blank to keep the stored one; remove the row to delete the header."
              : " Leave the list empty for an unauthenticated server."}
          </FormDescription>
        </FormItem>
        <FormField
          control={form.control}
          name="allowedTools"
          render={({ field }) => (
            <AllowedToolsField
              tools={tools}
              loading={loadingTools}
              selected={field.value}
              onSelectedChange={field.onChange}
            />
          )}
        />
      </form>
    </Form>
  )
}
