import React from "react"
import axios from "axios"
import { toast } from "sonner"
import { Loader2, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import ConfirmDialog from "@/components/dashboard/confirm-dialog"
import LlmConfigForm from "@/components/dashboard/llm/forms/llm-config-form"
import type { LlmConfigFormSchemaType } from "@/components/dashboard/llm/forms/llm-config-form/llm-config-form-schema"
import type { LlmConfig } from "@/models/llm-config"
import type { FormCompletionResult } from "@/models/api/validation-error"
import {
  deleteLlmConfig,
  getLlmConfig,
  setLlmConfig,
} from "@/queries/api/llm.queries"
import { extractErrorMessage, extractValidationError } from "@/lib/api-errors"

export default function LlmConfigSection() {
  const [config, setConfig] = React.useState<LlmConfig | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)

  // Bumped on delete to remount the form with empty fields.
  const [formEpoch, setFormEpoch] = React.useState(0)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await getLlmConfig()
      setConfig(data)
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        setConfig(null)
      } else {
        toast.error(extractErrorMessage(error, "Failed to load LLM config"))
      }
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void load()
  }, [load])

  const onFormComplete = async (
    values: LlmConfigFormSchemaType
  ): Promise<FormCompletionResult<keyof LlmConfigFormSchemaType, boolean>> => {
    setSaving(true)
    try {
      const { data } = await setLlmConfig({
        provider: values.provider,
        model: values.model,
        apiKey: values.apiKey ?? "",
        baseUrl: values.baseUrl || undefined,
      })
      setConfig(data)
      toast.success("LLM configuration saved")
      return { result: true }
    } catch (error) {
      const validation =
        extractValidationError<keyof LlmConfigFormSchemaType>(error)
      if (validation) {
        return { error: validation }
      }
      toast.error(extractErrorMessage(error, "Failed to save LLM config"))
      return {}
    } finally {
      setSaving(false)
    }
  }

  const onDelete = async () => {
    try {
      await deleteLlmConfig()
      setConfig(null)
      setFormEpoch((epoch) => epoch + 1)
      toast.success("LLM configuration removed")
    } catch (error) {
      toast.error(extractErrorMessage(error, "Failed to remove LLM config"))
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-2">
          <div>
            <CardTitle>LLM connection</CardTitle>
            <CardDescription>
              The provider that powers this agent's reasoning. One connection
              per agent.
            </CardDescription>
          </div>
          {config ? (
            <Badge
              variant={
                config.setupStatus === "active" ? "default" : "secondary"
              }
            >
              {config.setupStatus}
            </Badge>
          ) : null}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading ? (
          <div className="space-y-4">
            {[0, 1, 2].map((i) => (
              <div key={i} className="space-y-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-9 w-full" />
              </div>
            ))}
          </div>
        ) : (
          <LlmConfigForm
            key={formEpoch}
            config={config}
            onFormComplete={onFormComplete}
          />
        )}
      </CardContent>
      <CardFooter className="justify-between gap-2">
        {config ? (
          <ConfirmDialog
            title="Remove LLM configuration?"
            description="The agent will not be able to reason until a new LLM is configured. The stored API key is deleted from the vault."
            confirmLabel="Remove"
            onConfirm={onDelete}
          >
            <Button type="button" variant="destructive" size="sm">
              <Trash2 />
              Remove
            </Button>
          </ConfirmDialog>
        ) : (
          <span />
        )}
        <Button
          type="submit"
          form="llm-config-form"
          disabled={saving || loading}
        >
          {saving ? <Loader2 className="animate-spin" /> : null}
          Save
        </Button>
      </CardFooter>
    </Card>
  )
}
