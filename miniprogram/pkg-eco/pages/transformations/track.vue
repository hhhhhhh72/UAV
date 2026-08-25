<template>
  <view class="page" :class="{ 'no-motion': noMotion }">
    <!-- Loading -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- Error -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- Empty -->
    <view v-else-if="!t" class="st">
      <u-empty description="暂无成果转化记录">
        <view class="stb" @tap="goBack">返回</view>
      </u-empty>
    </view>

    <!-- Content -->
    <template v-else>
      <!-- 概览 -->
      <view class="t-hero">
        <view class="t-back" aria-role="button" aria-label="返回上一页" @tap="goBack"></view>
        <view class="t-tag">{{ t.status === 'completed' ? '已完成' : '进行中' }}</view>
        <text class="t-title">{{ t.title }}</text>
        <text class="t-sub">阶段：{{ stageLabel }}<template v-if="rank"> · 已推进至第 {{ rank }}/4 阶段</template></text>
      </view>

      <!-- 转化阶段（虚线轨道 + 进度动画，自适应间距） -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">转化阶段</text></view>
        <view class="tr-flow">
          <view class="tr-track">
            <view class="tr-base"></view>
            <view class="tr-prog" :style="{ transform: flowReady ? 'scaleX(' + flowRatio + ')' : 'scaleX(0)' }"></view>
            <view class="tr-stages">
              <view
                v-for="(st, si) in stages"
                :key="si"
                class="tr-stage"
                :class="{ done: rank >= si + 1, cur: rank === si + 1 }"
              >
                <view class="tr-dot"></view>
                <text class="tr-stage-name">{{ st }}</text>
              </view>
            </view>
          </view>
          <view class="tr-meta">
            <text>{{ rank ? '已推进至第 ' + rank + '/4 阶段' : '转化阶段待确认' }}</text>
            <text v-if="rank" class="tr-cur">已完成 {{ pct }}%</text>
          </view>
        </view>
      </view>

      <!-- 当前进度 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">当前进度</text></view>
        <view class="t-row"><text class="t-k">阶段</text><text class="t-v" :class="stageCls">{{ stageLabel }}</text></view>
        <view class="t-row"><text class="t-k">进度描述</text><text class="t-v">{{ t.progress || '—' }}</text></view>
        <view class="t-row"><text class="t-k">合作单位</text><text class="t-v">{{ t.partner_id || '—' }}</text></view>
      </view>

      <!-- 里程碑 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">里程碑</text><text class="ms-count">共 {{ t.milestones.length }} 项</text></view>
        <view class="ms-list">
          <view v-for="(m, i) in t.milestones" :key="i" class="ms-item" :class="{ done: m.completed, cur: !m.completed && i === curIndex, todo: !m.completed && i > curIndex }">
            <view class="ms-ic"><text>{{ m.completed ? '✓' : (!m.completed && i === curIndex ? '…' : '○') }}</text></view>            <view class="ms-bd">
              <text class="ms-name">{{ m.name }}</text>
              <text class="ms-date">{{ (m.date || '').slice(0, 10) || '规划中' }}</text>
              <text v-if="m.evidence" class="ms-ev">{{ m.evidence }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 时间信息 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">时间信息</text></view>
        <view class="t-row"><text class="t-k">创建时间</text><text class="t-v">{{ (t.created_at || '').slice(0, 10) }}</text></view>
        <view class="t-row"><text class="t-k">更新时间</text><text class="t-v">{{ (t.updated_at || '').slice(0, 10) }}</text></view>
      </view>
      <view style="height: 160rpx"></view>

      <!-- 底部操作栏 -->
      <view class="bb">
        <template v-if="isOwner">
          <view class="bo" @tap="goBack">返回成果详情</view>
          <view class="bp" :class="{ disabled: submitting }" @tap="onAdvance">{{ submitting ? '推进中...' : '推进下一阶段' }}</view>
        </template>
        <template v-else>
          <view class="bo" @tap="goBack">返回成果详情</view>
          <view class="bp" @tap="onContact">联系发布方</view>
        </template>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request, getStoredUser } from '@/utils/request'
