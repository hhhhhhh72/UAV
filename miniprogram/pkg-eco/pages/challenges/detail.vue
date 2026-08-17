<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="难题详情" show-back :fixed="true" @back="goBack" />

    <!-- 骨架 -->
    <view v-if="loading" class="skw">
      <view class="sk-h"></view>
      <view class="sk-sec"><view class="sk-l w80"></view><view class="sk-l w100"></view><view class="sk-l w60"></view></view>
      <view class="sk-sec"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w80"></view></view>
    </view>

    <!-- 错误 -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- 空 -->
    <view v-else-if="!d" class="st">
      <u-empty description="该难题已下架或不存在">
        <view class="stb" @tap="goBack">返回列表</view>
      </u-empty>
    </view>

    <template v-else>
      <!-- 深蓝 Hero -->
      <view class="hero">
        <view class="h-tags">
          <text class="h-tag">{{ d.f }}</text>
          <text class="h-tag" :class="'st-' + d.stCls">{{ d.stLabel }}</text>
        </view>
        <text class="h-title">{{ d.t }}</text>
        <view class="h-budget">
          <text class="h-vl">{{ d.budgetText }}</text>
          <text class="h-lb">悬赏金额 · 择优揭榜</text>
        </view>
      </view>

      <!-- 信息卡 -->
      <view class="info">
        <view class="stat">
          <view class="si"><text class="sv">{{ d.budgetText }}</text><text class="sl">悬赏金额</text></view>
          <view class="si"><text class="sv">{{ d.daysLeft }}</text><text class="sl">剩余截止</text></view>
          <view class="si"><text class="sv">{{ d.dateShort }}</text><text class="sl">发布时间</text></view>
        </view>
        <view class="row">
          <view class="ic">主</view>
          <view class="it"><text class="il">发布单位</text><text class="iv">{{ d.organizer }}</text></view>
        </view>
        <view class="row">
          <view class="ic orange">域</view>
          <view class="it"><text class="il">所属领域</text><text class="iv">{{ d.f }}</text></view>
        </view>
        <view class="row">
          <view class="ic green">止</view>
          <view class="it"><text class="il">截止日期</text><text class="iv">{{ d.deadlineText }}</text></view>
        </view>
      </view>

      <!-- 难题描述 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">难题描述</text></view>
        <text class="p">{{ d.desc || '暂无描述' }}</text>
      </view>

      <!-- 攻关要求 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">攻关要求</text></view>
        <view v-if="d.reqs.length" class="ul">
          <view v-for="(r, i) in d.reqs" :key="i" class="li">
            <view class="dot"></view>
            <text class="li-t">{{ r }}</text>
          </view>
        </view>
        <text v-else class="p dim">暂无具体指标，详情以发布单位沟通为准</text>
      </view>

      <!-- 揭榜流程 -->
      <view class="sec">
        <view class="sh"><view class="sd"></view><text class="sht">揭榜流程</text></view>
        <view class="steps">
          <view class="step"><view class="no">1</view><text class="txt">提交<br/>揭榜意向</text></view>
          <view class="step"><view class="no">2</view><text class="txt">协会<br/>审核对接</text></view>
          <view class="step"><view class="no">3</view><text class="txt">达成合作<br/>攻关结题</text></view>
        </view>
      </view>
      <view style="height: 100px"></view>

      <!-- 底部操作栏（布局对齐 achievements/detail：底部吸附 + safe-area + 42px 按钮） -->
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" aria-role="button" :aria-label="isFav ? '取消收藏' : '收藏'" @tap="toggleFav"><text class="bit">{{ isFav ? '♥' : '♡' }}</text></view>
        <button class="bo" open-type="share" hover-class="bo-hover" hover-start-time="0" hover-stay-time="300">转发</button>
        <view class="bp" @tap="onContact">联系发布者</view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { MOCK_CHALLENGES } from '@/utils/mockChallenges'
import { useReduceMotion } from '@/utils/motion'

const loading = ref(true)
const err = ref(false)
const d = ref(null)
const isFav = ref(false)
const statusBarHeight = ref(20)
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
let id = ''
/* 收藏持久化：本地存储兜底（后端收藏接口就绪前的纯前端实现），按难题 id 去重 */
const FAV_KEY = 'challenge_favs'
const loadFavs = () => {
  try {
    const v = uni.getStorageSync(FAV_KEY)
    return Array.isArray(v) ? new Set(v) : new Set()
  } catch (e) { return new Set() }
}
const favs = loadFavs()
const saveFavs = () => {
  try { uni.setStorageSync(FAV_KEY, [...favs]) } catch (e) { /* 忽略 */ }
}

