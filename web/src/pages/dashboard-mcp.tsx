import McpConfigSection from "@/components/dashboard/mcp"
import SelfToolsSection from "@/components/dashboard/self-tools"

export default function DashboardMcpPage() {
  return (
    <div className="w-full space-y-6">
      <SelfToolsSection />
      <McpConfigSection />
    </div>
  )
}
