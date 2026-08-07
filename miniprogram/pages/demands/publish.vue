<template>
  <view class="publish-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" @tap="onBack"><text class="back-sym">‹</text></view>
      <text class="page-title">{{ pageTitle }}</text>
      <view class="head-spacer"></view>
    </view>

    <!-- 步骤条 -->
    <view class="stepbar">
      <view
        v-for="(s, i) in steps"
        :key="i"
        class="step"
        :class="{ active: step === i + 1, done: step > i + 1 }"
      >
        <view class="step-num"><text>{{ step > i + 1 ? '✓' : i + 1 }}</text></view>
        <text class="step-label">{{ s }}</text>
      </view>
    </view>

    <!-- 步骤 3：预览提交 -->
    <view v-if="step === 3" class="form-card">
      <text class="card-title">确认发布信息</text>
      <text class="card-intro">信息提交后将进入协会审核，审核通过后才会公开展示。</text>
      <view class="preview-list">
        <view class="preview-row">
          <text class="pv-label">发布类型</text>
          <text class="pv-value">{{ typeName }}</text>
        </view>
        <view class="preview-row">
          <text class="pv-label">分类</text>
          <text class="pv-value">{{ form.category }}</text>
        </view>
        <view class="preview-row">
          <text class="pv-label">标题</text>
          <text class="pv-value">{{ form.title || '未填写' }}</text>
        </view>
        <view class="preview-row">
          <text class="pv-label">说明</text>
          <text class="pv-value">{{ form.desc || '未填写' }}</text>
        </view>
        <view class="preview-row">
          <text class="pv-label">{{ priceLabel }}</text>
          <text class="pv-value">{{ form.price || form.quote || '面议' }}</text>
        </view>
        <view class="preview-row">
          <text class="pv-label">{{ regionLabel }}</text>
          <text class="pv-value">{{ form.region || form.range || form.stock || '未填写' }}</text>
        </view>
        <view class="preview-row">
          <text class="pv-label">联系人</text>
          <text class="pv-value">{{ form.contact || '未填写' }}</text>
        </view>
      </view>
    </view>

    <!-- 步骤 1 / 2：表单 -->
    <view v-else class="form-card">
      <text class="card-title">{{ step === 1 ? '说明你要发布的内容' : '补充对接需要的信息' }}</text>
      <text class="card-intro">标有 * 的字段为必填项，请保证对接信息真实有效。</text>

      <!-- 步骤 1 -->
      <template v-if="step === 1">
        <text class="field-label">{{ categoryLabel }} <text class="req">*</text></text>
        <view class="pick-row">
          <view
            v-for="c in categoryOptions"
            :key="c"
            class="pick"
            :class="{ active: form.category === c }"
            @tap="form.category = c"
          >{{ c }}</view>
        </view>

        <text class="field-label">标题 <text class="req">*</text></text>
        <input v-model="form.title" class="field" placeholder="请输入清晰、具体的标题" />

        <text class="field-label">{{ descLabel }} <text class="req">*</text></text>
        <textarea v-model="form.desc" class="textarea" :placeholder="descPlaceholder"></textarea>
      </template>

      <!-- 步骤 2：需求 -->
      <template v-else-if="type === 'demand'">
        <view class="field-row-2">
          <view class="field-col">
            <text class="field-label">预算金额 <text class="req">*</text></text>
            <input v-model="form.price" class="field" type="digit" placeholder="例如 18000" />
          </view>
          <view class="field-col">
            <text class="field-label">作业地区 <text class="req">*</text></text>
            <input v-model="form.region" class="field" placeholder="例如 江津区" />
          </view>
        </view>
        <text class="field-label">截止日期 <text class="req">*</text></text>
        <input v-model="form.deadline" class="field" placeholder="例如 2026-08-12" />
        <text class="field-label">联系人 <text class="req">*</text></text>
        <input v-model="form.contact" class="field" placeholder="姓名或企业对接人" />
        <text class="field-label">附件</text>
        <text class="help">原型中模拟上传作业范围图、技术要求或现场照片。</text>
        <view class="attach-add" @tap="addAttachment">
          <text class="attach-plus">＋ 添加附件</text>
        </view>
      </template>

      <!-- 步骤 2：服务 -->
      <template v-else-if="type === 'service'">
        <text class="field-label">服务范围 <text class="req">*</text></text>
        <input v-model="form.range" class="field" placeholder="例如 重庆市内及周边" />
        <text class="field-label">报价方式 <text class="req">*</text></text>
        <view class="pick-row">
          <view
            v-for="q in quoteOptions"
            :key="q"
            class="pick"
            :class="{ active: form.quote === q }"
            @tap="form.quote = q"
          >{{ q }}</view>
        </view>
        <text class="field-label">设备与资质</text>
        <textarea v-model="form.qualification" class="textarea" placeholder="例如 M350 RTK、热成像载荷、保险及行业资质"></textarea>
        <text class="field-label">联系人 <text class="req">*</text></text>
        <input v-model="form.contact" class="field" placeholder="姓名或企业对接人" />
      </template>

      <!-- 步骤 2：商品 -->
      <template v-else>
        <view class="field-row-2">
          <view class="field-col">
            <text class="field-label">规格 / 型号 <text class="req">*</text></text>
            <input v-model="form.spec" class="field" placeholder="例如 M350 RTK" />
          </view>
          <view class="field-col">
            <text class="field-label">价格 <text class="req">*</text></text>
            <input v-model="form.price" class="field" placeholder="例如 1200 /天" />
          </view>
        </view>
        <text class="field-label">设备成色 <text class="req">*</text></text>
        <view class="pick-row">
          <view
            v-for="c in conditionOptions"
            :key="c"
            class="pick"
            :class="{ active: form.condition === c }"
            @tap="form.condition = c"
          >{{ c }}</view>
        </view>
        <text class="field-label">库存与所在地 <text class="req">*</text></text>
        <input v-model="form.stock" class="field" placeholder="例如 现货 3 套，九龙坡区" />
        <text class="field-label">联系人 <text class="req">*</text></text>
        <input v-model="form.contact" class="field" placeholder="姓名或企业对接人" />
      </template>
    </view>

    <!-- 底部操作栏 -->
    <view class="fixed-footer">
      <view v-if="step > 1" class="ghost-btn" @tap="prevStep">
        <text>上一步</text>
      </view>
      <view v-else class="ghost-btn" @tap="onBack">
        <text>取消</text>
      </view>
      <view class="next-btn" :class="step === 3 ? 'orange' : ''" @tap="nextStep">
        <text v-if="submitting">提交中...</text>
        <text v-else>{{ step === 3 ? '确认提交审核' : '下一步' }}</text>
      </view>
    </view>

    <!-- 提交成功页（覆盖态） -->
    <view v-if="submitted" class="success-mask">
      <view class="success-card">
        <view class="success-mark"><text class="success-sym">✓</text></view>
        <text class="success-title">已提交审核</text>
        <text class="success-desc">协会将在审核完成后通知您。审核通过后，信息会展示在供需大厅并参与智能匹配。</text>
        <view class="success-btn" @tap="goMine">查看我的发布</view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import { safeNavigateTo } from '../../utils/nav'
