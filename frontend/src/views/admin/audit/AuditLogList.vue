<template>
  <div class="admin-page">
    <div class="page-header">
      <div class="page-header-main">
        <h2>操作审计</h2>
        <span class="page-sub">谁、在什么时候、对什么对象、做了什么</span>
      </div>
      <span class="page-hint">勾选记录可导出，点行首箭头查看请求ID与元数据</span>
    </div>

    <CrudList
      resource="audit-logs"
      :columns="columns"
      :search-fields="searchFields"
      :creatable="false"
      :batch-delete="false"
      :selectable="true"
      :show-export="false"
      :api-function="listAudit"
      size="small"
      :scroll="{ x: 940 }"
      :expandable="expandable"
    >
      <!-- 时间：等宽数字，逐行对齐不抖动 -->
      <template #time="{ record }">
        <span class="cell-time">{{ formatTime(record.created_at) }}</span>
      </template>

      <!-- 操作人：等宽、过长省略，悬浮看完整 ID -->
      <template #actor="{ record }">
        <a-tooltip :content="record.actor_id" :disabled="!record.actor_id">
          <span class="cell-mono">{{ shortId(record.actor_id) }}</span>
        </a-tooltip>
      </template>

      <!-- 操作：中文为准，原始动作码放悬浮提示 -->
      <template #actionLabel="{ record }">
        <a-tooltip :content="record.action" :disabled="!record.action || actionLabel(record.action) === record.action">
          <span class="cell-mono">{{ actionLabel(record.action) }}</span>
        </a-tooltip>
      </template>

      <!-- 对象：资源类型 + 资源ID 合并成一列，长 ID 省略 -->
      <template #target="{ record }">
        <span class="cell-chip">{{ resourceLabel(record.resource_type) }}</span>
        <a-tooltip :content="record.resource_id" :disabled="!record.resource_id">
          <span class="cell-mono cell-dim">{{ shortId(record.resource_id) }}</span>
        </a-tooltip>
      </template>

      <!-- 结果：色点 + 中文，不用整块标签，减少表格里的色块噪音 -->
      <template #result="{ record }">
        <span class="cell-result">
          <i class="dot" :style="{ background: resultColor(record.result) }"></i>
          <span>{{ resultLabel(record.result) }}</span>
        </span>
      </template>

      <!-- 展开行：请求ID / 完整资源ID / 元数据（主表因此保持清爽） -->
      <template #expand="{ record }">
        <div class="row-detail">
          <div class="row-detail-item">
            <span class="row-detail-label">请求 ID</span>
            <span class="cell-mono">{{ record.request_id || '—' }}</span>
          </div>
          <div class="row-detail-item">
            <span class="row-detail-label">资源 ID</span>
            <span class="cell-mono">{{ record.resource_id || '—' }}</span>
          </div>
          <div class="row-detail-item row-detail-item--block">
            <span class="row-detail-label">元数据</span>
            <pre class="row-detail-meta">{{ formatMeta(record.metadata) }}</pre>
          </div>
        </div>
      </template>
      <!-- 多选后的批量动作：导出选中记录（审计留痕只读，导出是唯一有意义的批量操作） -->
      <template #batch="{ rows }">
        <a-button type="primary" size="small" @click="exportSelected(rows)">
          <template #icon><icon-download /></template>
          导出选中 CSV
        </a-button>
      </template>
    </CrudList>
  </div>
</template>

<script setup>
import CrudList from '../components/CrudList.vue'
import axios from '@/utils/http'

