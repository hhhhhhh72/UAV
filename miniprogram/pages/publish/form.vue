<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">{{ typeConfig.name }}</view>
    </view>

    <!-- 商品/课程两步进度条 -->
    <view v-if="typeConfig.stepped" class="pub-progress">
      <view class="pub-progress-top">
        <text>{{ typeConfig.steps[step] }}</text>
        <text>{{ step + 1 }}/{{ typeConfig.steps.length }}</text>
      </view>
      <view class="pub-progress-bar">
        <!-- 宽度按步骤数推导：写死 50/100 只在"恰好两步"时正确 -->
        <view :style="{ width: ((step + 1) / typeConfig.steps.length * 100) + '%' }"></view>
      </view>
    </view>

    <!-- 表单头部 -->
    <view class="pub-form-intro">
      <view class="pub-form-intro-h2">{{ typeConfig.name }}</view>
      <view class="pub-form-intro-p">{{ typeConfig.desc }}</view>
    </view>

    <!-- 分组区块 -->
    <view v-for="(section, si) in visibleSections" :key="si" class="pub-section">
      <view class="pub-section-title">{{ section.title }}</view>
      <view v-if="section.note" class="pub-section-note">{{ section.note }}</view>
      <view class="pub-form-card">
        <view v-for="(field, fi) in section.fields" :key="fi" class="pub-field">
          <view class="pub-field-label">{{ field[1] }}<text v-if="field[4]" class="pub-required">*</text></view>

          <!-- 分段选择：选项少且互斥（发布类型：商品 / 服务能力），直接平铺，
               省掉"点开抽屉再选一项再关掉"的三步 -->
          <view v-if="field[3] === 'segment'" class="pub-segment">
            <view
              v-for="opt in field[5]"
              :key="opt"
              class="pub-segment-item"
              :class="{ 'pub-segment-item--on': values[field[0]] === opt }"
              hover-class="pub-fade"
              @tap="pickSegment(field[0], opt)"
            >{{ opt }}</view>
          </view>

          <!-- 选择型：打开底部抽屉 -->
          <view
            v-else-if="field[3] === 'select'"
            class="pub-select-field"
            @tap="openSheet(field[0])"
          >
            <text :class="values[field[0]] ? 'pub-select-value' : 'pub-placeholder'">
              {{ values[field[0]] || field[2] }}
            </text>
            <text class="pub-arrow">›</text>
          </view>

          <!-- 多行文本 -->
          <textarea
            v-else-if="field[3] === 'textarea'"
            class="pub-input pub-input--textarea"
            :value="values[field[0]]"
            :placeholder="field[2]"
            placeholder-class="pub-placeholder"
            maxlength="300"
            @input="onInput(field[0], $event)"
          ></textarea>

          <!-- 单行文本 -->
          <input
            v-else
            class="pub-input"
            :value="values[field[0]]"
            :placeholder="field[2]"
            placeholder-class="pub-placeholder"
            :type="inputType(field[0])"
            @input="onInput(field[0], $event)"
          />
          <text v-if="unitOf(field[0])" class="pub-field-hint">{{ unitOf(field[0]) }}</text>
        </view>

        <!-- 上传区 -->
        <view v-if="section.upload">
          <view class="pub-upload-row">
            <!-- 删除按钮与详情图、需求附件保持一致。
                 此前顶部图集**只能加不能删**：选错一张、或编辑旧商品时回填了一张
                 打不开的图（如已失效的 http://tmp/... 本地路径），就没法去掉它，
                 只能整条放弃重发。 -->
            <view v-for="(photo, i) in photos" :key="i" class="pub-photo">
              <image v-if="photo && photo.src" :src="photo.src" mode="aspectFill" class="pub-photo-img" />
              <!-- 首图即列表封面：给它一个角标，用户才知道"哪一张会被别人第一眼看到"，
                   否则顺序调整没有任何可见反馈（列表里哪张在前全靠猜）。 -->
              <text v-if="i === 0" class="pub-photo-cover">封面</text>
              <text class="pub-file-del" @tap.stop="removePhoto(i)">×</text>
            </view>
            <view class="pub-add-photo" hover-class="pub-fade" @tap="addPhoto">＋</view>
          </view>
          <view class="pub-upload-tip">建议上传清晰实拍图。首图将作为列表封面——<text class="pub-upload-tip-strong">推荐 1:1（方图）或 3:4（竖图）</text>，横长图在列表里会被压扁、标题容易被挤没</view>
          <!-- 比例提示：只提示不拦截。供给大厅是按图片自身比例做瀑布流的，
               横长图仍能正常显示，只是卡片会变得很矮——拦下来反而挡住了正常的横构图实拍图。 -->
          <view v-if="coverHints.length" class="pub-upload-warn">
            <text v-for="(h, i) in coverHints" :key="i" class="pub-upload-warn-line">{{ h }}</text>
          </view>
          <!-- 需求附件材料：与现场资料同分区（图片下方，仅需求展示；PDF/图片 ≤10MB，最多 3 份） -->
          <template v-if="type === 'demand'">
            <view class="pub-upload-tip pub-upload-tip--files">附件材料（选填，PDF/图片，单个 ≤10MB，详情可下载）</view>
            <view class="pub-upload-row">
              <view v-for="(f, i) in files" :key="i" class="pub-photo pub-file">
                <text class="pub-file-name">{{ f.name }}</text>
                <text class="pub-file-del" @tap.stop="removeFile(i)">×</text>
              </view>
              <view v-if="files.length < 3" class="pub-add-photo" hover-class="pub-fade" @tap="addFile">＋</view>
            </view>
          </template>
        </view>

        <!-- 详情图上传区（仅商品）：与上面的顶部图集**分工不同**——
             上面那组是封面（列表/详情页第一眼看到的，首图做列表封面，上限 5 张）；
             这组是详情区长图（内部结构/铭牌/检测报告/实拍细节），铺在详情页往下翻的位置。 -->
        <view v-if="section.uploadDetail">
          <view class="pub-upload-tip">详情图（选填，最多 9 张）：会铺在商品详情页下方，用于展示内部结构、铭牌、检测报告、实拍细节</view>
          <view class="pub-upload-row">
            <view v-for="(photo, i) in detailPhotos" :key="i" class="pub-photo">
              <image v-if="photo && photo.src" :src="photo.src" mode="aspectFill" class="pub-photo-img" />
              <text class="pub-file-del" @tap.stop="removeDetailPhoto(i)">×</text>
            </view>
            <view v-if="detailPhotos.length < 9" class="pub-add-photo" hover-class="pub-fade" @tap="addDetailPhoto">＋</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 固定底部操作区 -->
    <view class="pub-sticky">
      <!-- 上一步：分步表单此前**只有前进没有后退**（step 全项目只被赋值 1 和 0，
           没有任何递减），填到第二步发现第一步的类型选错就只能退出重来。
           pkg-eco/pages/enterprise/register.vue 的多步表单一直有这个按钮，
           注释还写着"与发布页同款"——这里补上，两边才真的一致。 -->
      <view
        v-if="step > 0"
        class="pub-btn pub-btn--ghost"
        hover-class="pub-btn--active"
        @tap="prevAction"
      >上一步</view>
      <view class="pub-btn pub-btn--ghost" hover-class="pub-btn--active" @tap="saveDraft">保存草稿</view>
      <view
        class="pub-btn pub-btn--primary"
        :class="{ 'pub-btn--busy': editLoading }"
        hover-class="pub-btn--active"
        @tap="nextAction"
      >{{ editLoading ? '加载中…' : primaryText }}</view>
    </view>

    <!-- 选择底部抽屉 -->
    <view v-if="currentSheet" class="pub-overlay" @tap="closeSheet">
      <view class="pub-sheet" @tap.stop>
        <view class="pub-grab"></view>
        <view class="pub-sheet-head">
          <view class="pub-sheet-head-title">选择{{ currentSheet.label }}</view>
          <view class="pub-sheet-cancel" @tap="closeSheet">取消</view>
        </view>
        <view
          v-for="opt in currentSheet.options"
          :key="opt"
          class="pub-option"
          :class="{ 'pub-option--selected': values[currentSheet.id] === opt }"
          @tap="pickOption(currentSheet.id, opt)"
        >
          <text>{{ opt }}</text>
          <text v-if="values[currentSheet.id] === opt" class="pub-option-check">✓</text>
        </view>
      </view>
    </view>

    <!-- 底部黑色 toast -->
    <view v-if="toast" class="pub-toast">{{ toast }}</view>
  </view>
