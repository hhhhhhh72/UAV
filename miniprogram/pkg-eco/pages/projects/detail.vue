<template>
  <view class="page" :class="{ 'no-motion': noMotion }" :style="{ paddingTop: (statusBarHeight + 44) + 'px' }">
    <u-nav-bar title="课题详情" show-back :fixed="true" @back="goBack" />

    <!-- 骨架：深色 Hero 区 + 白色规格卡 + 内容行 -->
    <view v-if="loading" class="skw">
      <view class="sk-hero"></view>
      <view class="sk-card"><view class="sk-l w80"></view><view class="sk-l w60"></view><view class="sk-l w90"></view></view>
      <view class="sk-card"><view class="sk-l w40"></view><view class="sk-l w100"></view><view class="sk-l w70"></view></view>
    </view>

    <!-- 错误 -->
    <view v-else-if="err" class="st">
      <u-empty description="加载失败，请检查网络">
        <view class="stb" @tap="fetchData">重新加载</view>
      </u-empty>
    </view>

    <!-- 空 -->
    <view v-else-if="!d" class="st">
      <u-empty description="该课题已下架或不存在">
        <view class="stb" @tap="goBack">返回列表</view>
      </u-empty>
    </view>

    <template v-else>
      <!-- ===== 斜面切角 Hero（机翼式下缘：右浅左深）+ 任务档案装饰 ===== -->
      <view class="hero">
        <!-- 装饰层：细网格 / 轨道环 / 青绿航迹点 / 虚线航迹弧 / 十字准星（仅 Hero 内，低透明度） -->
        <view class="deco">
          <view class="orb orb1"></view>
          <view class="orb orb2"></view>
          <view class="orb-dot"></view>
          <view class="arc"></view>
          <view class="cross"><view class="c-v"></view><view class="c-h"></view></view>
        </view>

        <!-- 标签行 -->
        <view class="h-tags">
          <view class="h-tag field"><view class="h-dot"></view><text>{{ d.f }}</text></view>
          <view class="h-tag" :class="'st-' + d.stCls"><text>{{ d.stLabel }}</text></view>
        </view>

        <text class="h-title">{{ d.t }}</text>

        <view class="h-money">
          <text class="h-vl" :class="{ face: d.budgetText === '面议' }">{{ d.budgetText }}</text>
          <text class="h-lb">攻关经费 · 择优资助</text>
        </view>

        <!-- 数据带 -->
        <view class="h-rule"></view>
        <view class="h-stats">
          <view class="h-si"><text class="h-sv">{{ d.dlTxt }}</text><text class="h-sl">剩余截止</text></view>
          <view class="h-si"><text class="h-sv">{{ d.memberCount }} 家</text><text class="h-sl">参与单位</text></view>
          <view class="h-si"><text class="h-sv">{{ d.dateShort }}</text><text class="h-sl">发布日期</text></view>
        </view>
      </view>

      <!-- 任务参数规格行 -->
      <view class="spec-card">
        <view class="sp-row"><text class="sp-l">牵头单位</text><text class="sp-v">{{ d.leadOrg }}</text></view>
        <view class="sp-row"><text class="sp-l">所属领域</text><view class="sp-v sp-field"><view class="sp-dot"></view><text>{{ d.f }}</text></view></view>
        <view class="sp-row"><text class="sp-l">截止日期</text><text class="sp-v">{{ d.ddl }}</text></view>
      </view>

      <!-- ===== 01 · 课题概述 ===== -->
      <view class="sec" data-idx="0" :class="{ seen: seen.has('0') }">
        <view class="sech">
          <text class="ghost">01</text>
          <view class="sec-hr">
            <text class="sht">课题概述</text>
            <view class="rule"></view>
          </view>
        </view>
        <text class="sec-bd">{{ d.desc || '暂无简介' }}</text>
      </view>

      <!-- ===== 02 · 攻关要求（里程碑交付指标，青绿菱形标记） ===== -->
      <view class="sec" data-idx="1" :class="{ seen: seen.has('1') }">
        <view class="sech">
          <text class="ghost">02</text>
          <view class="sec-hr">
            <text class="sht">攻关要求</text>
            <view class="rule"></view>
          </view>
        </view>
        <view v-if="(d.milestones || []).length" class="rq">
          <view v-for="(m, i) in (d.milestones || [])" :key="i" class="rq-item">
            <view class="rq-dia"></view>
            <text class="rq-t">{{ m }}</text>
          </view>
        </view>
        <text v-else class="sec-bd dim">暂无攻关要求，详情以牵头单位发布为准</text>
      </view>

      <!-- ===== 03 · 攻关路线（航迹时间线：已完成实心蓝 / 当前青绿双环 / 未来虚线灰） ===== -->
      <view class="sec" data-idx="2" :class="{ seen: seen.has('2') }">
        <view class="sech">
          <text class="ghost">03</text>
          <view class="sec-hr">
            <text class="sht">攻关路线</text>
            <view class="rule"></view>
          </view>
        </view>
        <view class="rt">
          <view v-for="(p, i) in (d.timeline || [])" :key="i" class="rt-item">
            <view class="rt-rail">
              <view v-if="p.state === 'cur'" class="rt-node rt-cur">
                <view class="rt-cur-o"></view>
                <view class="rt-cur-i"></view>
              </view>
              <view v-else class="rt-node" :class="p.state === 'past' ? 'rt-done' : 'rt-future'"></view>
              <view v-if="i < (d.timeline || []).length - 1" class="rt-line" :class="p.state === 'past' ? 'rt-line-done' : 'rt-line-future'"></view>
            </view>
            <view class="rt-txt">
              <text class="rt-t" :class="p.state === 'cur' ? 'rt-t-cur' : (p.state === 'past' ? '' : 'rt-t-fut')">{{ p.label }}</text>
              <text class="rt-d" :class="{ 'rt-d-fut': p.state === 'future' }">{{ p.desc }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- ===== 04 · 参与单位（芯片网格：牵头蓝点高亮 + 溢出 +N） ===== -->
      <view v-if="(d.roster || []).length" class="sec" data-idx="3" :class="{ seen: seen.has('3') }">
        <view class="sech">
          <text class="ghost">04</text>
          <view class="sec-hr">
            <text class="sht">参与单位</text>
            <view class="rule"></view>
          </view>
        </view>
        <view class="chips">
          <view v-for="(c, i) in (d.chips || [])" :key="i" class="chip" :class="c.lead ? 'chip-lead' : ''">
            <view v-if="c.lead" class="chip-dot"></view>
            <text>{{ c.name }}</text>
          </view>
          <view v-if="d.moreCount > 0" class="chip chip-more"><text>+{{ d.moreCount }} 家</text></view>
        </view>
      </view>

      <!-- 读完提示 -->
      <view class="end-note"><text>· 共 {{ sectionCount }} 个章节 · 已读完全文 ·</text></view>
      <!-- 底部操作栏占位：栏高 144rpx + 呼吸间距，防固定栏遮挡结尾文案 -->
      <view class="bb-space"></view>

      <!-- 底部操作栏（固定）：双方形图标钮 + 全宽品牌渐变主按钮 -->
      <view class="bb">
        <view class="bi" :class="{ fv: isFav }" aria-role="button" :aria-label="isFav ? '取消收藏' : '收藏'" @tap="toggleFav"><view class="heart"></view></view>
        <button class="bo" open-type="share" hover-class="bo-hover" hover-start-time="0" hover-stay-time="300" aria-label="转发">转发</button>
        <view class="bp" @tap="onJoin">申请参与攻关</view>
      </view>
    </template>
  </view>
</template>

<script setup>
import { ref, reactive, computed, nextTick, getCurrentInstance } from 'vue'
import { onLoad, onReady, onUnload, onPageScroll, onShareAppMessage } from '@dcloudio/uni-app'
import { request } from '@/utils/request'
import { MOCK_PROJECTS } from '@/utils/mockProjects'
import { useReduceMotion } from '@/utils/motion'

const loading = ref(true)
const err = ref(false)
const d = ref(null)
const isFav = ref(false)
const statusBarHeight = ref(20)
const { noMotion, checkMotion } = useReduceMotion() // 减弱动效（无障碍）：装饰动画/位移缩放全关
const inst = getCurrentInstance()
let id = ''

/* 收藏持久化：本地存储兜底（后端收藏接口就绪前的纯前端实现），按课题 id 去重 */
const FAV_KEY = 'project_favs'
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

/* 滚动触达显示：首屏外的区块进入视口才轻浮（设计稿动效规划）
   双保险：IntersectionObserver 提前 32px 点亮（进场丝滑）+
           onPageScroll 按章节实测位置兜底点亮（个别机型观察器不回调时也绝不吞内容） */
const seen = reactive(new Set())
let obs = null
let winH = 667 // 视口高度（兜底值，onLoad 用系统信息覆盖）
let scrollY = 0 // 当前滚动位置（onPageScroll 维护）
let secTops = [] // 各章节顶部距页面顶部的距离（与 data-idx 同序，measureSecs 实测）
const forceSeen = () => { seen.add('0'); seen.add('1'); seen.add('2'); seen.add('3') }
/* 章节是否全部点亮：04 仅在有名册时渲染，无名册只看 0-2（防兜底循环永不短路） */
const allSeen = () => seen.has('0') && seen.has('1') && seen.has('2') &&
  (!(d.value && d.value.roster && d.value.roster.length) || seen.has('3'))
/* 测量章节位置：数据落位后调用；opacity/transform 不影响布局，隐藏态测量同样准确 */
const measureSecs = () => {
  try {
    uni.createSelectorQuery().in(inst?.proxy || inst).selectAll('.sec').boundingClientRect((rects) => {
      secTops = Array.isArray(rects) ? rects.map((r) => r.top + scrollY) : []
    }).exec()
  } catch (e) { secTops = [] }
}
const reObserve = () => {
  if (noMotion.value || !d.value) return
  if (obs) { try { obs.disconnect() } catch (e) { /* 忽略 */ } }
  try {
    obs = uni.createIntersectionObserver(inst?.proxy || inst, { thresholds: [0], initialRatio: 0 })
    obs.relativeToViewport({ bottom: 32 }).observe('.sec', (res) => {
      if (res && res.intersectionRatio > 0) {
        const idx = res.dataset && res.dataset.idx
        if (idx != null) seen.add(idx)
      }
    })
  } catch (e) { forceSeen() }
  measureSecs()
}

const daysLeft = (dt) => {
  if (!dt) return null
  const diff = new Date(dt) - new Date()
  return Number.isFinite(diff) ? Math.max(0, Math.ceil(diff / 86400000)) : null
}
/* 课题阶段（与 projects/list 同套）：规划中(橙) / 招募中(绿) / 进行中(蓝) / 已完成(灰)；
   招募或进行中且截止 ≤7 天 → 紧急(红)。curIdx 对应航迹时间线当前节点 */
const PHASE = {
  planning: { label: '规划中', cls: 'planning', idx: 0 },
  recruiting: { label: '招募中', cls: 'recruiting', idx: 1 },
  progress: { label: '进行中', cls: 'progress', idx: 2 },
  completed: { label: '已完成', cls: 'completed', idx: 3 },
}
const statusOf = (it) => {
  const s = String(it.status || '').toLowerCase()
  const ph = PHASE[s] || PHASE.planning
  const dl = daysLeft(it.end_date)
  const urgent = (s === 'recruiting' || s === 'progress') && dl != null && dl <= 7
  if (urgent) return { label: '紧急', cls: 'urgent', idx: ph.idx }
  return { label: ph.label, cls: ph.cls, idx: ph.idx }
}
const fmtMoney = (wan) => {
  if (wan == null || wan <= 0) return '面议'
  if (wan >= 1) return '¥' + (wan % 1 === 0 ? wan : wan.toFixed(1)) + '万'
  return '¥' + Math.round(wan * 10000)
}
/* 航迹时间线：4 阶段（规划中/招募中/进行中/结题验收）映射课题生命周期，
   当前阶段 = status；描述只使用真实日期（start/end），不伪造预期时间 */
const timelineOf = (it, cur) => {
  const s = (it.start_date || '').slice(0, 7)
  const e = (it.end_date || '').slice(0, 7)
  const defs = [
    { label: '规划中', cur: '立项 · 招募筹备中', past: s ? s + ' 立项启动' : '立项启动', future: '待进入' },
    { label: '招募中', cur: e ? '招募攻关团队 · 截止 ' + e : '招募攻关团队', past: '立项完成', future: '待招募完成' },
    { label: '进行中', cur: '联合攻关执行中', past: '攻关团队已组建', future: '待进入攻关期' },
    { label: '结题验收', cur: e ? '结题验收 · ' + e : '结题验收', past: '攻关执行完成', future: '待进入' },
  ]
  return defs.map((p, i) => ({
    label: p.label,
    state: i < cur ? 'past' : i === cur ? 'cur' : 'future',
    desc: i < cur ? p.past : i === cur ? p.cur : p.future,
  }))
}
// 里程碑/参与单位：仅展示后端真实提供的数据；缺省走模板空态文案，
// 不伪造指标（协会公信力铁律：绝不把平台编造的技术规格当作牵头单位要求展示）
// 模板数组访问全部带 (x || []) 防御：一旦 d 不是 mapItem 产物（畸形响应/旧包缓存），
// 页面降级为空态而非渲染崩溃（曾出现渲染期 length of undefined 白屏）

const mapItem = (it) => {
  const st = statusOf(it)
  const wan = it.budget_fen != null ? it.budget_fen / 100 / 10000 : 0
  const members = Array.isArray(it.members) ? it.members : []
  const lead = it.lead_org || ''
  // 名册 = 牵头单位在首 + 参与单位随行（与牵头同名的成员去重，避免重复展示）
  const roster = []
  if (lead) roster.push({ name: lead, lead: true })
  members.forEach((m) => {
    if (m && m !== lead) roster.push({ name: m, lead: false })
  })
  let ms = []
  if (Array.isArray(it.milestones) && it.milestones.length) ms = it.milestones
  else if (typeof it.milestones === 'string' && it.milestones.trim()) {
    ms = it.milestones.split(/[\n;；]/).map((x) => x.trim()).filter(Boolean)
  }
  const dl = daysLeft(it.end_date)
  const end = (it.end_date || '').slice(0, 10)
  let ddl = '待定'
  // 截止日期畸形（end 在而 dl 算不出）时降级"待定"，绝不渲染 "剩 null 天"
  if (end) ddl = st.cls === 'completed' ? end + ' · 已结题' : end + ' · ' + (dl == null ? '待定' : '剩 ' + dl + ' 天')
  const dlTxt = st.cls === 'completed' ? '已结题' : (dl == null ? '待定' : dl + ' 天')
  return {
    id: it.id,
    t: it.title || '未命名课题',
    f: it.field || '其他',
    desc: it.description || '',
    budgetText: fmtMoney(wan),
    stLabel: st.label,
    stCls: st.cls,
    leadOrg: lead || '待定',
    // 参与单位计数含牵头单位（设计稿语义：Hero "6 家" = 牵头 + 5 成员，与 04 芯片总数一致）
    memberCount: roster.length,
    dateShort: (it.created_at || '').slice(5, 10),
    dlTxt,
    ddl,
    milestones: ms,
    // 已完成：结题验收也是过去时，全部节点按"已完成"实心渲染，不再高亮当前节点
    timeline: timelineOf(it, st.cls === 'completed' ? 4 : st.idx),
    roster, // 模板 v-if 与章节计数都依赖它：缺失会让 04 参与单位整章静默消失
    chips: roster.slice(0, 6),
    moreCount: Math.max(0, roster.length - 6),
  }
}

/* 章节计数：04 参与单位在无名册时不渲染，读完提示按实际章节数显示 */
const sectionCount = computed(() => 3 + (d.value && d.value.roster && d.value.roster.length ? 1 : 0))

const fetchData = async () => {
  loading.value = true
  err.value = false
  try {
    const res = await request({ url: '/api/v1/projects/' + encodeURIComponent(id) })
    const it = (res && res.data) || res
    if (it && it.id) d.value = mapItem(it)
    else {
      // 接口不可用时回退演示数据
      const mock = (MOCK_PROJECTS || []).find((x) => x.id === id)
      d.value = mock ? mapItem(mock) : null
    }
  } catch {
    // 回退演示数据
    const mock = (MOCK_PROJECTS || []).find((x) => x.id === id)
    d.value = mock ? mapItem(mock) : null
    if (!d.value) err.value = true
  } finally {
    loading.value = false
    nextTick(reObserve) // 数据落位后再挂观察器，保证 .sec 节点已渲染
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
  title: d.value ? '课题攻关：' + d.value.t : '低空经济生态服务平台 · 课题攻关',
  path: '/pkg-eco/pages/projects/detail?id=' + encodeURIComponent(id),
}))
const onJoin = () => {
  uni.showToast({ title: '申请已提交，协会将评估后与您对接', icon: 'none' })
}
/* 返回：分享直达/冷启动进入时无页面栈，回退列表兜底（与 list.vue 同款） */
const goBack = () => uni.navigateBack({ fail: () => uni.redirectTo({ url: '/pkg-eco/pages/projects/list' }) })

