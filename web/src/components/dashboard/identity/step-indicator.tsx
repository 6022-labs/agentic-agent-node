import React from "react"
import { Check } from "lucide-react"

import { cn } from "@/lib/utils"

export type WizardStep = {
  label: string
  done: boolean
}

export default function StepIndicator({
  steps,
  current,
}: {
  steps: WizardStep[]
  current: number
}) {
  return (
    <div className="flex items-center gap-2">
      {steps.map((step, i) => {
        const index = i + 1
        const active = index === current
        return (
          <React.Fragment key={step.label}>
            {i > 0 ? <div className="h-px min-w-4 flex-1 bg-border" /> : null}
            <div className="flex items-center gap-1.5">
              <span
                className={cn(
                  "flex size-7 shrink-0 items-center justify-center rounded-full text-xs",
                  step.done
                    ? "bg-primary text-primary-foreground"
                    : active
                      ? "border-2 border-primary font-medium"
                      : "bg-muted text-muted-foreground"
                )}
              >
                {step.done ? <Check className="size-4" /> : index}
              </span>
              <span
                className={cn(
                  "text-sm",
                  active ? "font-medium" : "text-muted-foreground"
                )}
              >
                {step.label}
              </span>
            </div>
          </React.Fragment>
        )
      })}
    </div>
  )
}
