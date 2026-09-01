<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="training-courses"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增课程"
      @add="openForm()"
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
          <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
          <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无课程数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="课程详情" :width="'min(640px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="课程名称" :span="2">{{ currentItem.title || '-' }}</a-descriptions-item>
          <a-descriptions-item label="证书类型">{{ certTypeLabel(currentItem.cert_type) }}</a-descriptions-item>
          <a-descriptions-item label="价格">{{ currentItem.price_fen ? '¥' + (currentItem.price_fen / 100).toLocaleString() : '-' }}</a-descriptions-item>
          <a-descriptions-item label="名额">{{ currentItem.max_students ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="已报名">{{ currentItem.enrolled_count ?? 0 }} 人</a-descriptions-item>
          <a-descriptions-item label="开始日期">{{ formatDate(currentItem.start_date) }}</a-descriptions-item>
          <a-descriptions-item label="结束日期">{{ formatDate(currentItem.end_date) }}</a-descriptions-item>
          <a-descriptions-item label="地点">{{ currentItem.location || '-' }}</a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="封面图" :span="2">
            <a-image
              v-if="currentItem.image"
              :src="currentItem.image"
              :width="160"
              :height="100"
              alt="封面图"
              :preview-props="{ src: currentItem.image }"
              fit="cover"
              :style="{ borderRadius: '8px' }"
            />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="描述" :span="2">{{ currentItem.description || '-' }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑课程' : '新增课程'" :width="'min(560px, 94vw)'" :on-before-cancel="guardClose">
      <a-form :model="form" layout="vertical">
        <a-form-item label="课程名称" required><a-input v-model="form.title" style="width: 100%" :aria-required="true" /></a-form-item>
        <a-form-item label="封面图">
          <a-upload class="avatar-upload" :show-file-list="false" :custom-request="uploadRequest" accept="image/*" :before-upload="beforeUpload">
            <a-avatar v-if="form.image" :image-url="form.image" :size="80" shape="square" />
            <a-button v-else type="outline">点击上传</a-button>
          </a-upload>
        </a-form-item>
        <a-form-item label="证书类型">
          <a-select v-model="form.cert_type" style="width: 100%">
            <a-option value="caac">CAAC 执照</a-option>
            <a-option value="utc_dji">大疆 UTC</a-option>
            <a-option value="gov_level">人社等级</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="上课地点"><a-input v-model="form.location" style="width: 100%" /></a-form-item>
        <a-form-item label="状态">
          <a-select v-model="form.status" style="width: 100%">
            <a-option value="draft">草稿</a-option>
            <a-option value="pending">审核中</a-option>
            <a-option value="published">已上架</a-option>
            <a-option value="closed">已下架</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="价格(元)">
          <a-input-number v-model="form.priceYuan" :min="0" hide-button style="width: 100%" placeholder="单位：元" />
        </a-form-item>
        <a-form-item label="名额">
          <a-input-number v-model="form.max_students" :min="0" hide-button style="width: 100%" placeholder="招生人数" />
        </a-form-item>
        <a-form-item label="开始日期">
          <a-date-picker v-model="form.start_date" value-format="YYYY-MM-DD" placeholder="选择开班日期" style="width: 100%" />
        </a-form-item>
        <a-form-item label="结束日期">
          <a-date-picker v-model="form.end_date" value-format="YYYY-MM-DD" placeholder="选择结课日期" style="width: 100%" />
        </a-form-item>
        <a-form-item label="课程描述"><RichEditor v-model="form.description" /></a-form-item>
        <a-divider style="margin: 8px 0 16px">机构与展示信息（小程序详情页）</a-divider>
        <a-form-item label="机构名称">
          <a-input v-model="form.org_name" style="width: 100%" placeholder="培训机构全称（小程序详情页显示）" />
        </a-form-item>
        <a-form-item label="区县">
          <a-input v-model="form.district" style="width: 100%" placeholder="如 渝北区" />
        </a-form-item>
        <a-form-item label="培训周期（天）">
          <a-input-number v-model="form.duration_days" :min="0" hide-button style="width: 100%" placeholder="培训天数" />
        </a-form-item>
        <a-form-item label="评分">
          <a-input v-model="form.rating" style="width: 100%" placeholder="如 4.9" />
        </a-form-item>
        <a-form-item label="评价数">
          <a-input-number v-model="form.review_count" :min="0" hide-button style="width: 100%" placeholder="累计评价条数" />
        </a-form-item>
        <a-form-item label="通过率（%）">
          <a-input v-model="form.pass_rate" style="width: 100%" placeholder="如 92（小程序'通过考试'统计）" />
        </a-form-item>
        <a-form-item label="机构年限（年）">
          <a-input-number v-model="form.years" :min="0" hide-button style="width: 100%" placeholder="成立年限（小程序'机构年限'统计）" />
        </a-form-item>
        <a-form-item label="标签">
          <a-select v-model="form.tags" mode="tags" placeholder="输入后回车，如 实操教学、小班授课" style="width: 100%" allow-clear />
        </a-form-item>
        <a-form-item label="课程类型">
          <a-select v-model="form.course_types" mode="tags" placeholder="输入后回车，如 飞行执照、行业应用" style="width: 100%" allow-clear />
        </a-form-item>
        <a-form-item label="证书图">
          <a-upload class="avatar-upload" :show-file-list="false" :custom-request="certUploadRequest" accept="image/*" :before-upload="beforeUpload">
            <a-avatar v-if="form.certificate" :image-url="form.certificate" :size="80" shape="square" />
            <a-button v-else type="outline">点击上传</a-button>
          </a-upload>
        </a-form-item>
        <a-divider style="margin: 8px 0 16px">联系信息</a-divider>
        <a-form-item label="报名电话">
          <a-input v-model="form.phone" style="width: 100%" placeholder="机构报名电话（小程序拨打/展示）" />
        </a-form-item>
        <a-form-item label="营业时间">
          <a-input v-model="form.business_hours" style="width: 100%" placeholder="如 周一至周日 09:00-18:00" />
        </a-form-item>
        <a-divider style="margin: 8px 0 16px">价格方案（小程序「培训参考价」档位）</a-divider>
        <a-form-item label="各档位价格（元/人）">
          <div v-for="(row, i) in form.prices" :key="'p' + i" class="price-row">
            <a-input v-model="row.name" placeholder="档位名（如 标准班）" style="flex: 1" />
            <a-input-number v-model="row.price" :min="0" hide-button style="width: 120px" placeholder="价格" />
            <a-button type="text" status="danger" size="small" @click="form.prices.splice(i, 1)">删除</a-button>
          </div>
          <div class="price-row">
            <a-button size="small" type="outline" @click="form.prices.push({ name: '', price: null })">＋ 添加档位</a-button>
          </div>
        </a-form-item>
        <a-form-item label="课程方案（可选）">
          <div v-for="(row, i) in form.courses" :key="'c' + i" class="price-row">
            <a-input v-model="row.name" placeholder="方案名" style="flex: 1" />
            <a-input-number v-model="row.price" :min="0" hide-button style="width: 120px" placeholder="价格" />
            <a-button type="text" status="danger" size="small" @click="form.courses.splice(i, 1)">删除</a-button>
          </div>
          <div class="price-row">
            <a-button size="small" type="outline" @click="form.courses.push({ name: '', price: null })">＋ 添加方案</a-button>
          </div>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="handleCancel">取消</a-button>
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
import axios, { getAuthHeader } from '@/utils/http'
import CrudList from '../components/CrudList.vue'
import RichEditor from '@/components/RichEditor.vue'

const uploadUrl = '/api/v1/upload'

const beforeUpload = (item) => {
  const file = item?.file || item
  const isImage = !!file.type && file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

// 封面图上传：动态读取最新 accessToken（避免 token 轮转后仍用旧值）
const uploadRequest = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  try {
    const res = await axios.post(uploadUrl, fd, { headers: getAuthHeader() })
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    form.image = url
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error(e?.response?.data?.error?.message || e?.response?.data?.message || '上传失败')
  }
}

// 证书图上传：与封面图同源（写入 form.certificate）
const certUploadRequest = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  try {
    const res = await axios.post(uploadUrl, fd, { headers: getAuthHeader() })
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    form.certificate = url
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error(e?.response?.data?.error?.message || e?.response?.data?.message || '上传失败')
  }
}

const crudRef = ref()
const api = useAdminApi('training-courses')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ published: 'green', pending: 'orangered', draft: 'gray', closed: 'gray' }[s] || 'gray')
const statusLabel = { published: '已上架', pending: '审核中', draft: '草稿', closed: '已下架' }
const certTypeLabel = (t) => ({ caac: 'CAAC 执照', utc_dji: '大疆 UTC', gov_level: '人社等级' }[t] || t || '-')

