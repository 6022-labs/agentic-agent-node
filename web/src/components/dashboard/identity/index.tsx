import React from "react"
import { toast } from "sonner"
import {
  ArrowRight,
  BadgePlus,
  ExternalLink,
  Fingerprint,
  Pencil,
  RefreshCw,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import EmptyState from "@/components/dashboard/empty-state"
import EnsDriftAlert from "@/components/dashboard/ens-drift-alert"
import IdentityWizard from "@/components/dashboard/identity/identity-wizard"
import {
  computeWizardStep,
  type WizardStepNumber,
} from "@/components/dashboard/identity/wizard-step"
import { getAppConfig } from "@/configs/app-config"
import { getEnsAppUrl } from "@/configs/networks"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import {
  checkAgentEns,
  checkAgentMint,
} from "@/queries/api/registration.queries"
import { extractErrorMessage } from "@/lib/api-errors"

const statusVariant = (status: string) => {
  if (status === "confirmed") return "default"
  if (status === "failed") return "destructive"
  return "secondary"
}

export default function IdentitySection() {
  const { registration: draft, loading, refresh } = useAgentIdentity()
  const [refreshing, setRefreshing] = React.useState(false)
  const [wizardOpen, setWizardOpen] = React.useState(false)
  const [wizardStep, setWizardStep] = React.useState<WizardStepNumber>(1)

  const isMinted = draft?.mintStatus === "confirmed"
  const isEnsRegistered = draft?.ensStatus === "confirmed"
  const isComplete = isMinted && isEnsRegistered

  const openWizard = (step?: WizardStepNumber) => {
    setWizardStep(step ?? computeWizardStep(draft))
    setWizardOpen(true)
  }

  const refreshStatus = async () => {
    if (!draft) return
    setRefreshing(true)
    try {
      await checkAgentMint()
      if (draft.mintStatus === "confirmed") {
        await checkAgentEns()
      }
      await refresh()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to refresh status"))
    } finally {
      setRefreshing(false)
    }
  }

  if (loading && !draft) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Identity</CardTitle>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="grid gap-3 sm:grid-cols-2">
            {[0, 1, 2, 3].map((i) => (
              <div key={i} className="space-y-1.5">
                <Skeleton className="h-3 w-16" />
                <Skeleton className="h-4 w-40" />
              </div>
            ))}
          </div>
          <div className="flex gap-2">
            <Skeleton className="h-5 w-16 rounded-full" />
            <Skeleton className="h-5 w-20 rounded-full" />
          </div>
        </CardContent>
      </Card>
    )
  }

  const wizard = (
    <IdentityWizard
      open={wizardOpen}
      onOpenChange={setWizardOpen}
      initialStep={wizardStep}
    />
  )

  // No draft yet — single guided entry point into the setup wizard.
  if (!draft) {
    return (
      <Card>
        <CardContent>
          <EmptyState
            icon={Fingerprint}
            title="No on-chain identity yet"
            description="Describe the agent, mint it as an NFT and publish its ENS records in a short guided setup."
            action={
              <Button onClick={() => openWizard(1)}>
                <BadgePlus />
                Set up identity
              </Button>
            }
          />
        </CardContent>
        {wizard}
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div>
            <CardTitle>Identity</CardTitle>
            <CardDescription>
              {isComplete
                ? "The agent is minted and its ENS records are published."
                : "Finish the setup to mint the agent and register its ENS."}
            </CardDescription>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={refreshStatus}
            disabled={refreshing}
          >
            <RefreshCw className={refreshing ? "animate-spin" : ""} />
            Refresh
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-5">
        <EnsDriftAlert />
        <div className="grid gap-3 sm:grid-cols-2">
          <Field label="Name" value={draft.name} />
          <Field label="Role" value={draft.role} />
          <Field label="Collection" value={draft.collectionAddress} mono />
          <Field
            label="Token id"
            value={draft.tokenId != null ? `#${draft.tokenId}` : "—"}
          />
          {draft.walletAddress ? (
            <Field label="Agent wallet" value={draft.walletAddress} mono />
          ) : null}
          {draft.ensDomain ? (
            <Field
              label="ENS domain"
              value={draft.ensDomain}
              mono
              href={getEnsAppUrl(getAppConfig().chainId, draft.ensDomain)}
            />
          ) : null}
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm text-muted-foreground">Mint</span>
          <Badge variant={statusVariant(draft.mintStatus)}>
            {draft.mintStatus}
          </Badge>
          <span className="ml-3 text-sm text-muted-foreground">ENS</span>
          <Badge variant={statusVariant(draft.ensStatus)}>
            {draft.ensStatus}
          </Badge>
          {!isComplete ? (
            <span className="ml-auto flex items-center gap-2">
              {!isMinted ? (
                <Button variant="ghost" size="sm" onClick={() => openWizard(1)}>
                  <Pencil />
                  Edit details
                </Button>
              ) : null}
              <Button size="sm" onClick={() => openWizard()}>
                Continue setup
                <ArrowRight />
              </Button>
            </span>
          ) : null}
        </div>
      </CardContent>
      {wizard}
    </Card>
  )
}

function Field({
  label,
  value,
  mono,
  href,
}: {
  label: string
  value: string
  mono?: boolean
  href?: string | null
}) {
  return (
    <div className="min-w-0">
      <p className="text-xs text-muted-foreground">{label}</p>
      {href ? (
        <a
          href={href}
          target="_blank"
          rel="noreferrer"
          className={`flex items-center gap-1 text-sm hover:underline ${mono ? "font-mono" : ""}`}
        >
          <span className="truncate">{value}</span>
          <ExternalLink className="size-3 shrink-0" />
        </a>
      ) : (
        <p className={`truncate text-sm ${mono ? "font-mono" : ""}`}>{value}</p>
      )}
    </div>
  )
}
