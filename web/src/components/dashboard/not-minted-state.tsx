import { Fingerprint } from "lucide-react"
import { Link } from "react-router-dom"

import { Button } from "@/components/ui/button"
import EmptyState from "@/components/dashboard/empty-state"

export default function NotMintedState({ description }: { description: string }) {
  return (
    <EmptyState
      icon={Fingerprint}
      title="Agent not minted yet"
      description={description}
      action={
        <Button asChild variant="outline">
          <Link to="/dashboard/identity">Finish the identity setup</Link>
        </Button>
      }
    />
  )
}
