<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="products"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增商品"
      @add="openForm()"
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
        <a-tag :color="statusColor(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <template v-if="record.status === 'pending'">
            <a-button type="text" status="success" size="small" @click="handleApprove(record, 'listed')">通过</a-button>
            <a-button type="text" status="danger" size="small" @click="handleApprove(record, 'removed')">驳回</a-button>
          </template>
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无商品" />
      </template>
    </CrudList>

    <!-- 新增 / 编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑商品' : '新增商品'" :width="520" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical" class="dialog-form">
        <a-form-item label="商品名称" required>
          <a-input v-model="form.title" placeholder="如：工业级六旋翼无人机 X6-28L" allow-clear style="width: 100%" />
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
          <a-input v-model="form.brand" placeholder="可选" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="型号">
          <a-input v-model="form.model" placeholder="可选" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="成色">
          <a-select v-model="form.condition" style="width: 100%">
            <a-option label="全新" value="new" />
            <a-option label="二手" value="used" />
          </a-select>
        </a-form-item>
        <a-form-item label="价格(元)">
          <a-input v-model="form.priceYuan" type="number" placeholder="0.00" style="width: 100%" />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option label="待审核" value="pending" />
            <a-option label="在售" value="listed" />
            <a-option label="已售" value="sold" />
            <a-option label="已下架" value="removed" />
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
          <a-input v-model="form.seller_name" placeholder="默认平台自营" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model="form.description" type="textarea" :auto-size="{ minRows: 3, maxRows: 6 }" placeholder="商品说明" style="width: 100%" />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="formVisible = false">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">保存</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { useAdminApi } from '@/api/admin/common'
import axios from '@/utils/http'
import CrudList from '../components/CrudList.vue'

const crudRef = ref()
const api = useAdminApi('products')

const typeLabel = (t) => ({ drone: '整机', part: '配件', repair: '维修服务', aerial: '航拍服务', test_fly: '试飞测试', calibration: '检测标定', airspace: '空域协调' }[t] || t || '-')
// 商品状态：pending=待审核（用户发布，通过后才上架）/ listed=在售 / sold=已售 / removed=已下架
const statusLabel = (s) => ({ pending: '待审核', listed: '在售', sold: '已售', removed: '已下架' }[s] || s || '-')
const statusColor = (s) => ({ pending: 'orange', listed: 'green', sold: 'gray', removed: 'gray' }[s] || 'gray')

// 批量动作：批量上架 / 批量下架——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'list', label: '批量上架', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'listed' }) },
  { key: 'remove', label: '批量下架', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'removed' }) }
]

const searchFields = [
  { key: 'status', label: '状态', type: 'select', width: 120, options: [
    { value: '', label: '全部状态' },
    { value: 'pending', label: '待审核' },
    { value: 'listed', label: '在售' },
    { value: 'sold', label: '已售' },
    { value: 'removed', label: '已下架' }
  ]},
  { key: 'prod_type', label: '类型', type: 'select', width: 140, options: [
    { value: '', label: '全部类型' },
    { value: 'drone', label: '整机' },
    { value: 'part', label: '配件' },
    { value: 'repair', label: '维修服务' },
    { value: 'aerial', label: '航拍服务' },
    { value: 'test_fly', label: '试飞测试' },
    { value: 'calibration', label: '检测标定' },
    { value: 'airspace', label: '空域协调' }
  ]}
]

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

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', prod_type: 'drone', brand: '', model: '', condition: 'new', priceYuan: '', status: 'listed', description: '', seller_name: '', images: [] })
const imageList = reactive([])

const resetForm = () => {
  form.id = ''; form.title = ''; form.prod_type = 'drone'; form.brand = ''; form.model = ''
  form.condition = 'new'; form.priceYuan = ''; form.status = 'listed'; form.description = ''
  form.seller_name = ''; form.images = []
  imageList.length = 0
}

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    form.id = row.id
    form.title = row.title || ''; form.prod_type = row.prod_type || 'drone'
    form.brand = row.brand || ''; form.model = row.model || ''
    form.condition = row.condition || 'new'
    form.priceYuan = ((row.price_fen || 0) / 100).toString()
    form.status = row.status || 'listed'; form.description = row.description || ''
    form.seller_name = row.seller_name || ''; form.images = row.images || []
    form.images.forEach(u => imageList.push({ name: u.split('/').pop(), url: u }))
  } else {
    formEdit.value = false
  }
  formVisible.value = true
}

// 图片上传（/api/v1/upload 返回相对 URL）
// 注意：Arco custom-request 的参数是 fileItem，原生 File 在 fileItem.file 上
const uploadImage = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
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
// 注意：f.url 是 Arco 的本地 blob 预览地址，不能入库；真实地址在响应 data.url 里
const onImageChange = (fileList) => {
  imageList.length = 0
  imageList.push(...fileList)
  form.images = fileList.map(f => f.response?.data?.url || f.response?.url || f.url).filter(Boolean)
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入商品名称'); return }
  formLoading.value = true
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
    if (formEdit.value) await api.update(form.id, payload)
    else await api.create(payload)
    Message.success('保存成功')
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '保存失败') }
  finally { formLoading.value = false }
}

// 审核快捷操作：通过（pending→listed 上架）/ 驳回（pending→removed 下架）——传完整行避免清空其他字段
const handleApprove = async (row, status) => {
  try {
    await api.update(row.id, { ...row, status })
    Message.success(status === 'listed' ? '已通过，商品已上架' : '已驳回，商品已下架')
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '操作失败') }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除商品',
    content: `确定删除商品「${row.title}」吗？`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error('删除失败') }
    }
  })
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }

.cover-img { border-radius: 6px; overflow: hidden; }
.no-image { color: #C9CDD4; font-size: 12px; }

.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }
</style>
