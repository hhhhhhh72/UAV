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
      <template #logo="{ record }">
        <a-image
          v-if="record.logo_url"
          :src="record.logo_url"
          :width="36"
          :height="36"
          fit="cover"
          :preview="false"
          :alt="(record.name || '院校') + ' Logo'"
          :style="{ borderRadius: '6px', display: 'block' }"
        />
        <span v-else>-</span>
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
          <a-descriptions-item label="合作模式">{{ coopTypeLabel[currentItem.coop_type] || coopTypeLabel.both }}</a-descriptions-item>
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
          <a-descriptions-item label="联系电话">{{ revealPII ? (currentItem.phone || '-') : maskPhone(currentItem.phone) }}</a-descriptions-item>
          <a-descriptions-item label="敏感信息" :span="2">
            <a-checkbox v-model="revealPII" :disabled="!currentItem">显示完整联系方式（谨慎操作）</a-checkbox>
          </a-descriptions-item>
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
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑院校' : '新增院校'" :width="'min(680px, 94vw)'" :on-before-cancel="guardClose">
      <a-form :model="form" layout="vertical">
        <div class="form-group-title">基本信息</div>
        <a-form-item label="院校名称" required><a-input v-model="form.name" :aria-required="true" :maxlength="100" style="width: 100%" /></a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="所在城市"><a-input v-model="form.city" placeholder="如：重庆" :maxlength="30" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="院校简称"><a-input v-model="form.short_name" placeholder="如：渝职院" :maxlength="30" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="合作模式">
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
            <a-form-item label="办学层次"><a-input v-model="form.level_tags" placeholder="办学层次，如：双一流 985 / 专科 示范校" :maxlength="50" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="院校类型"><a-input v-model="form.tagsText" placeholder="院校类型，逗号分隔，如：本科,综合类" :maxlength="100" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Logo">
          <a-upload class="avatar-upload" :show-file-list="false" :custom-request="uploadRequest" accept="image/*" :before-upload="beforeUpload">
            <a-image v-if="form.logo_url" :src="form.logo_url" :width="80" :height="80" fit="cover" :preview="false" alt="Logo 预览" :style="{ borderRadius: '8px', display: 'block' }" />
            <a-button v-else type="outline" :loading="uploadingLogo">点击上传</a-button>
          </a-upload>
          <span v-if="uploadingLogo" class="upload-status">上传中…</span>
        </a-form-item>

        <div class="form-group-title">图片</div>
        <a-form-item label="封面图">
          <a-upload class="avatar-upload" :show-file-list="false" :custom-request="uploadCover" accept="image/*" :before-upload="beforeUpload">
            <a-image v-if="form.cover" :src="form.cover" :width="140" :height="88" fit="cover" alt="封面预览" :preview-props="{ src: form.cover }" :style="{ borderRadius: '8px' }" />
            <a-button v-else type="outline" :loading="uploadingCover">点击上传</a-button>
          </a-upload>
          <span v-if="uploadingCover" class="upload-status">上传中…</span>
          <div class="form-tip">小程序院校列表/详情页展示的封面全景图</div>
        </a-form-item>
        <a-form-item label="校园环境图">
          <a-upload
            :file-list="photoList"
            list-type="picture-card"
            :limit="4"
            :custom-request="uploadPhoto"
            :before-upload="beforeUpload"
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
        <a-form-item label="硕博导师数"><a-input-number v-model="form.teacher_count" :min="0" :max="999999" hide-button style="width: 100%" /></a-form-item>

        <div class="form-group-title">联系方式与就业</div>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="联系电话"><a-input ref="phoneRef" v-model="form.phone" placeholder="如：023-88886666" :maxlength="20" style="width: 100%" /></a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="官网地址"><a-input ref="websiteRef" v-model="form.website" placeholder="如：https://www.cqxx.edu.cn" :maxlength="200" style="width: 100%" /></a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="就业率"><a-input ref="rateRef" v-model="form.graduate_rate" placeholder="如：98%" :maxlength="10" style="width: 100%" /></a-form-item>

        <div class="form-group-title">院校介绍</div>
        <a-form-item label="院校介绍"><a-input v-model="form.intro" type="textarea" :auto-size="{ minRows: 2, maxRows: 4 }" :maxlength="500" style="width: 100%" />
          <div class="form-tip">小程序院校详情页主展示文案</div>
        </a-form-item>
        <a-form-item label="描述"><a-input v-model="form.description" type="textarea" :rows="2" :maxlength="500" style="width: 100%" />
          <div class="form-tip">备用描述，仅当院校介绍为空时展示</div>
        </a-form-item>

        <div class="form-group-title">专业与合作</div>
        <a-form-item label="无人机专业">
          <view v-for="(row, i) in majorsRows" :key="i" class="dsl-row">
            <a-input v-model="row.name" placeholder="专业名称（如：飞行器设计与工程）" :maxlength="50" style="flex: 2; min-width: 160px" />
            <a-select v-model="row.degree" style="width: 100px">
              <a-option value="本科">本科</a-option>
              <a-option value="硕士">硕士</a-option>
              <a-option value="博士">博士</a-option>
            </a-select>
            <a-input-number v-model="row.duration" :min="1" :max="10" hide-button style="width: 70px" />
            <span class="dsl-unit">年制</span>
            <a-switch v-model="row.flagship" :checked-value="true" :unchecked-value="false" />
            <span class="dsl-unit">王牌</span>
            <a-button type="text" status="danger" size="small" @click="removeMajorRow(i)">删除</a-button>
          </view>
          <a-button type="outline" size="small" @click="addMajorRow">＋ 添加专业</a-button>
          <div class="form-tip">王牌 = 国家级特色专业；留空则小程序不展示该区块</div>
        </a-form-item>
        <a-form-item label="合作企业">
          <view v-for="(row, i) in partnersRows" :key="i" class="dsl-row">
            <a-input v-model="row.name" placeholder="企业名称（如：大疆创新）" :maxlength="50" style="flex: 2; min-width: 160px" />
            <a-input v-model="row.type" placeholder="合作类型（如：联合实验室）" :maxlength="30" style="flex: 1; min-width: 120px" />
            <a-button type="text" status="danger" size="small" @click="removePartnerRow(i)">删除</a-button>
          </view>
          <a-button type="outline" size="small" @click="addPartnerRow">＋ 添加企业</a-button>
          <div class="form-tip">留空则小程序不展示该区块</div>
        </a-form-item>
        <a-form-item label="特色专业"><a-input v-model="form.majorsText" placeholder="多个专业用逗号分隔，如：无人机应用技术,测绘地理信息" :maxlength="200" style="width: 100%" />
          <div class="form-tip">列表页特色专业展示；与上方结构化"无人机专业"二选一即可</div>
        </a-form-item>
        <a-form-item label="实训设施"><a-input v-model="form.facilitiesText" placeholder="多个设施用逗号分隔，如：实训基地,联合实验室" :maxlength="200" style="width: 100%" /></a-form-item>
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
import { maskPhone } from '@/utils/mask'
import { onBeforeRouteLeave, useRouter } from 'vue-router'
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

