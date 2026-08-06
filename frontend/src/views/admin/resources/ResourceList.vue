<template>
  <div class="list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-input v-model="filterParams.keyword" placeholder="搜索资源名称..." clearable style="width: 220px" @keyup.enter="onSearchSubmit" @clear="onSearchSubmit">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="filterParams.res_type" clearable placeholder="资源类型" style="width: 160px" @change="onSearchSubmit">
          <el-option label="全部" value="" />
          <el-option label="无人机" value="drone" />
          <el-option label="机场" value="airport" />
          <el-option label="试飞场地" value="test_site" />
          <el-option label="测试基地" value="test_base" />
          <el-option label="其他" value="other" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">搜索</el-button>
        <el-button @click="resetParams">重置</el-button>
        <div style="margin-left:auto"><el-button type="success" @click="handleAdd">新增资源</el-button></div>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border @selection-change="onSelectChange" @sort-change="onSortChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="160" sortable="custom" />
        <el-table-column prop="name" label="资源名称" min-width="180" />
        <el-table-column prop="res_type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="resTypeColor[row.res_type] || 'info'">{{ resTypeLabel[row.res_type] || row.res_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="model" label="型号/规格" width="160" />
        <el-table-column prop="location" label="地区" width="140" />
        <el-table-column label="费用" width="120">
          <template #default="{ row }">&yen;{{ formatPriceFen(row.price_fen) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusColor[row.status] || 'info'">{{ statusLabel[row.status] || row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="可见级别" width="130">
          <template #default="{ row }">
            <el-tag :type="visColor[row.visibility_level] || 'info'" size="small">{{ visLabel[row.visibility_level] || '公开' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无数据" /></template>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="total > 0">
      <el-pagination v-model:current-page="filterParams.page" v-model:page-size="filterParams.page_size" :page-sizes="[10,20,50]" :total="total" layout="total,sizes,prev,pager,next,jumper" background @size-change="loadData" @current-change="loadData" />
    </div>

    <el-dialog v-model="detailVisible" title="资源详情" width="640px" destroy-on-close>
      <template v-if="currentItem">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID" :span="2">{{ currentItem.id }}</el-descriptions-item>
          <el-descriptions-item label="资源名称">{{ currentItem.name }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="resTypeColor[currentItem.res_type] || 'info'" size="small">{{ resTypeLabel[currentItem.res_type] || currentItem.res_type }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="型号/规格">{{ currentItem.model || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地区">{{ currentItem.location || '-' }}</el-descriptions-item>
          <el-descriptions-item label="可见级别">{{ currentItem.visibility_level || '-' }}</el-descriptions-item>
          <el-descriptions-item label="费用">&yen;{{ formatPriceFen(currentItem.price_fen) }}</el-descriptions-item>
          <el-descriptions-item label="预约信息">{{ currentItem.booking_info || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusColor[currentItem.status] || 'info'" size="small">{{ statusLabel[currentItem.status] || currentItem.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="预约方式" :span="2">{{ currentItem.booking_info || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>

    <!-- Add/Edit Form Dialog -->
    <el-dialog v-model="formVisible" :title="formEdit ? '编辑资源' : '新增资源'" width="560px" destroy-on-close>
      <el-form :model="form" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="资源名称" required><el-input v-model="form.name" placeholder="输入名称" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="资源类型" required>
            <el-select v-model="form.res_type" style="width:100%">
              <el-option label="无人机" value="drone" /><el-option label="机场" value="airport" />
              <el-option label="试飞场地" value="test_site" /><el-option label="测试基地" value="test_base" />
              <el-option label="其他" value="other" />
            </el-select>
          </el-form-item></el-col>
          <el-col :span="12"><el-form-item label="型号规格"><el-input v-model="form.model" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="地区"><el-input v-model="form.location" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="费用(元)"><el-input-number v-model="form.priceYuan" :min="0" style="width:100%" :controls="false" placeholder="单位：元" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="可见级别">
            <el-select v-model="form.visibility_level" style="width:100%">
              <el-option label="公开（政府访客可见）" value="public" />
              <el-option label="会员可见" value="member" />
              <el-option label="副会长单位可见" value="partner" />
              <el-option label="仅协会管理员" value="admin" />
            </el-select>
          </el-form-item></el-col>
          <el-col :span="12"><el-form-item label="状态">
            <el-select v-model="form.status" style="width:100%">
              <el-option label="可用" value="available" /><el-option label="使用中" value="in_use" /><el-option label="维护中" value="maintenance" />
            </el-select>
          </el-form-item></el-col>
          <el-col :span="24"><el-form-item label="预约方式"><el-input v-model="form.booking_info" placeholder="如：工作日 9-18 点，需提前 3 天" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleFormSubmit" :loading="formLoading">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'

const api = useAdminApi('industry-resources')

const resTypeLabel = { drone: '无人机', airport: '机场', test_site: '试飞场地', flying_field: '试飞场地', test_base: '测试基地', other: '其他' }
const resTypeColor = { drone: 'success', airport: 'warning', test_site: 'warning', flying_field: 'info', test_base: '', other: '' }
const statusLabel = { available: '可用', in_use: '使用中', maintenance: '维护中' }
const visLabel = { public: '公开', member: '会员', partner: '副会长单位', admin: '仅管理员' }
const visColor = { public: 'info', member: 'primary', partner: 'warning', admin: 'danger' }
const statusColor = { available: 'success', in_use: 'warning', maintenance: 'danger' }

const formatPriceFen = (fen) => {
  if (fen == null) return '-'
  const yuan = fen / 100
  return yuan.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onSelectChange, resetParams } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { res_type: '' }
})

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

// ---- Add/Edit form ----
const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id:'', name:'', res_type:'', model:'', location:'', priceYuan:null, status:'available', booking_info:'', visibility_level:'public' })
const resetForm = () => Object.assign(form, { id:'', name:'', res_type:'', model:'', location:'', priceYuan:null, status:'available', booking_info:'', visibility_level:'public' })
const handleAdd = () => { resetForm(); formEdit.value = false; formVisible.value = true }
const handleEdit = (row) => {
  Object.assign(form, { ...row, priceYuan: row.price_fen ? Math.round(row.price_fen / 100 * 100) / 100 : null, quantity: row.quantity || 0 })
  formEdit.value = true; formVisible.value = true
}
const handleFormSubmit = async () => {
  if (!form.name) { ElMessage.warning('请输入资源名称'); return }
  formLoading.value = true
  try {
    const payload = { ...form }
    payload.price_fen = Math.round((form.priceYuan || 0) * 100)
    delete payload.priceYuan
    if (formEdit.value) {
      await api.update(form.id, payload); ElMessage.success('更新成功')
    } else {
      await api.create(payload); ElMessage.success('创建成功')
    }
    formVisible.value = false; loadData()
  } catch (e) { ElMessage.error(e?.response?.data?.message || '操作失败') } finally { formLoading.value = false }
}
const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除「${row.name}」吗？`, '提示', { type: 'warning' }).then(async () => {
    try { await api.delete(row.id); ElMessage.success('已删除'); loadData() } catch { ElMessage.error('删除失败') }
  }).catch(() => {})
}

onMounted(loadData)
</script>

<style scoped>
.list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
.pagination-wrap { display: flex; justify-content: flex-end; margin-top: 16px; background: #fff; border-radius: 8px; padding: 16px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
@media (max-width: 767px) { .search-bar { padding: 12px; } .search-row { flex-direction: column; align-items: stretch; } .table-wrap { overflow-x: auto; } }
</style>
