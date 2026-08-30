import React from "react"
import { toast } from "sonner"
import {
  ChevronLeft,
  ChevronRight,
  MessagesSquare,
  Pencil,
  Plus,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import EmptyState from "@/components/dashboard/empty-state"
import SessionDialog from "@/components/dashboard/memory/dialogs/session-dialog"
import { extractErrorMessage } from "@/lib/api-errors"
import { formatRelativeTime } from "@/lib/format"
import type { Conversation } from "@/models/conversation"
import { listConversations } from "@/queries/api/conversations.queries"

const ELEMENTS_PER_PAGE = 10

export default function SessionsTable() {
  const [loading, setLoading] = React.useState(true)
  const [conversations, setConversations] = React.useState<Conversation[]>([])
  const [page, setPage] = React.useState(1)
  const [totalElements, setTotalElements] = React.useState(0)
  const [dialogOpen, setDialogOpen] = React.useState(false)
  const [editing, setEditing] = React.useState<Conversation | undefined>()

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await listConversations(page, ELEMENTS_PER_PAGE)
      setConversations(data.conversations ?? [])
      setTotalElements(data.totalElements)
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to load memory sessions"))
    } finally {
      setLoading(false)
    }
  }, [page])

  React.useEffect(() => {
    void load()
  }, [load])

  const totalPages = Math.max(Math.ceil(totalElements / ELEMENTS_PER_PAGE), 1)

  const openCreate = () => {
    setEditing(undefined)
    setDialogOpen(true)
  }

  const openEdit = (conversation: Conversation) => {
    setEditing(conversation)
    setDialogOpen(true)
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={openCreate}>
          <Plus />
          New session
        </Button>
      </div>

      {loading ? (
        <div className="space-y-2">
          <Skeleton className="h-9 w-full" />
          <Skeleton className="h-9 w-full" />
          <Skeleton className="h-9 w-full" />
        </div>
      ) : conversations.length === 0 ? (
        <EmptyState
          icon={MessagesSquare}
          title="No memory session yet"
          description="Sessions appear as soon as the agent holds a conversation, and can be created by hand."
        />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Id</TableHead>
              <TableHead>Messages</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {conversations.map((conversation) => (
              <TableRow key={conversation.conversationId}>
                <TableCell>
                  {conversation.name ?? (
                    <span className="text-muted-foreground">Untitled</span>
                  )}
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {conversation.conversationId}
                </TableCell>
                <TableCell>{conversation.messageCount}</TableCell>
                <TableCell>
                  {formatRelativeTime(conversation.updatedAt)}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="ghost"
                    size="icon"
                    title="Edit"
                    onClick={() => openEdit(conversation)}
                  >
                    <Pencil />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">
          Page {page} of {totalPages} · {totalElements} session
          {totalElements === 1 ? "" : "s"}
        </span>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={loading || page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            <ChevronLeft />
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={loading || page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
            <ChevronRight />
          </Button>
        </div>
      </div>

      <SessionDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        conversation={editing}
        onSaved={load}
      />
    </div>
  )
}
