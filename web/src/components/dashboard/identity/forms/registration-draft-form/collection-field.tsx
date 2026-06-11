import type { UseFormReturn } from "react-hook-form"

import {
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
import CreateCollectionDialog from "@/components/dashboard/identity/dialogs/create-collection-dialog"
import type { AgentCollection } from "@/models/agent-collection"
import { getAppConfig } from "@/configs/app-config"
import { truncateEthAddress } from "@/lib/utils"
import type { RegistrationDraftFormSchemaType } from "./registration-draft-form-schema"

type Props = {
  form: UseFormReturn<RegistrationDraftFormSchemaType>
  collections: AgentCollection[] | null
  onReloadCollections: () => Promise<AgentCollection[]>
}

export default function CollectionField({
  form,
  collections,
  onReloadCollections,
}: Props) {
  return (
    <FormField
      control={form.control}
      name="collectionAddress"
      render={({ field }) => {
        const selected = collections?.find((c) => c.address === field.value)
        // A saved draft can reference a collection the manager no longer
        // lists (e.g. chain switch) — keep it selectable.
        const orphanValue =
          field.value && collections && !selected ? field.value : null
        return (
          <FormItem>
            <div className="flex items-center justify-between">
              <FormLabel>Collection</FormLabel>
              {getAppConfig().collectionCreatorAddress ? (
                <CreateCollectionDialog
                  onCreated={async ({ address, name }) => {
                    const loaded = await onReloadCollections()
                    const resolved =
                      address ?? loaded.find((c) => c.name === name)?.address
                    if (resolved) {
                      form.setValue("collectionAddress", resolved, {
                        shouldValidate: true,
                        shouldDirty: true,
                      })
                    }
                  }}
                />
              ) : null}
            </div>
            {collections === null || collections.length > 0 ? (
              <Select
                value={field.value}
                onValueChange={field.onChange}
                disabled={collections === null}
              >
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue
                      placeholder={
                        collections === null
                          ? "Loading collections…"
                          : "Select a collection"
                      }
                    />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {collections?.map((collection) => (
                    <SelectItem
                      key={collection.address}
                      value={collection.address}
                    >
                      {collection.name} ·{" "}
                      {truncateEthAddress(collection.address)}
                      {collection.moderatorCount > 0 ? " · moderated" : ""}
                    </SelectItem>
                  ))}
                  {orphanValue ? (
                    <SelectItem value={orphanValue}>
                      {truncateEthAddress(orphanValue)} (not in catalog)
                    </SelectItem>
                  ) : null}
                </SelectContent>
              </Select>
            ) : (
              <FormControl>
                <Input placeholder="0x…" {...field} />
              </FormControl>
            )}
            <FormDescription>
              {selected
                ? selected.description ||
                  "The AgentCollection this agent is minted into."
                : "The AgentCollection this agent is minted into. Moderated collections require a moderator's approval to mint."}
            </FormDescription>
            <FormMessage />
          </FormItem>
        )
      }}
    />
  )
}
