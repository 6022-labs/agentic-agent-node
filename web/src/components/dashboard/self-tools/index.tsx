import React from "react"
import { toast } from "sonner"
import { Wrench } from "lucide-react"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Switch } from "@/components/ui/switch"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import EmptyState from "@/components/dashboard/empty-state"
import type { SelfTool } from "@/models/self-tool"
import { listSelfTools, updateSelfTool } from "@/queries/api/self-tools.queries"
import { extractErrorMessage } from "@/lib/api-errors"

export default function SelfToolsSection() {
  const [tools, setTools] = React.useState<SelfTool[]>([])
  const [loading, setLoading] = React.useState(true)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await listSelfTools()
      setTools(data.tools ?? [])
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to load built-in tools"))
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void load()
  }, [load])

  const onToggle = async (tool: SelfTool, enabled: boolean) => {
    setTools((prev) =>
      prev.map((t) => (t.id === tool.id ? { ...t, enabled } : t))
    )
    try {
      await updateSelfTool(tool.id, { enabled })
    } catch (error) {
      setTools((prev) =>
        prev.map((t) => (t.id === tool.id ? { ...t, enabled: tool.enabled } : t))
      )
      toast.error(extractErrorMessage(error, "Failed to update tool"))
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Built-in tools</CardTitle>
        <CardDescription>
          Tools the agent can call on itself during reasoning. Disabled tools
          are not offered to the model.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="space-y-2">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        ) : tools.length === 0 ? (
          <EmptyState
            icon={Wrench}
            title="No built-in tools"
            description="The runtime has not registered any built-in tools yet."
          />
        ) : (
          <div className="max-h-[28rem] overflow-y-auto">
            <Table>
              <TableHeader className="sticky top-0 z-10 bg-card">
                <TableRow>
                  <TableHead>Tool</TableHead>
                  <TableHead className="w-24 text-right">Enabled</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tools.map((tool) => (
                  <TableRow key={tool.id}>
                    <TableCell>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span
                            tabIndex={0}
                            className="cursor-help font-mono text-xs font-medium underline decoration-dotted underline-offset-4"
                          >
                            {tool.name}
                          </span>
                        </TooltipTrigger>
                        <TooltipContent className="max-w-sm">
                          {tool.description}
                        </TooltipContent>
                      </Tooltip>
                    </TableCell>
                    <TableCell className="text-right">
                      <Switch
                        checked={tool.enabled}
                        onCheckedChange={(enabled) => onToggle(tool, enabled)}
                        aria-label={`Toggle ${tool.name}`}
                      />
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
