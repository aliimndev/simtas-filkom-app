import { AlertCircle } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'

/** Daftar tindakan yang perlu dilakukan mahasiswa (warning card). */
export function PendingActionsCard({ actions }: { actions: string[] }) {
  if (actions.length === 0) return null

  return (
    <Card className="border-warning/30 bg-warning-50/70">
      <CardContent className="p-4">
        <p className="flex items-center gap-2 text-sm font-semibold text-warning-700">
          <AlertCircle className="h-4 w-4" /> Yang perlu Anda lakukan
        </p>
        <ul className="mt-2 space-y-1.5 text-sm text-warning-900">
          {actions.map((a, i) => (
            <li key={i} className="flex items-baseline gap-2">
              <span className="h-1 w-1 shrink-0 -translate-y-0.5 rounded-full bg-warning" />
              {a}
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  )
}
