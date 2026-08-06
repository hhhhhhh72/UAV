<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索课程标题..." allow-clear style="width: 220px" @press-enter="onSearchSubmit" @clear="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="draft">草稿</a-option>
              <a-option value="published">已发布</a-option>
              <a-option value="closed">已关闭</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>搜索</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="success" @click="handleAdd">新增</a-button>
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
        :row-selection="rowSelection"
      >
        <template #title="{ record }">
          <span class="cell-title">{{ record.title || '-' }}</span>
        </template>
        <template #price="{ record }">
          <span>{{ record.price_fen ? '¥' + (record.price_fen / 100).toLocaleString() : '-' }}</span>
        </template>
        <template #maxStudents="{ record }">
          <span>{{ record.max_students ?? '-' }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel[record.status] || record.status || '-' }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无课程数据" />
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

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="课程详情" :width="640" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="课程名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="分类">{{ currentItem.category || '-' }}</a-descriptions-item>
          <a-descriptions-item label="价格">{{ currentItem.price_fen ? '¥' + (currentItem.price_fen / 100).toLocaleString() : '-' }}</a-descriptions-item>
          <a-descriptions-item label="名额">{{ currentItem.max_students ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="已报名">{{ currentItem.enrolled_count ?? 0 }} 人</a-descriptions-item>
          <a-descriptions-item label="开始日期">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束日期">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑课程' : '新增课程'" :width="560" @ok="submitForm" @cancel="formVisible = false">
      <a-form :model="form" :label-col-flex="90">
        <a-row :gutter="16">
          <a-col :span="16">
            <a-form-item label="课程名称" required><a-input v-model="form.title" /></a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="证书类型">
              <a-select v-model="form.cert_type" style="width: 100%">
                <a-option value="caac">CAAC 执照</a-option>
                <a-option value="utc_dji">大疆 UTC</a-option>
                <a-option value="gov_level">人社等级</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="上课地点"><a-input v-model="form.location" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="状态">
              <a-select v-model="form.status" style="width: 100%">
                <a-option value="draft">草稿</a-option>
                <a-option value="published">已发布</a-option>
                <a-option value="closed">已关闭</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="价格(元)">
              <a-input-number v-model="form.priceYuan" :min="0" hide-button style="width: 100%" placeholder="单位：元" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="名额">
              <a-input-number v-model="form.max_students" :min="0" hide-button style="width: 100%" placeholder="招生人数" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="开始日期">
              <a-date-picker v-model="form.start_date" value-format="YYYY-MM-DD" placeholder="选择开班日期" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="结束日期">
              <a-date-picker v-model="form.end_date" value-format="YYYY-MM-DD" placeholder="选择结课日期" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="课程描述"><a-input v-model="form.description" type="textarea" :rows="3" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">提交</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('training-courses')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ 'published': 'green', 'draft': 'orangered', 'closed': 'gray' }[s] || 'gray')
const statusLabel = { published:'已发布', draft:'草稿', closed:'已关闭' }

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { status: '' }
})

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '课程名称', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '分类', dataIndex: 'category', width: 120 },
  { title: '价格', dataIndex: 'price_fen', slotName: 'price', width: 100, align: 'right' },
  { title: '名额', dataIndex: 'max_students', slotName: 'maxStudents', width: 80, align: 'right' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false); const formEdit = ref(false); const formLoading = ref(false)
const form = reactive({ id: '', title: '', cert_type: 'caac', priceYuan: null, max_students: null, location: '', start_date: '', end_date: '', status: 'draft', description: '' })
const resetForm = () => Object.assign(form, { id: '', title: '', cert_type: 'caac', priceYuan: null, max_students: null, location: '', start_date: '', end_date: '', status: 'draft', description: '' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => {
  Object.assign(form, {
    ...r,
    priceYuan: r.price_fen ? Math.round(r.price_fen / 100 * 100) / 100 : null,
    max_students: r.max_students ?? null
  })
  formEdit.value = true
  formVisible.value = true
}
const submitForm = async () => {
  if (!form.title) { Message.warning('请输入课程名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    p.price_fen = Math.round((form.priceYuan || 0) * 100)
    p.max_students = p.max_students ?? 0
    delete p.priceYuan
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除课程',
    content: '确定删除该课程吗？删除后不可恢复',
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(r.id)
        Message.success('已删除')
        loadData()
      } catch { Message.error('删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
