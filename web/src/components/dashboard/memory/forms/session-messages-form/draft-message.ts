import type { ConversationMessageRole } from "@/models/enums/conversation-message-role"
import { newId } from "@/lib/utils"

export type DraftMessage = {
  key: string
  role: ConversationMessageRole
  content: string
}

export const emptyMessage = (): DraftMessage => ({
  key: newId(),
  role: "user",
  content: "",
})
