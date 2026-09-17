<template>
  <div class="page">
    <!-- 商品回收站：同一个列表组件、同一个接口，只多带一个 deleted=1。
         不新开页面——管理端已有 46 条路由，回收站不值得第 47 条。
         :key="mode" 强制重建 CrudList，让 useListRequest 用新的 defaultParams 重新初始化。 -->
    <a-radio-group v-model="mode" type="button" size="small" class="mode-switch">
      <a-radio value="active">商品管理</a-radio>
      <a-radio value="recycle">回收站</a-radio>
    </a-radio-group>
    <CrudList
      :key="mode"
      ref="crudRef"
      resource="products"
      :columns="columns"
      :search-fields="mode === 'active' ? searchFields : []"
      :batch-actions="mode === 'active' ? batchActions : []"
      :default-params="mode === 'recycle' ? { deleted: 1 } : {}"
      :creatable="mode === 'active'"
      :batch-delete="mode === 'active'"
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
          alt="商品封面"
        />
        <span v-else class="no-image">无图</span>
      </template>
      <!-- 类型优先显示 category：prod_type 枚举只有 7 个粗类目，
           "巡检/测绘/植保/应急"不在其中，只显示 prod_type 会把这四类误标成"维修服务" -->
      <template #prodType="{ record }">
        <span>{{ (record.category || '').trim() || typeLabel(record.prod_type) }}</span>
      </template>
      <template #condition="{ record }">
        <span>{{ record.condition === 'used' ? '二手' : '全新' }}</span>
      </template>
      <template #delivery="{ record }">
        <span>{{ deliveryLabel(record.delivery) }}</span>
      </template>
      <template #price="{ record }">
        <span>{{ record.price_fen ? '¥' + (record.price_fen / 100).toLocaleString() : '面议' }}</span>
      </template>
      <template #status="{ record }">
        <a-tag :color="statusColor(record.status)" size="small">{{ statusLabel(record.status) }}</a-tag>
      </template>
      <!-- 审核状态是与上架状态**正交**的独立维度：一件商品可以"审核通过但已下架"，
           也可以"被驳回但上架状态还是未上架"。两列分开才看得出区别。 -->
      <template #checkStatus="{ record }">
        <a-space :size="4" direction="vertical" fill>
          <a-tag :color="checkColor(record.check_status)" size="small">{{ checkLabel(record.check_status) }}</a-tag>
          <a-tooltip v-if="record.check_status === 'rejected' && record.check_reason" :content="record.check_reason">
            <span class="reject-reason">{{ record.check_reason }}</span>
          </a-tooltip>
        </a-space>
      </template>
      <template #actions="{ record }">
        <a-space :size="4">
          <template v-if="mode === 'recycle'">
            <a-button type="text" status="success" size="small" @click="handleRestore(record)">恢复</a-button>
          </template>
          <template v-else>
            <!-- 审核是独立动作：走 /review 端点，驳回必须填原因、服务端留审核人时间并写审计。
                 此前"通过/驳回"只是给 PUT 传一个 status，驳回写的是 removed——
                 与"卖家主动下架"同值，卖家分不清，也没有原因和审计。 -->
            <template v-if="record.check_status === 'pending' || record.check_status === 'rejected'">
              <a-button type="text" status="success" size="small" @click="handleReview(record, 'passed')">通过</a-button>
              <a-button type="text" status="danger" size="small" @click="openReject(record)">驳回</a-button>
            </template>
            <a-button type="text" size="small" @click="openForm(record)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="handleDelete(record)">删除</a-button>
          </template>
        </a-space>
      </template>
      <template #empty>
        <a-empty description="暂无商品" />
      </template>
    </CrudList>

    <!-- 新增 / 编辑弹窗 -->
    <a-modal v-model:visible="formVisible" :title="formEdit ? '编辑商品' : '新增商品'" :width="'min(520px, 94vw)'" :on-before-cancel="guardClose">
      <a-form :model="form" layout="vertical" class="dialog-form">
        <a-form-item label="商品名称" required>
          <a-input v-model="form.title" :aria-required="true" placeholder="如：工业级六旋翼无人机 X6-28L" allow-clear style="width: 100%" />
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
        <!-- 交付方式决定买家下单时是否必须填收货地址（自提不需要）：
             与后端 service.OrderNeedsReceiver 同一套判定 -->
        <a-form-item label="交付方式">
          <a-select v-model="form.delivery" style="width: 100%">
            <a-option label="未指定（按商品类型判断）" value="" />
            <a-option label="自提" value="pickup" />
            <a-option label="同城配送" value="city" />
            <a-option label="物流发货" value="logistics" />
            <a-option label="可协商" value="negotiable" />
          </a-select>
        </a-form-item>
        <!-- 价格方式必须显式选：此前 price_fen=0 一个值同时表示"面议"和"填了 0 元" -->
        <a-form-item label="价格方式">
          <a-radio-group v-model="form.priceMode" type="button">
            <a-radio value="fixed">明码标价</a-radio>
            <a-radio value="negotiable">面议</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item label="价格(元)">
          <a-input
            ref="priceRef"
            v-model="form.priceYuan"
            type="number"
            :disabled="form.priceMode === 'negotiable'"
            :placeholder="form.priceMode === 'negotiable' ? '面议商品无需填价' : '0.00'"
            style="width: 100%"
          />
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
            :before-upload="beforeUpload"
            :custom-request="uploadImage"
            @change="onImageChange"
          />
        </a-form-item>
        <!-- 详情图：与「商品图片」（顶部图集，首图作列表封面）分工不同——
             这组铺在详情页往下翻的位置，用于内部结构/铭牌/检测报告/实拍细节 -->
        <a-form-item label="详情图">
          <a-upload
            :file-list="detailImageList"
            list-type="picture-card"
            :limit="9"
            :before-upload="beforeUpload"
            :custom-request="uploadImage"
            @change="onDetailImageChange"
          />
          <template #extra>
            <span class="form-extra">选填，最多 9 张。与上方商品图片是两组独立的图，会铺在商品详情页下方。</span>
          </template>
        </a-form-item>
        <a-form-item label="卖家">
          <a-input v-model="form.seller_name" placeholder="默认平台自营" allow-clear style="width: 100%" />
        </a-form-item>
        <a-form-item label="描述">
          <RichEditor v-model="form.description" />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="handleCancel">取消</a-button>
        <a-button type="primary" :loading="formLoading" @click="submitForm">保存</a-button>
      </template>
    </a-modal>

    <!-- 驳回商品：原因必填，会展示给卖家（「我的发布」里可见） -->
    <a-modal
      v-model:visible="rejectVisible"
      title="驳回商品"
      :width="'min(480px, 94vw)'"
      :ok-loading="rejectSubmitting"
      ok-text="确认驳回"
      @ok="submitReject"
    >
      <p class="reject-tip">驳回原因会展示给卖家，请写清楚需要补充或修改什么。</p>
      <p class="reject-title">{{ rejectTarget && rejectTarget.title }}</p>
      <a-textarea
        v-model="rejectReason"
        placeholder="如：型号铭牌照片不清晰，请重新上传"
        :max-length="200"
        show-word-limit
        :auto-size="{ minRows: 3, maxRows: 6 }"
      />
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
import RichEditor from '@/components/RichEditor.vue'

