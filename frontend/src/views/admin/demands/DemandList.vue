<template>
  <div class="demand-list-page">
    <!-- 搜索过滤区 -->
    <div class="search-bar">
      <div class="search-row">
        <el-input
          v-model="filterParams.keyword"
          placeholder="搜索需求标题..."
          clearable
          style="width: 240px"
          @keyup.enter="onSearchSubmit"
          @clear="onSearchSubmit"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="filterParams.status" clearable style="width: 140px" @change="onSearchSubmit">
          <el-option label="全部" value="all" />
          <el-option label="待审核" value="pending" />
          <el-option label="已发布" value="published" />
          <el-option label="已完成" value="completed" />
          <el-option label="已取消" value="cancelled" />
          <el-option label="已驳回" value="rejected" />
        </el-select>

        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
      </div>
    </div>

    <!-- 撮合统计条 -->
    <div class="stats-bar">
      <div class="stat"><span class="stat-num">{{ stats.total }}</span><span class="stat-label">需求总数</span></div>
      <div class="stat warn"><span class="stat-num">{{ stats.pending }}</span><span class="stat-label">待审核</span></div>
      <div class="stat ok"><span class="stat-num">{{ stats.published }}</span><span class="stat-label">已公开</span></div>
      <div class="stat done"><span class="stat-num">{{ stats.completed }}</span><span class="stat-label">已完成</span></div>
      <div class="stat rate"><span class="stat-num">{{ stats.rate }}%</span><span class="stat-label">完成率</span></div>
      <div class="stat amount"><span class="stat-num">¥{{ stats.offlineAmount }}</span><span class="stat-label">撮合成交额(本页)</span></div>
    </div>

    <!-- 批量操作栏 -->
    <div class="batch-bar" v-if="selectedIds.length > 0">
      <span class="batch-info">已选择 <b>{{ selectedIds.length }}</b> 项</span>
      <el-button type="success" :icon="Check" @click="batchApprove">批量通过</el-button>
      <el-button type="danger" :icon="CloseBold" @click="batchReject">批量驳回</el-button>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrap">
      <el-table
        v-loading="loading"
        :data="listData"
        row-key="id"
        stripe border
        @selection-change="onSelectChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="40" />

        <el-table-column prop="title" label="需求标题" min-width="200" sortable="custom">
          <template #default="{ row }">
            <span class="cell-title">{{ row.title || '无标题' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="publisher_name" label="发布者" width="130" />
        <el-table-column prop="biz_type" label="业务类型" width="120">
          <template #default="{ row }">{{ bizTypeLabel(row.biz_type) }}</template>
        </el-table-column>

        <el-table-column prop="district" label="地区" width="110" />

        <el-table-column prop="budget_fen" label="预算" width="110" sortable="custom" align="right">
          <template #default="{ row }">
            {{ row.budget_fen ? '¥' + Number(row.budget_fen / 100).toLocaleString() : '-' }}
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="提交时间" width="160" sortable="custom">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>

        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <template v-if="row.status === 'pending'">
              <el-divider direction="vertical" />
              <el-button link type="success" size="small" @click="handleApprove(row)">通过</el-button>
              <el-button link type="danger" size="small" @click="handleReject(row)">驳回</el-button>
            </template>
            <template v-else-if="row.status === 'published' || row.status === 'completed'">
              <el-divider direction="vertical" />
              <el-button link type="warning" size="small" @click="handleClose(row)">关闭</el-button>
              <el-button link type="success" size="small" @click="handleAmount(row)">登记金额</el-button>
            </template>
          </template>
        </el-table-column>

        <template #empty><el-empty description="暂无需求数据" /></template>
      </el-table>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination
        v-model:current-page="filterParams.page"
        v-model:page-size="filterParams.page_size"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @size-change="loadData"
        @current-change="loadData"
      />
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="需求详情" width="600px" :close-on-click-modal="false">
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="标题" :span="2">{{ currentItem.title || '-' }}</el-descriptions-item>
          <el-descriptions-item label="发布者">{{ currentItem.publisher_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="业务类型">{{ bizTypeLabel(currentItem.biz_type) }}</el-descriptions-item>
          <el-descriptions-item label="地区">{{ currentItem.district || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="预算">{{ currentItem.budget_fen ? '¥' + (currentItem.budget_fen / 100).toLocaleString() : '-' }}</el-descriptions-item>
          <el-descriptions-item label="线下成交金额">{{ currentItem.offline_amount_fen ? '¥' + (currentItem.offline_amount_fen / 100).toLocaleString() : '-' }}</el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ formatDate(currentItem.created_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="currentItem.biz_fields && currentItem.biz_fields.reject_reason" label="驳回/关闭原因" :span="2">
            {{ currentItem.biz_fields.reject_reason }}
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
        </el-descriptions>

      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Check, CloseBold } from '@element-plus/icons-vue'
import { showSuccessToast, showFailToast } from '@/utils/feedback'
import { ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { getDemandList, approveDemand, rejectDemand, closeDemand, setOfflineAmount } from '@/api/admin/demand'

const bizTypeLabel = (t) => ({
  aerial_photo: '航拍摄影', mapping: '测绘', inspection: '巡检',
  agriculture: '植保', logistics: '物流配送', training: '培训',
  competition: '赛事', other: '其他'
}[t] || t || '-')

const statusLabel = (s) => ({
  pending: '待审核', published: '已发布', matched: '已匹配',
  completed: '已完成', cancelled: '已取消', rejected: '已驳回'
}[s] || s || '-')

const statusTagType = (s) => ({
  published: 'success', matched: '', completed: 'success',
  pending: 'warning', cancelled: 'info', rejected: 'danger'
}[s] || 'info')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth()+1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

// 撮合统计（分类计数基于当前页数据；需求总数取接口 total）
const stats = computed(() => {
  const rows = listData.value || []
  const pending = rows.filter((x) => x.status === 'pending').length
  const published = rows.filter((x) => x.status === 'published').length
  const completed = rows.filter((x) => x.status === 'completed').length
  const rate = rows.length ? Math.round((completed / rows.length) * 100) : 0
  const offlineAmount = rows.reduce((s, x) => s + (x.offline_amount_fen || 0), 0) / 100
  return {
    total: total.value || 0,
    pending, published, completed, rate,
    offlineAmount: offlineAmount.toLocaleString('zh-CN', { minimumFractionDigits: 2 }),
  }
})

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: getDemandList,
  idKey: 'id',
  // 默认全量视角：通过/驳回后数据不因筛选变化而"消失"
  defaultParams: { status: 'all' }
})

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (d) => { currentItem.value = d; detailVisible.value = true }

const handleApprove = async (item) => {
  try {
    await approveDemand(item.id)
    showSuccessToast('审核通过')
    item.status = 'published'
    detailVisible.value = false
    loadData()
  } catch (e) { showFailToast(errMsg(e)) }
}

// 统一错误提示：后端 fail 格式为 {error:{code,message}}，逐层取
const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const handleReject = async (item) => {
  try {
    const { value } = await ElMessageBox.prompt('请填写驳回理由（发布者可见，可据此修改后重提）', '驳回需求', {
      confirmButtonText: '确认驳回',
      cancelButtonText: '取消',
      inputPlaceholder: '如：信息不完整 / 违规内容 / 重复发布',
      inputValidator: (v) => (v && v.trim() ? true : '驳回理由必填'),
    })
    await rejectDemand(item.id, value.trim())
    showSuccessToast('已驳回')
    item.status = 'rejected'
    detailVisible.value = false
    loadData()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      // 带真实错误原因，便于定位（如：HTTP 状态码 + 后端 message）
      const msg = errMsg(e)
      showFailToast(msg)
      console.error('[驳回需求失败]', e)
    }
  }
}

// 登记线下成交金额（联系对接模式：平台撮合价值度量）
const handleAmount = async (item) => {
  try {
    const { value } = await ElMessageBox.prompt('登记该需求线下成交金额（元）', '登记成交金额', {
      confirmButtonText: '确认登记',
      cancelButtonText: '取消',
      inputPlaceholder: '如：12000',
      inputPattern: /^\d+(\.\d{1,2})?$/,
      inputErrorMessage: '请输入有效金额',
    })
    const fen = Math.round(Number(value) * 100)
    await setOfflineAmount(item.id, fen)
    showSuccessToast('已登记 ¥' + value)
    item.offline_amount_fen = fen
    loadData()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') showFailToast(errMsg(e))
  }
}

// 关闭已公开需求（发布者失联/虚假信息/线下已成交）
const handleClose = async (item) => {
  try {
    const { value } = await ElMessageBox.prompt('关闭后需求从大厅下架，请填写关闭原因', '关闭需求', {
      confirmButtonText: '确认关闭',
      cancelButtonText: '取消',
      inputPlaceholder: '如：线下已成交 / 信息失实',
      inputValidator: (v) => (v && v.trim() ? true : '关闭原因必填'),
    })
    await closeDemand(item.id, value.trim())
    showSuccessToast('已关闭')
    item.status = 'cancelled'
    loadData()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') showFailToast(errMsg(e))
  }
}

const batchApprove = () => {
  selectedIds.value.forEach(id => approveDemand(id).catch(() => {}))
  showSuccessToast('批量通过已提交')
  loadData()
}

const batchReject = () => {
  selectedIds.value.forEach(id => rejectDemand(id, '').catch(() => {}))
  showSuccessToast('批量驳回已提交')
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.demand-list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }

/* 撮合统计条 */
.stats-bar { display: flex; gap: 32px; background: #fff; border-radius: 8px; padding: 14px 20px; margin-bottom: 12px; box-shadow: 0 1px 3px rgba(0,0,0,.06); }
.stat { display: flex; align-items: baseline; gap: 8px; }
.stat-num { font-size: 22px; font-weight: 700; color: var(--el-text-color-primary); }
.stat.warn .stat-num { color: var(--el-color-warning); }
.stat.ok .stat-num { color: var(--el-color-success); }
.stat.done .stat-num { color: var(--el-color-primary); }
.stat.rate .stat-num { color: var(--el-color-info); }
.stat.amount .stat-num { color: var(--el-color-success); }
.stat-label { font-size: 13px; color: var(--el-text-color-secondary); }
.batch-bar { background: var(--el-color-primary-light-9); border: 1px solid var(--el-color-primary-light-5); border-radius: 8px; padding: 10px 16px; margin-bottom: 16px; display: flex; align-items: center; gap: 12px; }
.batch-info { font-size: 13px; color: var(--el-text-color-secondary); margin-right: auto; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.cell-title { font-weight: 500; color: var(--el-text-color-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; max-width: 300px; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.review-actions { text-align: center; padding-top: 16px; }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
