<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar :title="done ? '发布结果' : '发布难题'" show-back :fixed="true" @back="goBack" />

    <!-- ═══ 成功态 ═══ -->
    <view v-if="done" class="succ">
      <view class="succ-ring"><u-icon name="success" size="40px" color="#fff" /></view>
      <text class="succ-title">提交成功</text>
      <text class="succ-desc">难题已提交，协会审核通过后将在广场展示
审核进度将通过消息中心通知</text>
      <view class="receipt">
        <view class="rt">发布回执 <text class="rt-no">{{ receiptNo }}</text></view>
        <view class="rrow"><text>难题标题</text><text class="rv">{{ summary.title }}</text></view>
        <view class="rrow"><text>所属领域</text><text class="rv">{{ summary.field }}</text></view>
        <view class="rrow"><text>悬赏金额</text><text class="rv">{{ summary.money }}</text></view>
        <view class="rrow"><text>截止日期</text><text class="rv">{{ summary.deadline }}</text></view>
        <view class="rrow"><text>状态</text><text class="rv cl-wait">待审核</text></view>
      </view>
      <view class="btn-main" @tap="goList">返回列表</view>
      <view class="btn-ghost" @tap="resetForm">再发布一条</view>
    </view>

    <!-- ═══ 表单态 ═══ -->
    <template v-else>
      <view class="form-wrap">
        <view class="form-card">
          <view class="form-group">
            <view class="form-label">难题标题 <text class="required">*</text></view>
            <input class="form-input" :class="{ err: errors.title }" v-model="form.title" placeholder="一句话说明技术难题 / 攻关目标" placeholder-class="ph" :maxlength="60" @input="clearError('title')" />
            <text v-if="errors.title" class="form-err">{{ errors.title }}</text>
          </view>
          <view class="form-group">
            <view class="form-label">所属领域 <text class="required">*</text></view>
            <view class="pills">
              <text
                v-for="f in FIELDS"
                :key="f"
                class="pill"
                :class="{ act: form.field === f }"
                @tap="form.field = form.field === f ? '' : f; clearError('field')"
              >{{ f }}</text>
            </view>
            <text v-if="errors.field" class="form-err">{{ errors.field }}</text>
          </view>
          <view class="form-group">
            <view class="form-label">悬赏金额</view>
            <view class="money-row">
              <input class="form-input money-input" :class="{ err: errors.money }" v-model="form.money" type="number" placeholder="500000" placeholder-class="ph" :disabled="form.free" @input="clearError('money')" />
              <text class="money-unit">元</text>
              <view class="free-chip" :class="{ act: form.free }" @tap="toggleFree; clearError('money')">面议</view>
            </view>
            <text v-if="errors.money" class="form-err">{{ errors.money }}</text>
          </view>
          <view class="form-group">
            <view class="form-label">截止日期 <text class="required">*</text></view>
            <picker mode="date" :value="form.deadline" :start="todayStr" @change="onDateChange">
              <view class="picker-value" :class="{ ph: !form.deadline, err: errors.deadline }">{{ form.deadline || '请选择截止日期' }}</view>
            </picker>
            <text v-if="errors.deadline" class="form-err">{{ errors.deadline }}</text>
          </view>
          <view class="form-group">
            <view class="form-label">详细描述 <text class="required">*</text></view>
            <textarea class="form-textarea" :class="{ err: errors.desc }" v-model="form.desc" placeholder="技术背景、指标要求、期望成果、合作方式…" placeholder-class="ph" :maxlength="500" @input="clearError('desc')" />
            <text v-if="errors.desc" class="form-err">{{ errors.desc }}</text>
          </view>
        </view>

        <view class="notice">
          <text class="notice-line">· 仅<span class="hl">会员企业</span>可发布，提交后由协会审核通过后对外展示</text>
          <text class="notice-line">· 悬赏金额平台不代收，洽谈成功后线下结算</text>
          <text class="notice-line">· 发布者联系方式由平台凭会员身份对接，不公开征集</text>
        </view>
      </view>

      <!-- 底部提交（布局对齐 detail 底部操作栏：底部吸附 + safe-area + 42px 按钮） -->
      <view class="submit-bar">
        <view class="submit-btn" :class="{ disabled: submitting }" @tap="handleSubmit">
          {{ submitting ? '提交中...' : '提交发布' }}
        </view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, authStorage } from '@/utils/request'