/* 动作中文映射（覆盖后端 internal 下全部 s.audit 调用点）；未命中时原样显示英文码 */
const ACTION_LABELS = {
  accept_intent: '确认接单',
  accept_work_order: '接单',
  add_cert: '新增证书',
  apply_aftersale: '申请售后',
  approve_certificate: '证书通过',
  approve_demand: '需求审核通过',
  approve_pilot: '飞手认证通过',
  assign_worker: '派单',
  attach_enterprise_doc: '上传企业材料',
  batch_approve_demands: '批量通过需求',
  batch_review_enterprise: '批量审核企业',
  book_resource: '预约资源',
  broadcast_message: '群发消息',
  cancel_demand: '取消需求',
  cancel_intent: '取消意向',
  cancel_work_order: '取消工单',
  close_demand: '关闭需求',
  complete_demand: '需求完成',
  complete_enrollment: '结业完成',
  complete_enrollment_cert_failed: '结业（出证失败）',
  complete_work_order: '完成工单',
  create_achievement: '新增成果',
  create_article: '发布资讯',
  create_case: '新增案例',
  create_competition: '创建赛事',
  create_compliance_doc: '新增合规文档',
  create_compliance_standard: '新增团体标准',
  create_demand: '发布需求',
  create_emergency_dispatch: '创建应急调度',
  create_emergency_resource: '新增应急资源',
  create_enterprise: '创建企业',
  create_event: '创建活动',
  create_expert: '新增专家',
  create_job: '发布职位',
  create_message: '创建消息',
  create_portfolio: '新增作品集',
  create_post: '发布帖子',
  create_project_app: '提交项目申报',
  create_rd_challenge: '发布研发难题',
  create_report: '发布报告',
  create_research_project: '创建科研项目',
  create_resource: '新增产业资源',
  create_service_listing: '发布服务',
  create_trade_order: '创建订单',
  delete_achievement: '删除成果',
  delete_article: '删除资讯',
  delete_case: '删除案例',
  delete_expert: '删除专家',
  delete_order: '删除订单',
  delete_report: '删除报告',
  delete_user: '删除用户',
  escrow_deposit: '托管金充值',
  escrow_freeze: '托管金冻结',
  escrow_refund: '托管金退款',
  escrow_release: '托管金释放',
  export_demands: '导出需求',
  export_resource: '导出数据',
  import_members: '导入会员',
  join_research_project: '加入科研项目',
  login_sms: '短信登录',
  login_wechat: '微信登录',
  logout: '退出登录',
  pay_and_enroll: '付费报名',
  pay_trade_order: '支付订单',
  register_user: '用户注册',
  reject_certificate: '证书驳回',
  reject_intent: '拒绝意向',
  review_aftersale: '售后审核',
  review_demand: '需求审核',
  review_enrollment: '报名审核',
  review_enterprise: '企业认证审核',
  review_project_app: '项目申报审核',
  review_study_tour_enrollment: '研学报名审核',
  rework_work_order: '工单返工',
  save_services_config: '保存服务配置',
  set_offline_amount: '设置线下金额',
  signing_callback: '合同签署回调',
  start_work_order: '开始工单',
  study_tour_enroll: '研学报名',
  submit_demand: '提交需求',
  submit_enterprise: '提交企业认证',
  update_article: '更新资讯',
  update_enterprise: '更新企业',
  update_join_status: '更新加入状态',
  update_platform_config: '更新平台配置',
  update_user_role: '修改用户角色',
  upload_file: '上传文件',
  void_contract: '作废合同'
}

/* 资源类型中文名（库中实际出现 + 代码中已埋点） */
const RESOURCE_LABELS = {
  auth: '登录鉴权',
  file: '文件',
  demand: '需求',
  demands: '需求',
  enterprise: '企业',
  trade_order: '订单',
  csv: '导出',
  user: '用户',
  case: '案例',
  escrow: '托管金',
  article: '资讯',
  enrollment: '课程报名',
  intent: '对接意向',
  study_tour_enrollment: '研学报名',
  certificate: '证书',
  competition: '赛事',
  work_order: '工单',
  certified_pilot: '飞手',
  achievement: '成果',
  association_member: '协会会员',
  compliance_doc: '合规文档',
  config: '平台配置',
  contract: '合同',
  emergency_dispatch: '应急调度',
  emergency_resource: '应急资源',
  event: '活动',
  expert: '专家',
  industry_resource: '产业资源',
  job: '招聘职位',
  message: '消息',
  messages: '消息',
  platform: '平台',
  portfolio: '作品集',
  post: '帖子',
  project_app: '项目申报',
  project_join_request: '项目加入申请',
  rd_challenge: '研发难题',
  report: '行业报告',
  research_project: '科研项目',
  resource: '资源',
  service_listing: '服务展示',
  services_config: '服务配置',
  standard: '团体标准',
  system: '系统',
  url: '链接'
}

