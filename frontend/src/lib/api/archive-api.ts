import apiClient from './client'
import type { Archive, ArchiveStats } from '@/types/archive'
import type { PaginatedResponse } from '@/types/api'

export const archiveApi = {
  async search(params?: { q?: string; year?: number; field_of_study?: string; study_program?: string; page?: number; per_page?: number }): Promise<PaginatedResponse<Archive>> {
    const res = await apiClient.get('/archives', { params })
    return res.data
  },

  async get(id: string): Promise<Archive> {
    const res = await apiClient.get(`/archives/${id}`)
    return res.data.data
  },

  async stats(): Promise<ArchiveStats> {
    const res = await apiClient.get('/archives/stats')
    return res.data.data
  },

  // Used by the future graduation/archive flow (Job 17 student page); the
  // current v1.0 UI only lists/searches/downloads archives.
  async create(
    thesisId: string,
    file: File,
    data: { abstract_id: string; abstract_en?: string; keywords: string[]; graduation_year: number },
    onProgress?: (percent: number) => void,
    signal?: AbortSignal,
  ): Promise<Archive> {
    const form = new FormData()
    form.append('file', file)
    form.append('abstract_id', data.abstract_id)
    if (data.abstract_en) form.append('abstract_en', data.abstract_en)
    form.append('keywords', data.keywords.join(','))
    form.append('graduation_year', String(data.graduation_year))
    const res = await apiClient.post(`/theses/${thesisId}/archive`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      signal,
      onUploadProgress: (e) => {
        if (onProgress && e.total) {
          onProgress(Math.round((e.loaded * 100) / e.total))
        }
      },
    })
    return res.data.data
  },

  async getByThesis(thesisId: string): Promise<Archive | null> {
    const res = await apiClient.get(`/theses/${thesisId}/archive`)
    return res.data.data
  },

  /** Fetch a short-lived presigned URL (30 min), then open it in a new tab. */
  async getDownloadUrl(id: string): Promise<string> {
    const res = await apiClient.get(`/archives/${id}/download`)
    return res.data.data.download_url
  },

  async download(id: string): Promise<void> {
    const url = await this.getDownloadUrl(id)
    if (typeof window !== 'undefined') {
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  },
}
