// 32 random bytes as a 0x-hex string — raw material for EIP-3009 and Permit2 nonces.
export function randomBytes32Hex(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(32))
  return (
    "0x" +
    Array.from(bytes)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("")
  )
}
