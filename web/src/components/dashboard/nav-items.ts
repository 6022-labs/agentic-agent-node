import {
  Brain,
  BrainCircuit,
  Coins,
  Fingerprint,
  MessagesSquare,
  Wrench,
  type LucideIcon,
} from "lucide-react"

export type NavItem = {
  title: string
  path: string
  icon: LucideIcon
}

export const configureNavItems: NavItem[] = [
  { title: "Identity", path: "/dashboard/identity", icon: Fingerprint },
  { title: "LLM", path: "/dashboard/llm", icon: Brain },
  { title: "Tools", path: "/dashboard/mcp", icon: Wrench },
  { title: "Memory", path: "/dashboard/memory", icon: BrainCircuit },
  { title: "Payments", path: "/dashboard/payments", icon: Coins },
]

export const testNavItems: NavItem[] = [
  { title: "Playground", path: "/dashboard/playground", icon: MessagesSquare },
]

const allNavItems = [...configureNavItems, ...testNavItems]

export function navTitleForPath(pathname: string): string {
  return allNavItems.find((item) => item.path === pathname)?.title ?? ""
}
