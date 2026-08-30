import React from "react"
import { toast } from "sonner"
import { ArrowLeft, ArrowRight, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import { TRANSACTION_STATUS } from "@/models/enums/transaction-status"
import { isDraftComplete } from "@/models/agent-registration"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { saveAgentRegistration } from "@/queries/api/registration.queries"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"
import RegistrationDraftForm, {
  REGISTRATION_DRAFT_FORM_ID,
} from "./forms/registration-draft-form"
import type { RegistrationDraftFormSchemaType } from "./forms/registration-draft-form/registration-draft-form-schema"
import { useAgentCollections } from "./use-agent-collections"
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
  const [savingDraft, setSavingDraft] = React.useState(false)
  const [draftValues, setDraftValues] =
    React.useState<RegistrationDraftFormSchemaType | null>(null)

  // Collections are per-chain: reload them whenever the form's chain changes.
  const { collections, loadCollections } = useAgentCollections(
    draftValues?.chainId
  )

  const onDraftComplete = async (
    values: RegistrationDraftFormSchemaType
  ): Promise<FormCompletionResult<keyof RegistrationDraftFormSchemaType>> => {
    setSavingDraft(true)
    try {
      const { data } = await saveAgentRegistration({
        chainId: values.chainId,
        collectionAddress: values.collectionAddress,
        name: values.name,
        image: values.image,
        attributes: Object.fromEntries(
          values.attributes.map((attribute) => [
            attribute.key.trim(),
            attribute.value,
          ])
        ),
      })
      setRegistration(data)
      toast.success("Draft saved")
      return {}
    } catch (error) {
      const validation =
        extractValidationError<keyof RegistrationDraftFormSchemaType>(error)
      // Only fields the form renders can carry an inline error; the rest are toasted.
      if (validation && draftValues && validation.field in draftValues) {
        return { error: validation }
      }
      toast.error(extractErrorMessage(error, "Failed to save draft"))
      return {}
    } finally {
      setSavingDraft(false)
    }
  }

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
          <>
            <RegistrationDraftForm
              draft={registration}
              collections={collections}
              onReloadCollections={loadCollections}
              onValueChange={setDraftValues}
              onFormComplete={onDraftComplete}
            />
            <div className="flex justify-end gap-2">
              <Button
                type="submit"
                form={REGISTRATION_DRAFT_FORM_ID}
                variant={registration ? "outline" : "default"}
                disabled={savingDraft}
              >
                {savingDraft ? <Loader2 className="animate-spin" /> : null}
                {registration ? "Update draft" : "Save draft"}
              </Button>
              <Button
                type="button"
                onClick={() => setStep(2)}
                disabled={!isDraftComplete(registration) || savingDraft}
              >
                Continue
                <ArrowRight />
              </Button>
            </div>
          </>
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