import {
  HALL_CATEGORIES, isLoggedIn, currentUserName,
  getPosts, savePosts,
} from '../../utils/hallData'

const type = ref('demand') // demand | service | product
const step = ref(1)
const submitting = ref(false)
const submitted = ref(false)

const form = ref({
  category: '', title: '', desc: '',
  price: '', region: '', deadline: '', contact: '', range: '',
  quote: '', qualification: '', spec: '', condition: '', stock: '',
})

const steps = ['基础信息', '业务信息', '预览提交']

const typeName = computed(() => ({ demand: '需求', service: '服务能力', product: '商品设备' })[type.value])
const pageTitle = computed(() => '发布' + typeName.value)

const categoryOptions = computed(() => {
  const kind = type.value === 'demand' ? 'demand' : type.value
  const arr = HALL_CATEGORIES[kind] || []
  return arr.filter((c) => c !== '全部')
})
const categoryLabel = computed(() => ({ demand: '需求分类', service: '服务分类', product: '商品品类' })[type.value])
const descLabel = computed(() => ({ demand: '需求描述', service: '能力说明', product: '商品说明' })[type.value])
const descPlaceholder = computed(() => ({
  demand: '请补充场景、范围、规格或交付说明',
  service: '请说明服务能力、团队配置与交付标准',
  product: '请说明设备规格、成色与售后保障',
})[type.value])

const priceLabel = computed(() => (type.value === 'demand' ? '预算金额' : type.value === 'service' ? '报价方式' : '价格'))
const regionLabel = computed(() => (type.value === 'demand' ? '作业地区' : type.value === 'service' ? '服务范围' : '库存与所在地'))