</template>

<script setup>
import { safeBack } from '../../utils/nav'
import { ref, computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { TYPES, getPost, upsertPost, draftPosts, saveFormState, prodTypeToOption } from '../../utils/publishData'
import { isServiceProdType } from '../../utils/enums'
import { useSafeTop } from '../../utils/safeTop'
import { requireLogin, request, getErrorMessage } from '../../utils/request'

const { topPad, initSafeTop } = useSafeTop(true)

const props = defineProps({}) // 无 props，页面通过路由参数驱动
void props

const type = ref('')
const step = ref(0)
const values = ref({})
const photos = ref([])
// 详情图（详情区长图）。与 photos（顶部图集）是两组独立的图，见模板注释。
const detailPhotos = ref([])
const sheetId = ref('')
const toast = ref('')
const toastTimer = ref(null)
const resumeId = ref('')
// 编辑**后端已发布**的商品时携带其 id（与 resumeId 的"本地草稿"语义不同）：
// 预览页据此走 PATCH /api/v1/products/{id} 而不是 POST。
const editBackendId = ref('')
const editLoading = ref(false)

const typeConfig = computed(() => TYPES[type.value] || null)

// 发布类型：'商品' | '服务能力'，由第一步的「发布类型」显式选定。
// 取代此前"从选中的类型选项反推是不是服务"的做法——那种反推正是表单混乱的根源：
// 成色/品牌/物流对服务没意义，服务类目/区域/计价单位对实物没意义，
// 硬塞进一个表单只能靠一层层条件显隐打补丁。
const isServiceProduct = computed(() => type.value === 'product' && values.value.bizKind === '服务能力')

// 字段在当前选择下是否应出现。渲染与必填校验共用本函数，
// 避免出现"界面不显示、校验却仍拦截提交"的割裂。
function fieldVisible(f) {
  // 分支字段：f[7] 是 scope（'goods' | 'service'），不属于当前发布类型就不渲染。
  const scope = f[7]
  if (scope) {
    // 还没选发布类型时，两分支的字段一个都不显示——
    // 否则默认按"商品"渲染，用户会先看到一半用不上的字段再被换掉。
    if (!String(values.value.bizKind || '').trim()) return false
    if (scope === 'goods' && isServiceProduct.value) return false
    if (scope === 'service' && !isServiceProduct.value) return false
  }
  // 自定义服务区域：只有选了「其他」才出现。
  // 它标了 required=true，因此"必填"只在可见时生效——隐藏时不会被必填校验拦住。
  if (f[0] === 'regionOther') return String(values.value.region || '').trim() === '其他'
  return true
}

const visibleSections = computed(() => {
  const t = typeConfig.value
  if (!t) return []
  const secs = t.stepped
    ? step.value === 0 ? [t.sections[0]] : t.sections.slice(1)
    : t.sections
  // 复制分区对象（保留 upload / uploadDetail / note），只替换 fields
  return secs.map((s) => Object.assign({}, s, { fields: (s.fields || []).filter(fieldVisible) }))
})

// 某字段此刻是否真的在表单里（已按步骤 + 服务类隐藏规则过滤）。
// 用于提交前的语义自检：只对**用户看得见**的字段做校验。
// 曾经的 bug：价格方式自检无条件执行，而「价格方式」在第二步才渲染，
// 第一步点"下一步"就被一句"请选择价格方式"拦住，可界面上根本没有这个字段。
function fieldInView(id) {
  return visibleSections.value.some((s) => (s.fields || []).some((f) => f[0] === id))
}

const currentSheet = computed(() => {
  const t = typeConfig.value
  if (!t || !sheetId.value) return null
  let found = null
  t.sections.forEach((s) => {
    ;(s.fields || []).forEach((f) => {
      if (f[0] === sheetId.value) found = { id: f[0], label: f[1], options: f[5] || [] }
    })
  })
  return found
})

const primaryText = computed(() => {
  const t = typeConfig.value
  if (!t) return '预览发布'
  if (t.stepped && step.value === 0) return '下一步'
  return '预览发布'
})

// 数值型输入
function inputType(id) {
  const numeric = ['budget', 'budget_min', 'budget_max', 'pilot_count', 'price', 'stock', 'quota', 'duration', 'contact']
  return numeric.includes(id) ? 'number' : 'text'
}
function unitOf(id) {
  if (id === 'budget' || id === 'budget_min' || id === 'budget_max' || id === 'price') return '元'
  if (id === 'pilot_count') return '人'
  if (id === 'stock') return '件'
  if (id === 'quota') return '人'
  return ''
}

function onInput(id, e) {
  values.value[id] = e.detail.value
}
function openSheet(id) {
  sheetId.value = id
}
function closeSheet() {
  sheetId.value = ''
}
function pickOption(id, val) {
  values.value[id] = val
  // 服务区域改选为非「其他」时清掉此前手填的自定义值：
  // 那栏随即隐藏，留着旧值会把它照样提交上去。
  if (id === 'region' && val !== '其他') values.value.regionOther = ''
  sheetId.value = ''
}

// 分段选择（发布类型 商品/服务能力）。
// 切换时必须清掉另一分支已填的值：不清的话，先填商品再切到服务，
// 成色/品牌/交付方式会跟着提交——界面上看不见，后端却收到一堆该类型不该有的字段。
const GOODS_ONLY_KEYS = ['productType', 'condition', 'brand', 'delivery']
const SERVICE_ONLY_KEYS = ['serviceType', 'category', 'unit', 'region', 'regionOther']
function pickSegment(id, val) {
  if (id === 'bizKind' && String(values.value.bizKind || '') !== val) {
    const drop = val === '服务能力' ? GOODS_ONLY_KEYS : SERVICE_ONLY_KEYS
    drop.forEach((k) => { values.value[k] = '' })
  }
  values.value[id] = val
}
/* 需求附件：选文件（PDF/图片）→ 存临时路径，预览页发布时上传 */
const files = ref([])
function addFile() {
  if (files.value.length >= 3) return
  uni.chooseMessageFile({
    count: 3 - files.value.length,
    type: 'file',
    success: (res) => {
      const picked = (res.tempFiles || []).map((f) => ({ name: (f.name || f.path || '').split('/').pop(), path: f.path })).filter((f) => f.path)
      files.value.push(...picked)
    },
  })
}
function removeFile(i) { files.value.splice(i, 1) }

/* 主图比例建议：1:1（方图）～3:4（竖图）。
   为什么是这两个：供给大厅是推荐流，按图片自身比例瀑布流展示（见 utils/hallData.js
   的 coverRatio）。横长图会把卡片压得很矮、标题被挤没；过于窄长的竖图会占掉一整屏。
   阈值比大厅的裁剪区间 [0.5, 2.0] 略紧，是"建议"而非"限制"——超出只是不理想，
   仍然能发布、也仍然按它自己的比例正常渲染。 */
const COVER_RATIO_MIN = 0.72
const COVER_RATIO_MAX = 1.08
function ratioHint(photo) {
  const w = photo && photo.w
  const h = photo && photo.h
  if (!w || !h) return ''
  const r = w / h
  if (r > COVER_RATIO_MAX) {
    return '第 ' + (photos.value.indexOf(photo) + 1) + ' 张是横图（' + w + '×' + h + '），列表里会显得很矮，建议裁成 1:1 或 3:4'
  }
  if (r < COVER_RATIO_MIN) {
    return '第 ' + (photos.value.indexOf(photo) + 1) + ' 张偏高窄（' + w + '×' + h + '），列表里会占掉很高一屏，建议裁成 3:4'
  }
  return ''
}
/* coverHints 只对已经读到尺寸的图给提示；读不到（格式不支持等）就当作合规，不打扰用户。 */
const coverHints = computed(() => photos.value.map(ratioHint).filter(Boolean))

function addPhoto() {
  if (photos.value.length >= 5) return
  const pick = (paths) => {
    const start = photos.value.length
    photos.value.push(...paths.map((p) => ({ src: p, w: 0, h: 0 })))
    showToast('已添加 ' + paths.length + ' 张图片')
    // 读原图尺寸用于比例提示。失败静默跳过——提示是锦上添花，不能因此挡住发布。
    paths.forEach((src, i) => {
      if (typeof uni.getImageInfo !== 'function') return
      uni.getImageInfo({
        src,
        success: (info) => {
          const photo = photos.value[start + i]
          if (photo) {
            photo.w = info.width || 0
            photo.h = info.height || 0
          }
        },
      })
    })
  }
  if (typeof uni.chooseMedia === 'function') {
    uni.chooseMedia({
      count: 5 - photos.value.length,
      mediaType: ['image'],
      success: (res) => pick(res.tempFiles.map((f) => f.tempFilePath)),
    })
  } else {
    uni.chooseImage({
      count: 5 - photos.value.length,
      success: (res) => pick(res.tempFilePaths),
    })
  }
}
const DETAIL_IMG_MAX = 9

function removePhoto(i) { photos.value.splice(i, 1) }
function removeDetailPhoto(i) { detailPhotos.value.splice(i, 1) }

// 详情图选图：上限 9 张（比顶部图集宽松——详情图通常是长图序列）
function addDetailPhoto() {
  if (detailPhotos.value.length >= DETAIL_IMG_MAX) {
    showToast('详情图最多 ' + DETAIL_IMG_MAX + ' 张')
    return
  }
  const pick = (paths) => {
    detailPhotos.value.push(...paths.map((p) => ({ src: p })))
    showToast('已添加 ' + paths.length + ' 张详情图')
  }
  const count = DETAIL_IMG_MAX - detailPhotos.value.length
  if (typeof uni.chooseMedia === 'function') {
    uni.chooseMedia({
      count,
      mediaType: ['image'],
      success: (res) => pick(res.tempFiles.map((f) => f.tempFilePath)),
    })
  } else {
    uni.chooseImage({
      count,
      success: (res) => pick(res.tempFilePaths),
    })
  }
}

function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => {
    toast.value = ''
  }, 2200)
}

