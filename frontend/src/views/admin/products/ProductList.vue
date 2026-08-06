<template>
  <div class="page">
    <!-- 搜索 + 新增 -->
    <a-card :bordered="false" class="search-card">
      <a-form layout="horizontal" :model="filterParams" class="search-form">
        <a-space wrap>
          <a-form-item label="类型" class="form-item">
            <a-select v-model="filterParams.prod_type" style="width: 140px" allow-clear @change="onSearchSubmit">
              <a-option label="全部类型" value="" />
              <a-option label="整机" value="drone" />
              <a-option label="配件" value="part" />
              <a-option label="维修服务" value="repair" />
              <a-option label="航拍服务" value="aerial" />
              <a-option label="试飞测试" value="test_fly" />
              <a-option label="检测标定" value="calibration" />
              <a-option label="空域协调" value="airspace" />
            </a-select>
          </a-form-item>
          <a-button type="primary" @click="onSearchSubmit"><template #icon><icon-search /></template>查询</a-button>
          <a-button type="primary" @click="showCreate"><template #icon><icon-plus /></template>新增商品</a-button>
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
        <template #cover="{ record }">
          <a-image
            v-if="Array.isArray(record.images) && record.images[0]"
            :src="record.images[0]"
            :preview-props="{ srcList: record.images }"
            width="56"
            height="56"
            fit="cover"
            class="cover-img"
          />
          <span v-else class="no-image">无图</span>
        </template>
        <template #prodType="{ record }">
          <span>{{ typeLabel(record.prod_type) }}</span>
        </template>
        <template #condition="{ record }">
          <span>{{ record.condition === 'used' ? '二手' : '全新' }}</span>
        </template>
        <template #price="{ record }">
          <span>{{ ((record.price_fen || 0) / 100).toFixed(2) }}</span>
        </template>
        <template #status="{ record }">
          <a-tag :color="record.status === 'listed' ? 'green' : 'gray'" size="small">{{ record.status === 'listed' ? '在售' : record.status }}</a-tag>
        </template>
        <template #actions="{ record }">
          <a-space :size="4">
            <a-button type="text" size="small" @click="showEdit(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="onDelete(record)">删除</a-button>
          </a-space>
        </template>
        <template #empty>
          <a-empty description="暂无商品" />
        </template>
      </a-table>
    </a-card>

    <!-- 新增 / 编辑弹窗 -->
    <a-modal v-model:visible="dialog.visible" :title="dialog.isEdit ? '编辑商品' : '新增商品'" :width="520">
      <a-form :model="form" layout="horizontal" class="dialog-form">
        <a-form-item label="商品名称" required>
          <a-input v-model="form.title" placeholder="如：工业级六旋翼无人机 X6-28L" allow-clear />
        </a-form-item>
        <a-form-item label="类型">
          <a-select v-model="form.prod_type" style="width: 100%">
            <a-option label="整机" value="drone" />
            <a-option label="配件" value="part" />
            <a-option label="维修服务" value="repair" />
            <a-option label="航拍服务" value="aerial" />
            <a-option label="试飞测试" value="test_fly" />
            <a-option label="检测标定" value="calibration" />
            <a-option label="空域协调" value="airspace" />
          </a-select>
        </a-form-item>
        <a-form-item label="品牌">
          <a-input v-model="form.brand" placeholder="可选" allow-clear />
        </a-form-item>
        <a-form-item label="型号">
          <a-input v-model="form.model" placeholder="可选" allow-clear />
        </a-form-item>
        <a-form-item label="成色">
          <a-select v-model="form.condition" style="width: 100%">
            <a-option label="全新" value="new" />
            <a-option label="二手" value="used" />
          </a-select>
        </a-form-item>
        <a-form-item label="价格(元)">
          <a-input v-model="form.priceYuan" type="number" placeholder="0.00" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option label="在售" value="listed" />
            <a-option label="下架" value="removed" />
          </a-select>
        </a-form-item>
        <a-form-item label="商品图片">
          <a-upload
            :file-list="imageList"
            list-type="picture-card"
            :limit="6"
            :custom-request="uploadImage"
            @change="onImageChange"
          />
        </a-form-item>
        <a-form-item label="卖家">
          <a-input v-model="form.seller_name" placeholder="默认平台自营" allow-clear />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model="form.description" type="textarea" :auto-size="{ minRows: 3, maxRows: 6 }" placeholder="商品说明" />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="dialog.visible = false">取消</a-button>
        <a-button type="primary" :loading="dialog.loading" @click="handleSubmit">保存</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { Message, Modal } from '@arco-design/web-vue'
