import McpConfigSection from "@/components/dashboard/mcp"
import SelfToolsSection from "@/components/dashboard/self-tools"

export default function DashboardMcpPage() {
  return (
    <div className="grid w-full items-start gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)]">
      <SelfToolsSection />
      <McpConfigSection />
    </div>
  )
}
