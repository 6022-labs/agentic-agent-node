import * as React from "react"
import { ChevronDown } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

export type MultiSelectOption = {
  value: string
  // Defaults to the value; pass a node to style it (e.g. monospace).
  label?: React.ReactNode
  description?: React.ReactNode
}

type Props = {
  options: MultiSelectOption[]
  selected: string[]
  onSelectedChange: (selected: string[]) => void
  // Trigger label when nothing is selected.
  placeholder?: string
  // Plural noun used in the "3 of 12 …" summary.
  itemsLabel?: string
  disabled?: boolean
  className?: string
}

export function MultiSelect({
  options,
  selected,
  onSelectedChange,
  placeholder = "None selected",
  itemsLabel = "selected",
  disabled,
  className,
}: Props) {
  const toggle = (value: string, checked: boolean) =>
    onSelectedChange(
      checked
        ? [...selected, value]
        : selected.filter((entry) => entry !== value)
    )

  const summary = () => {
    if (selected.length === 0) return placeholder
    if (selected.length === options.length) {
      return `All ${options.length} ${itemsLabel}`
    }
    return `${selected.length} of ${options.length} ${itemsLabel}`
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="outline"
          disabled={disabled}
          className={cn("w-full justify-between font-normal", className)}
        >
          <span className={cn(selected.length === 0 && "text-muted-foreground")}>
            {summary()}
          </span>
          <ChevronDown className="opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      {/* Scrolls instead of growing: the list can hold dozens of options. */}
      <DropdownMenuContent
        align="start"
        className="max-h-72 w-(--radix-dropdown-menu-trigger-width) overflow-y-auto"
      >
        <DropdownMenuItem
          onSelect={(event) => {
            event.preventDefault()
            onSelectedChange(options.map((option) => option.value))
          }}
        >
          Select all
        </DropdownMenuItem>
        <DropdownMenuItem
          onSelect={(event) => {
            event.preventDefault()
            onSelectedChange([])
          }}
        >
          Clear
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        {options.map((option) => (
          <DropdownMenuCheckboxItem
            key={option.value}
            checked={selected.includes(option.value)}
            // Keeps the menu open so several options can be picked in one go.
            onSelect={(event) => event.preventDefault()}
            onCheckedChange={(checked) => toggle(option.value, checked)}
          >
            <span className="min-w-0">
              <span className="block truncate">{option.label ?? option.value}</span>
              {option.description ? (
                <span className="block truncate text-xs text-muted-foreground">
                  {option.description}
                </span>
              ) : null}
            </span>
          </DropdownMenuCheckboxItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
