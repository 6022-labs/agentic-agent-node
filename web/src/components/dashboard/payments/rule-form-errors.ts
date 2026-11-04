import axios from "axios"

import { isPlainAddressPattern } from "@/lib/token-format"

// isAssetTaken reports the runtime's refusal of a second rule on one token.
export const isAssetTaken = (error: unknown): boolean =>
  axios.isAxiosError(error) &&
  error.response?.status === 409 &&
  (error.response.data as { code?: string } | undefined)?.code ===
    "asset_already_has_rule"

// toFormField maps a runtime error path ("config.asset", "windows[0].maxAmount") to a rendered control, else null.
export function toFormField<Field extends string>(
  path: string,
  fields: readonly string[]
): Field | null {
  const field = path.replace(/^config\./, "").replace(/\[(\d+)\]/g, ".$1")
  return fields.includes(field.split(".")[0]) ? (field as Field) : null
}

// normalizePattern lowercases a plain address, as the runtime stores and matches it.
export const normalizePattern = (pattern: string): string =>
  isPlainAddressPattern(pattern) ? pattern.toLowerCase() : pattern
