import type { ReactNode } from "react"

import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"

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
        {/* Focusable so keyboard users can reveal the hint too. */}
        <span
          tabIndex={0}
          className={cn(
            "rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
            className
          )}
        >
          {children}
        </span>
      </TooltipTrigger>
      <TooltipContent className="block max-w-sm p-2 text-xs select-text">
        {content}
      </TooltipContent>
    </Tooltip>
  )
}