import { MOCK_TRANSFORMS_BY_ACH, STAGE_SHORT, STAGE_RANK } from '@/utils/mockAchievements'
import { revealContactCopy } from '@/utils/contactMask'
import { useReduceMotion } from '@/utils/motion'

const achievementId = ref('')
const tid = ref('')
const t = ref(null)
const loading = ref(true)
const err = ref(false)
const uid = ref('')
const submitting = ref(false)
const flowReady = ref(false)

const STAGE_ORDER = ['lab', 'pilot', 'industrialized', 'listed']

const stages = ['实验室', '中试', '产业化', '已上市'] // 阶段字面同系统：已上市
const rank = computed(() => (t.value ? STAGE_RANK[t.value.stage] || 0 : 0))
// 进度语义统一（P1-1）：百分比与轨道同源——满宽 = 走完前 3 段（rank 1→0%，rank 4→100%）
// 文案用「已推进至第 X/4 阶段」表位置、百分比表阶段完成度，一屏一套真理
const pct = computed(() => (rank.value > 0 ? Math.round(((rank.value - 1) / 3) * 100) : 0))
const flowRatio = computed(() => (rank.value > 0 ? (rank.value - 1) / 3 : 0))
const isOwner = computed(() => !!uid.value && !!t.value && String(uid.value) === String(t.value.owner_id))
const nextStage = computed(() => (t.value ? STAGE_ORDER[rank.value] || '' : ''))
const stageLabel = computed(() => {
  const s = (t.value?.stage || '').toLowerCase()
  return STAGE_SHORT[s] || t.value?.stage || ''
})
// 四段签名色同列表页：灰/琥珀/绿/深蓝
const stageCls = computed(() => {
  const s = (t.value?.stage || '').toLowerCase()
  if (s === 'lab' || s === 'laboratory') return 'cl-la'
  if (s === 'pilot') return 'cl-pi'
  if (s === 'listed') return 'cl-li'
  return 'cl-in'
})
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
const curIndex = computed(() => {
  const list = t.value?.milestones || []
  const idx = list.findIndex((m) => !m.completed)
  return idx >= 0 ? idx : list.length - 1
})

const mapT = (it) => ({
  id: it.id,
  achievement_id: it.achievement_id || '',
  title: it.title || '成果转化',
  stage: it.stage || 'lab',
  progress: it.progress || '',
  partner_id: it.partner_id || '',
  contact_info: it.contact_info || '',
  status: it.status || 'active',
  created_at: it.created_at || '',
  updated_at: it.updated_at || '',
  milestones: Array.isArray(it.milestones) ? it.milestones : [],
})

// ===== 数据获取 =====
// 请求超时兜底：挂起/断网时 10s 内转为错误态，避免永久骨架（同详情页 withTimeout）
const withTimeout = (p, ms = 10000) => new Promise((resolve, reject) => {
  const timer = setTimeout(() => reject(new Error('请求超时，请重试')), ms)
  p.then(
    (v) => { clearTimeout(timer); resolve(v) },
    (e) => { clearTimeout(timer); reject(e) }
  )
})

// 接口替换点：GET /api/v1/transformations?achievement_id=
const fetchData = async () => {
  if (!achievementId.value) { loading.value = false; err.value = true; return }
  loading.value = true
  err.value = false
  try {
    const res = await withTimeout(request({ url: '/api/v1/transformations', data: { achievement_id: achievementId.value } }))
    const list = Array.isArray(res) ? res : (res?.data || [])
    const found = (tid.value ? list.find((x) => x.id === tid.value) : null) || list[0]
    applyT(found ? mapT(found) : null)
  } catch {
    useMock()
  } finally {
    loading.value = false
  }
}

const applyT = (v) => {
  t.value = v
  flowReady.value = false
  setTimeout(() => { flowReady.value = true }, 300)
}

const useMock = () => {
  // P2-6：生产环境绝不回退演示数据，失败即诚实错误态
  if (process.env.NODE_ENV === 'production') { err.value = true; return }
  const list = MOCK_TRANSFORMS_BY_ACH[achievementId.value] || []
  const found = (tid.value ? list.find((x) => x.id === tid.value) : null) || list[0]
  applyT(found ? mapT(found) : null)
}