// 批量动作：批量上架 / 批量下架——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'publish', label: '批量上架', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'published' }) },
  { key: 'close', label: '批量下架', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'closed' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索课程标题...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'pending', label: '审核中' },
    { value: 'draft', label: '草稿' },
    { value: 'published', label: '已上架' },
    { value: 'closed', label: '已下架' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '课程名称', dataIndex: 'title', slotName: 'title', minWidth: 200 },
  { title: '价格', dataIndex: 'price_fen', slotName: 'price', width: 110, align: 'right' },
  { title: '名额', dataIndex: 'max_students', slotName: 'maxStudents', width: 80 },
  { title: '状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (d) => { currentItem.value = d; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const form = reactive({
  id: '', title: '', cert_type: 'caac', priceYuan: null, max_students: null, location: '', start_date: '',
  end_date: '', status: 'draft', description: '', image: '',
  org_name: '', district: '', duration_days: null, rating: '', review_count: null, pass_rate: '', years: null,
  tags: [], course_types: [], certificate: '', phone: '', business_hours: '', prices: [], courses: [],
})

const resetForm = () => Object.assign(form, {
  id: '', title: '', cert_type: 'caac', priceYuan: null, max_students: null, location: '', start_date: '',
  end_date: '', status: 'draft', description: '', image: '',
  org_name: '', district: '', duration_days: null, rating: '', review_count: null, pass_rate: '', years: null,
  tags: [], course_types: [], certificate: '', phone: '', business_hours: '', prices: [], courses: [],
})

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      ...row,
      priceYuan: row.price_fen ? Math.round(row.price_fen / 100 * 100) / 100 : null,
      max_students: row.max_students ?? null,
      duration_days: row.duration_days ?? null,
      review_count: row.review_count ?? null,
      years: row.years ?? null,
      tags: Array.isArray(row.tags) ? row.tags.slice() : (row.tags ? String(row.tags).split(',') : []),
      course_types: Array.isArray(row.course_types) ? row.course_types.slice() : [],
      prices: (Array.isArray(row.prices) ? row.prices : []).map((p) => ({ name: p.name || '', price: p.price ?? null })),
      courses: (Array.isArray(row.courses) ? row.courses : []).map((p) => ({ name: p.name || '', price: p.price ?? null })),
    })
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