// 当前步骤必填校验
function requiredMissing() {
  if (!typeConfig.value) return []
  const missing = []
  // 用 visibleSections：它已按当前步骤取分区（与原 candidates 等价），
  // 并剔除服务类目下隐藏的成色——否则界面藏着、校验还在拦。
  visibleSections.value.forEach((s) => {
    ;(s.fields || []).forEach((f) => {
      if (f[4] && !String(values.value[f[0]] || '').trim()) missing.push(f[1])
    })
  })
  return missing
}

// 格式校验（手机号/座机、非负数字）：仅校验已填写的带 rule 字段；选填留空不拦截
const PHONE_RE = /^(?:\+?86)?1[3-9]\d{9}$|^0\d{2,3}-?\d{7,8}$/
const NUMBER_RE = /^\d+(?:\.\d{1,2})?$/
function formatInvalid() {
  const t = typeConfig.value
  if (!t) return []
  const bad = []
  t.sections.forEach((s) => {
    ;(s.fields || []).forEach((f) => {
      const v = String(values.value[f[0]] || '').trim()
      const rule = f[6]
      if (!v || !rule) return
      if (rule === 'phone' && !PHONE_RE.test(v)) bad.push(f[1] + '（电话格式）')
      else if (rule === 'number' && !NUMBER_RE.test(v)) bad.push(f[1] + '（需为数字）')
    })
  })
  return bad
}