const pad = (n) => (n < 10 ? '0' + n : '' + n)
const daysLeft = (dt) => {
  if (!dt) return null
  const diff = new Date(dt) - new Date()
  return Number.isFinite(diff) ? Math.max(0, Math.ceil(diff / 86400000)) : null
}
const statusOf = (it) => {
  const s = String(it.status || '').toLowerCase()
  const dl = daysLeft(it.deadline)
  if (s === 'closed' || s === '已截止') return { label: '已截止', cls: 'closed' }
  if (s === 'urgent' || s === '紧急' || (dl != null && dl <= 7)) return { label: '紧急', cls: 'urgent' }
  return { label: '进行中', cls: 'open' }
}
const fmtMoney = (wan) => {
  if (wan == null || wan <= 0) return '面议'
  if (wan >= 1) return '¥' + (wan % 1 === 0 ? wan : wan.toFixed(1)) + '万'
  return '¥' + Math.round(wan * 10000)
}
// 攻关要求：仅展示后端真实提供的 requirements；缺省时走模板空态文案，
// 不伪造指标（协会公信力铁律——绝不把平台编造的技术规格当作发布方要求展示）

const mapItem = (it) => {
  const dl = daysLeft(it.deadline)
  const st = statusOf(it)
  const wan = it.budget_fen != null ? it.budget_fen / 100 / 10000 : 0
  let reqs = []
  if (Array.isArray(it.requirements) && it.requirements.length) reqs = it.requirements
  else if (typeof it.requirements === 'string' && it.requirements.trim()) {
    reqs = it.requirements.split(/[\n;；]/).map((s) => s.trim()).filter(Boolean)
  }
  return {
    id: it.id,
    t: it.title || '未命名难题',
    f: it.field || '其他',
    desc: it.description || '',
    budgetText: fmtMoney(wan),
    stLabel: st.label,
    stCls: st.cls,
    daysLeft: dl == null ? '待定' : dl + ' 天',
    dateShort: (it.created_at || '').slice(5, 10),
    deadlineText: (it.deadline || '').slice(0, 10) + (st.cls === 'closed' ? ' · 已截止' : ' · 逾期不再受理'),
    organizer: it.poster_name || '协会会员企业',
    reqs,
  }
}

const fetchData = async () => {
  loading.value = true
  err.value = false
  try {
    const res = await request({ url: '/api/v1/challenges/' + encodeURIComponent(id) })
    const it = (res && res.data) || res
    if (it && it.id) d.value = mapItem(it)
    else {
      // 接口不可用时回退演示数据（仅开发环境）
      if (import.meta.env.DEV) {
        const mock = (MOCK_CHALLENGES || []).find((x) => x.id === id)
        d.value = mock ? mapItem(mock) : null
      } else {
        err.value = true
      }
    }
  } catch {
    // 回退演示数据（仅开发环境）
    if (import.meta.env.DEV) {
      const mock = (MOCK_CHALLENGES || []).find((x) => x.id === id)
      d.value = mock ? mapItem(mock) : null
    }
    if (!d.value) err.value = true
  } finally {
    loading.value = false
  }
}

const toggleFav = () => {
  isFav.value = !isFav.value
  if (isFav.value) {
    favs.add(id)
  } else {
    favs.delete(id)
  }
  saveFavs()
  uni.showToast({ title: isFav.value ? '已收藏' : '已取消收藏', icon: 'none' })
}
onShareAppMessage(() => ({
  title: d.value ? '研发难题：' + d.value.t : '低空经济生态服务平台 · 研发难题广场',
  path: '/pkg-eco/pages/challenges/detail?id=' + encodeURIComponent(id),
}))
const onContact = () => {
  uni.showToast({ title: '已通知协会，将为您对接发布企业', icon: 'none' })
}
const goBack = () => uni.navigateBack()

onLoad((options) => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  id = options?.id || ''
  isFav.value = favs.has(id)
  fetchData()
})
</script>