/* 结果中文映射：后端 result 记的是业务结果（created/approved/pending…），不是 ok/failed */
const RESULT_LABELS = {
  success: '成功',
  created: '已创建',
  uploaded: '已上传',
  updated: '已更新',
  saved: '已保存',
  sent: '已发送',
  imported: '已导入',
  registered: '已注册',
  submitted: '已提交',
  approved: '已通过',
  approve: '已通过',
  accepted: '已受理',
  completed: '已完成',
  'completed+cert_issued': '完成并出证',
  enrolled: '已报名',
  paid: '已支付',
  deposited: '已充值',
  frozen: '已冻结',
  released: '已释放',
  refunded: '已退款',
  exported: '已导出',
  booked: '已预约',
  received: '已签署',
  started: '已开始',
  closed: '已关闭',
  reworked: '已返工',
  pending: '待处理',
  reject: '已驳回',
  rejected: '已驳回',
  cancelled: '已取消',
  deleted: '已删除',
  voided: '已作废',
  logged_out: '已退出'
}

const actionLabel = (a) => ACTION_LABELS[a] || a || '—'
const resourceLabel = (t) => RESOURCE_LABELS[t] || t || '—'
const resultLabel = (r) => RESULT_LABELS[r] || r || '—'

/* 结果色点：正向=成功色 / 待处理=警示色 / 否定=危险色 / 退出等=中性灰 */
const resultColor = (r) => {
  const s = String(r || '').toLowerCase()
  if (['reject', 'rejected', 'cancelled', 'deleted', 'voided'].includes(s)) return 'var(--danger-color)'
  if (s === 'pending') return 'var(--warning-color)'
  if (['logged_out', 'logout'].includes(s)) return 'var(--text-tertiary)'
  if (!s) return 'var(--text-tertiary)'
  return 'var(--success-color)'
}

/* 长标识省略：保留首尾特征（悬浮可见完整值） */
const shortId = (v) => {
  if (!v) return '—'
  const s = String(v)
  return s.length > 24 ? s.slice(0, 12) + '…' + s.slice(-6) : s
}

const formatTime = (v) => {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return String(v).slice(0, 19).replace('T', ' ')
  const pad = (n) => String(n).padStart(2, '0')
  return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds())
}

const formatMeta = (m) => {
  if (m == null) return '—'
  try {
    const s = typeof m === 'string' ? m : JSON.stringify(m, null, 2)
    return !s || s === '{}' ? '—' : s
  } catch (e) {
    return String(m)
  }
}

/* 5 列自适应：宽度即最小宽度，超出由省略号与悬浮提示兜住 */
const columns = [
  { title: '时间', dataIndex: 'created_at', slotName: 'time', width: 168 },
  { title: '操作人', dataIndex: 'actor_id', slotName: 'actor', width: 170 },
  { title: '操作', dataIndex: 'action', slotName: 'actionLabel', width: 150 },
  { title: '对象', dataIndex: 'resource_id', slotName: 'target', width: 260 },
  { title: '结果', dataIndex: 'result', slotName: 'result', width: 110 }
]

const expandable = { title: '', width: 40 }