// 回上一步：只动步骤，不清 values —— 退回去改完再前进，之前填的仍在。
function prevAction() {
  if (step.value > 0) step.value -= 1
}

function nextAction() {
  // 编辑已发布商品时，回填未完成不允许提交，否则会把空表单覆盖上去
  if (editLoading.value) return
  // 需求预算区间自检：下限 > 上限直接拦截（与后端 VALIDATION_ERROR 一致，提交前提示）
  if (type.value === 'demand') {
    const mn = Number(values.value.budget_min) || 0
    const mx = Number(values.value.budget_max) || 0
    if (mn > 0 && mx > 0 && mn > mx) {
      showToast('预算下限不能大于上限')
      return
    }
  }
  // 商品价格方式自检（与后端 service.normalizeAndValidate 同一套口径，提交前就拦住）。
  // 必须限定 fieldInView('priceMode')：stepped 表单第一步只填基础信息，价格字段在第二步，
  // 此时不该拿一个还没露面的字段拦人。
  if (type.value === 'product' && fieldInView('priceMode')) {
    const mode = String(values.value.priceMode || '').trim()
    const price = String(values.value.price || '').trim()
    if (!mode) {
      showToast('请选择价格方式')
      return
    }
    if (mode === '明码标价' && !(Number(price) > 0)) {
      showToast('明码标价必须填写大于 0 的售价；不标价请选择「面议」')
      return
    }
    if (mode === '面议' && price) {
      showToast('选择面议时售价请留空')
      return
    }
  }
  const missing = requiredMissing()
  if (missing.length) {
    showToast('请先填写：' + missing[0])
    return
  }
  const bad = formatInvalid()
  if (bad.length) {
    showToast('请检查：' + bad[0])
    return
  }
  const t = typeConfig.value
  if (t.stepped && step.value === 0) {
    step.value = 1
    return
  }
  goPreview()
}