<style>
page {
  background: #f5f6f8;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #f5f6f8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===== 深蓝 Hero ===== */
.hero {
  position: relative;
  min-height: 180px;
  padding: 18px 16px 56px;
  color: #fff;
  background: linear-gradient(140deg, #17314A 0%, #0F2A44 45%, #0A66C2 165%);
  overflow: hidden;
  box-shadow: 0 6px 18px rgba(7, 77, 146, 0.22);
}
.h-tags { display: flex; gap: 6px; margin-bottom: 10px; }
.h-tag {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.92);
  color: #074D92;
}
.h-tag.st-open { color: #168A55; }
.h-tag.st-urgent { color: #D92D20; }
.h-tag.st-closed { color: #667085; }
.h-title { display: block; font-size: 19px; font-weight: 800; line-height: 1.35; }
.h-budget { display: flex; align-items: baseline; gap: 6px; margin-top: 10px; }
.h-vl { font-size: 26px; font-weight: 800; color: #FFD166; }
/* 悬赏金额数字：hero 的心脏——落位时 ios-pop 弹簧弹出一次（80ms 起播 .25s，总 330ms ≤ 400ms） */
.h-vl { animation: popIn .25s cubic-bezier(.34, 1.8, .64, 1) 80ms backwards; }
/* 95% 白：蓝端渐变上 ≥4.5:1（原 78% 仅约 3.6:1，AA 不达标） */
.h-lb { font-size: 11px; color: rgba(255, 255, 255, 0.95); }

/* ===== 信息卡 ===== */
.info {
  position: relative;
  z-index: 5;
  margin: -32px 12px 0;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
  padding: 12px;
}
.stat { display: flex; padding-bottom: 10px; border-bottom: 1px solid #F2F3F5; margin-bottom: 4px; }
.si { flex: 1; text-align: center; position: relative; }
.si + .si::before { content: ''; position: absolute; left: 0; top: 6px; bottom: 6px; width: 0.5px; background: #F0F0F0; }
.sv { font-size: 15px; font-weight: 800; color: #17212B; display: block; }
.sl { font-size: 11px; color: #667085; margin-top: 2px; display: block; }
.row { display: flex; align-items: flex-start; gap: 9px; padding: 8px 0; border-bottom: 0.5px solid #F5F5F5; }
.row:last-child { border-bottom: none; }
.ic {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: #EAF3FB;
  color: #0A66C2;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: none;
  font-size: 11px;
  font-weight: 700;
}
.ic.orange { background: #FFF0E6; color: #E96012; }
.ic.green { background: #E9F7F0; color: #168A55; }
.it { flex: 1; min-width: 0; }
/* 标签对比度：原 #98A2B3 白底 2.6:1 违 AA，升 #667085（4.9:1）+ 字号 9px→11px */
.il { font-size: 11px; color: #667085; display: block; margin-bottom: 1px; }
.iv { font-size: 12.5px; color: #17212B; font-weight: 500; line-height: 1.4; display: block; }

/* ===== 区块 ===== */
.sec {
  margin: 10px 12px 0;
  padding: 13px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
}
.sh { display: flex; align-items: center; gap: 7px; margin-bottom: 10px; }
.sd { width: 3px; height: 14px; background: #0A66C2; border-radius: 2px; }
.sht { font-size: 14.5px; font-weight: 700; color: #17212B; }
.p { font-size: 12.5px; color: #667085; line-height: 1.7; display: block; }
.p.dim { color: #5D6B82; } /* 空态灰：#5D6B82 白底 5.4:1（原 #98A2B3 2.6:1 违 AA），比正文 #667085 更稳但不失弱化感 */
.ul { display: flex; flex-direction: column; }
.li { display: flex; align-items: flex-start; gap: 8px; padding: 0 0 9px; }
.li:last-child { padding-bottom: 0; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: #0A66C2; margin-top: 6px; flex: none; }
.li-t { font-size: 12px; color: #667085; line-height: 1.6; flex: 1; }

/* ===== 揭榜流程 ===== */
.steps { display: flex; align-items: flex-start; }
.step { flex: 1; text-align: center; position: relative; }
.step .no {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 6px;
}
.step .txt { font-size: 11px; color: #667085; line-height: 1.4; }
.step::after {
  content: '';
  position: absolute;
  top: 13px;
  left: calc(50% + 16px);
  right: calc(-50% + 16px);
  height: 2px;
  background: #EAF3FB;
}
.step:last-child::after { display: none; }

/* ===== 底部操作栏（参考 achievements/detail 布局） ===== */
.bb {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 50;
  background: #fff;
  border-top: 0.5px solid #F0F0F0;
  display: flex;
  align-items: center;
  padding: 10px 16px;
  gap: 10px;
  padding-bottom: calc(10px + env(safe-area-inset-bottom));
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.04);
}
.bi {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: #667085; /* 未收藏心形：原 #98A2B3 2.6:1 低于非文本控件 3:1，升 #667085（4.9:1） */
}
.bi.fv { color: #ff3b30; }
.bit { font-size: 18px; line-height: 1; }
.bo {
  height: 42px;
  border-radius: 8px;
  border: 1px solid #0A66C2;
  background: #fff;
  color: #0A66C2;
  font-size: 13px;
  font-weight: 600;
  padding: 0 16px;
  margin: 0;
  line-height: 1;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
/* open-type=share 按钮：清除小程序 button 默认伪元素边框与 hover 底色，按压反馈走 hover-class（:active 对 button 不生效） */
.bo::after { border: none; }
.bo-hover { transform: scale(.95); background: #F4F8FC; }
.bp {
  flex: 1;
  height: 42px;
  border-radius: 8px;
  background: #0A66C2;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(10, 102, 194, 0.3);
}

/* ===== 骨架 ===== */
.skw { padding-top: 10px; }
.sk-h { height: 200px; margin: 12px; border-radius: 12px; background: #EDF0F3; }
.sk-sec { margin: 12px; padding: 16px; background: #fff; border-radius: 10px; }
.sk-l { height: 14px; background: #EDF0F3; border-radius: 4px; margin-bottom: 10px; }
.sk-l.w80 { width: 80%; }
.sk-l.w100 { width: 100%; }
.sk-l.w60 { width: 60%; }
.sk-l.w40 { width: 40%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120px 20px; }
.stb { padding: 8px 24px; border-radius: 8px; background: #0A66C2; color: #fff; font-size: 13px; font-weight: 500; margin-top: 16px; }

/* ===================== 动效规范（对齐全局动画规范） =====================
   白名单：仅 transform / opacity（小尺寸元素 color/background 过渡允许——仅重绘不重排）
   禁参与动画：top/left/width/height/margin（触发重排）、box-shadow/filter（低端安卓掉帧）
   时长：微反馈 150-200ms（按压按下 .08s 即时到位）/ 松手弹簧回位 .3s（ios-pop）/ 浮层 200-300ms / 页面级 ≤400ms；
        退场 = 进场 ×0.7 且必须存在
   曲线：两枚固定曲线——ios-pop cubic-bezier(.34,1.8,.64,1) 松手弹簧回弹（仅按压/弹出类 transform）+
        ios-decel cubic-bezier(.32,.72,0,1) 浮层流体减速（底部操作栏进场）；
        其余进场 ease-out / 退场 ease-in / 循环 linear；除这两枚外禁手写 cubic-bezier
   数量：入场错峰首屏可见项，整页编排 ≤400ms；循环动画任何时刻全页 ≤1 处
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) 入场编排（hero 0ms → 信息卡 50ms → 三个区块 60/100/140ms，总 390ms ≤ 400ms）
   backwards 填充 → 延迟期保持隐藏不闪跳 */
.hero { animation: fadeUp .25s ease-out backwards; }
.info { animation: fadeUp .25s ease-out .05s backwards; }
.sec { animation: fadeUp .25s ease-out 60ms backwards; }
/* 三个 .sec 相邻兄弟，用 + 选择器级联（不依赖 nth-child 数节点，nav 组件混入也安全） */
.sec + .sec { animation-delay: 100ms; }
.sec + .sec + .sec { animation-delay: 140ms; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-10px); } to { opacity: 1; transform: translateY(0); } }

/* 攻关要求条目：从左侧淡入，前 6 条 20ms 错峰（总 300ms ≤ 400ms；超出立即入场） */
.li { animation: liIn .2s ease-out backwards; }
.li:nth-child(1) { animation-delay: 0ms; }
.li:nth-child(2) { animation-delay: 20ms; }
.li:nth-child(3) { animation-delay: 40ms; }
.li:nth-child(4) { animation-delay: 60ms; }
.li:nth-child(5) { animation-delay: 80ms; }
.li:nth-child(6) { animation-delay: 100ms; }
@keyframes liIn { from { opacity: 0; transform: translateX(-10px); } to { opacity: 1; transform: translateX(0); } }

/* 2) 揭榜流程 3 步：依次淡入上移（40ms 错峰）+ 连线自左向右 scaleX 生长 + 步骤圈 ios-pop 弹簧弹出
   （总 380ms ≤ 400ms；scaleX/scale 均 transform，零重排；弹簧曲线仅作用于 transform） */
.step { animation: stIn .2s ease-out backwards; }
.step:nth-child(1) { animation-delay: 0ms; }
.step:nth-child(2) { animation-delay: 40ms; }
.step:nth-child(3) { animation-delay: 80ms; }
@keyframes stIn { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
.step::after { animation: lineGrow .25s ease-out backwards; transform-origin: left center; }
.step:nth-child(2)::after { animation-delay: 40ms; }
@keyframes lineGrow { from { transform: scaleX(0); } to { transform: scaleX(1); } }
.step .no { animation: popIn .3s cubic-bezier(.34, 1.8, .64, 1) backwards; } /* ios-pop：步骤圈弹簧弹出 */
.step:nth-child(2) .no { animation-delay: 40ms; }
.step:nth-child(3) .no { animation-delay: 80ms; }
@keyframes popIn { from { transform: scale(.6); opacity: 0; } to { transform: scale(1); opacity: 1; } }

/* 统计数字：信息卡落位后弹簧弹出（40ms 错峰，总 370ms ≤ 400ms；ios-pop 数字过冲回位） */
.sv { animation: popIn .25s cubic-bezier(.34, 1.8, .64, 1) 40ms backwards; }
.si:nth-child(2) .sv { animation-delay: 80ms; }
.si:nth-child(3) .sv { animation-delay: 120ms; }

/* 章节标题条：自顶部下压弹出（scaleY，transform 零重排） */
.sd { transform-origin: top; animation: barDrop .25s ease-out 60ms backwards; }
@keyframes barDrop { from { transform: scaleY(0); } to { transform: scaleY(1); } }

/* 底部操作栏：自下而上滑入（fixed 定位，transform 不影响布局；ios-decel 流体减速——sheet 落地感） */
.bb { animation: bbUp .3s cubic-bezier(.32, .72, 0, 1) backwards; }
@keyframes bbUp { from { opacity: 0; transform: translateY(24px); } to { opacity: 1; transform: translateY(0); } }

/* 骨架呼吸（加载中环境光；循环 1.2-1.6s linear） */
.sk-h, .sk-sec, .sk-l { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 3) 交互反馈：按压反馈（按下 .08s linear 即时到位；松手 .3s ios-pop 弹簧回位——与 list 同套手感） */
.stb { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; } /* ios-pop */
.stb:active { transform: scale(.94); opacity: .85; transition: transform .08s linear; }
.bi { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, color .2s ease; } /* ios-pop */
.bi:active { transform: scale(.9); background: #EAF3FB; transition: transform .08s linear; }
.bi.fv:active { background: #FDECEC; } /* 已收藏时按压给红色系反馈 */
.bo { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, color .2s ease; } /* ios-pop */
.bo:active { transform: scale(.95); background: #F4F8FC; transition: transform .08s linear; }
.bp { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; } /* ios-pop */
.bp:active { transform: scale(.95); opacity: .92; transition: transform .08s linear; }

/* 4) 状态过渡：收藏点亮时心形 ios-pop 弹簧弹出（scale .8→1 自然过冲回位，iOS 收藏手感）；取消收藏无反向动画 */
.bi.fv .bit { animation: heartPop .3s cubic-bezier(.34, 1.8, .64, 1); }
@keyframes heartPop { from { transform: scale(.8); } to { transform: scale(1); } }

/* 5) 视觉层级：信息卡与 Hero 重叠处补柔和投影，强化"卡片浮在 Hero 上"的层次（纯视觉，无布局影响） */
.info { box-shadow: 0 4px 12px rgba(16, 24, 40, 0.05); }

/* 6) "紧急"状态标签呼吸——全页唯一循环动画（语义反馈：紧急在呼吸）
   与骨架呼吸（skPulse）按 loading 状态互斥出现，任何时刻页面循环动画 ≤1 处 */
.h-tag.st-urgent { animation: urgPulse 1.6s linear infinite; }
@keyframes urgPulse { 0%, 100% { opacity: 1; } 50% { opacity: .72; } }

/* ===================== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===================== */
.page.no-motion .hero,
.page.no-motion .info,
.page.no-motion .sec,
.page.no-motion .li,
.page.no-motion .step,
.page.no-motion .bb,
.page.no-motion .sv,
.page.no-motion .h-vl,
.page.no-motion .sd,
.page.no-motion .step .no,
.page.no-motion .step::after,
.page.no-motion .bi.fv .bit { animation: none; } /* 装饰入场/状态弹出全关 */
.page.no-motion .sk-h, .page.no-motion .sk-sec, .page.no-motion .sk-l { animation: none; } /* 骨架呼吸关 */
.page.no-motion .h-tag.st-urgent { animation: none; } /* 紧急呼吸关（语义靠颜色传达，静态可见） */
.page.no-motion .stb:active,
.page.no-motion .bi:active,
.page.no-motion .bo:active,
.page.no-motion .bp:active { transform: none; } /* 按压微缩放关，保留颜色/透明度反馈 */
</style>
