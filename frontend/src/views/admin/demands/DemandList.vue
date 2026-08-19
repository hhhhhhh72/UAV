<template>
  <div class="page">
    <!-- 撮合统计（独立全量统计接口，不随翻页变化） -->
    <a-card :bordered="false" class="stat-card">
      <div class="stats-bar">
        <div class="stat"><span class="stat-num">{{ stats.total }}</span><span class="stat-label">需求总数</span></div>
        <div class="stat warn"><span class="stat-num">{{ stats.pending }}</span><span class="stat-label">待审核</span></div>
        <div class="stat ok"><span class="stat-num">{{ stats.published }}</span><span class="stat-label">已公开</span></div>
        <div class="stat done"><span class="stat-num">{{ stats.completed }}</span><span class="stat-label">已完成</span></div>
        <div class="stat rate"><span class="stat-num">{{ stats.rate }}%</span><span class="stat-label">完成率</span></div>
        <div class="stat amount"><span class="stat-num">¥{{ Number(stats.offline_amount).toLocaleString('zh-CN', { minimumFractionDigits: 2 }) }}</span><span class="stat-label">撮合成交额</span></div>
      </div>
    </a-card>

    <CrudList
      ref="crudRef"
      resource="demands"
      :columns="columns"
      :search-fields="searchFields"
      :default-params="defaultParams"
      :batch-actions="batchActions"
      :batch-delete="false"
    >
      <template #title="{ record }">
        <span class="cell-title">{{ record.title || '-' }}</span>
      </template>
      <template #bizType="{ record }">
        <a-tag color="arcoblue" size="small">{{ bizTypeLabel(record.biz_type) }}</a-tag>
      </template>
      <template #image="{ record }">
        <a-image
          v-if="record.images && record.images.length"
          :src="record.images[0]"
          :preview-src="record.images[0]"
          width="56"
          height="40"
          :style="{ objectFit: 'cover', borderRadius: '4px', display: 'block' }"
          alt="需求图片"
          @error="onImgError"
        />
        <span v-else class="img-empty">-</span>
      </template>
      <template #price="{ record }">
        <span>{{ record.budget_fen ? '¥' + (record.budget_fen / 100).toLocaleString() : '面议' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusTagColor(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #offlineAmount="{ record }">
        <span v-if="record.offline_amount_fen" class="amount-text">¥{{ (record.offline_amount_fen / 100).toLocaleString() }}</span>
        <span v-else class="amount-empty">-</span>
      </template>
      <template #createdAt="{ record }">
        <span class="time-text">{{ formatDate(record.created_at) }}</span>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
          <template v-if="record.status === 'pending'">
            <a-button type="text" status="success" size="small" @click="handleApprove(record)">通过</a-button>
            <a-button type="text" status="danger" size="small" @click="openInputModal('reject', record)">驳回</a-button>
          </template>
          <template v-else-if="record.status === 'published' || record.status === 'completed'">
            <a-button type="text" status="warning" size="small" @click="openInputModal('close', record)">关闭</a-button>
            <a-button type="text" size="small" @click="openInputModal('amount', record)">登记金额</a-button>
          </template>
          <template v-else-if="record.status === 'cancelled' || record.status === 'rejected'">
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </template>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无需求数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="需求详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="需求标题" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型">{{ bizTypeLabel(currentItem.biz_type) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTagColor(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="预算">{{ currentItem.budget_fen ? '¥' + (currentItem.budget_fen / 100).toLocaleString() : '面议' }}</a-descriptions-item>
          <a-descriptions-item label="成交金额">{{ currentItem.offline_amount_fen ? '¥' + (currentItem.offline_amount_fen / 100).toLocaleString() : '-' }}</a-descriptions-item>
          <a-descriptions-item label="地区" :span="2">{{ currentItem.district || '-' }}</a-descriptions-item>
          <a-descriptions-item label="联系人">{{ currentItem.contact || '-' }}</a-descriptions-item>
          <a-descriptions-item label="提交时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item v-if="currentItem.images && currentItem.images.length" label="图片" :span="2">
            <div class="detail-imgs">
              <a-image
                v-for="(img, i) in currentItem.images"
                :key="i"
                :src="img"
                width="100"
                height="72"
                :style="{ objectFit: 'cover', borderRadius: '4px', marginRight: '8px' }"
                :preview-src="img"
                alt="需求图片"
                @error="onImgError"
              />
            </div>
          </a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 输入弹窗（驳回理由 / 关闭原因 / 登记金额） -->
    <a-modal v-model:visible="inputModal.visible" :title="inputModal.title" :width="440" @ok="confirmInputModal" @cancel="inputModal.visible = false">
      <a-input
        ref="inputModalRef"
        v-model="inputModal.value"
        :placeholder="inputModal.placeholder"
        :type="inputModal.type === 'amount' ? 'number' : 'text'"
        allow-clear
      />
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import axios from '@/utils/http'
import { approveDemand, rejectDemand, closeDemand, setOfflineAmount, deleteDemand } from '@/api/admin/demand'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const defaultParams = { status: 'all' }

// 图片加载失败（如 /static/ 相对路径 404）时隐藏，不显示破图
const onImgError = (e) => { e.target.style.display = 'none' }

const bizTypeLabel = (t) => ({
  cable_inspection: '工业巡检',
  plant_transport: '植保运输',
  spray_pesticide: '农药喷洒',
  trade_lease: '租赁服务',
  clean_paint: '清洗保洁',
  other: '其他'
}[t] || t || '-')

const statusLabel = (s) => ({
  pending: '待审核', published: '已公开', completed: '已完成',
  cancelled: '已取消', rejected: '已驳回'
}[s] || s || '-')

const statusTagColor = (s) => ({
  pending: 'orangered', published: 'arcoblue', completed: 'green',
  cancelled: 'gray', rejected: 'red'
}[s] || 'gray')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 撮合统计：独立全量统计接口，不随列表分页/筛选变化
const stats = ref({ total: 0, pending: 0, published: 0, completed: 0, cancelled: 0, rejected: 0, rate: 0, offline_amount: 0 })

const loadStats = async () => {
  try {
    const res = await axios.get('/api/v1/admin/demands/stats')
    const d = res?.data?.data || res?.data || {}
    stats.value = {
      total: d.total || 0, pending: d.pending || 0, published: d.published || 0,
      completed: d.completed || 0, cancelled: d.cancelled || 0, rejected: d.rejected || 0,
      rate: d.rate || 0, offline_amount: d.offline_amount || 0,
    }
  } catch (e) { /* 统计失败不阻塞列表 */ }
}

// 后端 listAdminDemands 仅支持 status 过滤，keyword 无效已移除
const searchFields = [
  { key: 'status', label: '状态', type: 'select', options: [
    { value: 'all', label: '全部' },
    { value: 'pending', label: '待审核' },
    { value: 'published', label: '已公开' },
    { value: 'completed', label: '已完成' },
    { value: 'cancelled', label: '已取消' },
    { value: 'rejected', label: '已驳回' }
  ]}
]

// 批量动作：批量通过/批量驳回（传完整行数据，逐行调用审核接口；完成后同步刷新统计条）
const batchActions = [
  { key: 'approve', label: '批量通过', status: 'success', api: async (row) => { await approveDemand(row.id); loadStats() } },
  { key: 'reject', label: '批量驳回', status: 'danger', api: async (row) => { await rejectDemand(row.id, ''); loadStats() } }
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '需求标题', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '图片', dataIndex: 'images', slotName: 'image', width: 90 },
  { title: '类型', dataIndex: 'biz_type', slotName: 'bizType', width: 100 },
  { title: '预算', dataIndex: 'budget_fen', slotName: 'price', width: 110, align: 'right' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '成交金额', dataIndex: 'offline_amount_fen', slotName: 'offlineAmount', width: 110, align: 'right' },
  { title: '提交时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (d) => { currentItem.value = d; detailVisible.value = true }

// 输入弹窗（驳回/关闭/金额共用）
const inputModal = reactive({ visible: false, type: '', title: '', placeholder: '', value: '', record: null })
const inputModalRef = ref()

const openInputModal = (type, record) => {
  const cfg = {
    reject: { title: '驳回需求', placeholder: '请填写驳回理由（发布者可见，可据此修改后重提）' },
    close: { title: '关闭需求', placeholder: '关闭后需求从大厅下架，请填写关闭原因' },
    amount: { title: '登记成交金额', placeholder: '登记该需求线下成交金额（元），如：12000' },
  }[type]
  Object.assign(inputModal, { visible: true, type, record, value: '', ...cfg })
}

const confirmInputModal = async () => {
  const { type, value, record } = inputModal
  if (!value || !value.trim()) { Message.warning('请填写内容'); return }
  try {
    if (type === 'reject') {
      await rejectDemand(record.id, value.trim())
      Message.success('已驳回')
      record.status = 'rejected'
    } else if (type === 'close') {
      await closeDemand(record.id, value.trim())
      Message.success('已关闭')
      record.status = 'cancelled'
    } else if (type === 'amount') {
      // 金额校验：NaN/Infinity/非正数/超大值一律拦截，避免 null/负数/溢出值入库
      const n = Number(value)
      if (!Number.isFinite(n) || n <= 0 || n > 99999999.99) {
        Message.error('成交金额需为大于 0 且不超过 99999999.99 的数字（元）')
        inputModalRef.value && inputModalRef.value.focus && inputModalRef.value.focus()
        return
      }
      const fen = Math.round(n * 100)
      await setOfflineAmount(record.id, fen)
      Message.success('已登记 ¥' + value)
      record.offline_amount_fen = fen
    }
    inputModal.visible = false
    crudRef.value?.reload(); loadStats()
  } catch (e) { Message.error(errMsg(e)) }
}

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const handleApprove = async (item) => {
  try {
    await approveDemand(item.id)
    Message.success('审核通过')
    item.status = 'published'
    detailVisible.value = false
    crudRef.value?.reload(); loadStats()
  } catch (e) { Message.error(errMsg(e)) }
}

// 删除已取消/已驳回需求
const handleDelete = (item) => {
  Modal.confirm({
    title: '删除需求',
    content: `确定删除该需求吗？删除后不可恢复（${item.title || item.id}）`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteDemand(item.id)
        Message.success('已删除')
        crudRef.value?.reload(); loadStats()
      } catch (e) { Message.error(errMsg(e)) }
    }
  })
}

onMounted(loadStats)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.stat-card { margin-bottom: 16px; }

.stats-bar {
  display: flex;
  gap: 0;
}

.stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 4px 16px;
  border-right: 1px solid #EEF1F4;
}

.stat:last-child { border-right: none; }

.stat-num {
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text-1);
  line-height: 1.2;
}

.stat-label {
  font-size: 12px;
  color: #86909C;
}

.stat.warn .stat-num { color: #E96012; }
.stat.ok .stat-num { color: #168A55; }
.stat.done .stat-num { color: #168A55; }
.stat.rate .stat-num { color: #165DFF; }
.stat.amount .stat-num { color: #E96012; font-size: 20px; }

.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.amount-text { color: #E96012; font-weight: 500; }
.amount-empty { color: #C9CDD4; }
.time-text { color: #86909C; font-size: 12px; }

.detail-imgs { display: flex; flex-wrap: wrap; gap: 8px; }

.img-empty { color: #C9CDD4; }
</style>
