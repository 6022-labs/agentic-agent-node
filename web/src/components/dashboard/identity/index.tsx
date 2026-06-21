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

import { Button, buttonVariants } from "@/components/ui/button"
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
import EnsDriftAlert from "./ens-drift-alert"
import IdentityWizard from "@/components/dashboard/identity/identity-wizard"
import {
  computeWizardStep,
  type WizardStepNumber,
} from "@/components/dashboard/identity/wizard-step"
import {
  getEnsAppUrl,
  getOpenSeaUrl,
  getRaribleUrl,
  getScannerAddressUrl,
} from "@/configs/networks"
import { useAgentIdentity } from "@/contexts/agent-identity-context"
import {
  checkAgentEns,
  checkAgentMint,
} from "@/queries/api/registration.queries"
import { extractErrorMessage } from "@/lib/api-errors"
import { ipfsImageUrl } from "@/lib/ipfs"

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

  // Reconcile from chain: pick up a mint (or moderator-approved proposal), then
  // its ENS, off the fresh status rather than the stale draft.
  const reconcileStatus = React.useCallback(async () => {
    const { data: mint } = await checkAgentMint()
    if (mint.mintStatus === "confirmed") {
      await checkAgentEns()
    }
    await refresh()
  }, [refresh])

  const refreshStatus = async () => {
    if (!draft) return
    setRefreshing(true)
    try {
      await reconcileStatus()
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to refresh status"))
    } finally {
      setRefreshing(false)
    }
  }

  const autoReconciledRef = React.useRef(false)
  React.useEffect(() => {
    if (autoReconciledRef.current) return
    if (draft?.mintStatus !== "submitted") return
    autoReconciledRef.current = true
    reconcileStatus().catch(() => {})
  }, [draft?.mintStatus, reconcileStatus])

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

  const imageUrl = ipfsImageUrl(draft.images?.default)
  const openSeaUrl = getOpenSeaUrl(
    draft.chainId,
    draft.collectionAddress,
    draft.tokenId
  )
  const raribleUrl = getRaribleUrl(
    draft.chainId,
    draft.collectionAddress,
    draft.tokenId
  )

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
        <div className="flex flex-col gap-6 sm:flex-row sm:items-stretch">
          {imageUrl || openSeaUrl || raribleUrl ? (
            <div className="flex w-full shrink-0 flex-col gap-2 sm:w-64">
              {imageUrl ? (
                <img
                  src={imageUrl}
                  alt={draft.name}
                  className="w-full rounded-lg border"
                />
              ) : null}
              {openSeaUrl || raribleUrl ? (
                <div className="grid grid-cols-2 gap-2">
                  {openSeaUrl ? (
                    <a
                      href={openSeaUrl}
                      target="_blank"
                      rel="noreferrer"
                      className={buttonVariants({
                        variant: "outline",
                        size: "sm",
                        className: "w-full",
                      })}
                    >
                      OpenSea
                      <ExternalLink />
                    </a>
                  ) : null}
                  {raribleUrl ? (
                    <a
                      href={raribleUrl}
                      target="_blank"
                      rel="noreferrer"
                      className={buttonVariants({
                        variant: "outline",
                        size: "sm",
                        className: "w-full",
                      })}
                    >
                      Rarible
                      <ExternalLink />
                    </a>
                  ) : null}
                </div>
              ) : null}
            </div>
          ) : null}
          <div className="flex flex-1 flex-col justify-center gap-6">
            <div className="space-y-3">
              <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
                <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
                  {draft.name}
                </h2>
                {draft.role ? (
                  <Badge variant="secondary" className="capitalize">
                    {draft.role}
                  </Badge>
                ) : null}
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-sm text-muted-foreground">Mint</span>
                <Badge variant={statusVariant(draft.mintStatus)}>
                  {draft.mintStatus}
                </Badge>
                <span className="ml-2 text-sm text-muted-foreground">ENS</span>
                <Badge variant={statusVariant(draft.ensStatus)}>
                  {draft.ensStatus}
                </Badge>
                {!isComplete ? (
                  <span className="ml-auto flex items-center gap-2">
                    {!isMinted ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => openWizard(1)}
                      >
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
              {draft.mintStatus === "submitted" && draft.proposalId != null ? (
                <p className="text-sm text-muted-foreground">
                  Awaiting a moderator's approval of the mint proposal — this
                  can take a while. Check back later.
                </p>
              ) : null}
            </div>

            <dl className="divide-y rounded-lg border">
              <Row
                label="Collection"
                value={draft.collectionAddress}
                mono
                href={getScannerAddressUrl(
                  draft.chainId,
                  draft.collectionAddress
                )}
              />
              <Row
                label="Token id"
                value={draft.tokenId != null ? `#${draft.tokenId}` : "—"}
              />
              {draft.walletAddress ? (
                <Row
                  label="Agent wallet"
                  value={draft.walletAddress}
                  mono
                  href={getScannerAddressUrl(
                    draft.chainId,
                    draft.walletAddress
                  )}
                />
              ) : null}
              {draft.ensDomain ? (
                <Row
                  label="ENS domain"
                  value={draft.ensDomain}
                  mono
                  href={getEnsAppUrl(draft.chainId, draft.ensDomain)}
                />
              ) : null}
            </dl>
          </div>
        </div>
      </CardContent>
      {wizard}
    </Card>
  )
}

function Row({
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
    <div className="grid grid-cols-[7.5rem_1fr] items-baseline gap-4 px-4 py-3.5 sm:grid-cols-[9rem_1fr]">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="min-w-0">
        {href ? (
          <a
            href={href}
            target="_blank"
            rel="noreferrer"
            className={`flex items-center gap-1 text-sm hover:underline ${mono ? "font-mono" : ""}`}
          >
            <span className="truncate">{value}</span>
            <ExternalLink className="size-3.5 shrink-0" />
          </a>
        ) : (
          <span className={`block truncate text-sm ${mono ? "font-mono" : ""}`}>
            {value}
          </span>
        )}
      </dd>
    </div>
  )
}
