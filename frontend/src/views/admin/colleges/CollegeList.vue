<template>
  <div class="page">
    <CrudList
      ref="crudRef"
      resource="colleges"
      :columns="columns"
      :search-fields="searchFields"
      :batch-actions="batchActions"
      creatable
      add-label="新增院校"
      @add="openForm()"
    >
      <template #name="{ record }">
        <span class="cell-title">{{ record.name || '-' }}</span>
      </template>
      <template #coopType="{ record }">
        <span>{{ coopTypeLabel[record.coop_type] || coopTypeLabel.both }}</span>
      </template>
      <template #city="{ record }">
        <span>{{ record.city || record.region || '-' }}</span>
      </template>
      <template #majors="{ record }">
        <span>{{ arrText(record.majors) }}</span>
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
        <a-empty description="暂无院校数据" />
      </template>
    </CrudList>

    <!-- 详情弹窗 -->
    <a-modal v-model:visible="detailVisible" title="院校详情" :width="'min(640px, 94vw)'" :footer="false">
      <template v-if="currentItem">
        <a-descriptions :column="2" bordered size="medium">
          <a-descriptions-item label="院校名称" :span="2">{{ currentItem.name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="所在城市">{{ currentItem.city || currentItem.region || '-' }}</a-descriptions-item>
          <a-descriptions-item label="院校简称">{{ currentItem.short_name || '-' }}</a-descriptions-item>
          <a-descriptions-item label="分域">{{ coopTypeLabel[currentItem.coop_type] || coopTypeLabel.both }}</a-descriptions-item>
          <a-descriptions-item label="合作状态">
            <a-tag :color="statusTag(currentItem.status)" size="small">{{ statusLabel[currentItem.status] || currentItem.status || '-' }}</a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="层次标签" :span="2">{{ currentItem.level_tags || '-' }}</a-descriptions-item>
          <a-descriptions-item label="类型标签" :span="2">{{ arrText(currentItem.tags) || '-' }}</a-descriptions-item>
          <a-descriptions-item label="Logo" :span="2">
            <a-image
              v-if="currentItem.logo_url"
              :src="currentItem.logo_url"
              :width="120"
              :height="80"
              :preview-props="{ src: currentItem.logo_url }"
              fit="cover"
              alt="院校 Logo"
              :style="{ borderRadius: '8px' }"
            />
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="无人机专业数">{{ currentItem.major_count ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="合作企业数">{{ currentItem.partner_count ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="在读学生">{{ currentItem.student_count ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="硕博导师数">{{ currentItem.teacher_count ?? '-' }}</a-descriptions-item>
          <a-descriptions-item label="就业率">{{ currentItem.graduate_rate || '-' }}</a-descriptions-item>
          <a-descriptions-item label="联系电话">{{ currentItem.phone || '-' }}</a-descriptions-item>
          <a-descriptions-item label="官网" :span="2">
            <a-link v-if="currentItem.website" :href="currentItem.website" target="_blank">{{ currentItem.website }}</a-link>
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="院校介绍" :span="2">{{ currentItem.intro || currentItem.description || '-' }}</a-descriptions-item>
          <a-descriptions-item v-if="currentItem.majors_detail && currentItem.majors_detail.length" label="无人机专业" :span="2">
            <div v-for="m in currentItem.majors_detail" :key="m.name" class="detail-item">
              {{ m.name }} <span class="detail-tag">{{ m.degree || '本科' }} · {{ m.duration || 4 }}年制</span>
              <span v-if="m.flagship" class="detail-tag detail-tag--hot">王牌</span>
            </div>
          </a-descriptions-item>
          <a-descriptions-item v-if="currentItem.partners && currentItem.partners.length" label="合作企业" :span="2">
            <div v-for="p in currentItem.partners" :key="p.name" class="detail-item">
              {{ p.name }} <span class="detail-tag">{{ p.type || '合作单位' }}</span>
            </div>
          </a-descriptions-item>
          <a-descriptions-item label="特色专业" :span="2">{{ arrText(currentItem.majors) || '-' }}</a-descriptions-item>
          <a-descriptions-item label="实训设施" :span="2">{{ arrText(currentItem.facilities) || '-' }}</a-descriptions-item>
          <a-descriptions-item label="创建时间">{{ formatDate(currentItem.created_at) }}</a-descriptions-item>
        </a-descriptions>
      </template>
    </a-modal>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑院校' : '新增院校'" :width="'min(680px, 94vw)'" :before-close="beforeClose">
      <a-form :model="form" layout="vertical">
        <div class="form-group-title">基本信息</div>
        <a-form-item label="院校名称" required><a-input v-model="form.name" style="width: 100%" /></a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="所在城市"><a-input v-model="form.city" placeholder="如：重庆" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="院校简称"><a-input v-model="form.short_name" placeholder="如：渝职院" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="分域">
              <a-select v-model="form.coop_type" style="width: 100%">
                <a-option value="research">科研合作</a-option>
                <a-option value="talent">人才培养</a-option>
                <a-option value="both">综合</a-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="合作状态">
              <a-select v-model="form.status" style="width: 100%">
                <a-option value="active">合作中</a-option>
                <a-option value="inactive">已终止合作</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="层次标签"><a-input v-model="form.level_tags" placeholder="学历层次，如：双一流 985 / 专科 示范校" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="类型标签"><a-input v-model="form.tagsText" placeholder="院校类型，逗号分隔，如：本科,双一流" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Logo">
          <a-upload class="avatar-upload" :show-file-list="false" :custom-request="uploadRequest" accept="image/*" :before-upload="beforeUpload">
            <a-avatar v-if="form.logo_url" :image-url="form.logo_url" :size="80" shape="square" />
            <a-button v-else type="outline" :loading="uploadingLogo">点击上传</a-button>
          </a-upload>
        </a-form-item>

        <div class="form-group-title">图片</div>
        <a-form-item label="封面图">
          <a-upload class="avatar-upload" :show-file-list="false" :custom-request="uploadCover" accept="image/*" :before-upload="beforeUpload">
            <a-avatar v-if="form.cover" :image-url="form.cover" :size="80" shape="square" />
            <a-button v-else type="outline" :loading="uploadingCover">点击上传</a-button>
          </a-upload>
          <div class="form-tip">小程序院校列表/详情页展示的封面全景图</div>
        </a-form-item>
        <a-form-item label="校园环境图">
          <a-upload
            :file-list="photoList"
            list-type="picture-card"
            :limit="4"
            :custom-request="uploadPhoto"
            @change="onPhotoChange"
          />
          <div class="form-tip">最多 4 张，小程序详情页校园环境四格展示</div>
        </a-form-item>

        <div class="form-group-title">数据指标</div>
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="无人机专业数"><a-input-number v-model="form.major_count" :min="0" :max="999999" hide-button style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="合作企业数"><a-input-number v-model="form.partner_count" :min="0" :max="999999" hide-button style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="在读学生"><a-input-number v-model="form.student_count" :min="0" :max="999999" hide-button style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="硕博导师数"><a-input-number v-model="form.teacher_count" :min="0" :max="999999" hide-button style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="就业率"><a-input v-model="form.graduate_rate" placeholder="如：98%" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>

        <div class="form-group-title">联系方式</div>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="联系电话"><a-input v-model="form.phone" placeholder="如：023-88886666" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="官网地址"><a-input v-model="form.website" placeholder="如：https://www.cqxx.edu.cn" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>

        <div class="form-group-title">院校介绍</div>
        <a-form-item label="院校介绍"><a-input v-model="form.intro" type="textarea" :auto-size="{ minRows: 2, maxRows: 4 }" style="width: 100%" /></a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="2" style="width: 100%" /></a-form-item>

        <div class="form-group-title">专业与合作</div>
        <a-form-item label="无人机专业">
          <a-textarea ref="majorsRef" v-model="form.majorsDetailText" :auto-size="{ minRows: 2, maxRows: 5 }" placeholder="每行一个专业，格式：名称|学历|年制|王牌，如：&#10;飞行器设计与工程|本科|4|1&#10;无人机系统工程|本科|4|0&#10;飞行器控制与信息工程|硕士|3|0" style="width: 100%" />
          <div class="form-tip">学历：本科/硕士/博士；王牌填 1 表示该专业为国家级特色专业</div>
        </a-form-item>
        <a-form-item label="合作企业">
          <a-textarea ref="partnersRef" v-model="form.partnersText" :auto-size="{ minRows: 2, maxRows: 4 }" placeholder="每行一个企业，格式：名称|类型，如：&#10;大疆创新|联合实验室&#10;中航工业|实习基地" style="width: 100%" />
        </a-form-item>
        <a-form-item label="特色专业"><a-input v-model="form.majorsText" placeholder="多个专业用逗号分隔，如：无人机应用技术,测绘地理信息" style="width: 100%" /></a-form-item>
        <a-form-item label="实训设施"><a-input v-model="form.facilitiesText" placeholder="多个设施用逗号分隔，如：实训基地,联合实验室" style="width: 100%" /></a-form-item>
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
import axios, { getAuthHeader } from '@/utils/http'
import CrudList from '../components/CrudList.vue'

const uploadUrl = '/api/v1/upload'

const beforeUpload = (item) => {
  const file = item?.file || item
  const isImage = !!file.type && file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) { Message.error('只能上传图片文件'); return false }
  if (!isLt5M) { Message.error('图片不能超过 5MB'); return false }
  return true
}

// Logo 上传：动态读取最新 accessToken
const uploadingLogo = ref(false)
const uploadRequest = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  uploadingLogo.value = true
  try {
    const res = await axios.post(uploadUrl, fd, { headers: getAuthHeader() })
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    form.logo_url = url
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error(e?.response?.data?.error?.message || e?.response?.data?.message || '上传失败')
  } finally {
    uploadingLogo.value = false
  }
}

// 封面图上传
const uploadingCover = ref(false)
const uploadCover = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  uploadingCover.value = true
  try {
    const res = await axios.post(uploadUrl, fd, { headers: getAuthHeader() })
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    form.cover = url
    Message.success('上传成功')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error(e?.response?.data?.error?.message || e?.response?.data?.message || '上传失败')
  } finally {
    uploadingCover.value = false
  }
}

// 校园环境图（多图，最多 4 张）
const photoList = reactive([])
const uploadPhoto = async ({ fileItem, onSuccess, onError }) => {
  const fd = new FormData()
  fd.append('file', fileItem.file)
  try {
    const res = await axios.post(uploadUrl, fd, { headers: getAuthHeader() })
    const url = res?.data?.url || res?.url
    if (!url) throw new Error('上传失败')
    onSuccess && onSuccess(res)
  } catch (e) {
    onError && onError(e)
    Message.error(e?.response?.data?.error?.message || e?.response?.data?.message || '上传失败')
  }
}
// a-upload 列表变化（新增/移除）时同步 form.photos
// 注意：f.url 是 Arco 的本地 blob 预览地址，不能入库；真实地址在响应 data.url 里。
// 编辑态旧照片以 {name, url} 推入（无 response），其 url 是真实地址——保留；
// 仅排除上传中（uploading）的 blob 项，防止编辑后旧照片静默丢失。
const onPhotoChange = (fileList) => {
  photoList.length = 0
  photoList.push(...fileList)
  form.photos = (fileList || [])
    .map((f) => f.response?.data?.url || f.response?.url || (f.status === 'uploading' ? '' : f.url))
    .filter(Boolean)
}

const crudRef = ref()
const api = useAdminApi('colleges')

const formatDate = (d) => {
  if (!d) return '-'
  const dt = new Date(d)
  if (isNaN(dt.getTime())) return '-'
  const p = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${p(dt.getMonth() + 1)}-${p(dt.getDate())} ${p(dt.getHours())}:${p(dt.getMinutes())}`
}

const statusTag = (s) => ({ active: 'green', inactive: 'gray' }[s] || 'gray')
const statusLabel = { active: '合作中', inactive: '已终止合作' }

// 批量动作：批量合作中 / 批量终止合作——传完整行数据避免清空其他字段
const batchActions = [
  { key: 'activate', label: '批量合作中', status: 'success', api: (row) => api.update(row.id, { ...row, status: 'active' }) },
  { key: 'close', label: '批量终止合作', status: 'warning', api: (row) => api.update(row.id, { ...row, status: 'inactive' }) }
]

const searchFields = [
  { key: 'keyword', label: '关键词', placeholder: '搜索院校名称...', width: 220 },
  { key: 'status', label: '状态', type: 'select', options: [
    { value: '', label: '全部' },
    { value: 'active', label: '合作中' },
    { value: 'inactive', label: '已终止合作' }
  ]}
]

const columns = [
  { title: 'ID', dataIndex: 'id', width: 160 },
  { title: '院校名称', dataIndex: 'name', slotName: 'name', minWidth: 200 },
  { title: '分域', dataIndex: 'coop_type', slotName: 'coopType', width: 100 },
  { title: '所在城市', dataIndex: 'city', slotName: 'city', width: 120 },
  { title: '特色专业', dataIndex: 'majors', slotName: 'majors', minWidth: 160 },
  { title: '合作状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const showDetail = (row) => { currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const majorsRef = ref()
const partnersRef = ref()
const coopTypeLabel = { research: '科研合作', talent: '人才培养', both: '综合' }
const arrText = (v) => (Array.isArray(v) ? v.join('、') : (v || ''))
const splitArr = (s) => String(s || '').split(/[,，、]/).map(x => x.trim()).filter(Boolean)
// 专业对象数组 ↔ 文本（每行：名称|学历|年制|王牌）
// 解析返回 { items, errors }——坏行不再静默丢弃，逐行报错指明行号；
// 全角竖线（｜）归一为半角，避免整行被当成长名称的脏数据。
const majorsDetailText = (list) => (Array.isArray(list) ? list.map((m) => [m.name, m.degree || '本科', m.duration || 4, m.flagship ? 1 : 0].join('|')).join('\n') : '')
const parseMajorsDetail = (s) => {
  const items = []
  const errors = []
  String(s || '').replace(/｜/g, '|').split('\n').map(l => l.trim()).filter(Boolean).forEach((line, i) => {
    const p = line.split('|').map(x => x.trim())
    if (p.length > 4) { errors.push(`专业第 ${i + 1} 行字段过多（应为 名称|学历|年制|王牌 4 项）`); return }
    if (!p[0]) { errors.push(`专业第 ${i + 1} 行缺少名称`); return }
    const duration = Number(p[2])
    if (p[2] && (!Number.isInteger(duration) || duration < 1 || duration > 10)) {
      errors.push(`专业第 ${i + 1} 行年制需为 1-10 的整数（当前：${p[2]}）`); return
    }
    if (p[3] && p[3] !== '1' && p[3] !== '0') {
      errors.push(`专业第 ${i + 1} 行王牌需填 1 或 0（当前：${p[3]}）`); return
    }
    items.push({ name: p[0], degree: p[1] || '本科', duration: duration || 4, flagship: p[3] === '1', key: p[3] === '1' ? '国家级特色专业' : '' })
  })
  return { items, errors }
}
// 合作企业对象数组 ↔ 文本（每行：名称|类型）
const partnersText = (list) => (Array.isArray(list) ? list.map((p) => [p.name, p.type || '合作单位'].join('|')).join('\n') : '')
const parsePartners = (s) => {
  const items = []
  const errors = []
  String(s || '').replace(/｜/g, '|').split('\n').map(l => l.trim()).filter(Boolean).forEach((line, i) => {
    const p = line.split('|').map(x => x.trim())
    if (p.length > 2) { errors.push(`合作企业第 ${i + 1} 行字段过多（应为 名称|类型 2 项）`); return }
    if (!p[0]) { errors.push(`合作企业第 ${i + 1} 行缺少名称`); return }
    items.push({ icon: p[0].slice(0, 1), name: p[0], type: p[1] || '合作单位' })
  })
  return { items, errors }
}

const form = reactive({
  id: '', name: '', city: '', short_name: '', coop_type: 'both', status: 'active',
  level_tags: '', tagsText: '', logo_url: '', cover: '', photos: [],
  major_count: null, partner_count: null, teacher_count: null, student_count: null, graduate_rate: '',
  phone: '', website: '', intro: '', majorsDetailText: '', partnersText: '',
  majorsText: '', facilitiesText: '', region: '', description: '',
})
const resetForm = () => {
  photoList.length = 0
  Object.assign(form, {
    id: '', name: '', city: '', short_name: '', coop_type: 'both', status: 'active',
    level_tags: '', tagsText: '', logo_url: '', cover: '', photos: [],
    major_count: null, partner_count: null, teacher_count: null, student_count: null, graduate_rate: '',
    phone: '', website: '', intro: '', majorsDetailText: '', partnersText: '',
    majorsText: '', facilitiesText: '', region: '', description: '',
  })
}

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    Object.assign(form, {
      id: row.id, name: row.name || '', city: row.city || '', short_name: row.short_name || '',
      coop_type: row.coop_type || 'both', status: row.status || 'active',
      level_tags: row.level_tags || '', tagsText: arrText(row.tags),
      logo_url: row.logo_url || '', cover: row.cover || '', photos: row.photos || [],
      major_count: row.major_count ?? null, partner_count: row.partner_count ?? null,
      teacher_count: row.teacher_count ?? null, student_count: row.student_count ?? null,
      graduate_rate: row.graduate_rate || '',
      phone: row.phone || '', website: row.website || '', intro: row.intro || '',
      majorsDetailText: majorsDetailText(row.majors_detail), partnersText: partnersText(row.partners),
      majorsText: arrText(row.majors), facilitiesText: arrText(row.facilities),
      region: row.region || '', description: row.description || '',
    })
    ;(row.photos || []).forEach((u) => photoList.push({ name: u.split('/').pop(), url: u }))
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

const submitForm = async () => {
  if (!form.name) { Message.warning('请输入院校名称'); return }
  // 专业/合作企业：逐行校验，坏行明确报错（一次报至多 3 条）并聚焦对应输入框
  const majorsParsed = parseMajorsDetail(form.majorsDetailText)
  if (majorsParsed.errors.length > 0) {
    Message.error(majorsParsed.errors.slice(0, 3).join('；'))
    majorsRef.value && majorsRef.value.focus && majorsRef.value.focus()
    return
  }
  const partnersParsed = parsePartners(form.partnersText)
  if (partnersParsed.errors.length > 0) {
    Message.error(partnersParsed.errors.slice(0, 3).join('；'))
    partnersRef.value && partnersRef.value.focus && partnersRef.value.focus()
    return
  }
  // 联系方式/就业率格式校验（均选填，填了就要合法）
  if (form.phone && !/^[\d\-+() ]{5,20}$/.test(form.phone.trim())) { Message.warning('联系电话格式不正确'); return }
  if (form.website && !/^https?:\/\/\S+$/.test(form.website.trim())) { Message.warning('官网地址需以 http:// 或 https:// 开头'); return }
  if (form.graduate_rate && !/^\d{1,3}(\.\d{1,2})?%?$/.test(form.graduate_rate.trim())) { Message.warning('就业率格式不正确，如：98 或 98%'); return }
  formLoading.value = true
  try {
    const p = {
      id: form.id, name: form.name, coop_type: form.coop_type, region: form.region,
      logo_url: form.logo_url, status: form.status, description: form.description,
      majors: splitArr(form.majorsText), facilities: splitArr(form.facilitiesText),
      // 小程序院校页对齐字段
      city: form.city, short_name: form.short_name, level_tags: form.level_tags,
      tags: splitArr(form.tagsText), cover: form.cover, photos: form.photos || [],
      major_count: Number(form.major_count) || 0, partner_count: Number(form.partner_count) || 0,
      teacher_count: Number(form.teacher_count) || 0, student_count: Number(form.student_count) || 0,
      graduate_rate: form.graduate_rate.trim(), phone: form.phone.trim(), website: form.website.trim(),
      intro: form.intro, majors_detail: majorsParsed.items,
      partners: partnersParsed.items,
    }
    if (formEdit.value) {
      await api.update(form.id, p)
      Message.success('更新成功')
    } else {
      await api.create(p)
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

// 未保存守卫：Esc/点遮罩/点 X 关闭前，若表单有改动则确认，避免 20+ 字段输入全丢
let formSnapshot = ''
const beforeClose = () => {
  if (JSON.stringify(form) === formSnapshot) return true
  return new Promise((resolve) => {
    Modal.confirm({
      title: '放弃修改',
      content: '表单有未保存的修改，确定放弃吗？',
      okText: '放弃修改',
      cancelText: '继续编辑',
      onOk: () => resolve(true),
      onCancel: () => resolve(false),
    })
  })
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '删除院校',
    content: `确定删除院校「${row.name || row.id}」吗？删除后不可恢复`,
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

.form-tip {
  font-size: 12px;
  color: var(--color-text-2);
  line-height: 1.5;
  margin-top: 4px;
}

/* 表单分区：组间宽松留白 + 细分隔线，组内保持 Arco 默认紧凑节奏 */
.form-group-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-1);
  margin: 20px 0 4px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--color-border-2);
}
.form-group-title:first-child {
  margin-top: 0;
}

/* 详情弹窗内嵌列表项 */
.detail-item {
  font-size: 13px;
  color: var(--color-text-1);
  line-height: 1.8;
}
.detail-tag {
  display: inline-block;
  font-size: 12px;
  color: var(--color-text-3);
  background: var(--color-fill-2);
  border-radius: 4px;
  padding: 1px 8px;
  margin-left: 6px;
}
.detail-tag--hot {
  color: var(--color-warning-6);
  background: var(--color-warning-1);
}
</style>
