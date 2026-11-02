import type { ReactNode } from "react"

import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"

// HoverHint is the payments page's one hover style, the same as the audience badge's list.
export default function HoverHint({
  content,
  children,
  className,
}: {
  content: ReactNode
  children: ReactNode
  className?: string
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className={className}>{children}</span>
      </TooltipTrigger>
      <TooltipContent className="block max-w-sm p-2 text-xs select-text">
        {content}
      </TooltipContent>
    </Tooltip>
  )
}