import { useReduceMotion } from '@/utils/motion'

const FIELDS = ['飞控系统', '动力电池', 'AI算法', '通信链路', '新型材料', '载荷设备']

const done = ref(false)
const submitting = ref(false)
const statusBarHeight = ref(20)
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
const receiptNo = ref('')
const summary = ref({ title: '', field: '', money: '', deadline: '' })
const form = ref({
  title: '',
  field: '',
  money: '',
  free: false,
  deadline: '',
  desc: '',
})
const errors = ref({}) // 内联校验错误：提交时标记对应字段（红框 + 提示），输入时清除

const pad = (n) => (n < 10 ? '0' + n : '' + n)
const todayStr = computed(() => {
  const now = new Date()
  return now.getFullYear() + '-' + pad(now.getMonth() + 1) + '-' + pad(now.getDate())
})
const moneyText = computed(() => {
  if (form.value.free) return '面议'
  const n = Number(form.value.money)
  if (!n || n <= 0) return '—'
  if (n >= 10000) return '¥' + (n / 10000) + '万'
  return '¥' + n
})

const toggleFree = () => {
  form.value.free = !form.value.free
  if (form.value.free) form.value.money = ''
}

const onDateChange = (e) => {
  form.value.deadline = e.detail.value
  clearError('deadline')
}

/* 全量校验：一次标记所有未过项（内联红框提示），返回首错文案供 toast */
const validate = () => {
  const e = {}
  if (!form.value.title.trim()) e.title = '请输入难题标题'
  if (!form.value.field) e.field = '请选择所属领域'
  if (!form.value.free && !(Number(form.value.money) > 0)) e.money = '请输入悬赏金额或选择面议'
  if (!form.value.deadline) e.deadline = '请选择截止日期'
  if (!form.value.desc.trim()) e.desc = '请输入详细描述'
  errors.value = e
  return e.title || e.field || e.money || e.deadline || e.desc || ''
}
const clearError = (k) => {
  if (errors.value[k]) errors.value[k] = ''
}

const handleSubmit = async () => {
  if (submitting.value) return // 防重复点击双发
  const msg = validate()
  if (msg) {
    uni.showToast({ title: msg, icon: 'none' })
    return
  }
  const token = authStorage.getAccessToken()
  if (!token) {
    uni.showToast({ title: '请先登录后再发布', icon: 'none' })
    setTimeout(() => uni.navigateTo({ url: '/pages/login/index' }), 500)
    return
  }
  submitting.value = true
  try {
    const res = await request({
      url: '/api/v1/rd-challenges',
      method: 'POST',
      data: {
        title: form.value.title.trim(),
        field: form.value.field,
        description: form.value.desc.trim(),
        budget_fen: form.value.free ? 0 : Math.round(Number(form.value.money) * 100),
        deadline: form.value.deadline,
      },
    })
    const it = (res && res.data) || res
    receiptNo.value = 'RD' + (it?.id || Date.now()).toString().slice(-12)
    summary.value = {
      title: form.value.title.trim(),
      field: form.value.field,
      money: moneyText.value,
      deadline: form.value.deadline,
    }
    done.value = true
  } catch (e) {
    uni.showToast({ title: (e && (e.errMsg || e.message)) || '发布失败，请稍后重试', icon: 'none' })
  } finally {
    submitting.value = false
  }
}

const resetForm = () => {
  form.value = { title: '', field: '', money: '', free: false, deadline: '', desc: '' }
  errors.value = {}
  done.value = false
  submitting.value = false
}
const goList = () => {
  uni.navigateBack({
    fail: () => uni.redirectTo({ url: '/pkg-eco/pages/challenges/list' }),
  })
}
const goBack = () => uni.navigateBack({
  fail: () => uni.redirectTo({ url: '/pkg-eco/pages/challenges/list' }),
})

