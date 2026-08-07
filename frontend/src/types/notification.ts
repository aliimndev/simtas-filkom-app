export interface AppNotification {
  id: string
  title: string
  message: string
  type: string
  link?: string | null
  is_read: boolean
  read_at?: string | null
  created_at: string
}

export interface UnreadCount {
  unread_count: number
}
