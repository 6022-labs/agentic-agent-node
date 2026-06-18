import React from "react"
import { toast } from "sonner"
import { Loader2, Plus } from "lucide-react"

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
import AddMcpConnectionForm from "@/components/dashboard/mcp/forms/add-mcp-connection-form"
import type { AddMcpConnectionFormSchemaType } from "@/components/dashboard/mcp/forms/add-mcp-connection-form/add-mcp-connection-form-schema"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { createMcpConnection } from "@/queries/api/mcp.queries"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"

type Props = {
  onCreated: () => Promise<void> | void
}

export default function AddMcpConnectionDialog({ onCreated }: Props) {
  const [open, setOpen] = React.useState(false)
  const [saving, setSaving] = React.useState(false)

  const onFormComplete = async (
    values: AddMcpConnectionFormSchemaType
  ): Promise<FormCompletionResult<keyof AddMcpConnectionFormSchemaType>> => {
    setSaving(true)
    try {
      await createMcpConnection({
        name: values.name,
        url: values.url,
        authHeaderName: values.authHeaderName || undefined,
        authHeaderValue: values.authHeaderValue || undefined,
      })
      toast.success("MCP connection added")
      setOpen(false)
      await onCreated()
      return {}
    } catch (error) {
      const validation =
        extractValidationError<keyof AddMcpConnectionFormSchemaType>(error)
      if (validation) {
        return { error: validation }
      }
      toast.error(extractErrorMessage(error, "Failed to add MCP connection"))
      return {}
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <Plus />
          Add
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add MCP connection</DialogTitle>
          <DialogDescription>
            Header auth values are stored in the node vault and never returned.
          </DialogDescription>
        </DialogHeader>
        <AddMcpConnectionForm onFormComplete={onFormComplete} />
        <DialogFooter>
          <Button type="submit" form="mcp-create-form" disabled={saving}>
            {saving ? <Loader2 className="animate-spin" /> : null}
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