const onContact = () => {
  // 对接是产品引擎，但复制前置确认门槛（与详情页同一道门，见 utils/contactMask.js）——
  // 未经确认不把完整联系方式写进剪贴板；未公开时给去向，不给空话
  if (revealContactCopy(t.value?.contact_info)) return
  uni.showModal({
    title: '联系方式暂未公开',
    content: '发布方暂未公开联系方式。可返回成果详情页查看该成果的联系信息，或稍后重试。',
    confirmText: '知道了',
    showCancel: false,
  })
}
const goBack = () => uni.navigateBack()

// ===== 推进下一阶段（仅发布方可操作） =====
// 接口替换点：POST /api/v1/transformations/{id}/advance (body: stage / progress)
const onAdvance = () => {
  if (!isOwner.value || !t.value) return
  if (!nextStage.value) {
    uni.showToast({ title: '已达最终阶段', icon: 'none' })
    return
  }
  uni.showModal({
    title: '推进下一阶段',
    content: '将推进到「' + (STAGE_SHORT[nextStage.value] || nextStage.value) + '」，请填写当前进度说明：',
    editable: true,
    placeholderText: '如：完成小批量试产，进入量产验证',
    success: async (res) => {
      if (!res.confirm) return
      const progress = (res.content || '').trim()
      if (!progress) {
        uni.showToast({ title: '请填写进度说明', icon: 'none' })
        return
      }
      submitting.value = true
      try {
        await request({
          url: '/api/v1/transformations/' + encodeURIComponent(t.value.id) + '/advance',
          method: 'POST',
          data: { stage: nextStage.value, progress },
        })
        // 本地乐观联动：把当前第一个未完成里程碑标记为完成（对齐原型交互；后端 advance 不更新里程碑）
        const ms = (t.value.milestones || []).slice()
        const idx = ms.findIndex((m) => !m.completed)
        if (idx >= 0) ms[idx] = { ...ms[idx], completed: true, date: new Date().toISOString() }
        applyT({ ...t.value, stage: nextStage.value, progress, updated_at: new Date().toISOString(), milestones: ms })
        uni.showToast({ title: '已推进下一阶段', icon: 'success' })
        // 以后端为准同步阶段/进度；若后端里程碑未变，则保留本地完成标记
        await fetchData()
        const cur = t.value
        if (cur && cur.milestones && cur.milestones.length === ms.length) {
          const curIdx = cur.milestones.findIndex((m) => !m.completed)
          if (curIdx === idx && ms[idx] && ms[idx].completed) {
            const ms2 = cur.milestones.slice()
            ms2[idx] = { ...ms2[idx], completed: true }
            t.value = { ...cur, milestones: ms2 }
          }
        }
      } catch {
        uni.showToast({ title: '推进失败，请重试', icon: 'none' })
      } finally {
        submitting.value = false
      }
    },
  })
}

onLoad((options) => {
  if (options?.achievement_id) achievementId.value = decodeURIComponent(options.achievement_id)
  if (options?.id) tid.value = decodeURIComponent(options.id)
  const user = getStoredUser()
  uid.value = (user && (user.id || user.user_id)) || ''
  checkMotion()
  fetchData()
})
</script>

