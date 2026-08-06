import apiClient from './client'
import type { AppNotification, UnreadCount } from '@/types/notification'

export const notificationApi = {
  async list(limit = 20): Promise<AppNotification[]> {
    const res = await apiClient.get('/notifications', { params: { limit } })
    return res.data.data
  },

  async unreadCount(): Promise<number> {
    const res = await apiClient.get<{ data: UnreadCount }>('/notifications/unread-count')
    return res.data.data.unread_count
  },

  async markRead(id: string): Promise<void> {
    await apiClient.patch(`/notifications/${id}/read`)
  },

  async markAllRead(): Promise<void> {
    await apiClient.patch('/notifications/read-all')
  },
}