const quoteOptions = ['按项目报价', '按天报价', '面议']
const conditionOptions = ['全新', '九成新', '二手良好']

/* ================= 步骤逻辑 ================= */
function initForm() {
  form.value = {
    category: type.value === 'demand' ? '巡检' : type.value === 'service' ? '巡检' : '整机租赁',
    title: '', desc: '',
    price: '', region: '', deadline: '', contact: '', range: '',
    quote: type.value === 'service' ? '按项目报价' : '', qualification: '',
    spec: '', condition: type.value === 'product' ? '全新' : '', stock: '',
  }
}

function validateStep() {
  if (step.value === 1) {
    if (!form.value.title.trim()) return '请填写标题'
    if (!form.value.desc.trim()) return '请填写内容说明'
    return ''
  }
  if (step.value === 2) {
    if (type.value === 'demand') {
      if (!form.value.price.trim()) return '请填写预算金额'
      if (!form.value.region.trim()) return '请填写作业地区'
      if (!form.value.deadline.trim()) return '请填写截止日期'
      if (!form.value.contact.trim()) return '请填写联系人'
    } else if (type.value === 'service') {
      if (!form.value.range.trim()) return '请填写服务范围'
      if (!form.value.contact.trim()) return '请填写联系人'
    } else {
      if (!form.value.spec.trim()) return '请填写规格型号'
      if (!form.value.price.trim()) return '请填写价格'
      if (!form.value.stock.trim()) return '请填写库存与所在地'
      if (!form.value.contact.trim()) return '请填写联系人'
    }
    return ''
  }
  return ''
}

function nextStep() {
  if (submitting.value) return
  const err = validateStep()
  if (err) {
    uni.showToast({ title: err, icon: 'none' })
    return
  }
  if (step.value < 3) {
    step.value++
    return
  }
  submit()
}

function prevStep() {
  if (step.value > 1) step.value--
}

async function submit() {
  if (submitting.value) return
  submitting.value = true
  try {
    // 需求走后端创建；服务/商品本期为模拟提交
    if (type.value === 'demand' && isLoggedIn()) {
      try {
        const budgetNum = parseFloat(form.value.price)
        await request({
          url: '/api/v1/demands',
          method: 'POST',
          data: {
            title: form.value.title,
            biz_type: bizTypeValue(),
            budget: Number.isFinite(budgetNum) ? budgetNum : 0,
            district: form.value.region,
            description: form.value.desc,
            contact: form.value.contact,
            publisher_name: currentUserName(),
          },
        })
      } catch (e) { /* 后端不可用：降级到本地存储 */ }
    }
    const posts = getPosts()
    posts.unshift({
      id: 'm' + Date.now(),
      type: typeName.value,
      title: form.value.title,
      status: '待审核',
      date: '刚刚',
    })
    savePosts(posts)
    await new Promise((resolve) => setTimeout(resolve, 400))
    submitting.value = false
    submitted.value = true
  } catch (e) {
    submitting.value = false
    uni.showToast({ title: '提交失败，请稍后重试', icon: 'none' })
  }
}

function bizTypeValue() {
  const cat = form.value.category
  const map = {
    巡检: 'cable_inspection', 植保: 'plant_transport', 农药: 'spray_pesticide',
    吊运: 'plant_transport', 航拍: 'other', 测绘: 'other', 应急: 'other',
    租赁: 'trade_lease',
  }
  return map[cat] || 'other'
}

function addAttachment() {
  uni.showToast({ title: '已模拟添加附件', icon: 'none' })
}

const goMine = () => safeNavigateTo('/pages/demands/mine')
const onBack = () => uni.navigateBack()

onLoad((options) => {
  const t = (options && options.type) || 'demand'
  if (t === 'service' || t === 'product' || t === 'demand') {
    type.value = t
  }
  initForm()
  // 预填联系人
  const u = currentUserName()
  if (u !== '微信用户') form.value.contact = u
})
</script>

<style scoped>
.publish-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(160rpx + env(safe-area-inset-bottom));
  box-sizing: border-box;
}