// 保存草稿：写入本地“我的草稿”
function saveDraft() {
  const t = typeConfig.value
  if (!t) return
  const id = resumeId.value || 'draft-' + Date.now()
  // 记住草稿 id：继续「预览发布」时携带该 id，发布时 upsert 覆盖草稿，
  // 避免同内容"草稿 + 发布"两条并存（曾现：存草稿后发布，my-posts 显示两条）
  resumeId.value = id
  // 草稿同时记录附件临时路径（恢复草稿时回填）
  values.value.__files = files.value.map((f) => ({ name: f.name, path: f.path }))
  const post = {
    id,
    type: type.value,
    label: t.short,
    title: values.value.title || '未命名发布内容',
    status: '草稿',
    statusKey: 'draft',
    date: '保存于 刚刚',
    meta: [],
    note: '内容已保存为草稿，可继续编辑后再次提交。',
    values: Object.assign({}, values.value),
    photoCount: photos.value.length,
    detailPhotoCount: detailPhotos.value.length,
  }
  upsertPost(post)
  showToast('草稿已保存，可在「我的草稿」继续编辑')
}

function goPreview() {
  const state = {
    type: type.value,
    step: step.value,
    values: Object.assign({}, values.value),
    photoCount: photos.value.length,
    detailPhotoCount: detailPhotos.value.length,
    // 图片真实临时路径（发布时上传到服务器；历史数据只有数量无路径，过滤为空）
    photos: photos.value.map((p) => p && p.src).filter(Boolean),
    // 详情图同样带临时路径过去（预览页负责上传）
    detailPhotos: detailPhotos.value.map((p) => p && p.src).filter(Boolean),
    // 需求附件临时路径（发布时上传，历史草稿无此字段为空）
    files: files.value.map((f) => f.path).filter(Boolean),
    resumeId: resumeId.value || '',
    // 非空表示这是"编辑已发布商品"，预览页走 PATCH
    editBackendId: editBackendId.value || '',
  }
  saveFormState(state)
  const q = encodeURIComponent(JSON.stringify(state))
  uni.navigateTo({ url: '/pages/publish/preview?state=' + q })
}

