// 开发态视觉 fixture（仅用于本地视觉验收，绝不进入发布包）
//
// 用途：按 Codex 修订指令「五、验证步骤」，需要在微信开发者工具里分别验收
// 企业已认证 / 认证飞手 / 普通个人 / 未登录 四种身份画面。
//
// 规则（对齐修订指令「四、代码组织与数据限制」）：
//  1. 默认关闭：仅当 storage 显式设置 `mine_fixture = 'enterprise' | 'pilot' | 'individual'` 才生效；
//  2. 发布剔除：开发环境专用的条件编译区块只会进入本地产物，正式构建不含此文件调用逻辑；
//  3. 无权限变化：只影响 mine 页展示的 view model，不写真实 user/token，不改认证接口；
//  4. 不写入 storage 作为真实数据：fixture 值只存在于本次会话内存。
import { getStoredUser } from './request'

// 生产环境（非 DEV）直接返回 null，杜绝进入正式包
// #ifndef DEV
const FIXTURE_AVAILABLE = false
// #endif

// #ifdef DEV
const FIXTURE_AVAILABLE = true

const FIXTURES = {
  // 企业已认证：重庆云翎科技有限公司 / enterprise / approved
  enterprise: {
    user: { name: '重庆云翎科技有限公司', role: 'enterprise', phone: '138****5621' },
    enterpriseStatus: 'approved',
    pilotStatus: '',
    overview: { publish: '6', talk: '2', certText: '已通过' },
    device: { bound: '2', online: '1', flights: '12' },
  },
  // 认证飞手：张航 / individual + pilot approved
  pilot: {
    user: { name: '张航', role: 'individual', phone: '138****5621', isAuth: true },
    enterpriseStatus: '',
    pilotStatus: 'approved',
    overview: { certText: '已通过', flights: '12', certs: '2' },
    device: { bound: '2', online: '1', flights: '12' },
  },
  // 普通个人：张航 / individual（实名认证为演示写死状态，无需 fixture 字段）
  individual: {
    user: { name: '张航', role: 'individual', phone: '138****5621' },
    enterpriseStatus: '',
    pilotStatus: '',
    overview: { authText: '已认证' },
  },
}
// #endif

// 读取当前 fixture。返回 { active, scenario } 或 { active: false }。
// 调用方据此在 dev 下替换身份相关数据；正式构建恒为 inactive。
export function getMineFixture() {
  if (!FIXTURE_AVAILABLE) return { active: false, scenario: null }

  // 开发环境：由显式 storage 开关激活，默认关闭
  const key = uni.getStorageSync('mine_fixture') || ''
  const scenario = FIXTURES[key]
  if (!scenario) return { active: false, scenario: null }

  // fixture 只派生展示数据；真实登录态仍以 storage 为准
  const realUser = getStoredUser()
  return {
    active: true,
    scenario: key,
    data: scenario,
    userOverride: realUser ? null : scenario.user,
  }
}