onLoad(() => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  // 发布者身份由后端凭 Token 关联（createRDChallenge 只认 authenticatedActor），前端不收集联系方式
})
</script>

<style>
page {
  background: #F4F6F8;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: 100px;
}

/* ===== 表单 ===== */
.form-wrap { padding: 12px; }
.form-card {
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
  padding: 12px;
  box-shadow: 0 3px 12px rgba(16, 24, 40, 0.05); /* 与 list 卡片同层级投影 */
}
.form-group { margin-bottom: 16px; }
.form-group:last-child { margin-bottom: 0; }
.form-label { font-size: 13px; font-weight: 600; color: #344054; margin-bottom: 6px; }
.required { color: #D92D20; font-size: 12px; margin-left: 2px; }
.form-input {
  width: 100%;
  box-sizing: border-box;
  min-height: 42px;
  padding: 10px 12px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  font-size: 13px;
  color: #17212B;
  background: #fff;
}
.form-textarea {
  width: 100%;
  box-sizing: border-box;
  min-height: 100px;
  padding: 10px 12px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  font-size: 13px;
  color: #17212B;
  background: #fff;
  line-height: 1.6;
}
.ph { color: #98A2B3; } /* 占位符 2.6:1（豁免档）：原 #c8c9cc 1.6:1 过淡 */
.form-err { display: block; font-size: 12px; color: #D92D20; margin-top: 4px; } /* 内联校验提示 */
.form-input.err, .form-textarea.err, .picker-value.err { border-color: #D92D20; }
.pills { display: flex; flex-wrap: wrap; gap: 8px; }
.pill {
  min-height: 36px; /* 触控目标：30px→36px（与 list 筛选 chip 同步） */
  padding: 0 13px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  background: #fff;
  color: #667085;
  font-size: 13px; /* 对齐 list 筛选面板 .p-chip */
  display: inline-flex;
  align-items: center;
}
.pill.act { color: #fff; border-color: #0A66C2; background: #0A66C2; font-weight: 600; }
.money-row { display: flex; align-items: center; gap: 8px; }
.money-input { flex: 1; }
.money-unit { font-size: 12px; color: #667085; flex: none; }
.free-chip {
  flex: none;
  min-height: 36px; /* 触控目标：30px→36px */
  padding: 0 13px;
  border-radius: 6px;
  border: 1px solid #E4E7EC;
  background: #fff;
  color: #667085;
  font-size: 13px; /* 对齐 list 筛选面板 .p-chip */
  display: inline-flex;
  align-items: center;
}
.free-chip.act { color: #0A66C2; border-color: #0A66C2; font-weight: 600; background: #F4F8FC; }
.picker-value {
  min-height: 42px;
  padding: 10px 12px;
  border: 1px solid #E4E7EC;
  border-radius: 6px;
  font-size: 13px;
  color: #17212B;
  background: #fff;
  box-sizing: border-box;
  display: flex;
  align-items: center;
}
.picker-value.ph { color: #98A2B3; } /* 占位符 2.6:1（豁免档）：原 #c8c9cc 1.6:1 过淡 */
.form-err { display: block; font-size: 12px; color: #D92D20; margin-top: 4px; } /* 内联校验提示 */
.form-input.err, .form-textarea.err, .picker-value.err { border-color: #D92D20; }

/* ===== 须知 ===== */
.notice {
  margin-top: 10px;
  background: #F4F8FC;
  border: 1px solid #DCEAF7;
  border-radius: 8px;
  padding: 10px 12px;
}
.notice-line { display: block; font-size: 12px; color: #667085; line-height: 1.8; }
.hl { color: #0A66C2; font-weight: 600; }

/* ===== 底部提交 ===== */
.submit-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 10px 16px;
  padding-bottom: calc(10px + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 0.5px solid #F0F0F0;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.04); /* 对齐 detail 底部操作栏 */
}
.submit-btn {
  height: 42px; /* 对齐 detail .bp 按钮体系 */
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(10, 102, 194, 0.3);
}
.submit-btn.disabled { background: #98A2B3; box-shadow: none; }

/* ===== 成功态 ===== */
.succ { padding: 44px 28px; display: flex; flex-direction: column; align-items: center; text-align: center; }
.succ-ring {
  width: 84px;
  height: 84px;
  border-radius: 50%;
  background: linear-gradient(135deg, #34c759, #5bd88a);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 14px 36px rgba(52, 199, 89, 0.35);
  margin-bottom: 20px;
}
.succ-title { font-size: 20px; font-weight: 700; color: #17212B; margin-bottom: 8px; }
.succ-desc { font-size: 13px; color: #667085; line-height: 1.7; margin-bottom: 18px; white-space: pre-line; }
.receipt {
  width: 100%;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
  padding: 14px;
  margin-bottom: 22px;
  text-align: left;
  box-sizing: border-box;
}
.rt {
  font-size: 12px;
  color: #667085;
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px dashed #EBEDF0;
  display: flex;
  justify-content: space-between;
}
.rt-no { color: #0A66C2; font-weight: 600; }
.rrow { display: flex; justify-content: space-between; font-size: 13px; padding: 5px 0; color: #667085; gap: 12px; }
.rv { color: #17212B; font-weight: 600; text-align: right; flex: 1; }
.cl-wait { color: #E96012; }
.btn-main {
  width: 100%;
  height: 42px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(10, 102, 194, 0.3);
  box-sizing: border-box;
}
.btn-ghost {
  width: 100%;
  height: 42px;
  border-radius: 8px;
  border: 1px solid #0A66C2;
  background: #fff;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  margin-top: 10px;
  box-sizing: border-box;
}

/* ===================== 动效规范（对齐全局动画规范，与 list/detail 同套） =====================
   白名单：仅 transform / opacity（小尺寸元素 color/background 过渡允许——仅重绘不重排）
   禁参与动画：top/left/width/height/margin（触发重排）、box-shadow/filter（低端安卓掉帧）
   时长：微反馈 150-200ms（按压按下 .08s 即时到位）/ 松手弹簧回位 .3s（ios-pop）/ 浮层 200-300ms / 页面级 ≤400ms；
        退场 = 进场 ×0.7 且必须存在
   曲线：两枚固定曲线——ios-pop cubic-bezier(0.16,1,0.3,1) 松手柔顺减速（仅按压/弹出类 transform）+
        ios-decel cubic-bezier(.32,.72,0,1) 浮层流体减速（提交栏进场）；
        其余进场 ease-out / 退场 ease-in / 循环 linear；除这两枚外禁手写 cubic-bezier
   数量：入场错峰首屏可见项，整页编排 ≤400ms；循环动画任何时刻全页 ≤1 处（本页无循环动画）
   可打断：本页无浮层/弹层，表单态↔成功态用 v-if 互斥渲染，入场动画单次执行天然不排队
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) 表单态入场编排：表单卡 0ms → 6 组字段 30ms 步进错峰（30~180ms）+ 200ms 动画 = 380ms ≤ 400ms；
   提交栏与卡片同步自下而上滑入（fixed 定位 translateY，不重排） */
.form-card { animation: fadeUp .25s ease-out backwards; }
.form-group { animation: fadeUp .2s ease-out backwards; }
.form-group:nth-child(1) { animation-delay: 30ms; }
.form-group:nth-child(2) { animation-delay: 60ms; }
.form-group:nth-child(3) { animation-delay: 90ms; }
.form-group:nth-child(4) { animation-delay: 120ms; }
.form-group:nth-child(5) { animation-delay: 150ms; }
.form-group:nth-child(6) { animation-delay: 180ms; }
.notice { animation: fadeUp .2s ease-out 220ms backwards; }
.submit-bar { animation: bbUp .3s cubic-bezier(.32, .72, 0, 1) backwards; } /* ios-decel：提交栏 sheet 落地感 */
@keyframes fadeUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
@keyframes bbUp { from { opacity: 0; transform: translateY(24px); } to { opacity: 1; transform: translateY(0); } }

/* 2) 成功态入场编排：圆环 ios-pop 弹簧落位（本页唯一作者时刻，.35s 过冲回位）→ 标题 40ms → 描述 80ms → 回执 120ms
   → 主按钮 160ms → 次按钮 190ms（总 390ms ≤ 400ms；backwards 填充，延迟期保持隐藏不闪跳） */
.succ-ring { animation: ringIn .35s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
.succ-title { animation: fadeUp .2s ease-out 40ms backwards; }
.succ-desc { animation: fadeUp .2s ease-out 80ms backwards; }
.receipt { animation: fadeUp .2s ease-out 120ms backwards; }
.btn-main { animation: fadeUp .2s ease-out 160ms backwards; }
.btn-ghost { animation: fadeUp .2s ease-out 190ms backwards; }
@keyframes ringIn { from { opacity: 0; transform: scale(.8); } to { opacity: 1; transform: scale(1); } }

/* 3) 状态过渡：领域 chip / 面议 chip 选中与 list 筛选面板同套（200ms 平滑 + ios-pop 微弹过冲回位） */
.pill, .free-chip { transition: background .2s ease, border-color .2s ease, color .2s ease, transform .3s cubic-bezier(0.16, 1, 0.3, 1); } /* ios-pop：松手弹簧回位 */
.pill:active, .free-chip:active { transform: scale(.94); transition: transform .08s linear; } /* 按下即时到位，其余按压变化同步走即时 */
.pill.act, .free-chip.act { animation: chipPop .3s cubic-bezier(0.16, 1, 0.3, 1); } /* ios-pop：选中微弹带轻微过冲回位 */
@keyframes chipPop { from { transform: scale(.9); } to { transform: scale(1); } }

/* 4) 交互反馈：按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位——与 list/detail 同套手感） */
.submit-btn { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease, background .2s ease; } /* ios-pop；含 disabled 切换平滑 */
.submit-btn:active { transform: scale(.95); opacity: .92; transition: transform .08s linear; }
.submit-btn.disabled:active { transform: none; opacity: 1; } /* 提交中不响应按压 */
.picker-value { transition: background .2s ease, transform .3s cubic-bezier(0.16, 1, 0.3, 1); } /* ios-pop */
.picker-value:active { background: #F4F8FC; transform: scale(.99); transition: transform .08s linear; }
.btn-main { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), opacity .15s ease, background .2s ease; } /* ios-pop */
.btn-main:active { transform: scale(.95); opacity: .92; transition: transform .08s linear; }
.btn-ghost { transition: transform .3s cubic-bezier(0.16, 1, 0.3, 1), background .2s ease, color .2s ease; } /* ios-pop */
.btn-ghost:active { background: #F4F8FC; transform: scale(.96); transition: transform .08s linear; }

/* ===================== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===================== */
.page.no-motion .form-card,
.page.no-motion .form-group,
.page.no-motion .notice,
.page.no-motion .submit-bar,
.page.no-motion .succ-ring,
.page.no-motion .succ-title,
.page.no-motion .succ-desc,
.page.no-motion .receipt,
.page.no-motion .btn-main,
.page.no-motion .btn-ghost { animation: none; } /* 装饰入场全关 */
.page.no-motion .pill.act,
.page.no-motion .free-chip.act { animation: none; } /* 选中微弹属缩放，关闭；选中色保留 */
.page.no-motion .pill:active,
.page.no-motion .free-chip:active,
.page.no-motion .picker-value:active,
.page.no-motion .submit-btn:active,
.page.no-motion .btn-main:active,
.page.no-motion .btn-ghost:active { transform: none; } /* 按压微缩放关闭，保留颜色/透明度反馈 */
</style>
