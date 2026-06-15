// Mirrors the runtime's TransactionStatus lifecycle, shared by mint and ENS.
export const TRANSACTION_STATUS = {
  pending: "pending",
  submitted: "submitted",
  confirmed: "confirmed",
  failed: "failed",
} as const
