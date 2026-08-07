'use client'

import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Dialog } from '@/components/ui/dialog'
import type { TitleChangeRequest } from '@/types/title-change'

/** Dialog konfirmasi pembatalan permintaan perubahan judul yang masih PENDING. */
export function CancelRequestDialog({
  target,
  onClose,
  onConfirm,
  isPending,
  isError,
  errorMessage,
}: {
  target: TitleChangeRequest | null
  onClose: () => void
  onConfirm: (id: string) => void
  isPending: boolean
  isError: boolean
  errorMessage: string | null
}) {
  return (
    <Dialog
      open={Boolean(target)}
      onClose={onClose}
      title="Batalkan Permintaan Perubahan Judul?"
      description="Permintaan yang sedang diproses akan ditarik dan tidak dapat dikembalikan."
    >
      {target && (
        <div className="space-y-4">
          <div className="rounded-lg bg-muted/50 p-3 text-sm">
            <p className="text-xs font-medium text-muted-foreground">Judul yang diajukan</p>
            <p className="mt-0.5">{target.requested_title}</p>
          </div>
          {isError && <Alert variant="danger">{errorMessage}</Alert>}
          <div className="flex flex-wrap justify-end gap-2">
            <Button type="button" variant="ghost" onClick={onClose}>Kembali</Button>
            <Button
              type="button"
              variant="danger"
              loading={isPending}
              onClick={() => onConfirm(target.id)}
            >
              Ya, Batalkan
            </Button>
          </div>
        </div>
      )}
    </Dialog>
  )
}
