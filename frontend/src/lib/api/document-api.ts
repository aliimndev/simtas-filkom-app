import apiClient from './client'
import type { ThesisDocument } from '@/types/document'
import type { PaginatedResponse } from '@/types/api'

export const documentApi = {
  async list(thesisId: string, params?: { page?: number; per_page?: number; document_type?: string; status?: string }): Promise<PaginatedResponse<ThesisDocument>> {
    const res = await apiClient.get(`/theses/${thesisId}/documents`, { params })
    return res.data
  },

  async get(thesisId: string, id: string): Promise<ThesisDocument> {
    const res = await apiClient.get(`/theses/${thesisId}/documents/${id}`)
    return res.data.data
  },

  async history(thesisId: string, params?: { page?: number; per_page?: number; document_type?: string }): Promise<PaginatedResponse<ThesisDocument>> {
    const res = await apiClient.get(`/theses/${thesisId}/documents/history`, { params })
    return res.data
  },

  /**
   * Upload a PDF document with an optional progress callback (Job 21).
   * The callback receives a 0–100 percentage based on axios onUploadProgress.
   */
  async upload(
    thesisId: string,
    file: File,
    documentType: string,
    chapterNumber?: number,
    onProgress?: (percent: number) => void,
    signal?: AbortSignal,
  ): Promise<ThesisDocument> {
    const form = new FormData()
    form.append('file', file)
    form.append('document_type', documentType)
    if (chapterNumber != null) {
      form.append('chapter_number', String(chapterNumber))
    }
    const res = await apiClient.post(`/theses/${thesisId}/documents`, form, {
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

  async review(id: string, decision: 'approved' | 'revision_required', notes?: string): Promise<ThesisDocument> {
    const res = await apiClient.patch(`/documents/${id}/review`, { decision, notes })
    return res.data.data
  },

  /**
   * Fetch a short-lived presigned URL (15 min) from the backend, then open it.
   * Opening the API endpoint directly would fail: a fresh browser tab cannot
   * attach the Authorization header. (Job 21)
   */
  async getDownloadUrl(thesisId: string, id: string): Promise<string> {
    const res = await apiClient.get(`/theses/${thesisId}/documents/${id}/download`)
    return res.data.data.download_url
  },

  async download(thesisId: string, id: string): Promise<void> {
    const url = await this.getDownloadUrl(thesisId, id)
    if (typeof window !== 'undefined') {
      window.open(url, '_blank', 'noopener,noreferrer')
    }
  },
}
