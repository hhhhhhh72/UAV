// 富文本字段的纯文本兜底：列表卡片/摘要等场景直接插值 HTML 会带出 <p> 等标签，
// stripHtml 剥掉标签并还原常见实体，保留文本内容。
export function stripHtml(input) {
  if (!input || typeof input !== 'string') return ''
  return input
    .replace(/<[^>]*>/g, '')
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/s+/g, ' ')
    .trim()
}