// 价格档位/方案数组序列化（{name, price}）；空行丢弃，价格转数字
const planRows = (rows) => (Array.isArray(rows) ? rows : [])
  .filter((r) => String(r.name || '').trim())
  .map((r) => ({ name: String(r.name).trim(), price: Number(r.price) || 0 }))

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入课程名称'); return }
  formLoading.value = true
  try {
    const payload = { ...form }
    // 空值提交 null：价格/名额未填时不再写 0 假数据（用户显式输入 0 仍保留 0）
    payload.price_fen = form.priceYuan == null ? null : Math.round(form.priceYuan * 100)
    payload.max_students = form.max_students ?? null
    payload.duration_days = form.duration_days ?? null
    payload.review_count = form.review_count ?? null
    payload.years = form.years ?? null
    payload.prices = planRows(form.prices)
    payload.courses = planRows(form.courses)
    delete payload.priceYuan
    if (formEdit.value) {
      await api.update(form.id, payload)
      Message.success('更新成功')
    } else {
      await api.create(payload)
      Message.success('创建成功')
    }
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
}

// 未保存守卫：X/Esc/遮罩/取消 关闭前，表单有改动则确认（onBeforeCancel 返回 false 阻断关闭）
let formSnapshot = ''
const guardClose = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  Modal.confirm({
    title: '放弃修改',
    content: '表单有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => { formVisible.value = false },
  })
  return false
}
// 底部取消按钮：走守卫，确认无改动/放弃修改后才真正关闭
const handleCancel = () => { if (guardClose()) formVisible.value = false }

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除课程',
    content: `确定删除该课程吗？（${row.title || row.id}）`,
    okText: '删除',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已删除')
        crudRef.value?.reload()
      } catch (e) { Message.error(e?.response?.data?.message || '删除失败') }
    }
  })
}
</script>

<style scoped>
.cell-title {
  font-weight: 500;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
  max-width: 300px;
}
.price-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.price-row .arco-btn { flex: none; }
</style>