onPageScroll((e) => {
  scrollY = e?.scrollTop ?? 0
  // 兜底点亮：章节顶越过视口底即标记。观察器正常时它先 32px 触发（滚动达帧更早）；观察器失效时这里是内容可见的唯一保证
  if (noMotion.value || allSeen()) return
  if (secTops.length) {
    secTops.forEach((top, i) => { if (top != null && scrollY + winH > top) seen.add(String(i)) })
  } else if (scrollY > 300) {
    forceSeen() // 测量也失败：已明显滚动仍未点亮 → 全量可见，绝不吞内容
  }
})

onLoad((options) => {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 20
    winH = sys.windowHeight || 667
  } catch (e) { /* 保持默认 */ }
  checkMotion()
  id = options?.id || ''
  isFav.value = favs.has(id)
  fetchData()
})
onReady(reObserve)
onUnload(() => {
  if (obs) { try { obs.disconnect() } catch (e) { /* 忽略 */ } }
})
</script>

<style>
page {
  background: #F5F6F8;
}
</style>
<style scoped>
.page {
  min-height: 100vh;
  background: #F5F6F8;
  padding-bottom: env(safe-area-inset-bottom);
}

/* ===================== 设计体系（任务档案式 Mission Dossier，还原 docs/设计参考/projects-detail） =====================
   概念：像读一份飞行任务档案一样读课题。斜面 Hero（机翼式下缘）+ 编号章节 + 航迹时间线
   色板：深海渐变 #17314A→#0F2A44→#0A66C2（Hero）；青绿 #1DD4A8 强调（经费数字/菱形/当前节点）
   层级：L1 Hero 标题 38rpx/800  L2 区块题 30rpx/700  L3 正文 24rpx  L4 标注 20rpx
   幽灵编号：84rpx/800 #E3E9F1 档案章节号；装饰仅 Hero 内（网格/轨道环/虚线航迹/十字准星） */
