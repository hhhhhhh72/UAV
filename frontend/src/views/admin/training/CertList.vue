<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索持有人或证书编号..." allow-clear style="width: 260px" @press-enter="onSearchSubmit" @clear="onSearchSubmit" />
          </a-form-item>
          <a-form-item label="状态" class="form-item">
            <a-select v-model="filterParams.status" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="valid">有效</a-option>
              <a-option value="expired">已过期</a-option>
              <a-option value="revoked">已吊销</a-option>
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
        <template #certNo="{ record }">
          <span class="cell-mono">{{ record.cert_no || '-' }}</span>
        </template>
        <template #issueDate="{ record }">
          <span class="time-text">{{ formatDate(record.issue_date) }}</span>
        </template>
        <template #expireDate="{ record }">
          <span class="time-text">{{ formatDate(record.expire_date) }}</span>
        </template>
        <template #image="{ record }">
          <a-image
            v-if="record.image_url"
            :src="fullUrl(record.image_url)"
            :preview="true"
            width="44"
            height="44"
            fit="cover"
            style="border-radius: 4px; cursor: pointer;"
          />
          <span v-else>-</span>
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
          <a-empty description="暂无证书数据" />
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
    <a-modal v-model:visible="detailVisible" title="证书详情" :width="600" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="证书编号">{{ currentItem.cert_no || '-' }}</a-descriptions-item>
          <a-descriptions-item label="证书类型">{{ currentItem.cert_type || '-' }}</a-descriptions-item>
          <a-descriptions-item label="持有人">{{ currentItem.holder_name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="签发日期">{{ formatDate(currentItem.issue_date) }}</a-descriptions-item>
          <a-descriptions-item label="有效期至">{{ formatDate(currentItem.expire_date) }}</a-descriptions-item>
          <a-descriptions-item label="发证机构" :span="2">{{ currentItem.issuer || '-' }}</a-descriptions-item>
          <a-descriptions-item label="备注" :span="2">{{ currentItem.remark || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑证书' : '新增证书'" :width="560" @ok="submitForm" @cancel="formVisible = false">
      <a-form :model="form" :label-col-flex="90">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="证书编号" required><a-input v-model="form.cert_number" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="证书类型"><a-input v-model="form.cert_type" placeholder="caac / utc_dji / gov_level" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="发证机构"><a-input v-model="form.issuer_org" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="签发日期">
              <a-date-picker v-model="form.issue_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="有效期至">
              <a-date-picker v-model="form.expire_date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="pending">待审核</a-option>
            <a-option value="approved">已通过</a-option>
            <a-option value="rejected">已驳回</a-option>
          </a-select>
        </a-form-item>
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

const api = useAdminApi('certificates')

// 相对路径图片补全后端地址（vite/nginx 已代理 /uploads）
const fullUrl = (u) => (u && u.startsWith('http') ? u : (import.meta.env.VITE_API_TARGET || 'http://localhost:8080') + (u || ''))

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ 'valid': 'green', 'expired': 'orangered', 'revoked': 'red' }[s] || 'gray')
const statusLabel = { 'valid': '有效', 'expired': '已过期', 'revoked': '已吊销' }

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
  { title: '证书编号', dataIndex: 'cert_number', slotName: 'certNo', width: 180 },
  { title: '持有人ID', dataIndex: 'user_id', minWidth: 140 },
  { title: '证书类型', dataIndex: 'cert_type', width: 120 },
  { title: '签发日期', dataIndex: 'issue_date', slotName: 'issueDate', width: 120 },
  { title: '有效期至', dataIndex: 'expire_date', slotName: 'expireDate', width: 120 },
  { title: '等级', dataIndex: 'level', width: 80 },
  { title: '发证机构', dataIndex: 'issuer_org', minWidth: 140 },
  { title: '证书图片', dataIndex: 'image_url', slotName: 'image', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)

const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false); const formEdit = ref(false); const formLoading = ref(false)
const form = reactive({ id: '', cert_number: '', cert_type: '', issuer_org: '', issue_date: '', expire_date: '', status: 'pending' })
const resetForm = () => Object.assign(form, { id: '', cert_number: '', cert_type: '', issuer_org: '', issue_date: '', expire_date: '', status: 'pending' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, r); formEdit.value = true; formVisible.value = true }
const submitForm = async () => {
  if (!form.cert_number) { Message.warning('请输入证书编号'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (r) => {
  Modal.confirm({
    title: '删除证书',
    content: '确定删除该证书吗？删除后不可恢复',
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

.cell-mono { font-family: 'Courier New', monospace; font-size: 13px; }
.time-text { color: #86909C; font-size: 12px; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