function goBack() {
  safeBack()
}

// 发布类型两个下拉各自的选项（预置服务区域要用它的第 6 位）
const REGION_OPTIONS = (TYPES.product.sections
  .reduce((acc, s) => acc.concat(s.fields || []), [])
  .find((f) => f[0] === 'region') || [])[5] || []
// 后端交付方式枚举 → 表单中文（与 preview.vue mapDelivery 互为逆映射）
const DELIVERY_LABELS = { pickup: '自提', city: '同城配送', logistics: '物流发货', negotiable: '可协商' }

// 编辑**后端已发布**的商品：从服务端拉取并回填（卖家在商品详情页点「编辑商品」进来，
// mall/detail.vue 跳 ?type=product&editId=xxx）。
//
// 此前这个函数**根本不存在**——onLoad 里调用了它却没有定义，卖家点「编辑商品」
// 必然抛 ReferenceError，页面填不上任何内容。同时 editLoading 声明了也从未使用。
async function loadBackendProduct(id) {
  editLoading.value = true
  try {
    const p = await request({ url: '/api/v1/products/' + encodeURIComponent(id) })
    // 分支由后端 prod_type 决定，回填后表单只显示该分支的字段
    const svc = isServiceProdType(p.prod_type)
    const vals = {
      bizKind: svc ? '服务能力' : '商品',
      // 商品与服务各有一个类型下拉，各自只在一侧出现，所以另一个留空
      productType: svc ? '' : prodTypeToOption(p.prod_type),
      serviceType: svc ? prodTypeToOption(p.prod_type) : '',
      title: p.title || '',
      condition: p.condition === 'used' ? '二手 95 新' : '全新',
      // 表单品牌/型号是合并输入，preview.vue 的 splitBrand 按 '/' 拆回去
      brand: [p.brand, p.model].filter(Boolean).join(' / '),
      category: p.category || '',
      priceMode: p.price_mode === 'negotiable' ? '面议' : '明码标价',
      price: p.price_fen > 0 ? String(p.price_fen / 100) : '',
      unit: p.unit || '',
      delivery: DELIVERY_LABELS[p.delivery] || '',
      region: '',
      regionOther: '',
      description: p.description || '',
    }
    // 服务区域：命中预置范围就直接用；否则落「其他」+ 手填原值。
    // 老数据里有"重庆及西南/川渝地区"这类自由文本，不回退就会在编辑保存时被静默清空。
    const raw = String(p.region || '').trim()
    if (raw) {
      if (REGION_OPTIONS.indexOf(raw) >= 0 && raw !== '其他') vals.region = raw
      else { vals.region = '其他'; vals.regionOther = raw }
    }
    values.value = vals
    // 图片：服务端存的是 URL，直接当 src 用（preview.vue 的 isServerImage 会跳过重复上传）
    photos.value = (Array.isArray(p.images) ? p.images : []).filter(Boolean).map((src) => ({ src }))
    detailPhotos.value = (Array.isArray(p.detail_images) ? p.detail_images : []).filter(Boolean).map((src) => ({ src }))
    step.value = 0
  } catch (e) {
    showToast(getErrorMessage(e) || '商品信息加载失败')
  } finally {
    editLoading.value = false
  }
}

