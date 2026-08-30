import React from "react"
import { toast } from "sonner"
import { Wrench } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Switch } from "@/components/ui/switch"
import { Table, TableBody, TableCell, TableRow } from "@/components/ui/table"
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
        prev.map((t) =>
          t.id === tool.id ? { ...t, enabled: tool.enabled } : t
        )
      )
      toast.error(extractErrorMessage(error, "Failed to update tool"))
    }
  }

  const groups = React.useMemo(() => {
    const byGroup = new Map<string, SelfTool[]>()
    for (const tool of tools) {
      const group = tool.group || "Other"
      byGroup.set(group, [...(byGroup.get(group) ?? []), tool])
    }
    return [...byGroup.entries()].sort(([a], [b]) => a.localeCompare(b))
  }, [tools])

  return (
    <Card className="lg:h-full lg:min-h-0 lg:overflow-hidden">
      <CardHeader>
        <CardTitle>Built-in tools</CardTitle>
        <CardDescription>
          Built-in tools the agent can call on itself during reasoning.
        </CardDescription>
      </CardHeader>
      <CardContent className="lg:flex lg:min-h-0 lg:flex-1 lg:flex-col">
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
          <div className="max-h-[28rem] overflow-y-auto lg:max-h-none lg:min-h-0 lg:flex-1">
            <Table>
              <TableBody>
                {groups.map(([group, groupTools]) => (
                  <React.Fragment key={group}>
                    <TableRow className="border-b-0 hover:bg-transparent">
                      <TableCell
                        colSpan={2}
                        className="sticky top-0 z-10 bg-card pt-4 pb-2"
                      >
                        <Badge
                          variant="secondary"
                          className="text-sm font-semibold"
                        >
                          {group}
                        </Badge>
                      </TableCell>
                    </TableRow>
                    {groupTools.map((tool) => (
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
                        <TableCell className="w-24 text-right">
                          <Switch
                            checked={tool.enabled}
                            onCheckedChange={(enabled) =>
                              onToggle(tool, enabled)
                            }
                            aria-label={`Toggle ${tool.name}`}
                          />
                        </TableCell>
                      </TableRow>
                    ))}
                  </React.Fragment>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
