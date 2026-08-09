'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Bell } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'
import { notificationApi } from '@/lib/api/notification-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { formatDateTime } from '@/lib/utils/date'
import { cn } from '@/lib/utils/cn'
import type { AppNotification } from '@/types/notification'

const listKey = ['notifications'] as const

export function NotificationBell() {
  const accessToken = useAuthStore((s) => s.accessToken)
  const [open, setOpen] = useState(false)
  const queryClient = useQueryClient()

  const unread = useQuery({
    queryKey: ['notifications', 'unread'],
    queryFn: notificationApi.unreadCount,
    enabled: Boolean(accessToken),
    refetchInterval: 30_000,
  })

  const list = useQuery({
    queryKey: listKey,
    queryFn: () => notificationApi.list(20),
    enabled: Boolean(accessToken) && open,
  })

  const markAll = useMutation({
    mutationFn: notificationApi.markAllRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread'] })
    },
  })

  const markOne = useMutation({
    mutationFn: notificationApi.markRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['notifications', 'unread'] })
    },
  })

  const count = unread.data ?? 0

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="relative rounded-lg p-2 text-foreground transition-colors hover:bg-muted"
        aria-label={`Notifikasi${count > 0 ? `, ${count} belum dibaca` : ''}`}
      >
        <Bell className="h-5 w-5" />
        {count > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-danger px-1 text-[10px] font-bold leading-none text-white">
            {count > 99 ? '99+' : count}
          </span>
        )}
      </button>

      {open && (
        <>
          <div aria-hidden className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute right-0 top-full z-50 mt-1.5 w-80 overflow-hidden rounded-lg border border-border bg-card shadow-lg">
            <div className="flex items-center justify-between border-b border-border px-4 py-2.5">
              <p className="text-sm font-semibold">Notifikasi</p>
              {count > 0 && (
                <button
                  type="button"
                  onClick={() => markAll.mutate()}
                  className="text-xs text-primary hover:underline"
                >
                  Tandai semua dibaca
                </button>
              )}
            </div>

            <div className="max-h-80 overflow-y-auto">
              {list.isLoading && <p className="p-4 text-sm text-muted-foreground">Memuat…</p>}
              {list.isError && <p className="p-4 text-sm text-danger-700">Gagal memuat notifikasi.</p>}
              {list.data?.length === 0 && (
                <p className="p-4 text-sm text-muted-foreground">Tidak ada notifikasi.</p>
              )}
              {list.data?.map((n) => {
                const onClick = () => {
                  if (!n.is_read) markOne.mutate(n.id)
                  setOpen(false)
                }
                const body = <NotificationBody n={n} />
                return (
                  <NotificationItem key={n.id} isRead={n.is_read} onClick={onClick}>
                    {n.link ? <Link href={n.link}>{body}</Link> : body}
                  </NotificationItem>
                )
              })}
            </div>
          </div>
        </>
      )}
    </div>
  )
}

function NotificationItem({
  isRead,
  onClick,
  children,
}: {
  isRead: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <div
      className={cn(
        'cursor-pointer border-b border-border px-4 py-3',
        !isRead && 'bg-primary/5',
      )}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          onClick()
        }
      }}
    >
      {children}
    </div>
  )
}

function NotificationBody({ n }: { n: AppNotification }) {
  return (
    <>
      <div className="flex items-start justify-between gap-2">
        <p className="text-sm font-medium">{n.title}</p>
        <span className="shrink-0 text-[10px] text-muted-foreground">{formatDateTime(n.created_at)}</span>
      </div>
      <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">{n.message}</p>
    </>
  )
}
