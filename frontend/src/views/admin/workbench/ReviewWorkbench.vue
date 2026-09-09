<template>
  <div class="admin-page">
    <div class="page-header">
      <div>
        <h2>审核待办</h2>
        <p class="page-sub">平台各业务模块的待处理审核事项，点击卡片进入对应管理页处理</p>
      </div>
      <a-button :loading="loading" @click="loadAll">
        <template #icon><icon-refresh /></template>
        刷新
      </a-button>
    </div>

    <!-- 汇总条 -->
    <div class="wb-summary" :class="{ 'wb-summary--zero': !loading && totalPending === 0 }">
      <div class="wb-summary-main">
        <span class="wb-summary-num">{{ loading ? '—' : totalPending }}</span>
        <span class="wb-summary-unit">项待处理</span>
      </div>
      <span class="wb-summary-hint">{{ summaryHint }}</span>
    </div>

    <!-- 待办卡片 -->
    <a-row :gutter="[16, 16]">
      <a-col v-for="item in items" :key="item.key" :xs="24" :sm="12" :md="8" :lg="6">
        <div
          class="wb-card"
          :class="{ 'wb-card--zero': !loading && !item.count && !item.error, 'wb-card--error': !!item.error }"
          @click="go(item)"
        >
          <div class="wb-card-head">
            <span class="wb-card-title">{{ item.title }}</span>
            <span v-if="!loading && item.count > 0" class="wb-card-badge">待处理</span>
          </div>
          <div class="wb-card-count">{{ loading ? '—' : (item.error ? '!' : item.count) }}</div>
          <div class="wb-card-foot">
            <span class="wb-card-desc">{{ item.error || item.desc }}</span>
            <span class="wb-card-go">{{ item.error ? '重试' : '去处理' }} ›</span>
          </div>
        </div>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminApi } from '@/api/admin/common'

const router = useRouter()

/**
 * 待办数据源：各资源的管理端列表接口 + 待审状态值（与后端 domain 常量一致）
 * total 取自分页响应的 total 字段（page_size=1 只取数量，不拉数据）。
 */
const SOURCES = [
  { key: 'enterprises', title: '企业认证', desc: '企业实名认证申请', resource: 'enterprises', status: 'submitted', link: '/admin/enterprises' },
  { key: 'demands', title: '需求审核', desc: '需求大厅发布审核', resource: 'demands', status: 'pending', link: '/admin/demands' },
  { key: 'courses', title: '课程审核', desc: '培训课程上架审核', resource: 'training-courses', status: 'pending', link: '/admin/training' },
  { key: 'certificates', title: '证书审核', desc: '资质证书申请审核', resource: 'certificates', status: 'pending', link: '/admin/certs' },
  { key: 'pilots', title: '飞手认证', desc: '认证飞手申请审核', resource: 'certified-pilots', status: 'pending', link: '/admin/talent?tab=pilots' },
  { key: 'competitions', title: '赛事审核', desc: '企业发布赛事审核', resource: 'competitions', status: 'pending', link: '/admin/competition' },
  { key: 'projects', title: '项目申报', desc: '课题项目申报审核', resource: 'project-applications', status: 'submitted', link: '/admin/projects' }
]

const items = ref(SOURCES.map((s) => ({ ...s, count: 0, error: '' })))
const loading = ref(false)

const totalPending = computed(() => items.value.reduce((sum, i) => sum + (i.error ? 0 : i.count), 0))

const summaryHint = computed(() => {
  if (loading.value) return '正在统计各模块待办…'
  if (totalPending.value === 0) return '当前没有待处理事项'
  const busy = items.value.filter((i) => i.count > 0).map((i) => i.title)
  return '待处理模块：' + busy.join('、')
})

async function loadAll() {
  loading.value = true
  await Promise.all(
    items.value.map(async (item) => {
      try {
        const api = useAdminApi(item.resource)
        const res = await api.list({ status: item.status, page: 1, page_size: 1 })
        item.count = Number((res && res.total) || 0)
        item.error = ''
      } catch (e) {
        item.count = 0
        item.error = '统计失败'
      }
    })
  )
  loading.value = false
}

function go(item) {
  if (item.error) {
    loadAll()
    return
  }
  router.push(item.link)
}

onMounted(loadAll)
</script>

<style scoped>
.admin-page {
  padding: 20px;
}
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}
.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-1);
}
.page-sub {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--color-text-3);
}

/* 汇总条 */
.wb-summary {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 18px 22px;
  margin-bottom: 20px;
  border-radius: 12px;
  background: linear-gradient(135deg, #eef5fd 0%, #f7fbff 100%);
  border: 1px solid #d8e8f8;
}
.wb-summary--zero {
  background: linear-gradient(135deg, #f2fbf6 0%, #f8fdfa 100%);
  border-color: #cdeada;
}
.wb-summary-main {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-shrink: 0;
}
.wb-summary-num {
  font-size: 34px;
  font-weight: 700;
  line-height: 1;
  color: #0a66c2;
  font-variant-numeric: tabular-nums;
}
.wb-summary--zero .wb-summary-num {
  color: #0b6b41;
}
.wb-summary-unit {
  font-size: 14px;
  color: var(--color-text-2);
}
.wb-summary-hint {
  font-size: 13px;
  color: var(--color-text-3);
  line-height: 1.6;
}


/* 待办卡片 */
.wb-card {
  height: 100%;
  padding: 18px;
  border-radius: 12px;
  border: 1px solid var(--color-border-2);
  background: var(--color-bg-2);
  cursor: pointer;
  transition: transform 0.16s ease, box-shadow 0.16s ease, border-color 0.16s ease;
}
.wb-card:hover {
  transform: translateY(-2px);
  border-color: #a9cdf0;
  box-shadow: 0 8px 20px rgba(10, 102, 194, 0.12);
}
.wb-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.wb-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-1);
}
.wb-card-badge {
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  color: #b54708;
  background: #fff4e5;
  border: 1px solid #f5cd8a;
  flex-shrink: 0;
}
.wb-card-count {
  margin: 14px 0 12px;
  font-size: 30px;
  font-weight: 700;
  line-height: 1;
  color: #0a66c2;
  font-variant-numeric: tabular-nums;
}
.wb-card-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.wb-card-desc {
  font-size: 12px;
  color: var(--color-text-3);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.wb-card-go {
  font-size: 12px;
  color: #0a66c2;
  flex-shrink: 0;
}

/* 无待办：弱化 */
.wb-card--zero {
  background: var(--color-fill-1);
  border-color: var(--color-border-1);
}
.wb-card--zero .wb-card-count {
  color: var(--color-text-4);
}
.wb-card--zero .wb-card-go {
  color: var(--color-text-3);
}
.wb-card--zero:hover {
  border-color: var(--color-border-2);
  box-shadow: none;
  transform: none;
}

/* 统计失败 */
.wb-card--error .wb-card-count {
  color: #d92d20;
}
.wb-card--error .wb-card-desc {
  color: #d92d20;
}
</style>
