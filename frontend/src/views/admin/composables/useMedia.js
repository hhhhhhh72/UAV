import axios from '@/utils/http'
import { showFailToast } from 'vant'

/**
 * Normalize media URLs to avoid mixed-content / relative-path issues.
 */
export function normalizeMediaUrl(raw) {
  if (!raw || typeof raw !== 'string') return raw
  const url = raw.trim()
  if (!url) return url
  if (url.startsWith('data:') || url.startsWith('blob:')) return url
  if (url.startsWith('uploads/')) return `/${url}`
  if (url.startsWith('/')) return url
  if (url.startsWith('http://') || url.startsWith('https://')) {
    try {
      const u = new URL(url)
      const host = u.hostname
      const port = u.port
      const isLocalish =
        host === 'localhost' ||
        host === '127.0.0.1' ||
        host === '0.0.0.0' ||
        host === '172.17.0.1' ||
        port === '8090'
      if (isLocalish) {
        return `${u.pathname}${u.search}${u.hash}`
      }
    } catch (e) {
      // ignore
    }
    return url
  }
  return url
}

/**
 * Upload a single file via /api/upload and return the normalised URL.
 * Supports both { file: File } object and direct File/Blob object.
 */
export async function uploadFile(file) {
  const formData = new FormData()
  // Handle both { file: File } wrapper and direct File/Blob
  const actualFile = file.file || file
  if (!actualFile) {
    showFailToast('没有可上传的文件')
    return null
  }
  formData.append('file', actualFile)
  try {
    const res = await axios.post('/api/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    if (res.data.success) {
      return normalizeMediaUrl(res.data.url)
    } else {
      showFailToast('上传失败')
      return null
    }
  } catch (err) {
    console.error(err)
    showFailToast('上传出错')
    return null
  }
}
