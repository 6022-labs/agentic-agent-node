import IdentityNotice from "@/components/dashboard/identity-notice"
import PaymentPolicySection from "@/components/dashboard/payments"

export default function DashboardPaymentsPage() {
  return (
    <div className="w-full">
      <IdentityNotice />
      <PaymentPolicySection />
    </div>
  )
}
