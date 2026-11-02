import PaymentReceivePolicySection from "@/components/dashboard/payments/receive-section"
import PaymentSpendPolicySection from "@/components/dashboard/payments/spend-section"

export default function DashboardPaymentsPage() {
  return (
    <div className="grid w-full items-start gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] lg:items-stretch">
      <PaymentReceivePolicySection />
      <PaymentSpendPolicySection />
    </div>
  )
}
