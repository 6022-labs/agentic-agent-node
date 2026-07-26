import { Skeleton } from "@/components/ui/skeleton"
import {
  FormDescription,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { MultiSelect } from "@/components/ui/multi-select"
import type { McpServerTool } from "@/models/mcp-connection"

type Props = {
  tools: McpServerTool[] | null
  loading: boolean
  selected: string[]
  onSelectedChange: (selected: string[]) => void
}

export default function AllowedToolsField({
  tools,
  loading,
  selected,
  onSelectedChange,
}: Props) {
  return (
    <FormItem>
      <FormLabel>Allowed tools</FormLabel>
      {loading ? (
        <Skeleton className="h-9 w-full" />
      ) : tools === null ? (
        <FormDescription>
          Load the server's tools to choose which ones the agent may call.
        </FormDescription>
      ) : tools.length === 0 ? (
        <FormDescription>This server exposes no tool.</FormDescription>
      ) : (
        <MultiSelect
          options={tools.map((tool) => ({
            value: tool.name,
            label: <span className="font-mono text-xs">{tool.name}</span>,
            description: tool.description,
          }))}
          selected={selected}
          onSelectedChange={onSelectedChange}
          placeholder="No tool allowed"
          itemsLabel="tools"
        />
      )}
      <FormMessage />
    </FormItem>
  )
}
