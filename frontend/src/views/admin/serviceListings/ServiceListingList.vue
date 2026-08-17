<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="service-listings"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增服务能力"
      @add="openForm()"
    >
      <template #cover="{ record }">
        <a-image
          v-if="record.image"
          :src="record.image"
          width="56"
          height="56"
          fit="cover"
          class="cover-img"
        />
        <span v-else class="no-image">无图</span>
      </template>
      <template #price="{ record }">
        <span>{{ priceLabel(record.price_fen) }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="record.status === 'published' ? 'green' : 'gray'" size="small">{{ record.status === 'published' ? '已上架' : '已下架' }}</a-tag>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无服务能力" />
      </template>
    </CrudList>

    <!-- 新增 / 编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑服务能力' : '新增服务能力'" :width="520" @cancel="formVisible = false">
      <a-form :model="form" layout="vertical" class="dialog-form">
        <a-form-item label="服务标题" required>
          <a-input v-model="form.title" placeholder="如：桥梁与光伏设施精细化巡检" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="服务商名称" required>
          <a-input v-model="form.provider_name" placeholder="如：重庆翼航科技有限公司" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="服务分类">
          <a-select v-model="form.category" style="width: 100%">
            <a-option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="服务区域">
          <a-input v-model="form.region" placeholder="如：重庆 · 渝北区 / 服务重庆及周边" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="报价(元)">
          <a-input v-model="form.priceYuan" type="number" placeholder="0 表示面议" style="width: 100%" />
        </a-form-item>
        <a-form-item label="报价单位">
          <a-input v-model="form.unit" placeholder="如：次、项、平方公里、亩、公里" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="封面图">
          <a-upload
            :file-list="imageList"
            list-type="picture-card"
            :limit="1"
            :custom-request="uploadImage"
            @change="onImageChange"
          />
        </a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option label="上架" value="published" />
            <a-option label="下架" value="offline" />
          </a-select>
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model="form.description" type="textarea" :auto-size="{ minRows: 3, maxRows: 6 }" placeholder="服务能力说明" style="width: 100%" />
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
const api = useAdminApi('service-listings')

/* 服务分类值域（与小程序供需大厅 hallData.js 对齐） */
const CATEGORIES = ['巡检', '航拍', '测绘', '应急', '植保']

const priceLabel = (fen) => (!fen ? '面议' : (fen / 100).toFixed(2))

// 批量动作：批量上架 / 批量下架——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '批量上架', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) },
  { key: 'offline', label: '批量下架', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'offline' }) }
]

// 搜索过滤：keyword 命中标题/服务商/描述，category 精确匹配
const searchFields = [
  { key: 'keyword', label: '关键词', type: 'input', width: 220, placeholder: '标题 / 服务商 / 描述' },
  { key: 'category', label: '分类', type: 'select', width: 140, options: [
    { value: '', label: '全部分类' },
    ...CATEGORIES.map(c => ({ value: c, label: c }))
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 200 },
  { title: '封面', dataIndex: 'image', slotName: 'cover', width: 90 },
  { title: '服务标题', dataIndex: 'title', minWidth: 160 },
  { title: '分类', dataIndex: 'category', width: 90 },
  { title: '服务商', dataIndex: 'provider_name', width: 140 },
  { title: '区域', dataIndex: 'region', width: 120 },
  { title: '报价(元)', dataIndex: 'price_fen', slotName: 'price', width: 100, align: 'right' },
  { title: '单位', dataIndex: 'unit', width: 90 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
]

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({ id: '', title: '', provider_name: '', category: '巡检', region: '', priceYuan: '', unit: '', image: '', status: 'published', description: '' })
const imageList = reactive([])

const resetForm = () => {
  form.id = ''; form.title = ''; form.provider_name = ''; form.category = '巡检'
  form.region = ''; form.priceYuan = ''; form.unit = ''; form.image = ''
  form.status = 'published'; form.description = ''
  imageList.length = 0
}

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    form.id = row.id
    form.title = row.title || ''; form.provider_name = row.provider_name || ''
    form.category = row.category || '巡检'; form.region = row.region || ''
    form.priceYuan = ((row.price_fen || 0) / 100).toString()
    form.unit = row.unit || ''; form.image = row.image || ''
    form.status = row.status || 'published'; form.description = row.description || ''
    if (form.image) imageList.push({ name: form.image.split('/').pop(), url: form.image })
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

// a-upload 列表变化（新增/移除）时同步 form.image
// 注意：f.url 是 Arco 的本地 blob 预览地址，不能入库；真实地址在响应 data.url 里
const onImageChange = (fileList) => {
  imageList.length = 0
  imageList.push(...fileList)
  const imgs = fileList.map(f => f.response?.data?.url || f.response?.url || f.url).filter(Boolean)
  form.image = imgs[0] || ''
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入服务标题'); return }
  if (!form.provider_name) { Message.warning('请输入服务商名称'); return }
  formLoading.value = true
  const payload = {
    title: form.title,
    provider_name: form.provider_name,
    category: form.category,
    region: form.region,
    price_fen: Math.round(parseFloat(form.priceYuan || 0) * 100),
    unit: form.unit,
    image: form.image,
    status: form.status,
    description: form.description,
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

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除服务能力',
    content: `确定删除服务「${row.title}」吗？`,
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
