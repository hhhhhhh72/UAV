// 需求接单全链路进度：把「需求 / 意向 / 工单」三个实体的状态串成一条线。
//
// 后端是一条跨三张表的链路（domain/models.go:231-236 需求 6 态、:296-300 工单 5 态）：
//
//   需求  pending 待审核 → published 公示中 → assigned 已接单（同时生成工单）
//         └→ 工单  pending 待开始 → ongoing 进行中 → awaiting_accept 待验收 → completed 已完成
//
// 此前前端**没有任何一处**把它们串起来：用户想知道"现在到哪了"只能翻消息通知
//（work_order.go 每一步确实都推了通知，但通知是流水、不是状态）。
// 这里给出唯一的映射来源，需求详情页与工单详情页共用，避免两边各推一套。

export const DEMAND_FLOW_STEPS = [
  { key: 'created', label: '已提交' },
  { key: 'published', label: '审核通过' },
  { key: 'assigned', label: '已接单' },
  { key: 'ongoing', label: '作业中' },
  { key: 'accepting', label: '待验收' },
  { key: 'completed', label: '已完成' },
]

// 负面终止态：线性进度条表达不了"第几步失败"，返回 tone 让页面改显示状态提示。
const TERMINAL_DEMAND = { rejected: '需求未通过审核', cancelled: '需求已下架' }

/**
 * @param {string} demandStatus    需求状态（demand.status）
 * @param {string} workOrderStatus 工单状态（work_order.status），没有工单传空
 * @returns {{ steps: Array, activeIndex: number, tone: 'active'|'danger'|'muted', note: string }}
 *          activeIndex 为当前停在第几步（0 基）；i < activeIndex 的步骤视为已完成。
 *          tone 非 'active' 时页面应改用 note 展示，而不是画进度条。
 */
export function demandFlow(demandStatus, workOrderStatus) {
  const d = String(demandStatus || '').trim()
  const w = String(workOrderStatus || '').trim()

  if (TERMINAL_DEMAND[d]) {
    return { steps: DEMAND_FLOW_STEPS, activeIndex: -1, tone: d === 'rejected' ? 'danger' : 'muted', note: TERMINAL_DEMAND[d], currentLabel: '' }
  }
  if (w === 'cancelled') {
    return { steps: DEMAND_FLOW_STEPS, activeIndex: -1, tone: 'muted', note: '工单已取消', currentLabel: '' }
  }

  let idx = 0 // 已提交
  // 用 Math.max 逐级抬高，避免后面的判断把前面的进度压回去
  if (d === 'published' || d === 'assigned' || d === 'completed') idx = Math.max(idx, 1) // 审核通过
  // assigned 的定义就是"已接受意向并生成工单"（domain/models.go:233），
  // 所以需求状态本身就足以判定到"已接单"——**不能依赖工单是否拉取成功**，
  // 否则需求详情页在工单请求返回前（或失败时）会把进度错画成"审核通过"。
  if (d === 'assigned' || d === 'completed' || w) idx = Math.max(idx, 2)
  if (w === 'ongoing') idx = Math.max(idx, 3)
  if (w === 'awaiting_accept') idx = Math.max(idx, 4)
  if (w === 'completed' || d === 'completed') idx = Math.max(idx, 5)
  const cur = DEMAND_FLOW_STEPS[idx]
  return { steps: DEMAND_FLOW_STEPS, activeIndex: idx, tone: 'active', note: '', currentLabel: cur ? cur.label : '' }
}
