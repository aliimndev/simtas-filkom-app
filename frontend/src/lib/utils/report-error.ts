// Minimal Sentry error reporter — no SDK dependency.
// Sends error events to the Sentry Envelope API when NEXT_PUBLIC_SENTRY_DSN
// is configured; no-op otherwise. Use in ErrorBoundary and catch blocks.
//
// ponytail: manual envelope POST instead of @sentry/nextjs; covers the 90%
// (reporting errors) without the SDK's build-time source-map integration.
// Add @sentry/nextjs when you need stack-trace de-minification in prod.

const DSN = process.env.NEXT_PUBLIC_SENTRY_DSN

let _dsn: { origin: string; key: string; project: string } | null | undefined

function parseDsn(dsn: string) {
  try {
    const u = new URL(dsn)
    const parts = u.pathname.split('/').filter(Boolean)
    return {
      origin: u.origin,
      key: u.username,
      project: parts[parts.length - 1],
    }
  } catch {
    return null
  }
}

function uuid() {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16)
  })
}

export function reportError(error: unknown, context?: Record<string, unknown>) {
  if (!DSN) return
  if (_dsn === undefined) _dsn = parseDsn(DSN)
  if (!_dsn) return

  const eventId = uuid()
  const event = {
    event_id: eventId,
    timestamp: new Date().toISOString(),
    platform: 'javascript',
    level: 'error',
    release: process.env.NEXT_PUBLIC_APP_VERSION,
    contexts: context ? { extra: context } : undefined,
    exception: {
      values: [
        {
          type: error instanceof Error ? error.name : 'Error',
          value: error instanceof Error ? error.message : String(error),
          stacktrace: error instanceof Error && error.stack
            ? { frames: error.stack.split('\n').slice(1, 8).map((l) => ({ filename: l.trim() })) }
            : undefined,
        },
      ],
    },
  }

  const header = JSON.stringify({ event_id: eventId, sent_at: new Date().toISOString() })
  const itemHeader = JSON.stringify({ type: 'event', length: new TextEncoder().encode(JSON.stringify(event)).length })
  const body = `${header}\n${itemHeader}\n${JSON.stringify(event)}`

  fetch(`${_dsn.origin}/api/${_dsn.project}/envelope/`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-sentry-envelope',
      'X-Sentry-Auth': `Sentry sentry_version=7, sentry_key=${_dsn.key}`,
    },
    body,
  }).catch(() => {})
}
