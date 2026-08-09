'use client'

import { useState } from 'react'
import { CheckCircle2, XCircle } from 'lucide-react'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Dialog } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type { TitleChangeRequest } from '@/types/title-change'

export type ReviewDecision = 'approve' | 'reject'

/** Dialog konfirmasi review perubahan judul — validasi catatan hidup di sini. */
export function ReviewDialog({
  target,
  onClose,
  isPending,
  pendingDecision,
  isError,
  errorMessage,
  onReview,
}: {
  target: TitleChangeRequest | null
  onClose: () => void
  isPending: boolean
  pendingDecision: ReviewDecision | null
  isError: boolean
  errorMessage: string | null
  onReview: (decision: ReviewDecision, notes: string) => void
}) {
  const [notes, setNotes] = useState('')
  const [notesTouched, setNotesTouched] = useState(false)

  // Konten dialog hanya dirender saat `target` ada (lihat `{target && …}`),
  // sehingga state catatan otomatis bersih setiap dialog dibuka.
  const notesInvalid = notesTouched && !notes.trim()

  function run(decision: ReviewDecision) {
    if (!target) return
    if (decision === 'reject' && !notes.trim()) {
      setNotesTouched(true)
      return
    }
    onReview(decision, notes.trim())
  }

  return (
    <Dialog
      open={Boolean(target)}
      onClose={onClose}
      title="Konfirmasi Perubahan Judul"
      description="Tinjau pengajuan mahasiswa sebelum mengambil keputusan."
    >
      {target && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
            <div className="rounded-lg bg-muted/50 p-3">
              <p className="text-xs font-medium text-muted-foreground">Judul Sebelumnya</p>
              <p className="mt-0.5 leading-snug">{target.previous_title}</p>
            </div>
            <div className="rounded-lg bg-primary-50 p-3">
              <p className="text-xs font-medium text-primary-700">Judul Baru</p>
              <p className="mt-0.5 leading-snug">{target.requested_title}</p>
            </div>
          </div>
          {target.reason && (
            <div className="rounded-lg border border-border p-3 text-sm">
              <p className="text-xs font-medium text-muted-foreground">Alasan Mahasiswa</p>
              <p className="mt-0.5">{target.reason}</p>
            </div>
          )}
          <div>
            <Label htmlFor="review-notes">Catatan Pembimbing</Label>
            <Textarea
              id="review-notes"
              rows={3}
              placeholder="Catatan untuk mahasiswa (wajib jika menolak)"
              value={notes}
              invalid={notesInvalid}
              onChange={(e) => {
                setNotes(e.target.value)
                if (notesTouched) setNotesTouched(false)
              }}
            />
            {notesInvalid && <p className="mt-1 text-xs text-danger">Catatan wajib diisi saat menolak perubahan judul.</p>}
          </div>
          {isError && <Alert variant="danger">{errorMessage}</Alert>}
          <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-4">
            <Button type="button" variant="ghost" onClick={onClose}>
              Kembali
            </Button>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="danger"
                loading={isPending && pendingDecision === 'reject'}
                disabled={isPending}
                onClick={() => run('reject')}
              >
                <XCircle className="h-4 w-4" /> Tolak
              </Button>
              <Button
                type="button"
                variant="success"
                loading={isPending && pendingDecision === 'approve'}
                disabled={isPending}
                onClick={() => run('approve')}
              >
                <CheckCircle2 className="h-4 w-4" /> Setujui
              </Button>
            </div>
          </div>
        </div>
      )}
    </Dialog>
  )
}