onLoad((options) => {
  initSafeTop()
  const t = options && options.type
  if (t && TYPES[t]) {
    type.value = t
  }
  // 编辑**后端已发布**的商品：从服务端拉取回填（卖家在商品详情页点"编辑"进来）。
  // 保存后会退回待审核，协会重新审核通过才再次上架——见 service.UpdateMyProduct。
  if (options && options.editId) {
    editBackendId.value = String(options.editId)
    loadBackendProduct(editBackendId.value)
  }
  // 编辑已有草稿/发布（历史数据只存数量，无真实路径，恢复为占位）
  if (options && options.id) {
    const post = getPost(options.id)
    if (post) {
      resumeId.value = post.id
      values.value = Object.assign({}, post.values || {})
      photos.value = Array.from({ length: post.photoCount || 0 }, () => ({}))
      detailPhotos.value = Array.from({ length: post.detailPhotoCount || 0 }, () => ({}))
      const savedFiles = (post.values && post.values.__files) || []
      files.value = savedFiles.map((f) => ({ name: f.name, path: f.path })).filter((f) => f.path)
    }
  }
  // 恢复发布首页“继续编辑”的草稿（原型 resumeDraft 行为）
  if (options && options.resume) {
    const posts = draftPosts()
    if (posts.length) {
      const p = posts[0]
      type.value = p.type || t
      resumeId.value = p.id
      values.value = Object.assign({}, p.values || {})
      photos.value = Array.from({ length: p.photoCount || 0 }, () => ({}))
      detailPhotos.value = Array.from({ length: p.detailPhotoCount || 0 }, () => ({}))
    }
  }
  if (type.value && !typeConfig.value) type.value = ''
})

onShow(() => {
  // 未登录每次进入都拦截（从登录页返回会继续跳转，必须登录后才能使用发布表单）
  requireLogin('请先登录后再发布')
})
</script>

