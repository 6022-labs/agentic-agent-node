import PaymentReceivePolicySection from "@/components/dashboard/payments/receive-section"
import PaymentSpendPolicySection from "@/components/dashboard/payments/spend-section"

export default function DashboardPaymentsPage() {
  return (
    <div className="w-full space-y-6">
      <PaymentReceivePolicySection />
      <PaymentSpendPolicySection />
    </div>
  )
}
