import React from "react"
import { useForm, useWatch, type DefaultValues } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"

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
import type { AgentCollection } from "@/models/agent-collection"
import { AGENT_ROLES } from "@/models/enums/agent-role"
import { getMintableChainIds } from "@/configs/app-config"
import { findEvmNetwork } from "@/models/networks"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { normalizeEnsLabel } from "@/lib/ens-label"
import {
  registrationDraftFormSchema,
  type RegistrationDraftFormSchemaType,
} from "./registration-draft-form-schema"
import CollectionField from "./collection-field"
import AttributeFields from "./attribute-fields"

const chainLabel = (chainId: string) =>
  findEvmNetwork(`eip155:${chainId}`)?.name ?? `Chain ${chainId}`

const buildDefaults = (
  draft: AgentRegistration | null,
  fallbackChainId: string
): DefaultValues<RegistrationDraftFormSchemaType> => ({
  chainId: draft?.chainId ?? fallbackChainId,
  collectionAddress: draft?.collectionAddress ?? "",
  name: draft?.name ?? "",
  // images.default holds the generated IPFS CID, not the base; leaving this
  // empty on re-edit keeps the stored image.
  image: undefined,
  // Row 0 is the pinned mandatory "role" attribute.
  attributes: [
    {
      key: "role",
      value:
        AGENT_ROLES.find((role) => role === draft?.attributes?.["role"]) ?? "",
    },
    ...Object.entries(draft?.attributes ?? {})
      .filter(([key]) => key !== "role")
      .map(([key, value]) => ({ key, value })),
  ],
})

// The Save button lives in the wizard footer and submits by this id.
export const REGISTRATION_DRAFT_FORM_ID = "registration-draft-form"

type Props = {
  draft: AgentRegistration | null
  collections: AgentCollection[] | null
  onReloadCollections: () => Promise<AgentCollection[]>
  onValueChange: (values: RegistrationDraftFormSchemaType) => void
  onFormComplete: (
    values: RegistrationDraftFormSchemaType
  ) => Promise<FormCompletionResult<keyof RegistrationDraftFormSchemaType>>
}

export default function RegistrationDraftForm({
  draft,
  collections,
  onReloadCollections,
  onValueChange,
  onFormComplete,
}: Props) {
  const mintableChainIds = React.useMemo(() => getMintableChainIds(), [])

  const form = useForm<RegistrationDraftFormSchemaType>({
    resolver: zodResolver(registrationDraftFormSchema),
    defaultValues: buildDefaults(draft, mintableChainIds[0] ?? ""),
  })

  // Re-sync when the loaded draft changes (re-edit), not on every render —
  // defaultValues only seed the first mount.
  React.useEffect(() => {
    form.reset(buildDefaults(draft, mintableChainIds[0] ?? ""))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [draft])

  // Collections are per-chain, so the wizard reloads them from the emitted values.
  const selectedChainId = useWatch({ control: form.control, name: "chainId" })

  const watched = useWatch({ control: form.control })
  React.useEffect(() => {
    onValueChange(form.getValues())
  }, [watched, form, onValueChange])

  const onSubmit = async (values: RegistrationDraftFormSchemaType) => {
    const { error } = await onFormComplete(values)

    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  return (
    <Form {...form}>
      <form
        id={REGISTRATION_DRAFT_FORM_ID}
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-4"
      >
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
          onReloadCollections={onReloadCollections}
        />
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input
                  placeholder="my-agent"
                  {...field}
                  onChange={(e) =>
                    field.onChange(normalizeEnsLabel(e.target.value))
                  }
                />
              </FormControl>
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
                {draft ? " Leave empty to keep the current image." : ""}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <AttributeFields form={form} />
      </form>
    </Form>
  )
}