// 校园环境图（多图，最多 4 张；成功由卡片缩略图反馈，不逐张弹 toast）
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
// 排除上传中（uploading）与失败（error）的 blob 项，防止 blob 地址入库。
const onPhotoChange = (fileList) => {
  photoList.length = 0
  photoList.push(...fileList)
  form.photos = (fileList || [])
    .map((f) => f.response?.data?.url || f.response?.url || (f.status === 'uploading' || f.status === 'error' ? '' : f.url))
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
  { title: 'Logo', dataIndex: 'logo_url', slotName: 'logo', width: 80 },
  { title: '院校名称', dataIndex: 'name', slotName: 'name', minWidth: 200 },
  { title: '合作模式', dataIndex: 'coop_type', slotName: 'coopType', width: 100 },
  { title: '所在城市', dataIndex: 'city', slotName: 'city', width: 120 },
  { title: '特色专业', dataIndex: 'majors', slotName: 'majors', minWidth: 160 },
  { title: '合作状态', dataIndex: 'status', slotName: 'status', width: 100 },
  { title: '操作', slotName: 'actions', width: 200, fixed: 'right' }
]

const detailVisible = ref(false)
const currentItem = ref(null)
const revealPII = ref(false)
const showDetail = (row) => { revealPII.value = false; currentItem.value = row; detailVisible.value = true }

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const phoneRef = ref()
const websiteRef = ref()
const rateRef = ref()
const majorsError = ref('')
const partnersError = ref('')
const coopTypeLabel = { research: '科研合作', talent: '人才培养', both: '综合' }
const arrText = (v) => (Array.isArray(v) ? v.join('、') : (v || ''))
const splitArr = (s) => String(s || '').split(/[,，、]/).map(x => x.trim()).filter(Boolean)
// 专业对象数组 ↔ 文本（每行：名称|学历|年制|王牌）
// 解析返回 { items, errors }——坏行不再静默丢弃，逐行报错指明行号；
// 结构化行编辑器：专业/合作企业（替代管道 DSL 文本域，格式错误架构上不可能）
const majorsRows = reactive([])
const partnersRows = reactive([])
const addMajorRow = () => majorsRows.push({ name: '', degree: '本科', duration: 4, flagship: false })
const removeMajorRow = (i) => majorsRows.splice(i, 1)
const addPartnerRow = () => partnersRows.push({ name: '', type: '' })
const removePartnerRow = (i) => partnersRows.splice(i, 1)
const loadRows = (row) => {
  majorsRows.length = 0
  partnersRows.length = 0
  ;(row.majors_detail || []).forEach((m) => majorsRows.push({ name: m.name || '', degree: m.degree || '本科', duration: m.duration || 4, flagship: !!m.flagship }))
  ;(row.partners || []).forEach((p) => partnersRows.push({ name: p.name || '', type: p.type || '' }))
}

