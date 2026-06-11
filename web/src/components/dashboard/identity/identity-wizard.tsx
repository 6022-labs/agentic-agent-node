import React from "react"
import { ArrowLeft } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import { TRANSACTION_STATUS } from "@/models/agent-registration"
import RegistrationDraftForm from "./forms/registration-draft-form"
import StepIndicator from "./step-indicator"
import MintStep from "./steps/mint-step"
import EnsStep from "./steps/ens-step"
import { computeWizardStep, type WizardStepNumber } from "./wizard-step"

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialStep?: WizardStepNumber
}

export default function IdentityWizard({
  open,
  onOpenChange,
  initialStep,
}: Props) {
  const { registration, setRegistration } = useAgentIdentity()
  const [step, setStep] = React.useState<WizardStepNumber>(1)

  React.useEffect(() => {
    if (open) setStep(initialStep ?? computeWizardStep(registration))
    // Re-derive only when the dialog opens, not on every status refresh.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const isMinted = registration?.mintStatus === TRANSACTION_STATUS.confirmed

  const steps = [
    { label: "Details", done: registration != null },
    { label: "Mint", done: !!isMinted },
    {
      label: "ENS",
      done: registration?.ensStatus === TRANSACTION_STATUS.confirmed,
    },
  ]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {registration ? "Identity setup" : "Create identity"}
          </DialogTitle>
          <DialogDescription>
            Describe the agent, mint it on-chain, then publish its ENS records.
          </DialogDescription>
        </DialogHeader>

        <StepIndicator steps={steps} current={step} />

        {step === 1 ? (
          <RegistrationDraftForm
            draft={registration}
            onSaved={setRegistration}
            onContinue={() => setStep(2)}
          />
        ) : null}

        {step === 2 && registration ? (
          <div className="space-y-4">
            <MintStep draft={registration} onAdvance={() => setStep(3)} />
            {!isMinted ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setStep(1)}
                className="text-muted-foreground"
              >
                <ArrowLeft />
                Edit details
              </Button>
            ) : null}
          </div>
        ) : null}

        {step === 3 && registration ? (
          <EnsStep draft={registration} onClose={() => onOpenChange(false)} />
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
