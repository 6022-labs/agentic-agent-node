// The runtime stores allowed tools as regexes, full-matched against the tool
// name; no pattern allows every tool. These helpers mirror that on the console
// side so a selection round-trips.

export function matchesToolPatterns(
  toolName: string,
  patterns: string[]
): boolean {
  if (patterns.length === 0) return true

  return patterns.some((pattern) => {
    try {
      return new RegExp(`^(?:${pattern})$`).test(toolName)
    } catch {
      return false
    }
  })
}

// toolNamePattern escapes a tool name so it is matched literally.
export function toolNamePattern(toolName: string): string {
  return toolName.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")
}
