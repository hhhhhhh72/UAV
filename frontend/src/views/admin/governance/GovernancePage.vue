<template>
  <div class="admin-page">
    <div class="page-header">
      <div class="page-header-main">
        <h2>审核与审计</h2>
        <span class="page-sub">{{ current.desc }}</span>
      </div>
      <a-select v-model="tab" class="module-select" size="small" @change="onChange">
        <a-option v-for="m in modules" :key="m.key" :value="m.key" :label="m.label">{{ m.label }}</a-option>
      </a-select>
    </div>

    <ReviewWorkbench v-if="tab === 'workbench'" />
    <AuditLogList v-else-if="canAudit" />
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ReviewWorkbench from '../workbench/ReviewWorkbench.vue'
import AuditLogList from '../audit/AuditLogList.vue'
import { useAuth } from '../composables/useAuth'

const { isPlatformAdmin } = useAuth()
const route = useRoute()
const router = useRouter()

/* 模块清单：操作审计仅平台管理员可见（与后端接口权限一致） */
const MODULES = [
  { key: 'workbench', label: '审核待办', desc: '各业务模块的待处理审核事项，点卡片进入对应管理页处理' },
  { key: 'audit', label: '操作审计', desc: '谁、在什么时候、对什么对象、做了什么（点行首箭头查看请求ID与元数据）', platformOnly: true }
]

const modules = computed(() => MODULES.filter((m) => !m.platformOnly || isPlatformAdmin.value))
const canAudit = computed(() => isPlatformAdmin.value)

const normalizeTab = (v) => (String(v || '') === 'audit' && isPlatformAdmin.value ? 'audit' : 'workbench')
const tab = ref(normalizeTab(route.query.tab))
const current = computed(() => modules.value.find((m) => m.key === tab.value) || modules.value[0])

/* 双向同步：URL ?tab= 与下拉选择一致，刷新/分享后仍停在同一个模块 */
watch(() => route.query.tab, (v) => {
  const t = normalizeTab(v)
  if (t !== tab.value) tab.value = t
})
watch(tab, (t) => {
  if (String(route.query.tab || '') !== t) router.replace({ path: '/admin/governance', query: { tab: t } })
})

const onChange = (t) => router.replace({ path: '/admin/governance', query: { tab: t } })
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
  min-width: 0;
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
.module-select {
  width: 160px;
  flex-shrink: 0;
}
</style>
