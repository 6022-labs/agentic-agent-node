import React from "react"
import { useForm, useWatch } from "react-hook-form"
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
import { LLM_PROVIDERS, type LlmConfig } from "@/models/llm-config"
import type { FormCompletionResult } from "@/models/api/validation-error"
import { extractAmazonPlatformRegionFromUrl } from "@/lib/utils"
import {
  createLlmConfigFormSchema,
  type LlmConfigFormSchemaType,
} from "./llm-config-form-schema"

// The Save button lives in the card footer and submits by this id.
export const LLM_CONFIG_FORM_ID = "llm-config-form"

type Props = {
  config: LlmConfig | null
  onFormComplete: (
    values: LlmConfigFormSchemaType
  ) => Promise<FormCompletionResult<keyof LlmConfigFormSchemaType, boolean>>
  onValueChange: (values: LlmConfigFormSchemaType) => void
}

export default function LlmConfigForm({
  config,
  onFormComplete,
  onValueChange,
}: Props) {
  // A ref, so the resolver sees the saved provider after each save without remounting the form.
  const currentProvider = React.useRef(config?.provider)
  React.useEffect(() => {
    currentProvider.current = config?.provider
  }, [config?.provider])

  const form = useForm<LlmConfigFormSchemaType>({
    resolver: (values, context, options) =>
      zodResolver(createLlmConfigFormSchema(currentProvider.current))(
        values,
        context,
        options
      ),
    defaultValues: {
      provider: config?.provider ?? "",
      model: config?.model ?? "",
      apiKey: "",
      baseUrl: config?.baseUrl ?? "",
    },
  })

  const selectedProvider = useWatch({ control: form.control, name: "provider" })
  const baseUrl = useWatch({ control: form.control, name: "baseUrl" })
  const provider = LLM_PROVIDERS.find((p) => p.value === selectedProvider)
  const isSelfHosted = provider?.kind === "self-hosted"
  const keepsKey = !!config && config.provider === selectedProvider
  const needsBaseUrl = isSelfHosted || !!provider?.requiresBaseUrl
  const isBedrock = selectedProvider === "aws.bedrock"
  const detectedRegion = isBedrock
    ? extractAmazonPlatformRegionFromUrl(baseUrl ?? "")
    : undefined

  // A self-hosted provider never takes a key: drop one typed before switching to it.
  React.useEffect(() => {
    if (isSelfHosted && form.getValues("apiKey")) form.setValue("apiKey", "")
  }, [form, isSelfHosted])

  const watched = useWatch({ control: form.control })
  React.useEffect(() => {
    onValueChange(form.getValues())
  }, [watched, form, onValueChange])

  const onSubmit = async (values: LlmConfigFormSchemaType) => {
    const { error, result: saved } = await onFormComplete(values)

    if (error) {
      form.setError(error.field, { type: "server", message: error.message })
      return
    }

    // The key was stored in the vault; don't keep it around in the field.
    if (saved) form.setValue("apiKey", "")
  }

  return (
    <Form {...form}>
      <form
        id={LLM_CONFIG_FORM_ID}
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-4"
      >
        <FormField
          control={form.control}
          name="provider"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Provider</FormLabel>
              <Select value={field.value} onValueChange={field.onChange}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a provider" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {LLM_PROVIDERS.map((provider) => (
                    <SelectItem key={provider.value} value={provider.value}>
                      {provider.label}
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
          name="model"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Model</FormLabel>
              <FormControl>
                <Input placeholder="claude-opus-4-8" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        {needsBaseUrl ? (
          <FormField
            control={form.control}
            name="baseUrl"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Server base URL</FormLabel>
                <FormControl>
                  <Input
                    placeholder={
                      isBedrock
                        ? "https://bedrock-runtime.us-east-1.amazonaws.com"
                        : "https://my-ollama-server.com:11434"
                    }
                    {...field}
                  />
                </FormControl>
                {isBedrock ? (
                  <FormDescription>
                    {detectedRegion
                      ? `AWS region detected: ${detectedRegion}`
                      : "Use the regional Bedrock runtime endpoint; the AWS region is read from it."}
                  </FormDescription>
                ) : null}
                <FormMessage />
              </FormItem>
            )}
          />
        ) : null}
        {!isSelfHosted ? (
          <FormField
            control={form.control}
            name="apiKey"
            render={({ field }) => (
              <FormItem>
                <FormLabel>API key</FormLabel>
                <FormControl>
                  <Input
                    type="password"
                    placeholder={keepsKey ? "•••••• (unchanged)" : "sk-…"}
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  Stored in the node vault, never returned by the API.
                  {keepsKey
                    ? " Leave blank to keep the current key."
                    : " A new provider needs its own key."}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        ) : null}
      </form>
    </Form>
  )
}
