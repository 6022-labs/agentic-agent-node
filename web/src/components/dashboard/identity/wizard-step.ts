import type { AgentRegistration } from "@/models/agent-registration"
import { TRANSACTION_STATUS } from "@/models/enums/transaction-status"

export type WizardStepNumber = 1 | 2 | 3

// Where setup left off: details → mint → ENS.
export function computeWizardStep(
  registration: AgentRegistration | null
): WizardStepNumber {
  if (!registration) return 1
  if (registration.mintStatus !== TRANSACTION_STATUS.confirmed) return 2
  return 3
}