/* 注：设计稿标注色 #98A2B3（弱文本）白灰底对比 2.6:1 违 AA，落码统一升 #5D6B82（≈5.4:1），层级不变 */

/* ===== 斜面切角 Hero：clip-path 机翼式下缘（右 82% 浅 / 左 100% 深），低版本基础库降级为直角不破坏内容 ===== */
.hero {
  position: relative;
  padding: 52rpx 32rpx 120rpx;
  background: linear-gradient(141deg, #17314A 0%, #0F2A44 45%, #0A66C2 100%); /* 141deg = 设计稿渐变轴 x2=0.8 / y2=1 */
  clip-path: polygon(0 0, 100% 0, 100% 82%, 0 100%);
  overflow: hidden;
}

/* 装饰层：全部 CSS 绘制，零位图；仅 Hero 内、低透明度（设计语言板：装饰资产仅 2 类） */
.deco { position: absolute; left: 0; top: 0; width: 100%; height: 100%; pointer-events: none; }
/* 细网格：竖线 150rpx / 横线 152rpx */
.deco::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  background-image:
    repeating-linear-gradient(90deg, rgba(255, 255, 255, 0.045) 0, rgba(255, 255, 255, 0.045) 3rpx, transparent 3rpx, transparent 150rpx),
    repeating-linear-gradient(0deg, rgba(255, 255, 255, 0.035) 0, rgba(255, 255, 255, 0.035) 3rpx, transparent 3rpx, transparent 152rpx);
}
/* 轨道环：双层同心圆（白 13% / 7%），右上角，青绿航迹点落在环上 */
.orb {
  position: absolute;
  border-radius: 50%;
  border: 5rpx solid #fff;
  box-sizing: border-box;
}
.orb1 { width: 216rpx; height: 216rpx; left: 520rpx; top: 16rpx; opacity: 0.13; }
.orb2 { width: 372rpx; height: 372rpx; left: 442rpx; top: -62rpx; opacity: 0.07; }
.orb-dot {
  position: absolute;
  width: 11rpx;
  height: 11rpx;
  border-radius: 50%;
  background: #1DD4A8;
  opacity: 0.85;
  left: 721rpx;
  top: 70rpx;
}
/* 虚线航迹弧：左下角大圆弧（dashed border，裁切于 Hero 斜边内） */
.arc {
  position: absolute;
  width: 500rpx;
  height: 500rpx;
  border-radius: 50%;
  border: 5rpx dashed rgba(29, 212, 168, 0.22);
  box-sizing: border-box;
  left: -214rpx;
  top: 344rpx;
}
/* 十字准星 */
.cross { position: absolute; left: 684rpx; top: 366rpx; width: 28rpx; height: 28rpx; }
.c-v { position: absolute; left: 12rpx; top: 0; width: 4rpx; height: 28rpx; background: rgba(255, 255, 255, 0.17); }
.c-h { position: absolute; left: 0; top: 12rpx; width: 28rpx; height: 4rpx; background: rgba(255, 255, 255, 0.17); }