import { useListRequest } from '@/hooks/useListRequest'
import { useAdminApi } from '@/api/admin/common'
import axios from '@/utils/http'

const api = useAdminApi('products')

const typeLabel = (t) => ({ drone: '整机', part: '配件', repair: '维修服务', aerial: '航拍服务', test_fly: '试飞测试', calibration: '检测标定', airspace: '空域协调' }[t] || t || '-')

const { listData, loading, filterParams, loadData, onSearchSubmit } = useListRequest({
  apiFunction: api.list,
  idKey: 'id',
  defaultParams: { prod_type: '' }
})

const columns = [
  { title: 'ID', dataIndex: 'id', width: 200 },
  { title: '图片', dataIndex: 'images', slotName: 'cover', width: 90 },
  { title: '商品名称', dataIndex: 'title', minWidth: 160 },
  { title: '类型', dataIndex: 'prod_type', slotName: 'prodType', width: 90 },
  { title: '品牌', dataIndex: 'brand', width: 100 },
  { title: '型号', dataIndex: 'model', width: 100 },
  { title: '成色', dataIndex: 'condition', slotName: 'condition', width: 80 },
  { title: '价格(元)', dataIndex: 'price_fen', slotName: 'price', width: 110, align: 'right' },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '卖家', dataIndex: 'seller_name', width: 120 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
]

const dialog = reactive({ visible: false, loading: false, isEdit: false, id: '' })
const form = reactive({ title: '', prod_type: 'drone', brand: '', model: '', condition: 'new', priceYuan: '', status: 'listed', description: '', seller_name: '', images: [] })
const imageList = reactive([])

const resetForm = () => {
  form.title = ''; form.prod_type = 'drone'; form.brand = ''; form.model = ''
  form.condition = 'new'; form.priceYuan = ''; form.status = 'listed'; form.description = ''
  form.seller_name = ''; form.images = []
  imageList.length = 0
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
  form.seller_name = row.seller_name || ''; form.images = row.images || []
  form.images.forEach(u => imageList.push({ name: u.split('/').pop(), url: u }))
  dialog.visible = true
}

// 图片上传（/api/v1/upload 返回相对 URL）
const uploadImage = async ({ file, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await axios.post('/api/v1/upload', fd)
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error('图片上传失败')
  }
}

// a-upload 列表变化（新增/移除）时同步 form.images
const onImageChange = (fileList) => {
  imageList.length = 0
  imageList.push(...fileList)
  form.images = fileList.map(f => f.url || f.response?.data?.url || f.response?.url).filter(Boolean)
}

const handleSubmit = async () => {
  if (!form.title) { Message.warning('请输入商品名称'); return }
  dialog.loading = true
  const payload = {
    title: form.title,
    prod_type: form.prod_type,
    brand: form.brand,
    model: form.model,
    condition: form.condition,
    price_fen: Math.round(parseFloat(form.priceYuan || 0) * 100),
    status: form.status,
    description: form.description,
    seller_name: form.seller_name,
    images: form.images
  }
  try {
    if (dialog.isEdit) await api.update(dialog.id, payload)
    else await api.create(payload)
    Message.success('保存成功')
    dialog.visible = false
    loadData()
  } catch (e) { Message.error(e?.response?.data?.message || '保存失败') }
  finally { dialog.loading = false }
}

const onDelete = (row) => {
  Modal.confirm({
    title: '删除商品',
    content: `确定删除商品「${row.title}」吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已删除')
        loadData()
      } catch (e) { Message.error('删除失败') }
    }
  })
}

onMounted(loadData)
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.search-card { margin-bottom: 16px; }

.search-form :deep(.arco-form-item) { margin-bottom: 0; }

.cover-img { border-radius: 6px; overflow: hidden; }
.no-image { color: #C9CDD4; font-size: 12px; }

.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }
</style>
