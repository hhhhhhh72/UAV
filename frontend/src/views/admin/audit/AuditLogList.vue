<template>
  <div class="admin-page">
    <div class="page-header">
      <h2>操作审计</h2>
      <span class="page-sub">记录管理端与业务的写操作（谁、何时、对什么、做了什么），用于合规追溯与问题排查</span>
    </div>

    <CrudList
      resource="audit-logs"
      :columns="columns"
      :search-fields="searchFields"
      :creatable="false"
      :batch-delete="false"
      :selectable="false"
      :api-function="listAudit"
    >
      <template #actionLabel="{ record }">
        <span class="audit-action">{{ actionLabel(record.action) }}</span>
        <span v-if="actionLabel(record.action) !== record.action" class="audit-action-raw">{{ record.action }}</span>
      </template>
      <template #result="{ record }">
        <a-tag :color="record.result === 'ok' || record.result === 'success' ? 'green' : 'gray'" size="small">
          {{ record.result || '-' }}
        </a-tag>
      </template>
      <template #time="{ record }">{{ formatTime(record.created_at) }}</template>
    </CrudList>
  </div>
</template>

<script setup>
import CrudList from '../components/CrudList.vue'
import axios from '@/utils/http'

/* 动作中文映射：覆盖当前后端已记录的动作类型（未命中时原样显示 + 灰色原文） */
const ACTION_LABELS = {
  login_wechat: '微信登录',
  login_password: '密码登录',
  logout: '退出登录',
  refresh_token: '刷新令牌',
  register: '用户注册',
  upload_file: '文件上传',
  create_demand: '发布需求',
  approve_demand: '需求审核通过',
  close_demand: '关闭需求',
  create_enterprise: '企业注册',
  review_enterprise: '企业认证审核',
  batch_review_enterprises: '批量企业审核',
  create_trade_order: '创建订单',
  pay_and_enroll: '付费报名',
  enroll: '课程报名',
  complete_enrollment: '结业完成',
  review_enrollment: '报名审核',
  export_resource: '导出数据',
  export_demands: '导出需求',
  export_enterprises: '导出企业',
  batch_approve_demands: '批量审批需求',
  approve_pilot: '飞手认证通过',
  reject_pilot: '飞手认证驳回',
  review_certificate: '证书审核',
  approve_certificate: '证书审批',
  save_services_config: '保存平台配置',
  update_config: '更新配置',
  escrow_deposit: '托管金充值',
  escrow_release: '托管金释放',
  escrow_refund: '托管金退款',
  create_review: '提交评价',
  review_work_order: '工单评价',
  accept_intent: '确认接单',
  create_work_order: '生成工单'
}

const actionLabel = (a) => ACTION_LABELS[a] || a || '-'

const formatTime = (v) => {
  if (!v) return '-'
  const d = new Date(v)
  if (isNaN(d.getTime())) return String(v).slice(0, 19).replace('T', ' ')
  const pad = (n) => String(n).padStart(2, '0')
  return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds())
}

const columns = [
  { title: '时间', dataIndex: 'created_at', slotName: 'time', width: 170 },
  { title: '操作人', dataIndex: 'actor_id', width: 200 },
  { title: '操作', dataIndex: 'action', slotName: 'actionLabel', width: 180 },
  { title: '资源类型', dataIndex: 'resource_type', width: 130 },
  { title: '资源ID', dataIndex: 'resource_id', width: 220 },
  { title: '结果', dataIndex: 'result', slotName: 'result', width: 100 },
  { title: '请求ID', dataIndex: 'request_id', width: 180 }
]

const searchFields = [
  { key: 'actor_id', label: '操作人', placeholder: '用户ID，如 user-xxx', width: 220 },
  { key: 'action', label: '操作', placeholder: '如 approve_demand', width: 180 },
  { key: 'resource_type', label: '资源类型', placeholder: '如 demand / enterprise', width: 180 },
  { key: 'range', label: '时间范围', type: 'range', startKey: 'start', endKey: 'end', width: 260 }
]

/* 审计日志为只读查询接口，用自定义 apiFunction 覆盖 CRUD list */
const listAudit = (params) => axios.get('/api/v1/admin/audit-logs', { params }).then((r) => r.data)
</script>

<style scoped>
.admin-page {
  padding: 20px;
}
.page-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
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
.audit-action {
  font-weight: 600;
  color: var(--color-text-1);
}
.audit-action-raw {
  margin-left: 6px;
  font-size: 12px;
  color: var(--color-text-3);
}
</style>
