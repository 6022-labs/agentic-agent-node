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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import ConfirmDialog from "@/components/dashboard/confirm-dialog"
import EmptyState from "@/components/dashboard/empty-state"
import McpConnectionDialog from "@/components/dashboard/mcp/dialogs/mcp-connection-dialog"
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
    <Card className="lg:h-full lg:min-h-0 lg:overflow-hidden">
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>MCP connections</CardTitle>
            <CardDescription>
              External tool servers the agent can call during reasoning.
            </CardDescription>
          </div>
          <McpConnectionDialog onSaved={load} />
        </div>
      </CardHeader>
      <CardContent className="lg:flex lg:min-h-0 lg:flex-1 lg:flex-col">
        {loading ? (
          <div className="space-y-2">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : connections.length === 0 ? (
          <EmptyState
            icon={Plug}
            title="No MCP connections yet"
            description="Connect an external tool server so the agent can call its tools while reasoning."
          />
        ) : (
          <div className="max-h-[28rem] overflow-y-auto lg:max-h-none lg:min-h-0 lg:flex-1">
            <Table>
              <TableHeader className="sticky top-0 z-10 bg-card">
                <TableRow>
                  <TableHead>Connection</TableHead>
                  <TableHead>Server URL</TableHead>
                  <TableHead>Tools</TableHead>
                  <TableHead>Auth</TableHead>
                  <TableHead className="w-24" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {connections.map((connection) => (
                  <TableRow key={connection.id}>
                    <TableCell className="font-medium">
                      {connection.name}
                    </TableCell>
                    <TableCell>
                      <span
                        className="font-mono text-xs text-muted-foreground"
                        title={connection.url}
                      >
                        {connection.url}
                      </span>
                    </TableCell>
                    <TableCell>
                      {connection.allowedToolPatterns?.length ? (
                        <Badge variant="secondary">
                          {connection.allowedToolPatterns.length} allowed
                        </Badge>
                      ) : (
                        <Badge variant="outline">all</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {connection.headerAuthKeys?.length ? (
                        <Badge variant="secondary">auth</Badge>
                      ) : (
                        <Badge variant="outline">none</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <McpConnectionDialog
                        connection={connection}
                        onSaved={load}
                      />
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
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
