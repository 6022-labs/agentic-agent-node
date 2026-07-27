import { useForm } from "react-hook-form"
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
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  transferIdentityFormSchema,
  type TransferIdentityFormSchemaType,
} from "./transfer-identity-form-schema"

type Props = {
  onFormComplete: (
    values: TransferIdentityFormSchemaType
  ) => Promise<FormCompletionResult<keyof TransferIdentityFormSchemaType>>
}

export default function TransferIdentityForm({ onFormComplete }: Props) {
  const form = useForm<TransferIdentityFormSchemaType>({
    resolver: zodResolver(transferIdentityFormSchema),
    defaultValues: { to: "" },
  })

  const onSubmit = async (values: TransferIdentityFormSchemaType) => {
    const { error } = await onFormComplete(values)

    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
    }
  }

  return (
    <Form {...form}>
      <form
        id="transfer-identity-form"
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-4"
      >
        <FormField
          control={form.control}
          name="to"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Recipient</FormLabel>
              <FormControl>
                <Input placeholder="0x…" {...field} />
              </FormControl>
              <FormDescription>
                The wallet that becomes the agent's owner.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}
