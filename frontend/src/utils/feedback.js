// 统一提示工具（Arco Design 实现，函数签名与旧 Element 版保持一致，
// useListRequest / 各页面 / 路由守卫均依赖这些导出）
import { Message, Modal } from '@arco-design/web-vue'

export function showToast(message) {
  Message.info(message)
}

export function showFailToast(message) {
  Message.error(message)
}

export function showSuccessToast(message) {
  Message.success(message)
}

export function showLoadingToast(message = '加载中...') {
  const text = (typeof message === 'object' && message !== null && message.message) ? message.message : message
  return Message.loading({ content: text, duration: 0 })
}

export function closeToast() {
  Message.clear()
}

// 确认弹窗：返回 Promise（确认 resolve / 取消 reject），兼容 ElMessageBox.confirm 调用方
export function showConfirmDialog({ title = '提示', message = '', confirmButtonText = '确定', cancelButtonText = '取消' } = {}) {
  return new Promise((resolve, reject) => {
    Modal.confirm({
      title,
      content: message,
      okText: confirmButtonText,
      cancelText: cancelButtonText,
      onOk: () => resolve('confirm'),
      onCancel: () => reject('cancel')
    })
  })
}
