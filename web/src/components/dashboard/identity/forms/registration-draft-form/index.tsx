import React from "react"
import { toast } from "sonner"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowRight, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { AgentRegistration } from "@/models/agent-registration"
import { AGENT_ROLES } from "@/models/enums/agent-role"
import { saveAgentRegistration } from "@/queries/api/registration.queries"
import { getMintableChainIds } from "@/configs/app-config"
import { findEvmNetwork } from "@/models/networks"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"
import {
  registrationDraftFormSchema,
  type RegistrationDraftFormSchemaType,
} from "./registration-draft-form-schema"
import { useAgentCollections } from "./use-agent-collections"
import CollectionField from "./collection-field"
import AttributeFields from "./attribute-fields"

const chainLabel = (chainId: string) =>
  findEvmNetwork(`eip155:${chainId}`)?.name ?? `Chain ${chainId}`

const isDraftComplete = (draft: AgentRegistration | null) =>
  !!draft &&
  !!draft.chainId &&
  !!draft.collectionAddress &&
  !!draft.name &&
  !!draft.role &&
  !!draft.images &&
  Object.keys(draft.images).length > 0

type Props = {
  draft: AgentRegistration | null
  onSaved: (registration: AgentRegistration) => void
  // Advance to the mint step; enabled once a draft has been saved.
  onContinue?: () => void
}

export default function RegistrationDraftForm({
  draft,
  onSaved,
  onContinue,
}: Props) {
  const [saving, setSaving] = React.useState(false)
  const mintableChainIds = React.useMemo(() => getMintableChainIds(), [])

  const form = useForm<RegistrationDraftFormSchemaType>({
    resolver: zodResolver(registrationDraftFormSchema),
    defaultValues: {
      chainId: draft?.chainId ?? mintableChainIds[0] ?? "",
      collectionAddress: draft?.collectionAddress ?? "",
      name: draft?.name ?? "",
      role: AGENT_ROLES.find((role) => role === draft?.role),
      // images.default holds the generated IPFS CID, not the base; always re-pick.
      image: undefined,
      attributes: Object.entries(draft?.attributes ?? {}).map(
        ([key, value]) => ({ key, value })
      ),
    },
  })

  // Collections are per-chain: reload them (and clear the stale selection) when
  // the chosen chain changes.
  const selectedChainId = form.watch("chainId")
  const { collections, loadCollections } = useAgentCollections(selectedChainId)

  const onSubmit = async (values: RegistrationDraftFormSchemaType) => {
    setSaving(true)
    try {
      const { data } = await saveAgentRegistration({
        chainId: values.chainId,
        collectionAddress: values.collectionAddress,
        name: values.name,
        role: values.role,
        image: values.image,
        attributes: Object.fromEntries(
          values.attributes.map((attribute) => [
            attribute.key.trim(),
            attribute.value,
          ])
        ),
      })
      toast.success("Draft saved")
      onSaved(data)
    } catch (error) {
      const validation =
        extractValidationError<keyof RegistrationDraftFormSchemaType>(error)
      if (validation && validation.field in form.getValues()) {
        form.setError(
          validation.field as keyof RegistrationDraftFormSchemaType,
          {
            type: "server",
            message: validation.message,
          }
        )
        return
      }
      toast.error(extractErrorMessage(error, "Failed to save draft"))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="chainId"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Chain</FormLabel>
              <Select
                value={field.value}
                onValueChange={(value) => {
                  field.onChange(value)
                  // The previous chain's collection no longer applies.
                  form.setValue("collectionAddress", "")
                }}
              >
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a chain" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {mintableChainIds.map((id) => (
                    <SelectItem key={id} value={id}>
                      {chainLabel(id)} ({id})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                The chain this agent's identity is minted on. Only chains the
                node can reach and mint on are listed.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <CollectionField
          form={form}
          chainId={selectedChainId}
          collections={collections}
          onReloadCollections={loadCollections}
        />
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="my-agent" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="role"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Role</FormLabel>
              <Select value={field.value} onValueChange={field.onChange}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a role" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {AGENT_ROLES.map((role) => (
                    <SelectItem key={role} value={role} className="capitalize">
                      {role}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="image"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Image</FormLabel>
              <FormControl>
                <Input
                  type="file"
                  accept="image/*"
                  onChange={(event) => field.onChange(event.target.files?.[0])}
                />
              </FormControl>
              <FormDescription>
                Framed into the NFT artwork and pinned to IPFS on save.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <AttributeFields form={form} />
        <div className="flex justify-end gap-2">
          <Button
            type="submit"
            variant={onContinue && draft ? "outline" : "default"}
            disabled={saving}
          >
            {saving ? <Loader2 className="animate-spin" /> : null}
            {draft ? "Update draft" : "Save draft"}
          </Button>
          {onContinue ? (
            <Button
              type="button"
              onClick={onContinue}
              disabled={!isDraftComplete(draft) || saving}
            >
              Continue
              <ArrowRight />
            </Button>
          ) : null}
        </div>
      </form>
    </Form>
  )
}