<style>
page { background: var(--color-bg); }
</style>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ===== 概览（刊头渐变：Gradient-Lock 唯一深色刊头） ===== */
.t-hero { position: relative; padding: 48rpx 40rpx 44rpx; background: linear-gradient(160deg, #0a3a6b, #074d92); color: #fff; }
.t-back { position: absolute; left: 8rpx; top: 12rpx; width: 88rpx; height: 88rpx; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23ffffff' stroke-width='2.2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M15 4l-8 8 8 8'/%3E%3C/svg%3E"); background-size: 44rpx; background-repeat: no-repeat; background-position: center; }
.t-back:active { opacity: .6; }
.t-tag { display: inline-block; font-size: 20rpx; padding: 4rpx 16rpx; border-radius: 8rpx; background: rgba(255,255,255,.2); color: #fff; margin-bottom: 16rpx; font-weight: 700; }
.t-title { font-size: 36rpx; font-weight: 700; line-height: 1.4; display: block; }
.t-sub { font-size: 24rpx; color: rgba(255,255,255,.82); margin-top: 12rpx; display: block; }

/* ===== 区块（对齐组卡片：描边 + 20rpx + 蓝调双层） ===== */
.sec { margin: 24rpx 32rpx; padding: 32rpx; background: #fff; border: 2rpx solid #E4E7EC; border-radius: 20rpx; box-shadow: 0 2rpx 6rpx rgba(10,30,60,.04), 0 12rpx 32rpx rgba(10,30,60,.05); }
.sh { display: flex; align-items: center; gap: 16rpx; margin-bottom: 24rpx; }
.sd { width: 8rpx; height: 36rpx; background: var(--color-primary); border-radius: 2rpx; flex-shrink: 0; }
.sht { font-size: 32rpx; font-weight: 700; color: var(--color-text); }
.ms-count { margin-left: auto; font-size: 24rpx; color: #667085; font-weight: 400; }

/* ===== 转化阶段：虚线轨道 + 进度动画（flex/gap 自适应间距） ===== */
.tr-flow { position: relative; display: flex; flex-direction: column; gap: 16rpx; }
.tr-track { position: relative; padding: 6rpx 0 10rpx; }
.tr-base { position: absolute; left: 12.5%; right: 12.5%; top: 10rpx; border-top: 2rpx dashed #DDE1E6; z-index: 0; }
.tr-prog { position: absolute; left: 12.5%; top: 9rpx; height: 4rpx; background: var(--color-primary); border-radius: 2rpx; z-index: 1; transform-origin: left center; transform: scaleX(0); transition: transform .4s cubic-bezier(0.16, 1, 0.3, 1); }
.tr-stages { display: flex; justify-content: space-between; position: relative; z-index: 2; }
.tr-stage { display: flex; flex-direction: column; align-items: center; gap: 8rpx; width: 25%; }
.tr-dot { width: 20rpx; height: 20rpx; border-radius: 50%; background: #fff; border: 4rpx solid var(--color-divider); box-sizing: border-box; transition: background .3s ease, border-color .3s ease, box-shadow .3s ease; }
.tr-stage.done .tr-dot { background: var(--color-primary); border-color: var(--color-primary); }
/* 当前点：静态实心 + 光晕环（同详情页；循环动效保留给「紧急」信号，见 DESIGN.md Motion 治理） */
.tr-stage.cur .tr-dot { background: var(--color-primary); border-color: var(--color-primary); box-shadow: 0 0 0 6rpx var(--color-primary-light); }
.tr-stage-name { font-size: 20rpx; color: #667085; }
.tr-stage.done .tr-stage-name, .tr-stage.cur .tr-stage-name { color: var(--color-primary); font-weight: 600; }
.tr-meta { display: flex; justify-content: space-between; align-items: center; font-size: 24rpx; color: #667085; }
.tr-cur { color: var(--color-primary); font-weight: 600; }

/* ===== 信息行 ===== */
.t-row { display: flex; gap: 12rpx; padding: 14rpx 0; border-bottom: 1rpx solid var(--color-divider); }
.t-row:last-child { border-bottom: none; }
.t-k { width: 140rpx; flex-shrink: 0; font-size: 24rpx; color: #667085; }
.t-v { flex: 1; font-size: 28rpx; color: var(--color-text); line-height: 1.5; }
/* 四段签名色同列表页：实验室灰 / 中试琥珀 / 产业化绿 / 已上市深蓝 */
.t-v.cl-la { color: #5D6B82; font-weight: 600; }
.t-v.cl-pi { color: #B45309; font-weight: 600; }
.t-v.cl-in { color: #1F7A48; font-weight: 600; }
.t-v.cl-li { color: #074D92; font-weight: 600; }

/* ===== 里程碑 ===== */
.ms-list { display: flex; flex-direction: column; }
/* 里程碑图标一律使用普通 view + text，禁止 cover-view（cover-view 层级独立，会穿透覆盖到底部操作栏等普通 view 之上） */
.ms-item { display: flex; gap: 16rpx; padding: 18rpx 0; position: relative; }
/* 连接线 = 上段（本项顶部 → 圆点中心）+ 下段（圆点中心 → 本项底部），两段拼成连续线。
   只锚定圆点中心（padding-top 18rpx + 半径 20rpx = 38rpx），与文本高度无关：换行/滚动不会错位或穿透。 */
.ms-item + .ms-item::before { content: ''; position: absolute; left: 19rpx; top: 0; height: 38rpx; width: 2rpx; background: var(--color-divider); z-index: 0; }
.ms-item:not(:last-child)::after { content: ''; position: absolute; left: 19rpx; top: 38rpx; bottom: 0; width: 2rpx; background: var(--color-divider); z-index: 0; }
.ms-ic { width: 40rpx; height: 40rpx; border-radius: 50%; flex: none; display: flex; align-items: center; justify-content: center; font-size: 24rpx; font-weight: 700; position: relative; z-index: 1; }
.ms-item.done .ms-ic { background: var(--color-success); color: #fff; }
.ms-item.cur .ms-ic { background: var(--color-primary); color: #fff; box-shadow: 0 0 0 6rpx var(--color-primary-light); }
.ms-item.todo .ms-ic { background: var(--color-bg); color: var(--color-text-placeholder); border: 1rpx solid var(--color-border); }
.ms-bd { flex: 1; min-width: 0; position: relative; z-index: 1; }
.ms-name { display: block; font-size: 28rpx; color: var(--color-text); line-height: 1.4; }
.ms-date { display: block; font-size: 20rpx; color: #667085; margin-top: 4rpx; }
.ms-ev { display: block; font-size: 20rpx; color: var(--color-primary); margin-top: 4rpx; }
.ms-item.todo .ms-name { color: var(--color-text-placeholder); }

/* ===== 底部操作栏 ===== */
.bb { position: sticky; bottom: 0; z-index: 60; background: #fff; border-top: 1rpx solid var(--color-divider); display: flex; align-items: center; padding: 20rpx 32rpx; gap: 20rpx; padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); box-shadow: 0 -2rpx 12rpx rgba(0,0,0,.04); }
.bo { height: 88rpx; border-radius: 16rpx; border: 2rpx solid var(--color-primary); background: #fff; color: var(--color-primary); font-size: 28rpx; font-weight: 600; padding: 0 32rpx; display: flex; align-items: center; flex-shrink: 0; }
.bo:active { background: var(--color-primary-light); }
.bp { flex: 1; height: 88rpx; border-radius: 16rpx; background: var(--color-primary); color: #fff; font-size: 28rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; }
.bp:active { transform: scale(.97); }
/* 禁用态成文：占位灰实底 + 去 gloss（DESIGN.md Disabled：占位灰文字，不做半透明化） */
.bp.disabled { background: var(--color-text-placeholder); box-shadow: none; }

/* ===== 骨架 ===== */
.skw { padding-top: 20rpx; }
.sk-h { height: 240rpx; background: linear-gradient(160deg, #0a3a6b, #074d92); animation: shimmer 1.5s ease; }
.sk-sec { margin: 24rpx 32rpx; padding: 32rpx; background: #fff; border: 2rpx solid #E4E7EC; border-radius: 20rpx; }
.sk-l { height: 28rpx; background: #EDF0F3; border-radius: 8rpx; margin-bottom: 16rpx; animation: shimmer 1.5s ease; }
.sk-l.w80 { width: 80%; }
.sk-l.w100 { width: 100%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }
@keyframes shimmer { 0%, 100% { opacity: 1; } 50% { opacity: .45; } }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 200rpx 40rpx; min-height: 800rpx; }
.stb { min-height: 88rpx; padding: 16rpx 48rpx; border-radius: 16rpx; background: var(--color-primary); color: #fff; font-size: 28rpx; font-weight: 500; display: flex; align-items: center; justify-content: center; }
.stb:active { opacity: .8; }

/* ===================== UI/UX 体验优化（仅新增/修改 wxss，不动模板/数据/逻辑） ===================== */
/* 动画统一 200-400ms；优先 transform/opacity；生产级轻量克制 */

/* 1) 入场动画：里程碑条目依次淡入 + 上移（20ms 错开，≤6 项；backwards 不阻塞点击态） */
.ms-item { animation: uiMsIn .3s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
.ms-item:nth-child(1) { animation-delay: 0ms; }
.ms-item:nth-child(2) { animation-delay: 20ms; }
.ms-item:nth-child(3) { animation-delay: 40ms; }
.ms-item:nth-child(4) { animation-delay: 60ms; }
.ms-item:nth-child(5) { animation-delay: 80ms; }
.ms-item:nth-child(6) { animation-delay: 100ms; }
@keyframes uiMsIn { from { opacity: 0; transform: translateY(10rpx); } to { opacity: 1; transform: translateY(0); } }

/* 区块轻量入场（一次，licensed 曲线） */
.sec { animation: uiSecIn .3s cubic-bezier(0.16, 1, 0.3, 1) backwards; }
@keyframes uiSecIn { from { opacity: 0; transform: translateY(10rpx); } to { opacity: 1; transform: translateY(0); } }

/* 2) 交互反馈：按钮按压轻微缩放 + 透明度（200ms） */
.bo { transition: transform .2s ease, background .2s ease, color .2s ease; }
.bo:active { transform: scale(.97); opacity: .9; }
.bp { transition: transform .2s ease, box-shadow .2s ease, opacity .2s ease; }
.bp:active { transform: scale(.97); opacity: .92; }
.stb { transition: transform .2s ease, opacity .2s ease; }
.stb:active { transform: scale(.95); opacity: .85; }

/* 3) 状态过渡：勾选图标出现（done 时 300ms 弹簧弹出，ios-pop 仅 transform）+ 图标容器过渡 */
.ms-ic { transition: transform .3s cubic-bezier(0.34, 1.8, 0.64, 1), background .3s ease, box-shadow .3s ease; }
.ms-item.done .ms-ic { animation: uiCheck .3s cubic-bezier(0.34, 1.8, 0.64, 1); }
@keyframes uiCheck { 0% { transform: scale(.5); opacity: 0; } 100% { transform: scale(1); opacity: 1; } }

/* 4) 渲染视觉细节 + 层级加固：连线在最底层、图标/文字在上、底部栏置顶（防穿透覆盖） */
.ms-item { position: relative; }
.ms-ic { position: relative; z-index: 1; }
.ms-bd { position: relative; z-index: 1; }
.ms-item + .ms-item::before { z-index: 0; }
.ms-item:not(:last-child)::after { z-index: 0; }
.bb { z-index: 60; }

/* ===== 【对齐组范式】主行动按钮 gloss：内高光 + 内厚度 + 品牌蓝晕（如需回退删除本块即可） ===== */
.bp { box-shadow: 0 4rpx 14rpx rgba(10,102,194,.25), inset 0 2rpx 0 rgba(255,255,255,.2), inset 0 -4rpx 10rpx rgba(7,77,146,.18); }
.stb { box-shadow: 0 4rpx 14rpx rgba(10,102,194,.25), inset 0 2rpx 0 rgba(255,255,255,.2), inset 0 -4rpx 10rpx rgba(7,77,146,.18); }

/* ===== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈（同 list.vue） ===== */
.page.no-motion .ms-item { animation: none; }
.page.no-motion .sec { animation: none; }
.page.no-motion .tr-prog { transition: none; }
.page.no-motion .tr-dot { transition: none; }
.page.no-motion .ms-ic { transition: none; animation: none; }
.page.no-motion .bo,
.page.no-motion .bp,
.page.no-motion .stb { transition: none; }
.page.no-motion .bo:active,
.page.no-motion .bp:active,
.page.no-motion .stb:active { transform: none; }
</style>
