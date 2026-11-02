import { ExternalLink, Globe } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { config } from "@/configs/env"
import { EVERYONE_PATTERN } from "@/models/payment-receive-policy"
import { audienceLabel } from "@/lib/token-format"

const PAYMENTS_DOCS_URL = `${config.docsBaseUrl}/docs/agent/payments`

// DocsLink points at the payments guide instead of explaining rules inline.
export function DocsLink() {
  return (
    <a
      href={PAYMENTS_DOCS_URL}
      target="_blank"
      rel="noreferrer"
      className="inline-flex items-center gap-1 underline-offset-4 hover:underline"
    >
      Learn more
      <ExternalLink className="size-3" />
    </a>
  )
}

// AudienceBadge folds a pattern list into one badge; its tooltip lists every pattern.
export function AudienceBadge({
  patterns,
  anyoneLabel = "anyone",
}: {
  patterns: string[]
  anyoneLabel?: string
}) {
  if (patterns.includes(EVERYONE_PATTERN)) {
    return (
      <Badge>
        <Globe />
        {anyoneLabel}
      </Badge>
    )
  }
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button type="button" aria-label={`Show all ${patterns.length}`}>
          <Badge variant="outline" className="font-mono">
            {audienceLabel(patterns, anyoneLabel)}
          </Badge>
        </button>
      </TooltipTrigger>
      <TooltipContent className="block max-w-sm p-2">
        <ul className="max-h-60 space-y-0.5 overflow-y-auto font-mono text-xs select-text">
          {patterns.map((pattern, index) => (
            <li key={`${pattern}-${index}`} className="break-all">
              {pattern}
            </li>
          ))}
        </ul>
      </TooltipContent>
    </Tooltip>
  )
}
