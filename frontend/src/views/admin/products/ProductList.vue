<template>
  <div class="product-list-page">
    <div class="search-bar">
      <div class="search-row">
        <el-select v-model="filterParams.prod_type" clearable style="width: 140px" @change="onSearchSubmit">
          <el-option label="全部类型" value="" />
          <el-option label="整机" value="drone" />
          <el-option label="配件" value="part" />
          <el-option label="维修服务" value="repair" />
        </el-select>
        <el-button type="primary" :icon="Search" @click="onSearchSubmit">查询</el-button>
        <el-button type="success" @click="showCreate">新增商品</el-button>
      </div>
    </div>

    <div class="table-wrap">
      <el-table v-loading="loading" :data="listData" row-key="id" stripe border>
        <el-table-column prop="id" label="ID" width="200" />
        <el-table-column prop="title" label="商品名称" min-width="160" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">{{ typeLabel(row.prod_type) }}</template>
        </el-table-column>
        <el-table-column prop="brand" label="品牌" width="100" />
        <el-table-column prop="model" label="型号" width="100" />
        <el-table-column label="成色" width="80">
          <template #default="{ row }">{{ row.condition === 'used' ? '二手' : '全新' }}</template>
        </el-table-column>
        <el-table-column label="价格(元)" width="110">
          <template #default="{ row }">{{ ((row.price_fen || 0) / 100).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'listed' ? 'success' : 'info'" size="small">{{ row.status === 'listed' ? '在售' : row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="seller_name" label="卖家" width="120" />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="showEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty><el-empty description="暂无商品" /></template>
      </el-table>
    </div>

    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑商品' : '新增商品'" width="520px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="商品名称" required><el-input v-model="form.title" placeholder="如：工业级六旋翼无人机 X6-28L" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.prod_type" style="width:100%">
            <el-option label="整机" value="drone" />
            <el-option label="配件" value="part" />
            <el-option label="维修服务" value="repair" />
          </el-select>
        </el-form-item>
        <el-form-item label="品牌"><el-input v-model="form.brand" placeholder="可选" /></el-form-item>
        <el-form-item label="型号"><el-input v-model="form.model" placeholder="可选" /></el-form-item>
        <el-form-item label="成色">
          <el-select v-model="form.condition" style="width:100%">
            <el-option label="全新" value="new" />
            <el-option label="二手" value="used" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格(元)"><el-input v-model="form.priceYuan" type="number" placeholder="0.00" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="在售" value="listed" />
            <el-option label="下架" value="removed" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" placeholder="商品说明" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible=false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="dialog.loading">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'
import axios from '@/utils/http'

const api = useAdminApi('products')

const typeLabel = (t) => ({ drone: '整机', part: '配件', repair: '维修服务' }[t] || t || '-')

const { listData, loading, filterParams, loadData, onSearchSubmit } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { prod_type: '' }
})

const dialog = reactive({ visible: false, loading: false, isEdit: false, id: '' })
const form = reactive({ title: '', prod_type: 'drone', brand: '', model: '', condition: 'new', priceYuan: '', status: 'listed', description: '' })

const resetForm = () => {
  form.title = ''; form.prod_type = 'drone'; form.brand = ''; form.model = ''
  form.condition = 'new'; form.priceYuan = ''; form.status = 'listed'; form.description = ''
}
const showCreate = () => { resetForm(); dialog.isEdit = false; dialog.visible = true }
const showEdit = (row) => {
  resetForm()
  dialog.isEdit = true; dialog.id = row.id
  form.title = row.title || ''; form.prod_type = row.prod_type || 'drone'
  form.brand = row.brand || ''; form.model = row.model || ''
  form.condition = row.condition || 'new'
  form.priceYuan = ((row.price_fen || 0) / 100).toString()
  form.status = row.status || 'listed'; form.description = row.description || ''
  dialog.visible = true
}

const handleSubmit = async () => {
  if (!form.title) { ElMessage.warning('请输入商品名称'); return }
  dialog.loading = true
  const payload = {
    title: form.title,
    prod_type: form.prod_type,
    brand: form.brand,
    model: form.model,
    condition: form.condition,
    price_fen: Math.round(parseFloat(form.priceYuan || 0) * 100),
    status: form.status,
    description: form.description
  }
  try {
    if (dialog.isEdit) await api.update(dialog.id, payload)
    else await api.create(payload)
    ElMessage.success('保存成功')
    dialog.visible = false
    loadData()
  } catch (e) { ElMessage.error(e?.response?.data?.message || '保存失败') }
  finally { dialog.loading = false }
}

const onDelete = async (row) => {
  try { await ElMessageBox.confirm(`确定删除商品「${row.title}」吗？`, '删除商品', { type: 'warning' }) }
  catch (e) { return }
  try {
    await api.delete(row.id)
    ElMessage.success('已删除')
    loadData()
  } catch (e) { ElMessage.error('删除失败') }
}

onMounted(loadData)
</script>

<style scoped>
.product-list-page { max-width: 1400px; margin: 0 auto; }
.search-bar { background: #fff; border-radius: 8px; padding: 16px 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); }
.search-row { display: flex; align-items: center; gap: 12px; }
.table-wrap { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.06); overflow: hidden; }
</style>