const crudRef = ref()
const api = useAdminApi('products')
// 'active' = 商品管理（在库）/ 'recycle' = 回收站（已下架进回收站的行）
const mode = ref('active')

const typeLabel = (t) => ({ drone: '整机', part: '配件', repair: '维修服务', aerial: '航拍服务', test_fly: '试飞测试', calibration: '检测标定', airspace: '空域协调' }[t] || t || '-')
// 商品状态：pending=待审核（用户发布，通过后才上架）/ listed=在售 / sold=已售 / removed=已下架
const statusLabel = (s) => ({ pending: '待审核', listed: '在售', sold: '已售', removed: '已下架' }[s] || s || '-')
const deliveryLabel = (d) => ({ pickup: '自提', city: '同城配送', logistics: '物流发货', negotiable: '可协商' }[d] || '未指定')
const statusColor = (s) => ({ pending: 'orange', listed: 'green', sold: 'gray', removed: 'gray' }[s] || 'gray')

// 批量上架 / 批量下架：走后端单次批量端点。
//
// 此前是逐行 PUT 整行（`api.update(row.id, { ...row, status })`）——只改一个 status
// 却把所有列写回去，并发编辑时后写覆盖先写。后端现在是一条条件 UPDATE：
// 只动 status 列，跳过已售与回收站的行，置为 listed 时还要求已过审。
const batchSetStatus = (ids, status) =>
  axios.post('/api/v1/admin/products/batch-status', { ids, status })

const batchActions = [
  { key: 'list', label: '批量上架', status: 'success', bulkApi: (ids) => batchSetStatus(ids, 'listed') },
  { key: 'remove', label: '批量下架', status: 'warning', bulkApi: (ids) => batchSetStatus(ids, 'removed') }
]

