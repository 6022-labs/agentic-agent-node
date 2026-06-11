import React from "react"
import { toast } from "sonner"
import { Plug, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import ConfirmDialog from "@/components/dashboard/confirm-dialog"
import EmptyState from "@/components/dashboard/empty-state"
import AddMcpConnectionDialog from "@/components/dashboard/mcp/dialogs/add-mcp-connection-dialog"
import type { McpConnection } from "@/models/mcp-connection"
import {
  deleteMcpConnection,
  listMcpConnections,
} from "@/queries/api/mcp.queries"
import { extractErrorMessage } from "@/lib/api-errors"

export default function McpConfigSection() {
  const [connections, setConnections] = React.useState<McpConnection[]>([])
  const [loading, setLoading] = React.useState(true)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await listMcpConnections()
      setConnections(data.connections ?? [])
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to load MCP connections"))
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void load()
  }, [load])

  const onDelete = async (id: string) => {
    try {
      await deleteMcpConnection(id)
      setConnections((prev) => prev.filter((c) => c.id !== id))
      toast.success("MCP connection removed")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to remove connection"))
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>MCP connections</CardTitle>
            <CardDescription>
              External tool servers the agent can call during reasoning.
            </CardDescription>
          </div>
          <AddMcpConnectionDialog onCreated={load} />
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <ul className="divide-y">
            {[0, 1, 2].map((i) => (
              <li
                key={i}
                className="flex items-center justify-between gap-4 py-3"
              >
                <div className="min-w-0 flex-1 space-y-2">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-3 w-56" />
                </div>
                <Skeleton className="size-8 rounded-md" />
              </li>
            ))}
          </ul>
        ) : connections.length === 0 ? (
          <EmptyState
            icon={Plug}
            title="No MCP connections yet"
            description="Connect an external tool server so the agent can call its tools while reasoning."
          />
        ) : (
          <ul className="divide-y">
            {connections.map((connection) => (
              <li
                key={connection.id}
                className="flex items-center justify-between gap-4 py-3"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="truncate font-medium">
                      {connection.name}
                    </span>
                    {connection.headerAuthKeys?.length ? (
                      <Badge variant="secondary">auth</Badge>
                    ) : null}
                  </div>
                  <p className="truncate text-xs text-muted-foreground">
                    {connection.url}
                  </p>
                </div>
                <ConfirmDialog
                  title="Remove MCP connection?"
                  description={`The agent will no longer be able to call tools from "${connection.name}".`}
                  confirmLabel="Remove"
                  onConfirm={() => onDelete(connection.id)}
                >
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label="Remove connection"
                  >
                    <Trash2 />
                  </Button>
                </ConfirmDialog>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  )
}
