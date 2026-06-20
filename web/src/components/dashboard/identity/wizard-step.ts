import {
  isDraftComplete,
  type AgentRegistration,
} from "@/models/agent-registration"
import { TRANSACTION_STATUS } from "@/models/enums/transaction-status"

export type WizardStepNumber = 1 | 2 | 3

// Where setup left off: details → mint → ENS. An incomplete draft stays on
// details so "Continue setup" can't jump to mint with missing fields.
export function computeWizardStep(
  registration: AgentRegistration | null
): WizardStepNumber {
  if (!isDraftComplete(registration)) return 1
  if (registration!.mintStatus !== TRANSACTION_STATUS.confirmed) return 2
  return 3
}
