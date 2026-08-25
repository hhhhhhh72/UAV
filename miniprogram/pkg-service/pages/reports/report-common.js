// 行业报告共享模块（方向 B · 蓝皮书刊物）
// 类型映射 / 日期格式化 / 正文预处理 / 详情缓存 / 内联 SVG 图标。
// 后端枚举（category）：whitepaper / research / analysis / other，与 admin 表单一致。
// 旧版小程序用中文值 + report_type/publish_date 字段，与后端不匹配（筛选失效、标签全显"其他"），
// 本模块统一收敛到后端真实契约。

export const REPORT_TYPES = [
  { value: '', label: '全部', en: '', cover: '' },
  { value: 'whitepaper', label: '白皮书', en: 'WHITE PAPER', cover: 'navy' },
  { value: 'research', label: '调研报告', en: 'RESEARCH REPORT', cover: 'teal' },
  { value: 'analysis', label: '行业分析', en: 'INDUSTRY ANALYSIS', cover: 'orange' },
  { value: 'other', label: '其他', en: 'OTHER', cover: 'slate' },
]

// 按后端枚举取类型定义；未知值兜底为「其他」样式，不抛错。
export function typeOf(value) {
  return REPORT_TYPES.find(function (t) { return t.value === value }) ||
    { value: value || '', label: value || '其他', en: '', cover: 'slate' }
}

// 后端时间格式（created_at: RFC3339）→ YYYY-MM-DD；空值返回 '-'。
export function formatDate(d) {
  if (!d) return '-'
  var dt = new Date(d)
  if (isNaN(dt.getTime())) return String(d)
  var p = function (n) { return (n < 10 ? '0' : '') + n }
  return dt.getFullYear() + '-' + p(dt.getMonth() + 1) + '-' + p(dt.getDate())
}

// 报告文件名（无 file_url 时不用）
export function fileName(report) {
  var url = report.file_url || ''
  if (!url) return ''
  var name = url.split('/').pop() || ''
  return name.split('?')[0] || '报告全文'
}

// 正文预处理：Content 可能是富文本 HTML（旧数据）或纯文本（新写入经后端消毒已去标签）。
// 返回 { kind: 'html'|'text'|'empty', html?, paras? }
// - html：防御性剔除脚本/事件属性后，给块级标签注入刊物排版内联样式（rich-text 无法被页面 CSS 穿透）。
// - text：按空行/换行切段，由页面用 text 组件排版（可做首字下沉）。
export function prepareContent(content) {
  var src = (content || '').trim()
  if (!src) return { kind: 'empty' }
  var looksHtml = /<[a-z][\s\S]*>/i.test(src)
  if (looksHtml) {
    return { kind: 'html', html: styleContentHtml(sanitizeContentHtml(src)) }
  }
  var paras = src
    .split(/\n{2,}|\n/)
    .map(function (s) { return s.trim() })
    .filter(Boolean)
  return paras.length ? { kind: 'text', paras: paras } : { kind: 'empty' }
}

function sanitizeContentHtml(html) {
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, '')
    .replace(/<style[\s\S]*?<\/style>/gi, '')
    .replace(/<iframe[\s\S]*?<\/iframe>/gi, '')
    .replace(/\son\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '')
    .replace(/<([a-z0-9]+)[^>]*\sclass\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '<$1') // 去 class，防旧样式干扰
}

// 数值镜像页面间距令牌（24=--space-md / 48=--space-xl / 16=--space-sm）；
// rich-text 内联样式不能用 var()，改令牌时需手工同步此处的字面量。
function styleContentHtml(html) {
  var block = 'margin:0 0 24rpx;font-size:30rpx;line-height:1.9;color:#1a1a1a;text-align:justify'
  var head = 'margin:48rpx 0 24rpx;font-family:Georgia,"Songti SC","STSong",SimSun,serif;font-size:34rpx;font-weight:bold;color:#074D92'
  return html
    .replace(/<p([ >])/gi, '<p style="' + block + '"$1')
    .replace(/<h([1-4])([ >])/gi, '<h$1 style="' + head + '"$2')
    .replace(/<li([ >])/gi, '<li style="margin-bottom:16rpx;font-size:30rpx;line-height:1.9;color:#1a1a1a"$1')
    .replace(/<img([ >])/gi, '<img style="max-width:100%;height:auto;border-radius:8rpx;margin:16rpx 0"$1')
    .replace(/<a([ >])/gi, '<a style="color:#0A66C2"$1')
}

// 详情页缓存：无公开按 ID 查询接口（决策②：列表传参），列表页导航前写入，
// 详情页按 id 读取。真机同会话内可靠；深链/分享场景待后端补公开详情接口。
let cachedReport = null
export function setReportCache(item) { cachedReport = item }
export function getReportCache(id) {
  return cachedReport && String(cachedReport.id) === String(id) ? cachedReport : null
}

// 内联 SVG 图标（项目约定：data-URI，不用 emoji/图片资源）。
export function svgUri(pathD, color) {
  return 'data:image/svg+xml,' + encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#' +
    color +
    '" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="' +
    pathD +
    '"/></svg>'
  )
}

// 返回箭头 / 搜索 / 下载（箭头入托盘线）/ 文件
export const ICON_BACK = svgUri('M15 5l-7 7 7 7', 'FFFFFF')
export const ICON_SEARCH_WHITE = svgUri('M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16z M21 21l-4.35-4.35', 'B9CBE8')
export const ICON_DOWNLOAD_NAVY = svgUri('M12 4v9m0 0l-3.5-3.5M12 13l3.5-3.5 M5 19.5h14', '074D92')
export const ICON_FILE = svgUri('M6 3h8l4 4v14H6z M14 3v4h4', '0A66C2')
