<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏 -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">{{ typeConfig.name }}</view>
      <view class="pub-nav-action" hover-class="pub-fade" :style="{ marginRight: capsuleGap + 'px' }" @tap="saveDraft">存草稿</view>
    </view>

    <!-- 商品/课程两步进度条 -->
    <view v-if="typeConfig.stepped" class="pub-progress">
      <view class="pub-progress-top">
        <text>{{ typeConfig.steps[step] }}</text>
        <text>{{ step + 1 }}/{{ typeConfig.steps.length }}</text>
      </view>
      <view class="pub-progress-bar">
        <view :style="{ width: (step === 0 ? 50 : 100) + '%' }"></view>
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

          <!-- 选择型：打开底部抽屉 -->
          <view
            v-if="field[3] === 'select'"
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
            <view v-for="(photo, i) in photos" :key="i" class="pub-photo">
              <image v-if="photo && photo.src" :src="photo.src" mode="aspectFill" class="pub-photo-img" />
            </view>
            <view class="pub-add-photo" hover-class="pub-fade" @tap="addPhoto">＋</view>
          </view>
          <view class="pub-upload-tip">建议上传清晰实拍图，首图将作为列表封面</view>
        </view>
      </view>
    </view>

    <!-- 固定底部操作区 -->
    <view class="pub-sticky">
      <view class="pub-btn pub-btn--ghost" hover-class="pub-btn--active" @tap="saveDraft">保存草稿</view>
      <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="nextAction">{{ primaryText }}</view>
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
import { onLoad } from '@dcloudio/uni-app'
import { TYPES, getPost, upsertPost, draftPosts, saveFormState } from '../../utils/publishData'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, capsuleGap, initSafeTop } = useSafeTop(true)

const props = defineProps({}) // 无 props，页面通过路由参数驱动
void props

const type = ref('')
const step = ref(0)
const values = ref({})
const photos = ref([])
const sheetId = ref('')
const toast = ref('')
const toastTimer = ref(null)
const resumeId = ref('')

const typeConfig = computed(() => TYPES[type.value] || null)

const visibleSections = computed(() => {
  const t = typeConfig.value
  if (!t) return []
  if (t.stepped) {
    if (step.value === 0) return [t.sections[0]]
    return t.sections.slice(1)
  }
  return t.sections
})

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
  const numeric = ['budget', 'price', 'stock', 'quota', 'duration', 'contact']
  return numeric.includes(id) ? 'number' : 'text'
}
function unitOf(id) {
  if (id === 'budget' || id === 'price') return '元'
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
  sheetId.value = ''
}
function addPhoto() {
  if (photos.value.length >= 5) return
  const pick = (paths) => {
    photos.value.push(...paths.map((p) => ({ src: p })))
    showToast('已添加 ' + paths.length + ' 张图片')
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
function showToast(text) {
  toast.value = text
  if (toastTimer.value) clearTimeout(toastTimer.value)
  toastTimer.value = setTimeout(() => {
    toast.value = ''
  }, 2200)
}

// 当前步骤必填校验
function requiredMissing() {
  const t = typeConfig.value
  if (!t) return []
  const candidates = t.stepped
    ? step.value === 0 ? [t.sections[0]] : t.sections.slice(1)
    : t.sections
  const missing = []
  candidates.forEach((s) => {
    ;(s.fields || []).forEach((f) => {
      if (f[4] && !String(values.value[f[0]] || '').trim()) missing.push(f[1])
    })
  })
  return missing
}

function nextAction() {
  const missing = requiredMissing()
  if (missing.length) {
    showToast('请先填写：' + missing[0])
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
    // 图片真实临时路径（发布时上传到服务器；历史数据只有数量无路径，过滤为空）
    photos: photos.value.map((p) => p && p.src).filter(Boolean),
    resumeId: resumeId.value || '',
  }
  saveFormState(state)
  const q = encodeURIComponent(JSON.stringify(state))
  uni.navigateTo({ url: '/pages/publish/preview?state=' + q })
}

function goBack() {
  safeBack()
}

onLoad((options) => {
  initSafeTop()
  const t = options && options.type
  if (t && TYPES[t]) {
    type.value = t
  }
  // 编辑已有草稿/发布（历史数据只存数量，无真实路径，恢复为占位）
  if (options && options.id) {
    const post = getPost(options.id)
    if (post) {
      resumeId.value = post.id
      values.value = Object.assign({}, post.values || {})
      photos.value = Array.from({ length: post.photoCount || 0 }, () => ({}))
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
    }
  }
  if (type.value && !typeConfig.value) type.value = ''
})
</script>

<style scoped>
@import './pub-style.css';

.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
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
