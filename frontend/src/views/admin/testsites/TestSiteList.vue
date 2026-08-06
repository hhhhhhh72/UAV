<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="关键词" class="form-item">
            <a-input v-model="filterParams.keyword" placeholder="搜索场地名称..." allow-clear style="width: 220px" @press-enter="onSearchSubmit">
              <template #prefix><icon-search /></template>
            </a-input>
          </a-form-item>
          <a-form-item label="场地类型" class="form-item">
            <a-select v-model="filterParams.site_type" placeholder="场地类型" style="width: 150px" @change="onSearchSubmit">
              <a-option value="">全部</a-option>
              <a-option value="flying_field">飞行场地</a-option>
              <a-option value="lab">实验室</a-option>
              <a-option value="indoor">室内场地</a-option>
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button @click="resetParams">重置</a-button>
          <a-button type="primary" status="success" style="margin-left: auto" @click="handleAdd">新增</a-button>
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
        @sorter-change="handleSortChange"
      >
        <template #name="{ record }">
          <span class="cell-name">{{ record.name || '-' }}</span>
        </template>
        <template #type="{ record }">
          <a-tag :color="typeTag(record.site_type)" size="small">{{ typeLabel(record.site_type) }}</a-tag>
        </template>
        <template #price="{ record }">
          <span class="cell-amount">{{ formatMoney(record.price_fen) }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="statusTag(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showDetail(record)">详情</a-button>
            <a-button type="text" size="small" @click="handleEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无场地数据" />
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
    <a-modal v-model:visible="detailVisible" title="场地详情" :width="640" :footer="false" :mask-closable="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="场地名称" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="场地类型">
            <a-tag :color="typeTag(currentItem.site_type)" size="small">{{ typeLabel(currentItem.site_type) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="地区">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="费用">{{ formatMoney(currentItem.price_fen) }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel(currentItem.status) }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="设施">{{ (currentItem.facilities || []).join('、') || '-' }}</a-descriptions-item>
          <a-descriptions-item label="配套设施" :span="2">{{ currentItem.facilities || '-' }}</a-descriptions-item>
          <a-descriptions-item label="使用规则" :span="2">{{ currentItem.booking_rule || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 表单弹窗（新增/编辑） -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑场地' : '新增场地'" :width="560" :mask-closable="false" :unmount-on-close="true" @close="resetForm">
      <a-form :model="form" layout="horizontal">
        <a-form-item label="场地名称"><a-input v-model="form.name" /></a-form-item>
        <a-form-item label="地点"><a-input v-model="form.location" /></a-form-item>
        <a-form-item label="类型">
          <a-select v-model="form.site_type" style="width: 100%">
            <a-option value="flying_field">飞行场地</a-option>
            <a-option value="lab">实验室</a-option>
            <a-option value="anechoic_chamber">消声室</a-option>
            <a-option value="wind_tunnel">风洞</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="available">可用</a-option>
            <a-option value="maintenance">维护中</a-option>
            <a-option value="reserved">已预约</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="3" /></a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">确定</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('test-sites')

const typeLabel = (t) => ({ flying_field: '飞行场地', lab: '实验室', indoor: '室内场地' }[t] || t || '-')
const typeTag = (t) => ({ flying_field: 'green', lab: 'orange', indoor: 'gray' }[t] || 'gray')

const statusLabel = (s) => ({ available: '可用', maintenance: '维护中', closed: '已关闭' }[s] || s || '-')
const statusTag = (s) => ({ available: 'green', maintenance: 'orange', closed: 'red' }[s] || 'gray')

const formatMoney = (fen) => {
  if (fen == null) return '-'
  const yuan = Number(fen) / 100
  return '¥' + yuan.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, resetParams } = useListRequest({
  apiFunction: api.list, idKey: 'id', defaultParams: { site_type: '' }
})

// a-table 行选择（兼容 useListRequest 的 selectedIds）
const rowSelection = computed(() => ({
  type: 'checkbox',
  showCheckedAll: true,
  selectedRowKeys: selectedIds.value,
  onChange: (keys) => { selectedIds.value = [...keys] }
}))

// Arco sorter-change → useListRequest.onSortChange（el-table 的 { prop, order } 形态）
const handleSortChange = (dataIndex, direction) => {
  onSortChange({
    prop: dataIndex,
    order: direction === 'ascend' ? 'ascending' : direction === 'descend' ? 'descending' : ''
  })
}

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160, sortable: { sortDirections: ['ascend', 'descend'] } },
  { title: '场地名称', dataIndex: 'name', slotName: 'name', minWidth: 180 },
  { title: '类型', dataIndex: 'site_type', slotName: 'type', width: 120 },
  { title: '地区', dataIndex: 'location', minWidth: 140 },
  { title: '费用', dataIndex: 'price_fen', slotName: 'price', width: 120, sortable: { sortDirections: ['ascend', 'descend'] } },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' },
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', name: '', site_type: 'flying_field', location: '', price_fen: 0, facilities: '', booking_rule: '', status: 'available' })
const resetForm = () => Object.assign(form, { id: '', name: '', site_type: 'flying_field', location: '', price_fen: 0, facilities: '', booking_rule: '', status: 'available' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (r) => { Object.assign(form, { ...r, price_fen: r.price_fen || 0 }); formEdit.value = true; formVisible.value = true }

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入场地名称'); return }
  formLoading.value = true
  try {
    const p = { ...form }
    formEdit.value ? await api.update(form.id, p) : await api.create(p)
    Message.success(formEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    loadData()
  } catch (e) { Message.error(errMsg(e)) }
  finally { formLoading.value = false }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '提示',
    content: `确定删除场地 "${row.name}" 吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try { await api.delete(row.id); Message.success('已删除'); loadData() } catch { Message.error('删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.cell-name { font-weight: 500; color: var(--color-text-1); }

.cell-amount { font-weight: 600; color: #E96012; }

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
</style>