/* 标签行：胶囊（领域 = 白透底 + 蓝点；状态 = 阶段色深底变体） */
.h-tags { display: flex; gap: 14rpx; position: relative; }
.h-tag {
  height: 44rpx;
  padding: 0 22rpx;
  border-radius: 22rpx;
  display: inline-flex;
  align-items: center;
  font-size: 22rpx;
  font-weight: 600;
}
.h-tag.field { background: rgba(255, 255, 255, 0.14); color: #fff; }
.h-dot { width: 14rpx; height: 14rpx; border-radius: 50%; background: #7FB8F0; margin-right: 10rpx; }
/* 深底状态色（浅色系，全部 ≥7:1；与白底阶段色同语义） */
.h-tag.st-planning { background: rgba(255, 177, 102, 0.16); color: #FFB678; }
.h-tag.st-recruiting { background: rgba(29, 212, 168, 0.16); color: #4BE3C0; }
.h-tag.st-progress { background: rgba(159, 201, 242, 0.16); color: #9FC9F2; }
.h-tag.st-completed { background: rgba(192, 201, 214, 0.16); color: #C6CEDB; }
.h-tag.st-urgent { background: rgba(255, 138, 128, 0.18); color: #FF8A80; }

.h-title {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-top: 58rpx;
  position: relative;
  font-size: 38rpx;
  font-weight: 800;
  line-height: 52rpx;
  letter-spacing: 0.01em;
  color: #fff;
}
/* 经费：青绿强调数字（Hero 心脏），标签弱白 */
.h-money { margin-top: 56rpx; display: flex; flex-direction: column; gap: 10rpx; position: relative; }
.h-vl {
  font-size: 58rpx;
  font-weight: 800;
  color: #1DD4A8;
  line-height: 1.1;
  font-feature-settings: "tnum" 1;
}
/* 面议非数字：不用金额的重量渲染（与 list.vue 同约定），避免误读为最高经费 */
.h-vl.face { font-size: 40rpx; font-weight: 700; }
.h-lb { font-size: 22rpx; color: rgba(255, 255, 255, 0.85); }

/* 数据带：发丝线 + 三栏（剩余截止 / 参与单位 / 发布日期） */
.h-rule { height: 4rpx; background: rgba(255, 255, 255, 0.16); margin: 40rpx 0 24rpx; position: relative; }
.h-stats { display: flex; position: relative; }
.h-si { flex: 1; display: flex; flex-direction: column; gap: 6rpx; }
.h-sv { font-size: 36rpx; font-weight: 700; color: #fff; font-feature-settings: "tnum" 1; }
.h-sl { font-size: 20rpx; color: rgba(255, 255, 255, 0.8); } /* 设计稿 72% → 80%：最亮角处仍 ≥4.5:1 */

/* ===== 任务参数规格行：白卡浮在斜面下，左标签右值 + 发丝行 ===== */
.spec-card {
  margin: 32rpx 24rpx 0;
  background: #fff;
  border-radius: 32rpx;
  padding: 26rpx 28rpx 18rpx;
  box-shadow: 0 16rpx 36rpx rgba(15, 27, 42, 0.06);
}
.sp-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  padding: 22rpx 0;
  border-top: 2rpx solid #F0F2F5;
}
.sp-row:first-child { border-top: none; padding-top: 10rpx; }
.sp-l { font-size: 22rpx; color: #667085; flex: none; }
.sp-v { font-size: 24rpx; font-weight: 600; color: #17212B; text-align: right; min-width: 0; font-feature-settings: "tnum" 1; }
.sp-field { display: inline-flex; align-items: center; gap: 10rpx; }
.sp-dot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #0d47a1; flex: none; }

/* ===== 编号章节（01-04）：幽灵大数字 + 区块题 + 细规则线 ===== */
.sec { margin: 88rpx 24rpx 0; }
.sech { display: flex; align-items: flex-end; }
.ghost {
  font-size: 84rpx;
  font-weight: 800;
  color: #E3E9F1;
  line-height: 76rpx;
  margin-right: 16rpx;
  margin-bottom: 4rpx;
  flex: none;
  font-feature-settings: "tnum" 1;
}
.sec-hr { flex: 1; min-width: 0; padding-bottom: 6rpx; }
.sht { display: block; font-size: 30rpx; font-weight: 700; color: #17212B; line-height: 42rpx; }
.rule { height: 2rpx; background: #EDF0F4; margin-top: 12rpx; }
.sec-bd {
  display: block;
  margin: 32rpx 0 0 108rpx;
  font-size: 24rpx;
  color: #667085;
  line-height: 40rpx;
}
.sec-bd.dim { color: #5D6B82; }

/* ===== 02 攻关要求：青绿菱形标记（rotate 45° 方块） ===== */
.rq { margin: 32rpx 0 0 108rpx; }
.rq-item { display: flex; align-items: flex-start; gap: 32rpx; padding-bottom: 36rpx; }
.rq-item:last-child { padding-bottom: 0; }
.rq-dia {
  width: 22rpx;
  height: 22rpx;
  background: #1DD4A8;
  transform: rotate(45deg);
  margin-top: 10rpx;
  flex: none;
}
.rq-t { flex: 1; font-size: 22rpx; color: #344054; line-height: 42rpx; }

/* ===== 03 攻关路线：航迹时间线（节点 34rpx 导轨列，线段在节点间生长） ===== */
.rt { margin: 32rpx 0 0 98rpx; }
.rt-item { display: flex; }
.rt-rail { width: 34rpx; display: flex; flex-direction: column; align-items: center; flex: none; }
/* 已完成节点：实心蓝；未来节点：白心灰环；当前节点：青绿双环（唯一强调色） */
.rt-node { width: 22rpx; height: 22rpx; border-radius: 50%; box-sizing: border-box; flex: none; position: relative; }
.rt-done { background: #0A66C2; }
.rt-future { background: #fff; border: 8rpx solid #D9DEE3; }
.rt-cur { width: 34rpx; height: 34rpx; display: flex; align-items: center; justify-content: center; }
.rt-cur-o {
  position: absolute;
  left: 0;
  top: 0;
  width: 34rpx;
  height: 34rpx;
  border-radius: 50%;
  border: 6rpx solid rgba(29, 212, 168, 0.35);
  box-sizing: border-box;
}
.rt-cur-i { width: 20rpx; height: 20rpx; border-radius: 50%; border: 9rpx solid #1DD4A8; box-sizing: border-box; }
/* 线段：已完成实心蓝 / 未来虚线灰（repeating-gradient 10rpx 段 20rpx 空） */
.rt-line { width: 8rpx; flex: 1; min-height: 88rpx; margin: -2rpx 0; }
.rt-line-done { background: #0A66C2; }
.rt-line-future {
  background-image: repeating-linear-gradient(180deg, #D9DEE3 0, #D9DEE3 20rpx, transparent 20rpx, transparent 40rpx);
}
.rt-txt { flex: 1; min-width: 0; padding-left: 26rpx; padding-bottom: 56rpx; }
.rt-item:last-child .rt-txt { padding-bottom: 0; }
.rt-t { display: block; font-size: 26rpx; font-weight: 700; color: #17212B; line-height: 34rpx; }
.rt-t-cur { color: #0B7A5F; } /* 设计稿 #0F9D7C 3.3:1 违 AA，升深 #0B7A5F ≈5.1:1，青绿语义不变 */
.rt-t-fut { color: #667085; font-weight: 600; }
.rt-d { display: block; margin-top: 8rpx; font-size: 20rpx; color: #667085; line-height: 28rpx; }
.rt-d-fut { color: #5D6B82; }

/* ===== 04 参与单位：芯片网格（牵头 = 蓝点 + 蓝字高亮） ===== */
.chips { margin: 32rpx 0 0 108rpx; display: flex; flex-wrap: wrap; gap: 16rpx; }
.chip {
  height: 58rpx;
  padding: 0 22rpx;
  border-radius: 12rpx;
  background: #F4F6F8;
  display: inline-flex;
  align-items: center;
  font-size: 22rpx;
  color: #344054;
  max-width: 100%;
}
.chip text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.chip-lead { background: #EAF3FB; color: #0A66C2; font-weight: 600; }
.chip-dot { width: 12rpx; height: 12rpx; border-radius: 50%; background: #0A66C2; margin-right: 10rpx; flex: none; }
.chip-more { background: transparent; border: 2rpx dashed #D9DEE3; color: #667085; }

/* 读完提示（设计稿 em-dash 分隔 → 中点分隔，符合可见文案零 em-dash 规范） */
.end-note { margin: 88rpx 0 0; text-align: center; }
.end-note text { font-size: 20rpx; color: #5D6B82; letter-spacing: 0.04em; }
/* 底部操作栏占位：栏高 152rpx（32+88+32），占位 160rpx 留 8rpx 呼吸；安全区由 page padding-bottom 承担 */
.bb-space { height: 160rpx; }

/* ===== 底部操作栏（固定）：双方形图标钮 88rpx + 转发文字钮 + 全宽渐变主按钮（rx40 胶囊） ===== */
.bb {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 50;
  background: #fff;
  border-top: 2rpx solid #F0F0F0;
  box-shadow: 0 -6rpx 28rpx rgba(15, 27, 42, 0.07);
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 32rpx;
  padding-bottom: calc(32rpx + env(safe-area-inset-bottom));
}
.bi {
  width: 88rpx; /* 80→88rpx：同研发难题 .bi */
  height: 88rpx;
  border-radius: 16rpx;
  background: #F4F6F8;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.bi.fv { background: #FDECEC; }
/* 收藏心形：设计稿描边心形路径（stroke 4.5 / 圆角连接）以 SVG data-URI 落为背景——
   绘制图标而非 emoji（项目规范）；已收藏整颗实心红 + 浅红底，语义双通道传达；
   描边 #667085 灰底 4.9:1，高于非文本控件 3:1 */
.heart {
  width: 40rpx; /* 心形 20px：同研发难题 .bit（42×28 → 40×40，contain 保持原比例） */
  height: 40rpx;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 62'%3E%3Cpath d='M32,26C21,15,5,29,14,40C20,47,27,51,32,51C37,51,44,47,50,40C59,29,43,15,32,26Z' fill='none' stroke='%23667085' stroke-width='4.5' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: center;
  background-size: contain;
}
.bi.fv .heart {
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 62'%3E%3Cpath d='M32,26C21,15,5,29,14,40C20,47,27,51,32,51C37,51,44,47,50,40C59,29,43,15,32,26Z' fill='%23FF3B30' stroke='%23FF3B30' stroke-width='4.5' stroke-linejoin='round'/%3E%3C/svg%3E");
}
/* 「转发」文字按钮（同研发难题 .bo）：白底蓝描边，44px 高 13px/600 */
.bo {
  height: 88rpx;
  border-radius: 16rpx;
  border: 2rpx solid #0A66C2;
  background: #fff;
  color: #0A66C2;
  font-size: 26rpx;
  font-weight: 600;
  padding: 0 32rpx;
  margin: 0;
  line-height: 1;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.bo::after { border: none; }
.bo-hover { transform: scale(.95); background: #F4F8FC; }
.bp {
  flex: 1;
  height: 88rpx; /* 80→88rpx：同研发难题 .bp */
  border-radius: 40rpx;
  background: linear-gradient(90deg, #0A66C2 0%, #074D92 100%);
  color: #fff;
  font-size: 28rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 16rpx 28rpx rgba(10, 102, 194, 0.35);
}

/* ===== 骨架（深色 Hero 区 + 白色规格卡 + 内容行） ===== */
.skw { padding-top: 10rpx; }
.sk-hero {
  height: 604rpx;
  margin: 0;
  background: #E2E6EC;
  clip-path: polygon(0 0, 100% 0, 100% 82%, 0 100%); /* 骨架同构斜面 Hero：加载态即预告版式 */
}
.sk-card { margin: 32rpx 24rpx 0; padding: 32rpx; background: #fff; border-radius: 32rpx; }
.sk-l { height: 24rpx; background: #EDF0F3; border-radius: 6rpx; margin-bottom: 24rpx; }
.sk-l:last-child { margin-bottom: 0; }
.sk-l.w80 { width: 80%; }
.sk-l.w60 { width: 60%; }
.sk-l.w90 { width: 90%; }
.sk-l.w40 { width: 40%; }
.sk-l.w100 { width: 100%; }
.sk-l.w70 { width: 70%; }

/* ===== 状态 ===== */
.st { display: flex; flex-direction: column; align-items: center; padding: 120rpx 40rpx; }
.stb { padding: 16rpx 48rpx; border-radius: 999rpx; background: #0A66C2; color: #fff; font-size: 26rpx; font-weight: 500; margin-top: 32rpx; }

/* ===================== 动效规范（对齐全局动画规范，项目铁律优先） =====================
   白名单：仅 transform / opacity（小尺寸元素 color/background 过渡允许，仅重绘不重排）
   禁参与动画：top/left/width/height/margin（触发重排）、box-shadow/filter（低端安卓掉帧）
   时长：微反馈 150-200ms（按压按下 .08s 即时到位）/ 松手弹簧回位 .3s（ios-pop）/ 页面级 ≤400ms
   曲线：两枚固定曲线：ios-pop cubic-bezier(.34,1.8,.64,1) 松手弹簧回弹（仅按压/弹出类 transform）+
        ios-decel cubic-bezier(.32,.72,0,1) 浮层流体减速（滚动浮现/操作栏落地）；
        其余进场 ease-out / 退场 ease-in / 循环 linear；除这两枚外禁手写 cubic-bezier
   数量：入场 = 设计稿动效规划（Hero 整块淡入 → 轨道环 scale 收缩落位 → 经费数字 ios-pop 弹出），
        ≤400ms 一次收完；编号章节滚动触达依次 fade-up；航迹线自上而下 scaleY 生长；当前节点双环单次脉冲；
        循环动画全页 ≤1 处
   no-motion：系统减弱动效时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 */

/* 1) Hero 进场（设计稿动效规划三件套） */
.hero { animation: fadeUp .3s ease-out backwards; }
@keyframes fadeUp { from { opacity: 0; transform: translateY(-10px); } to { opacity: 1; transform: translateY(0); } }
/* 轨道环：scale 收缩落位（1.18→1，ios-decel 40ms 起） */
.orb { animation: ringIn .34s cubic-bezier(.32, .72, 0, 1) 40ms backwards; }
@keyframes ringIn { from { opacity: 0; transform: scale(1.18); } to { opacity: 1; transform: scale(1); } }
/* 经费数字：Hero 心脏 ios-pop 弹簧弹出（80ms 起播 .25s，总 330ms ≤ 400ms） */
.h-vl { animation: popIn .25s cubic-bezier(.34, 1.8, .64, 1) 80ms backwards; }
@keyframes popIn { from { transform: scale(.6); opacity: 0; } to { transform: scale(1); opacity: 1; } }
/* 规格卡：60ms 起播 .25s */
.spec-card { animation: fadeUp .25s ease-out 60ms backwards; }

/* 2) 编号章节滚动触达浮现：默认隐于 10px 之下，进入视口才轻浮 */
.sec:not(.seen) { opacity: 0; transform: translateY(10px); }
.sec.seen { animation: secReveal .3s cubic-bezier(.32, .72, 0, 1); }
@keyframes secReveal { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
.sec.seen .ghost { animation: fadeIn .3s ease-out 40ms backwards; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

/* 02 菱形条目：依次淡入（前 6 条 40ms 错峰） */
.sec.seen .rq-item { animation: rowIn .2s ease-out backwards; }
.sec.seen .rq-item:nth-child(2) { animation-delay: 40ms; }
.sec.seen .rq-item:nth-child(3) { animation-delay: 80ms; }
.sec.seen .rq-item:nth-child(4) { animation-delay: 120ms; }
.sec.seen .rq-item:nth-child(5) { animation-delay: 160ms; }
.sec.seen .rq-item:nth-child(6) { animation-delay: 200ms; }
@keyframes rowIn { from { opacity: 0; transform: translateX(-8px); } to { opacity: 1; transform: translateX(0); } }

/* 03 航迹线：自上而下 scaleY 生长（transform-origin top，零重排）；线段与节点同错峰，
   生长级联而下（节点 i 亮起后其下线段才开始生长） */
.sec.seen .rt-line { animation: lineGrowY .3s ease-out backwards; transform-origin: top center; }
.sec.seen .rt-item:nth-child(2) .rt-line { animation-delay: 80ms; }
.sec.seen .rt-item:nth-child(3) .rt-line { animation-delay: 160ms; }
.sec.seen .rt-item:nth-child(4) .rt-line { animation-delay: 240ms; }
@keyframes lineGrowY { from { transform: scaleY(0); } to { transform: scaleY(1); } }
.sec.seen .rt-node { animation: nodeIn .26s cubic-bezier(.34, 1.8, .64, 1) backwards; }
.sec.seen .rt-item:nth-child(2) .rt-node { animation-delay: 80ms; }
.sec.seen .rt-item:nth-child(3) .rt-node { animation-delay: 160ms; }
.sec.seen .rt-item:nth-child(4) .rt-node { animation-delay: 240ms; }
@keyframes nodeIn { from { transform: scale(.5); opacity: 0; } to { transform: scale(1); opacity: 1; } }
/* 当前节点双环：单次脉冲（设计稿规定，非循环） */
.sec.seen .rt-cur-o { animation: ringPulse .5s ease-out 300ms backwards; }
@keyframes ringPulse {
  0% { transform: scale(1); opacity: 1; }
  60% { transform: scale(1.2); opacity: 0.75; }
  100% { transform: scale(1); opacity: 1; }
}

/* 04 单位芯片：依次淡入 */
.sec.seen .chip { animation: rowIn .2s ease-out backwards; }
.sec.seen .chip:nth-child(2) { animation-delay: 40ms; }
.sec.seen .chip:nth-child(3) { animation-delay: 80ms; }
.sec.seen .chip:nth-child(4) { animation-delay: 120ms; }
.sec.seen .chip:nth-child(5) { animation-delay: 160ms; }
.sec.seen .chip:nth-child(6) { animation-delay: 200ms; }
.sec.seen .chip:nth-child(7) { animation-delay: 240ms; }

/* 底部操作栏：自下而上滑入（fixed 定位，transform 不影响布局；ios-decel 落地感），100ms 起收于 400ms */
.bb { animation: bbUp .3s cubic-bezier(.32, .72, 0, 1) 100ms backwards; }
@keyframes bbUp { from { opacity: 0; transform: translateY(24px); } to { opacity: 1; transform: translateY(0); } }

/* 骨架呼吸（加载中环境光；循环 1.2-1.6s linear） */
.sk-l, .sk-hero { animation: skPulse 1.4s linear infinite; }
@keyframes skPulse { 0%, 100% { opacity: 1; } 50% { opacity: .55; } }

/* 3) 交互反馈：按压按下 .08s linear 即时到位，松手 .3s ios-pop 弹簧回位 */
.stb { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; }
.stb:active { transform: scale(.94); opacity: .85; transition: transform .08s linear; }
.bi { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, color .2s ease; }
.bi:active { transform: scale(.9); background: #EAF3FB; transition: transform .08s linear; }
.bi.fv:active { background: #FDECEC; }
.bo { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), background .2s ease, color .2s ease; } /* ios-pop（同研发难题 .bo） */
.bo:active { transform: scale(.95); background: #F4F8FC; transition: transform .08s linear; }
.bp { transition: transform .3s cubic-bezier(.34, 1.8, .64, 1), opacity .15s ease; }
.bp:active { transform: scale(.95); opacity: .92; transition: transform .08s linear; }

/* 4) 状态过渡：收藏点亮时心形 ios-pop 弹簧弹出（scale .8→1 自然过冲回位）；取消收藏无反向动画 */
.bi.fv .heart { animation: heartPop .3s cubic-bezier(.34, 1.8, .64, 1); }
@keyframes heartPop { from { transform: scale(.8); } to { transform: scale(1); } }

/* 5) "紧急"状态标签呼吸：全页唯一循环动画（语义反馈：紧急在呼吸）
   与骨架呼吸按 loading 状态互斥出现，任何时刻页面循环动画 ≤1 处 */
.h-tag.st-urgent { animation: urgPulse 1.6s linear infinite; }
@keyframes urgPulse { 0%, 100% { opacity: 1; } 50% { opacity: .62; } }

/* ===================== 减弱动效适配（无障碍）：no-motion 时装饰动画全关、位移/缩放禁用，保留淡入与颜色反馈 ===================== */
.page.no-motion .hero,
.page.no-motion .orb,
.page.no-motion .h-vl,
.page.no-motion .spec-card,
.page.no-motion .bb { animation: none; }
.page.no-motion .sec:not(.seen) { opacity: 1; transform: none; } /* 滚动浮现关闭：全部直接可见 */
.page.no-motion .sec.seen { animation: none; }
.page.no-motion .sec.seen .ghost,
.page.no-motion .sec.seen .rq-item,
.page.no-motion .sec.seen .rt-line,
.page.no-motion .sec.seen .rt-node,
.page.no-motion .sec.seen .rt-cur-o,
.page.no-motion .sec.seen .chip { animation: none; } /* 区块内部微编排全关 */
.page.no-motion .sk-l, .page.no-motion .sk-hero { animation: none; } /* 骨架呼吸关 */
.page.no-motion .h-tag.st-urgent { animation: none; } /* 紧急呼吸关（语义靠颜色传达，静态可见） */
.page.no-motion .bi.fv .heart { animation: none; } /* 收藏心形弹出关 */
.page.no-motion .stb:active,
.page.no-motion .bi:active,
.page.no-motion .bo:active,
.page.no-motion .bo-hover,
.page.no-motion .bp:active { transform: none; } /* 按压微缩放关，保留颜色/透明度反馈 */
</style>