const form = reactive({
  id: '', name: '', city: '', short_name: '', coop_type: 'both', status: 'active',
  level_tags: '', tagsText: '', logo_url: '', cover: '', photos: [],
  major_count: null, partner_count: null, teacher_count: null, student_count: null, graduate_rate: '',
  phone: '', website: '', intro: '',
  majorsText: '', facilitiesText: '', region: '', description: '',
})
const resetForm = () => {
  photoList.length = 0
  majorsError.value = ''
  partnersError.value = ''
  majorsRows.length = 0
  partnersRows.length = 0
  Object.assign(form, {
    id: '', name: '', city: '', short_name: '', coop_type: 'both', status: 'active',
    level_tags: '', tagsText: '', logo_url: '', cover: '', photos: [],
    major_count: null, partner_count: null, teacher_count: null, student_count: null, graduate_rate: '',
    phone: '', website: '', intro: '',
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
      majorsText: arrText(row.majors), facilitiesText: arrText(row.facilities),
      region: row.region || '', description: row.description || '',
    })
    loadRows(row)
    ;(row.photos || []).forEach((u) => photoList.push({ name: u.split('/').pop(), url: u }))
  } else {
    formEdit.value = false
  }
  takeSnapshot()
  formVisible.value = true
}

const submitForm = async () => {
  majorsError.value = ''
  partnersError.value = ''
  if (!form.name) { Message.warning('请输入院校名称'); return }
  // 结构化行编辑器：仅校验每行名称非空（格式由控件架构保证）
  const majorsItems = majorsRows
    .map((r) => ({ name: String(r.name || '').trim(), degree: r.degree || '本科', duration: Number(r.duration) || 4, flagship: !!r.flagship, key: r.flagship ? '国家级特色专业' : '' }))
    .filter((m) => m.name)
  const partnersItems = partnersRows
    .map((p) => ({ icon: (String(p.name || '').trim() || '企').slice(0, 1), name: String(p.name || '').trim(), type: String(p.type || '').trim() || '合作单位' }))
    .filter((p) => p.name)
  const emptyMajor = majorsRows.find((r) => !String(r.name || '').trim())
  const emptyPartner = partnersRows.find((p) => !String(p.name || '').trim())
  if (emptyMajor) { Message.warning('请填写专业名称或删除空行'); return }
  if (emptyPartner) { Message.warning('请填写企业名称或删除空行'); return }
  // 联系方式/就业率格式校验（均选填，填了就要合法；错误聚焦对应输入框）
  if (form.phone && !/^[\d\-+() ]{5,20}$/.test(form.phone.trim())) {
    Message.warning('联系电话格式不正确')
    phoneRef.value && phoneRef.value.focus && phoneRef.value.focus()
    return
  }
  if (form.website && !/^https?:\/\/\S+$/.test(form.website.trim())) {
    Message.warning('官网地址需以 http:// 或 https:// 开头')
    websiteRef.value && websiteRef.value.focus && websiteRef.value.focus()
    return
  }
  // 就业率：≤100；小数（如 0.98）归一为 98%
  if (form.graduate_rate) {
    const gv = form.graduate_rate.trim()
    const m = /^(\d{1,3}(?:\.\d{1,2})?)%?$/.exec(gv)
    if (!m) {
      Message.warning('就业率格式不正确，如：98 或 98%')
      rateRef.value && rateRef.value.focus && rateRef.value.focus()
      return
    }
    let rate = Number(m[1])
    if (rate > 100) {
      Message.warning('就业率不能超过 100%')
      rateRef.value && rateRef.value.focus && rateRef.value.focus()
      return
    }
    if (rate > 0 && rate < 1) rate = rate * 100 // 0.98 → 98
    form.graduate_rate = String(rate) + '%'
  }
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
      intro: form.intro, majors_detail: majorsItems,
      partners: partnersItems,
    }
    if (formEdit.value) {
      await api.update(form.id, p)
      Message.success('更新成功')
    } else {
      await api.create(p)
      Message.success('创建成功')
    }
    takeSnapshot()
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
  } finally {
    formLoading.value = false
  }
}