/* ═══════ 头部 ═══════ */
.page-header {
  height: 56px;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title { flex: 1; font-size: 34rpx; font-weight: 700; color: #17212B; }
.head-spacer { width: 72rpx; }

/* ═══════ 步骤条 ═══════ */
.stepbar {
  display: flex;
  padding: 30rpx 36rpx 24rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
}
.step {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12rpx;
  color: #98A2B3;
  font-size: 22rpx;
}
.step:not(:last-child)::after {
  content: '';
  height: 2rpx;
  flex: 1;
  background: #E4E7EC;
  margin: 0 14rpx;
}
.step-num {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #E9EDF1;
  color: #667085;
  font-size: 20rpx;
  flex-shrink: 0;
}
.step.active, .step.done { color: #0A66C2; }
.step.active .step-num, .step.done .step-num { color: #fff; background: #0A66C2; }
.step-label { white-space: nowrap; }

/* ═══════ 表单卡片 ═══════ */
.form-card {
  margin: 24rpx 32rpx;
  padding: 8rpx 28rpx 32rpx;
  border-radius: 16rpx;
  background: #fff;
}
.card-title { display: block; font-size: 32rpx; font-weight: 700; color: #17212B; margin: 30rpx 0 8rpx; }
.card-intro { display: block; color: #667085; font-size: 22rpx; line-height: 1.6; }

.field-label {
  display: block;
  color: #344054;
  font-size: 24rpx;
  font-weight: 650;
  margin: 28rpx 0 14rpx;
}
.req { color: #D92D20; font-style: normal; }
.field {
  width: 100%;
  height: 84rpx;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  color: #17212B;
  font-size: 26rpx;
  padding: 0 20rpx;
  box-sizing: border-box;
}
.textarea {
  width: 100%;
  min-height: 168rpx;
  border: 1px solid #E4E7EC;
  border-radius: 12rpx;
  background: #fff;
  color: #17212B;
  font-size: 26rpx;
  padding: 20rpx;
  box-sizing: border-box;
  line-height: 1.6;
}
.field-row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 20rpx; }
.pick-row { display: flex; flex-wrap: wrap; gap: 14rpx; }
.pick {
  padding: 14rpx 20rpx;
  border: 1px solid #E4E7EC;
  border-radius: 10rpx;
  background: #fff;
  color: #667085;
  font-size: 24rpx;
}
.pick.active {
  background: #EAF3FB;
  border-color: #B9D6EF;
  color: #0A66C2;
  font-weight: 650;
}
.help {
  display: block;
  font-size: 22rpx;
  color: #667085;
  line-height: 1.6;
  margin: 12rpx 0;
}
.attach-add {
  display: inline-flex;
  align-items: center;
  padding: 16rpx 24rpx;
  border: 1px solid #B9D6EF;
  border-radius: 12rpx;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 24rpx;
  font-weight: 650;
  margin-top: 8rpx;
}

/* ═══════ 预览 ═══════ */
.preview-list { margin-top: 16rpx; }
.preview-row {
  display: flex;
  justify-content: space-between;
  gap: 20rpx;
  border-bottom: 1px solid #EEF1F4;
  padding: 22rpx 0;
  font-size: 24rpx;
}
.pv-label { color: #667085; flex-shrink: 0; }
.pv-value {
  text-align: right;
  font-weight: 650;
  color: #17212B;
  max-width: 65%;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ═══════ 底部操作栏 ═══════ */
.fixed-footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 32rpx calc(20rpx + env(safe-area-inset-bottom));
  display: flex;
  gap: 20rpx;
  background: #fff;
  border-top: 1px solid #EEF1F4;
  z-index: 18;
  box-shadow: 0 -2px 10px rgba(16, 24, 40, 0.05);
}
.fixed-footer > view {
  flex: 1;
  height: 84rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 700;
}
.ghost-btn { border: 1px solid #E4E7EC; background: #fff; color: #344054; }
.next-btn { background: #0A66C2; color: #fff; }
.next-btn.orange { background: #F97316; }

/* ═══════ 成功页 ═══════ */
.success-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 100;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 56rpx;
}
.success-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.success-mark {
  width: 136rpx;
  height: 136rpx;
  border-radius: 50%;
  color: #168A55;
  background: #E9F7F0;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 32rpx;
}
.success-sym { font-size: 60rpx; }
.success-title { font-size: 38rpx; font-weight: 700; color: #17212B; margin-bottom: 14rpx; }
.success-desc {
  max-width: 520rpx;
  color: #667085;
  font-size: 24rpx;
  line-height: 1.7;
  margin-bottom: 40rpx;
}
.success-btn {
  height: 80rpx;
  padding: 0 36rpx;
  border-radius: 12rpx;
  background: #0A66C2;
  color: #fff;
  font-size: 26rpx;
  font-weight: 650;
  line-height: 80rpx;
}
</style>
