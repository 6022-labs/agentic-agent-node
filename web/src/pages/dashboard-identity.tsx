import IdentitySection from "@/components/dashboard/identity"
import SoulSection from "@/components/dashboard/identity/soul-section"
import GuidanceSection from "@/components/dashboard/identity/guidance-section"

export default function DashboardIdentityPage() {
  return (
    <div className="w-full space-y-6">
      <IdentitySection />
      <SoulSection />
      <GuidanceSection />
    </div>
  )
}
