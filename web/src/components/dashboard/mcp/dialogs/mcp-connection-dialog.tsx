import React from "react"
import { toast } from "sonner"
import { Loader2, Pencil, Plus, RefreshCw } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import McpConnectionForm, {
  MCP_CONNECTION_FORM_ID,
} from "@/components/dashboard/mcp/forms/mcp-connection-form"
import {
  BEARER_PREFIX,
  headerAuthName,
  type McpConnectionFormSchemaType,
  type McpHeaderAuthType,
} from "@/components/dashboard/mcp/forms/mcp-connection-form/mcp-connection-form-schema"
import type { McpConnection, McpServerTool } from "@/models/mcp-connection"
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  createMcpConnection,
  discoverMcpServerTools,
  updateMcpConnection,
} from "@/queries/api/mcp.queries"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"
import { toolNamePattern } from "@/lib/tool-patterns"

type HeaderRow = { type: McpHeaderAuthType; name: string; value: string }

// toHeaderAuthMap keeps blank values: the runtime reads them as "use the stored value",
// so a blank bearer row must never become the bare prefix.
const toHeaderAuthMap = (rows: HeaderRow[] | undefined) => {
  if (!rows?.length) return undefined
  return Object.fromEntries(
    rows.map((row) => [
      headerAuthName(row),
      row.type === "bearer" && row.value
        ? `${BEARER_PREFIX}${row.value.trim()}`
        : row.value,
    ])
  )
}

type Props = {
  // When set, the dialog edits this connection (pencil trigger) instead of
  // adding one (Add trigger).
  connection?: McpConnection
  onSaved: () => Promise<void> | void
}

export default function McpConnectionDialog({ connection, onSaved }: Props) {
  const [open, setOpen] = React.useState(false)
  const [saving, setSaving] = React.useState(false)
  const [loadingTools, setLoadingTools] = React.useState(false)
  const [tools, setTools] = React.useState<McpServerTool[] | null>(null)
  const [values, setValues] = React.useState<McpConnectionFormSchemaType | null>(
    null
  )
  const editing = connection !== undefined

  // A different server (or credential) exposes different tools, so the loaded
  // list no longer applies.
  const onValueChange = React.useCallback(
    (next: McpConnectionFormSchemaType) => {
      setValues((previous) => {
        if (
          previous &&
          (previous.url !== next.url ||
            JSON.stringify(previous.headerAuth) !==
              JSON.stringify(next.headerAuth))
        ) {
          setTools(null)
        }
        return next
      })
    },
    []
  )

  const onLoadTools = async () => {
    if (!connection && !values?.url) return

    setLoadingTools(true)
    try {
      const { data } = await discoverMcpServerTools({
        connectionId: connection?.id,
        url: values?.url || undefined,
        headerAuth: toHeaderAuthMap(values?.headerAuth),
      })
      setTools(data.tools ?? [])
    } catch (error) {
      const validation = extractValidationError(error)
      toast.error(
        validation?.message ??
          extractErrorMessage(error, "Failed to list the server's tools")
      )
    } finally {
      setLoadingTools(false)
    }
  }

  const onFormComplete = async (
    values: McpConnectionFormSchemaType
  ): Promise<FormCompletionResult<keyof McpConnectionFormSchemaType>> => {
    setSaving(true)
    try {
      const payload = {
        name: values.name,
        url: values.url,
        allowedToolPatterns: values.allowedTools.map(toolNamePattern),
        // Blank values are forwarded: on edit the runtime reads them as "keep the stored one".
        headerAuth: toHeaderAuthMap(values.headerAuth),
      }
      if (connection) {
        await updateMcpConnection(connection.id, payload)
      } else {
        await createMcpConnection(payload)
      }
      toast.success(editing ? "MCP connection updated" : "MCP connection added")
      setOpen(false)
      await onSaved()
      return {}
    } catch (error) {
      const validation =
        extractValidationError<keyof McpConnectionFormSchemaType>(error)
      if (validation) {
        return { error: validation }
      }
      toast.error(
        extractErrorMessage(
          error,
          editing
            ? "Failed to update MCP connection"
            : "Failed to add MCP connection"
        )
      )
      return {}
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) setTools(null)
      }}
    >
      <DialogTrigger asChild>
        {editing ? (
          <Button variant="ghost" size="icon-sm" aria-label="Edit connection">
            <Pencil />
          </Button>
        ) : (
          <Button size="sm">
            <Plus />
            Add
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {editing ? "Edit MCP connection" : "Add MCP connection"}
          </DialogTitle>
          <DialogDescription>
            Load the server's tools to pick the ones the agent may call. Header
            auth values are stored in the node vault and never returned.
          </DialogDescription>
        </DialogHeader>
        {/* Remount per opening so an edited connection is re-derived fresh. */}
        {open ? (
          <McpConnectionForm
            connection={connection}
            tools={tools}
            loadingTools={loadingTools}
            onValueChange={onValueChange}
            onFormComplete={onFormComplete}
          />
        ) : null}
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            loading={loadingTools}
            disabled={saving || (!connection && !values?.url)}
            onClick={() => void onLoadTools()}
          >
            {loadingTools ? null : <RefreshCw />}
            {tools === null ? "Load tools" : "Reload tools"}
          </Button>
          <Button
            type="submit"
            form={MCP_CONNECTION_FORM_ID}
            disabled={saving || loadingTools || tools === null}
            title={tools === null ? "Load the server's tools first" : undefined}
          >
            {saving ? <Loader2 className="animate-spin" /> : null}
            {editing ? "Save" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
