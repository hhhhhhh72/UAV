import { ElMessage, ElMessageBox } from 'element-plus'

let loadingInstance = null

export function showToast(message) {
  ElMessage({ message, type: 'info' })
}

export function showFailToast(message) {
  ElMessage({ message, type: 'error' })
}

export function showSuccessToast(message) {
  ElMessage({ message, type: 'success' })
}

export function showLoadingToast(message = '加载中...') {
  loadingInstance = ElMessage({ message, type: 'info', duration: 0 })
}

export function closeToast() {
  if (loadingInstance) {
    loadingInstance.close()
    loadingInstance = null
  }
}

export function showConfirmDialog({ title = '提示', message = '', confirmButtonText = '确定', cancelButtonText = '取消' } = {}) {
  return ElMessageBox.confirm(message, title, {
    confirmButtonText,
    cancelButtonText,
    type: 'warning'
  })
}