const searchFields = [
  { key: 'status', label: '状态', type: 'select', width: 120, options: [
    { value: '', label: '全部状态' },
    { value: 'pending', label: '待审核' },
    { value: 'listed', label: '在售' },
    { value: 'sold', label: '已售' },
    { value: 'removed', label: '已下架' }
  ]},
  { key: 'check_status', label: '审核状态', type: 'select', width: 130, options: [
    { value: '', label: '全部审核状态' },
    { value: 'pending', label: '待审核' },
    { value: 'passed', label: '已通过' },
    { value: 'rejected', label: '已驳回' }
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
  { title: '交付方式', dataIndex: 'delivery', slotName: 'delivery', width: 100 },
  { title: '价格(元)', dataIndex: 'price_fen', slotName: 'price', width: 110, align: 'right' },
  { title: '上架状态', dataIndex: 'status', slotName: 'status', width: 90 },
  { title: '审核状态', dataIndex: 'check_status', slotName: 'checkStatus', width: 150 },
  { title: '卖家', dataIndex: 'seller_name', width: 120 },
  { title: '操作', slotName: 'actions', width: 140, fixed: 'right' },
]

const formVisible = ref(false)
const formEdit = ref(false)
const formLoading = ref(false)
const priceRef = ref()
const form = reactive({ id: '', title: '', prod_type: 'drone', brand: '', model: '', condition: 'new', delivery: '', priceMode: 'fixed', priceYuan: '', status: 'listed', description: '', seller_name: '', images: [], detail_images: [] })
const imageList = reactive([])
// 详情图列表（详情区长图）。与 imageList（顶部图集）独立。
const detailImageList = reactive([])

const resetForm = () => {
  form.id = ''; form.title = ''; form.prod_type = 'drone'; form.brand = ''; form.model = ''
  form.condition = 'new'; form.delivery = ''; form.priceMode = 'fixed'; form.priceYuan = ''; form.status = 'listed'; form.description = ''
  form.seller_name = ''; form.images = []; form.detail_images = []
  imageList.length = 0
  detailImageList.length = 0
}

const openForm = (row) => {
  resetForm()
  if (row) {
    formEdit.value = true
    form.id = row.id
    form.title = row.title || ''; form.prod_type = row.prod_type || 'drone'
    form.brand = row.brand || ''; form.model = row.model || ''
    form.condition = row.condition || 'new'
    form.delivery = row.delivery || ''
    form.priceMode = row.price_mode || 'fixed'
    // 面议商品价格恒为 0，输入框留空而不是显示 0
    form.priceYuan = form.priceMode === 'negotiable' ? '' : ((row.price_fen || 0) / 100).toString()
    form.status = row.status || 'listed'; form.description = row.description || ''
    form.seller_name = row.seller_name || ''; form.images = row.images || []
    form.images.forEach(u => imageList.push({ name: u.split('/').pop(), url: u }))
    form.detail_images = row.detail_images || []
    form.detail_images.forEach(u => detailImageList.push({ name: u.split('/').pop(), url: u }))
  } else {
    formEdit.value = false
  }
  formSnapshot = JSON.stringify(form)
  formVisible.value = true
}

// 图片上传（/api/v1/upload 返回相对 URL）
// 注意：Arco custom-request 的参数是 fileItem，原生 File 在 fileItem.file 上
// 上传前校验：仅图片类型、单张不超过 5MB
const beforeUpload = (file) => {
  if (!file.type || !file.type.startsWith('image/')) { Message.warning('仅支持上传图片文件'); return false }
  if (file.size > 5 * 1024 * 1024) { Message.warning('图片大小不能超过 5MB'); return false }
  return true
}

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
// 注意：f.url 是 Arco 的本地 blob 预览地址，不能入库；真实地址在响应 data.url 里。
// 仅排除上传中（uploading）的 blob 项，编辑态旧照片（无 response 但 url 是真实地址）保留。
const onImageChange = (fileList) => {
  imageList.length = 0
  imageList.push(...fileList)
  form.images = fileList
    .map((f) => f.response?.data?.url || f.response?.url || (f.status === 'uploading' ? '' : f.url))
    .filter(Boolean)
}

// 详情图列表变化：与 onImageChange 同逻辑，只是写到 form.detail_images
const onDetailImageChange = (fileList) => {
  detailImageList.length = 0
  detailImageList.push(...fileList)
  form.detail_images = fileList
    .map((f) => f.response?.data?.url || f.response?.url || (f.status === 'uploading' ? '' : f.url))
    .filter(Boolean)
}

const submitForm = async () => {
  if (!form.title) { Message.warning('请输入商品名称'); return }
  // 价格校验：NaN/Infinity/负数/超大值一律拦截，避免 null/负数/溢出值入库
  const negotiable = form.priceMode === 'negotiable'
  const price = negotiable ? 0 : Number(form.priceYuan)
  if (negotiable && String(form.priceYuan || '').trim() !== '') {
    Message.error('选择面议时请清空价格')
    return
  }
  if (!negotiable && (!Number.isFinite(price) || price <= 0 || price > 100000000)) {
    // 明码标价必须给出真实价格：0 元此前既表示"面议"又表示"填了 0"，现在必须堵死
    Message.error('明码标价需为 0-100000000 之间、且大于 0 的数字（元）；不标价请选「面议」')
    priceRef.value && priceRef.value.focus && priceRef.value.focus()
    return
  }
  formLoading.value = true
  const payload = {
    title: form.title,
    prod_type: form.prod_type,
    brand: form.brand,
    model: form.model,
    condition: form.condition,
    delivery: form.delivery,
    price_mode: form.priceMode,
    price_fen: Math.round(price * 100),
    status: form.status,
    description: form.description,
    seller_name: form.seller_name,
    images: form.images,
    detail_images: form.detail_images
  }
  try {
    if (formEdit.value) await api.update(form.id, payload)
    else await api.create(payload)
    Message.success('保存成功')
    formSnapshot = JSON.stringify(form)
    formVisible.value = false
    crudRef.value?.reload()
  } catch (e) { Message.error(e?.response?.data?.message || '保存失败') }
  finally { formLoading.value = false }
}

// 未保存守卫：X/遮罩/Esc 关闭前经 on-before-cancel 校验，若表单有改动则确认，避免输入全丢
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
const handleCancel = () => {
  if (guardClose()) formVisible.value = false
}

// 审核状态展示。它与上架状态是两个正交维度：
// 审核通过 → 上架状态自动变在售；被驳回的商品上架状态仍是"未上架"，不会被标成"已下架"。
const checkLabel = (s) => ({ pending: '待审核', passed: '已通过', rejected: '已驳回' }[s] || s || '待审核')
const checkColor = (s) => ({ pending: 'orange', passed: 'green', rejected: 'red' }[s] || 'gray')

// 审核走独立端点：服务端会写 reviewed_at/reviewed_by 并落审计。
// 此前"通过/驳回"只是给 PUT 传一个 status，驳回写 removed —— 与"卖家主动下架"同值。
const handleReview = async (row, checkStatus, reason = '') => {
  try {
    await axios.post(`/api/v1/admin/products/${encodeURIComponent(row.id)}/review`, {
      check_status: checkStatus,
      check_reason: reason
    })
    Message.success(checkStatus === 'passed' ? '已通过，商品已上架' : '已驳回')
    crudRef.value?.reload()
    return true
  } catch (e) {
    Message.error(e?.response?.data?.message || '操作失败')
    return false
  }
}

// 驳回原因：服务端强制必填（service.ReviewProduct 返回 400），这里做前置校验与输入体验。
// 原因会展示给卖家，所以文案要提示"写清楚需要改什么"。
const rejectVisible = ref(false)
const rejectTarget = ref(null)
const rejectReason = ref('')
const rejectSubmitting = ref(false)
const openReject = (row) => {
  rejectTarget.value = row
  rejectReason.value = ''
  rejectVisible.value = true
}
const submitReject = async () => {
  const reason = rejectReason.value.trim()
  if (!reason) {
    Message.warning('请填写驳回原因')
    return
  }
  rejectSubmitting.value = true
  try {
    if (await handleReview(rejectTarget.value, 'rejected', reason)) rejectVisible.value = false
  } finally {
    rejectSubmitting.value = false
  }
}

// 删除 = 进回收站（后端软删除，不是物理删除）。
// 有进行中订单时后端返回 409，必须把原因显示出来，而不是笼统的"删除失败"。
const handleDelete = (row) => {
  Modal.confirm({
    title: '移入回收站',
    content: `确定将商品「${row.title}」移入回收站吗？可在回收站标签页恢复。`,
    okText: '移入回收站',
    cancelText: '取消',
    onOk: async () => {
      try {
        await api.delete(row.id)
        Message.success('已移入回收站')
        crudRef.value?.reload()
      } catch (e) {
        Message.error(e?.response?.data?.message || '删除失败')
      }
    }
  })
}

// 回收站还原
const handleRestore = async (row) => {
  try {
    await axios.post(`/api/v1/admin/products/${encodeURIComponent(row.id)}/restore`)
    Message.success('已恢复')
    crudRef.value?.reload()
  } catch (e) {
    Message.error(e?.response?.data?.message || '恢复失败')
  }
}
</script>

<style scoped>
.page { max-width: 1400px; margin: 0 auto; }
.mode-switch { margin-bottom: 12px; }
.form-extra { color: #86909c; font-size: 12px; }
.reject-tip { color: #86909c; font-size: 13px; margin: 0 0 8px; }
.reject-title { color: #1d2129; font-weight: 500; margin: 0 0 10px; }
.reject-reason {
  display: inline-block; max-width: 140px; overflow: hidden; text-overflow: ellipsis;
  white-space: nowrap; color: #f53f3f; font-size: 12px;
}

.cover-img { border-radius: 6px; overflow: hidden; }
.no-image { color: #C9CDD4; font-size: 12px; }

.dialog-form :deep(.arco-form-item-label-col) { min-width: 88px; }
</style>
