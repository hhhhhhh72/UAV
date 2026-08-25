// 后端 API 地址配置（唯一配置点）
// 生产环境：正式域名（微信后台 request/uploadFile/downloadFile 合法域名已配置 https://api.cqnarc.cn）
export const BASE_URL = 'https://api.cqnarc.cn'

// 演示数据开关（严格仅开发环境生效：NODE_ENV==='development' 且此开关为 true 时
// 报告/成果等页面走本地 mock 数据；生产编译 NODE_ENV 恒为 production，恒走真实接口）。
export const FORCE_MOCK_REPORTS = false