// 未保存守卫：X/Esc/遮罩/取消 关闭前，表单有改动则确认（onBeforeCancel 返回 false 阻断关闭）
// 快照同时覆盖结构化行（majorsRows/partnersRows），行的增删改也触发守卫
let formSnapshot = ''
const takeSnapshot = () => { formSnapshot = JSON.stringify({ f: form, m: majorsRows, p: partnersRows }) }
const hasUnsaved = () => JSON.stringify({ f: form, m: majorsRows, p: partnersRows }) !== formSnapshot
const guardClose = () => {
  if (!hasUnsaved()) return true
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

// 路由级守卫：弹窗打开且有改动时，侧边栏切走也拦截（弹窗守卫不覆盖此路径）
const router = useRouter()
let pendingLeave = null
onBeforeRouteLeave((to) => {
  if (formVisible.value && hasUnsaved()) {
    pendingLeave = to
    Modal.confirm({
      title: '放弃修改',
      content: '表单有未保存的修改，确定离开吗？',
      okText: '放弃修改',
      cancelText: '继续编辑',
      onOk: () => {
        formVisible.value = false
        pendingLeave && router.push(pendingLeave.fullPath)
      },
    })
    return false
  }
  return true
})

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

/* 上传中状态（编辑态重传时按钮不可见，用文字提示） */
.upload-status {
  display: inline-block;
  font-size: 12px;
  color: var(--color-text-2);
  margin-left: 10px;
}

/* 结构化行编辑器：行内 flex 布局，间距紧凑 */
.dsl-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.dsl-unit {
  font-size: 12px;
  color: var(--color-text-3);
  white-space: nowrap;
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
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-2);
  background: var(--color-fill-2);
  border-radius: 4px;
  padding: 1px 8px;
  margin-left: 6px;
}
.detail-tag--hot {
  color: #B54708;
  background: #FFF0E6;
}
</style>
