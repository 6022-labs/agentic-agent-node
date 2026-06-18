import { useFieldArray, type UseFormReturn } from "react-hook-form"
import { Plus, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { RegistrationDraftFormSchemaType } from "./registration-draft-form-schema"

type Props = {
  form: UseFormReturn<RegistrationDraftFormSchemaType>
}

export default function AttributeFields({ form }: Props) {
  const attributeRows = useFieldArray({
    control: form.control,
    name: "attributes",
  })

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <Label>Attributes</Label>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => attributeRows.append({ key: "", value: "" })}
        >
          <Plus />
          Add
        </Button>
      </div>
      {attributeRows.fields.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Optional key/value traits stored in the NFT metadata.
        </p>
      ) : null}
      {attributeRows.fields.map((row, index) => (
        <div key={row.id} className="flex items-start gap-2">
          <FormField
            control={form.control}
            name={`attributes.${index}.key`}
            render={({ field }) => (
              <FormItem className="flex-1">
                <FormControl>
                  <Input placeholder="trait" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name={`attributes.${index}.value`}
            render={({ field }) => (
              <FormItem className="flex-1">
                <FormControl>
                  <Input placeholder="value" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            onClick={() => attributeRows.remove(index)}
            aria-label="Remove attribute"
          >
            <Trash2 />
          </Button>
        </div>
      ))}
    </div>
  )
}
