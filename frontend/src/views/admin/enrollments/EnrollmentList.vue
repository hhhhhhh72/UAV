<template>
  <div class="page">
    <!-- 搜索 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索姓名/电话..." allow-clear style="width: 200px" @press-enter="onSearchSubmit" @clear="onSearchSubmit" />
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>搜索</a-button>
          <a-button @click="resetParams">重置</a-button>
        </a-space>
      </a-form>
    </a-card>

    <!-- 数据表格 -->
    <a-card :bordered="false">
      <a-table
        :columns="columns"
        :data="listData"
        :loading="loading"
        row-key="id"
        :pagination="false"
      >
        <template #courseId="{ record }">
          <span>{{ record.course_id || '-' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="record.status === 'enrolled' ? 'green' : 'gray'" size="small">{{ record.status || '-' }}</a-tag>
        </template>
        <template #createdAt="{ record }">
          <span class="time-text">{{ formatDate(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
        </template>
        <template #empty>
          <a-empty description="暂无报名记录" />
        </template>
      </a-table>

      <div class="pagination-wrap" v-if="total > 0">
        <a-pagination
          v-model:current="filterParams.page"
          v-model:page-size="filterParams.page_size"
          :total="total"
          :page-size-options="[10, 20, 50]"
          show-total
          show-page-size
          @change="loadData"
        />
      </div>
    </a-card>

    <!-- 报名详情（含证件资料） -->
    <a-modal v-model:visible="detailVisible" title="报名详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="姓名" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="电话">{{ currentItem.phone || '-' }}</a-descriptions-item>
          <a-descriptions-item label="身份证号">{{ currentItem.id_card || '-' }}</a-descriptions-item>
          <a-descriptions-item label="性别">{{ currentItem.gender || '-' }}</a-descriptions-item>
          <a-descriptions-item label="生日">{{ currentItem.birthday || '-' }}</a-descriptions-item>
          <a-descriptions-item label="邮箱">{{ currentItem.email || '-' }}</a-descriptions-item>
          <a-descriptions-item label="学历">{{ currentItem.education || '-' }}</a-descriptions-item>
          <a-descriptions-item label="从业经验" :span="2">{{ currentItem.experience || '-' }}</a-descriptions-item>
          <a-descriptions-item label="课程ID" :span="2">{{ currentItem.course_id || '-' }}</a-descriptions-item>
          <a-descriptions-item label="证件照">
            <a-image v-if="currentItem.photo_url" :src="fullUrl(currentItem.photo_url)" :preview="true" width="64" height="64" fit="cover" style="border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="身份证照片">
            <a-image v-if="currentItem.id_card_image" :src="fullUrl(currentItem.id_card_image)" :preview="true" width="64" height="64" fit="cover" style="border-radius: 4px; cursor: pointer;" />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="无犯罪证明" :span="2">{{ currentItem.no_crime || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('enrollments')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = (n) => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}
const fullUrl = (u) => (u && u.startsWith('http') ? u : (import.meta.env.VITE_API_TARGET || 'http://localhost:8080') + (u || ''))

const { listData, loading, total, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: {},
})

const columns = [
  { title: '姓名', dataIndex: 'name', minWidth: 100 },
  { title: '电话', dataIndex: 'phone', width: 130 },
  { title: '身份证号', dataIndex: 'id_card', width: 190 },
  { title: '课程ID', dataIndex: 'course_id', slotName: 'courseId', minWidth: 180 },
  { title: '性别', dataIndex: 'gender', width: 70 },
  { title: '学历', dataIndex: 'education', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '报名时间', dataIndex: 'created_at', slotName: 'createdAt', width: 160 },
  { title: '操作', slotName: 'actions', width: 80, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