/* 下拉筛选：动作与资源类型用可搜索下拉，避免手敲英文码 */
const ACTION_OPTIONS = Object.keys(ACTION_LABELS)
  .sort()
  .map((k) => ({ label: ACTION_LABELS[k] + '（' + k + '）', value: k }))

const RESOURCE_OPTIONS = Object.keys(RESOURCE_LABELS)
  .sort()
  .map((k) => ({ label: RESOURCE_LABELS[k] + '（' + k + '）', value: k }))

const searchFields = [
  { key: 'actor_id', label: '操作人', placeholder: '用户 ID', width: 180 },
  { key: 'action', label: '操作', type: 'select', options: ACTION_OPTIONS, searchable: true, width: 200 },
  { key: 'resource_type', label: '资源类型', type: 'select', options: RESOURCE_OPTIONS, searchable: true, width: 190 },
  { key: 'range', label: '时间范围', type: 'range', startKey: 'start', endKey: 'end', width: 240 }
]

const listAudit = (params) => axios.get('/api/v1/admin/audit-logs', { params }).then((r) => r.data)

/* 导出选中记录：客户端生成 CSV（带 BOM，Excel 中文不乱码），列与表格同口径但用中文 */
const csvCell = (v) => {
  const s = v == null ? '' : String(v)
  return /[",\n]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s
}

const exportSelected = (rows) => {
  if (!rows || !rows.length) return
  const header = ['时间', '操作人', '操作', '对象类型', '资源ID', '结果', '请求ID']
  const lines = [header].concat(rows.map((r) => [
    formatTime(r.created_at),
    r.actor_id || '',
    actionLabel(r.action),
    resourceLabel(r.resource_type),
    r.resource_id || '',
    resultLabel(r.result),
    r.request_id || ''
  ]))
  const csv = '\ufeff' + lines.map((cells) => cells.map(csvCell).join(',')).join('\r\n')
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8;' }))
  const a = document.createElement('a')
  a.href = url
  a.download = '操作审计-选中-' + new Date().toISOString().slice(0, 10) + '.csv'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
</script>

<style scoped>
.admin-page {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.page-header-main {
  display: flex;
  align-items: baseline;
  gap: 12px;
  flex-wrap: wrap;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-1);
}
.page-sub {
  font-size: 13px;
  color: var(--color-text-3);
}
.page-hint {
  font-size: 12px;
  color: var(--color-text-3);
}

/* 单元格：等宽数字/标识 + 层级 + 省略 */
.cell-time {
  font-variant-numeric: tabular-nums;
  color: var(--color-text-2);
  white-space: nowrap;
}
.cell-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: var(--color-text-2);
}
.cell-dim {
  margin-left: 8px;
  color: var(--color-text-3);
}
.cell-chip {
  display: inline-block;
  padding: 0 6px;
  border-radius: 4px;
  background: var(--color-fill-2);
  color: var(--color-text-2);
  font-size: 12px;
  white-space: nowrap;
}

/* 结果：色点 + 文字（比整块 tag 更安静，颜色仍可扫读） */
.cell-result {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-2);
  white-space: nowrap;
}
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex: 0 0 auto;
}

/* 展开行详情：请求ID / 资源ID / 元数据 */
.row-detail {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 8px 24px;
  padding: 12px 16px;
  background: var(--color-fill-1);
  border-radius: 8px;
}
.row-detail-item {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}
.row-detail-item--block {
  grid-column: 1 / -1;
  align-items: flex-start;
}
.row-detail-label {
  flex: 0 0 auto;
  font-size: 12px;
  color: var(--color-text-3);
}
.row-detail-meta {
  margin: 0;
  max-height: 200px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-text-2);
  white-space: pre-wrap;
  word-break: break-all;
}
.row-detail-meta::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
.row-detail-meta::-webkit-scrollbar-thumb {
  background: var(--color-border-2);
  border-radius: 4px;
}
.row-detail-meta::-webkit-scrollbar-track {
  background: transparent;
}
</style>