<style scoped>
@import './pub-style.css';

/* 分段选择（发布类型：商品 / 服务能力）。
   走品牌色实心表示选中，与底部主按钮同一套视觉语言。 */
.pub-segment {
  display: flex;
  gap: 10px;
  margin-top: 2px;
}
.pub-segment-item {
  flex: 1;
  height: 44px;
  line-height: 44px;
  text-align: center;
  font-size: 15px;
  font-weight: 600;
  color: #5B6B7C;
  background: #F0F5FA;
  border: 1px solid transparent;
  border-radius: 24rpx;
  box-sizing: border-box;
}
.pub-btn--busy { opacity: 0.6; }

/* 第二步的底部是「上一步 + 保存草稿 + 预览发布」三个按钮：
   .pub-btn 基类的水平内边距是 15px，三个按钮并排时「保存草稿」只剩 53px 文本区
   （4 个字 14px 需要 56px），窄屏（320pt）会溢出。
   这里只收窄**本页**底栏里的 ghost，不去改共享的 pub-style.css
   —— 那个文件被 16 个页面 @import，动它会波及返回/取消等其他按钮。 */
.pub-sticky .pub-btn--ghost {
  padding: 0 4px;
  min-width: 0;
  white-space: nowrap;
}
.pub-segment-item--on {
  color: #fff;
  background: #0A66C2;
  border-color: #0A66C2;
  box-shadow: 0 7px 14px rgba(10, 102, 194, 0.20);
}

.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
}
/* 封面角标：复用品牌深空蓝，不引入新色；圆角沿用紧凑控件尺度。
   样式写在**本页**而不是 pub-style.css —— 后者被 16 个页面 @import，
   改它会波及需求/服务等所有发布页。 */
.pub-photo-cover {
  position: absolute;
  left: 0;
  bottom: 0;
  padding: 2rpx 10rpx;
  font-size: 20rpx;
  line-height: 1.6;
  color: #fff;
  background: rgba(10, 102, 194, 0.92);
  border-radius: 0 10rpx 0 12rpx;
}
/* 比例提示：橙色系＝"需要注意但不阻塞"，与安全橙语义一致（绿色只代表成功）。 */
.pub-upload-tip-strong { color: #B54708; font-weight: 600; }
.pub-upload-warn {
  margin-top: 8rpx;
  padding: 12rpx 16rpx;
  background: #FFFAEB;
  border: 1rpx solid #FEDF89;
  border-radius: 12rpx;
}
.pub-upload-warn-line {
  display: block;
  font-size: 22rpx;
  line-height: 1.6;
  color: #B54708;
}
/* 需求附件：与现场资料图片同排的文案盒（复用 pub-photo 尺寸） */
.pub-file {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  padding: 0 14rpx;
  box-sizing: border-box;
}
.pub-file-name {
  max-width: 150rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 22rpx;
  color: #344054;
}
.pub-file-del {
  font-size: 30rpx;
  color: #D92D20;
  line-height: 1;
}
/* 缩略图上的删除按钮必须绝对定位。
   根因：.pub-photo 是 62×62 + overflow:hidden，而 .pub-photo-img 用 width/height:100%
   占满整格；删除按钮若留在正常流里会被排到格子**下方**并被裁掉——编译产物侧证：
   form.wxss 里 .pub-file-del 只有 font-size/color/line-height，没有任何 position。
   结果是"存在但看不见的按钮"：注释说补上了删除，用户实际依然删不掉选错的封面。
   需求附件那排（.pub-file）是 flex 居中容器，删除按钮在流内正常显示，不受本规则影响。 */
.pub-photo .pub-file-del {
  position: absolute;
  top: 0;
  right: 0;
  z-index: 2;
  padding: 0 10rpx 6rpx;
  font-size: 26rpx;
  color: #fff;
  background: rgba(23, 33, 43, 0.55);
  border-radius: 0 7px 0 12rpx;
}
.pub-form-intro-h2 {
  font-size: 20px;
  margin: 0 0 4px;
  color: #17212B;
}
.pub-form-intro-p {
  font-size: 12px;
  color: #667085;
  margin: 0;
  line-height: 1.5;
}
.pub-fade { opacity: 0.6; }
</style>
