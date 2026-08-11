<template>
  <Layout :current="4">
    <view class="mine-page">
      <!-- ═══════ 顶部身份区 ═══════ -->
      <MineHeader
        :status-bar-h="statusBarH"
        :capsule-gap="capsuleGap"
        :unread-count="unreadCount"
        :vm="headerVm"
        @messages="goMessages"
        @settings="goSettings"
        @profile="handleUserClick"
        @cert="goPrimaryCert"
      />

      <!-- ═══════ 身份概览卡（登录后） ═══════ -->
      <view class="page-section" v-if="user">
        <view class="section-title">
          <text class="section-title-text">{{ overviewTitle }}</text>
          <text class="section-link" @tap="overviewLinkGo">{{ overviewLink }} ›</text>
        </view>

        <!-- 加载中：三格骨架 -->
        <view v-if="overviewLoading" class="card overview-card">
          <view class="ov-cell" v-for="i in 3" :key="i">
            <view class="ov-skel"></view>
            <view class="ov-skel-label"></view>
          </view>
        </view>

        <!-- 数据卡 -->
        <view v-else class="card overview-card">
          <view
            v-for="(c, i) in overviewCells"
            :key="i"
            class="ov-cell"
            hover-class="tap-fade"
            @tap="c.go"
          >
            <text class="ov-value">{{ c.value }}</text>
            <text class="ov-label">{{ c.label }}</text>
          </view>
          <view class="ov-note" v-if="overviewNote">
            <text class="ov-note-lead">{{ overviewNote.lead }}</text>
            <text class="ov-note-rest">{{ overviewNote.rest }}</text>
          </view>
        </view>
      </view>

      <!-- ═══════ 设备与飞行（仅企业 / 认证飞手） ═══════ -->
      <view class="page-section" v-if="showDeviceCard">
        <view class="section-title">
          <text class="section-title-text">设备与飞行</text>
          <text class="section-link" @tap="goDeviceManage">全部设备 ›</text>
        </view>
        <view class="card device-card">
          <!-- 左侧内容区 -->
          <view class="device-main" hover-class="tap-fade" @tap="goDeviceManage">
            <view class="device-icon">
              <image class="device-icon-img" :src="icons.drone" mode="aspectFit" />
            </view>
            <view class="device-copy">
              <template v-if="device">
                <text class="device-title">已绑定 {{ device.bound }} 台设备</text>
                <text class="device-sub">
                  <text class="device-online">{{ device.online }} 台在线</text>
                  <text class="device-sub-rest"> · 本月飞行 {{ device.flights }} 架次</text>
                </text>
              </template>
              <template v-else>
                <text class="device-title">暂未绑定设备</text>
                <text class="device-sub">绑定设备后可查看在线状态与飞行记录</text>
              </template>
            </view>
          </view>
          <!-- 右侧独立操作 -->
          <view class="device-more" hover-class="tap-fade" @tap="goDeviceManage">
            <text>管理</text>
            <text class="device-more-arrow">›</text>
          </view>
        </view>
      </view>

      <!-- ═══════ 我的业务（固定 3×2 顺序） ═══════ -->
      <view class="page-section">
        <view class="section-title">
          <text class="section-title-text">我的业务</text>
        </view>
        <view class="card">
          <MineQuickGrid :items="businessItems" @select="onBusinessSelect" />
        </view>
      </view>

      <!-- ═══════ 账号与服务 ═══════ -->
      <view class="page-section">
        <view class="section-title">
          <text class="section-title-text">账号与服务</text>
        </view>
        <view class="card">
          <MineCellGroup :items="accountItems" @select="onAccountSelect" />
        </view>
      </view>

      <!-- ═══════ 退出登录 ═══════ -->
      <view class="logout-card" v-if="user" @tap="doLogout">
        <text class="logout-text">退出登录</text>
      </view>

      <view class="bottom-spacer"></view>
    </view>
  </Layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import Layout from '@/components/Layout.vue'
import MineHeader from '@/components/mine/MineHeader.vue'
import MineQuickGrid from '@/components/mine/MineQuickGrid.vue'
import MineCellGroup from '@/components/mine/MineCellGroup.vue'
import { getStoredUser, request, authStorage, BASE_URL } from '../../utils/request'
import { getMineFixture } from '../../utils/mineFixture'

const icons = {
  drone: '/static/mine-icons/drone.svg',
}

const user = ref(null)
const unreadCount = ref(0)
const statusBarH = ref(20)
const capsuleGap = ref(0)

// 认证状态：由真实接口得出，不由前端任意切换
const pilotStatus = ref('')      // '' 未申请 / pending / approved / rejected
const pilotId = ref('')          // 飞手认证记录 ID（已认证跳档案用）
const enterpriseStatus = ref('') // '' 无企业 / draft / submitted / supplement_required / approved / rejected
const authStatus = ref('approved') // 实名认证：演示写死已认证（不接外部核验，无真实接口）

// 概览统计（真实可用则取，否则 0/暂无数据；仅开发 fixture 提供样例）
const overviewLoading = ref(false)
const overviewCounts = ref({})   // { publish, talk, certText, flights, certs }
const device = ref(null)         // null → 空态；{ bound, online, flights }

const roleLabels = {
  individual: '个人用户',
  enterprise: '企业用户',
  association_admin: '协会管理',
  platform_admin: '平台管理',
}

// ── 身份解析 ──
const identity = computed(() => {
  if (!user.value) return 'guest'
  const role = user.value.role
  if (role === 'association_admin' || role === 'platform_admin') return 'admin'
  if (role === 'enterprise') return 'enterprise'
  // individual 只有在飞手认证通过后才是「认证飞手」
  if (pilotStatus.value === 'approved') return 'pilot'
  return 'individual'
})

const enterpriseStatusText = computed(() => {
  const map = { draft: '草稿', submitted: '审核中', supplement_required: '需补充资料', approved: '企业认证已通过', rejected: '未通过' }
  return map[enterpriseStatus.value] || (enterpriseStatus.value ? enterpriseStatus.value : '去认证')
})

const pilotStatusText = computed(() => {
  const map = { pending: '审核中', approved: '飞手认证已通过', rejected: '未通过' }
  return map[pilotStatus.value] || (pilotStatus.value ? pilotStatus.value : '去申请')
})

// 实名认证短标签（概览格 / 认证信息 tail）
const authText = computed(() => {
  const map = { pending: '审核中', approved: '已认证', rejected: '被驳回' }
  return map[authStatus.value] || '未认证'
})

const isEnterpriseApproved = computed(() => identity.value === 'enterprise' && enterpriseStatus.value === 'approved')
const isPilotApproved = computed(() => identity.value === 'pilot' && pilotStatus.value === 'approved')

// ── MineHeader view model ──
const headerVm = computed(() => {
  const g = {
    name: '点击登录',
    initial: '?',
    avatar: '',
    badge: '',
    badgeClass: '',
    note: '登录后查看需求、对接意向与商城订单',
    showCertBar: false,
    certIcon: '',
    certMain: '',
    certState: '',
    certStateClass: '',
  }
  if (!user.value) return g

  const u = user.value
  g.name = u.name || u.phone || '微信用户'
  g.initial = (u.name || u.phone || '微').charAt(0).toUpperCase()
  // 头像字段兼容：本地缓存存 avatar（profile.vue 保存时写入），后端登录/me 返回 avatar_url
  g.avatar = avatarSrc(u.avatar || u.avatar_url)

  if (identity.value === 'enterprise') {
    g.badge = '企业账号'
    g.badgeClass = isEnterpriseApproved.value ? 'ok' : 'wait'
    g.note = `企业账号 · ${enterpriseStatusText.value}`
    g.showCertBar = true
    g.certIcon = '/static/mine-icons/enterprise.svg'
    g.certMain = '已加入 重庆市无人机产业协会'
    g.certState = isEnterpriseApproved.value ? '会员单位' : enterpriseStatusText.value
    g.certStateClass = isEnterpriseApproved.value ? 'ok' : 'wait'
  } else if (identity.value === 'pilot') {
    g.badge = '认证飞手'
    g.badgeClass = 'ok'
    g.note = '个人飞手 · 飞手认证已通过'
    g.showCertBar = true
    g.certIcon = '/static/mine-icons/drone.svg'
    g.certMain = 'CAAC 飞手认证'
    g.certState = '已认证'
    g.certStateClass = 'ok'
  } else if (identity.value === 'individual') {
    // 实名认证为演示写死状态：一律显示已认证
    g.badge = '个人用户'
    g.badgeClass = 'plain'
    g.showCertBar = true
    g.certIcon = '/static/mine-icons/certification.svg'
    g.note = '已实名认证 · 可申请升级飞手'
    g.certMain = '实名认证已通过'
    g.certState = '已认证'
    g.certStateClass = 'ok'
  } else {
    g.badge = roleLabels[u.role] || '平台账号'
    g.badgeClass = 'plain'
    g.note = u.name || '平台账号'
    g.showCertBar = true
    g.certIcon = '/static/mine-icons/certification.svg'
    g.certMain = '平台管理账号'
    g.certState = '—'
    g.certStateClass = 'plain'
  }
  return g
})

// ── 概览 ──
const overviewTitle = computed(() => {
  if (identity.value === 'enterprise') return '企业服务概览'
  if (identity.value === 'pilot') return '飞手服务概览'
  if (identity.value === 'individual') return '我的参与'
  return '服务概览'
})

const overviewLink = computed(() => {
  if (identity.value === 'enterprise') return '企业档案'
  if (identity.value === 'pilot') return '认证资料'
  return '账号资料'
})

const overviewLinkGo = () => {
  if (!requireLogin()) return
  if (identity.value === 'enterprise') return goEnterpriseCert()
  if (identity.value === 'pilot') return goPilotCert()
  return goProfile()
}

const overviewCells = computed(() => {
  const c = overviewCounts.value
  if (identity.value === 'enterprise') {
    return [
      { value: c.publish || '0', label: '我的发布', go: goMyDemands },
      { value: c.talk || '0', label: '洽谈会话', go: goIntents },
      { value: c.certText || enterpriseStatusText.value, label: '企业认证', go: goEnterpriseCert },
    ]
  }
  if (identity.value === 'pilot') {
    return [
      { value: c.certText || '已通过', label: '飞手认证', go: goPilotCert },
      { value: c.flights || '0', label: '飞行记录', go: goComingSoon },
      { value: c.certs || '0', label: '培训证书', go: goCertificates },
    ]
  }
  if (identity.value === 'individual') {
    return [
      { value: c.authText || authText.value, label: '实名认证', go: goAuth },
      { value: '0', label: '培训报名', go: goCourses },
      { value: '0', label: '商城订单', go: goOrders },
    ]
  }
  return []
})

const overviewNote = computed(() => {
  if (identity.value === 'enterprise') {
    return { lead: '会员权益已生效', rest: ' · 可发布供需与管理企业资源' }
  }
  if (identity.value === 'pilot') {
    return { lead: '飞手档案已完善', rest: ' · 可承接更多匹配任务' }
  }
  if (identity.value === 'individual') {
    // 实名认证为演示写死状态：一律显示已认证
    return { lead: '实名认证已通过', rest: '· 可申请飞手认证或企业入驻' }
  }
  return null
})

// ── 设备与飞行 ──
const showDeviceCard = computed(() => identity.value === 'enterprise' || identity.value === 'pilot')

// ── 我的业务（固定顺序不可改） ──
const businessItems = computed(() => [
  { icon: '/static/mine-icons/publish.svg', tone: 'publish', label: '我的发布', go: goMyDemands },
  { icon: '/static/mine-icons/intent.svg', tone: 'intent', label: '合作意向', go: goIntents },
  { icon: '/static/mine-icons/appointment.svg', tone: 'appointment', label: '我的预约', go: goComingSoon },
  { icon: '/static/mine-icons/enroll.svg', tone: 'enroll', label: '我的报名', go: goComingSoon },
  { icon: '/static/mine-icons/favorite.svg', tone: 'favorite', label: '我的收藏', go: goComingSoon },
  { icon: '/static/mine-icons/order.svg', tone: 'order', label: '商城订单', go: goOrders },
])

const onBusinessSelect = (i) => {
  const item = businessItems.value[i]
  if (item && item.go) item.go()
}

// ── 账号与服务（认证信息为聚合入口） ──
const certMenuText = computed(() => {
  if (identity.value === 'enterprise') return { desc: '企业认证与资质状态', tail: enterpriseStatusText.value, tailClass: certTailClass(enterpriseStatus.value, 'approved') }
  if (identity.value === 'pilot') return { desc: '飞手认证与执照状态', tail: pilotStatusText.value, tailClass: certTailClass(pilotStatus.value, 'approved') }
  if (identity.value === 'individual') {
    // 飞手认证申请中 / 被驳回：优先展示飞手状态（入口不能因状态消失；sync=请求失败占位不算）
    if (pilotStatus.value && pilotStatus.value !== 'sync') {
      return { desc: '飞手认证与执照状态', tail: pilotStatusText.value, tailClass: certTailClass(pilotStatus.value, 'approved') }
    }
    return { desc: '个人实名信息', tail: authText.value, tailClass: certTailClass(authStatus.value, 'approved') }
  }
  return { desc: '登录后查看认证状态', tail: '', tailClass: '' }
})

const accountItems = computed(() => [
  { icon: '/static/mine-icons/account.svg', tone: 'primary', label: '账号信息', desc: '手机号、微信绑定与个人资料', go: goProfile },
  { icon: '/static/mine-icons/certification.svg', tone: 'primary', label: '认证信息', desc: certMenuText.value.desc, tail: certMenuText.value.tail, tailClass: certMenuText.value.tailClass, go: goPrimaryCert },
  { icon: '/static/mine-icons/settings-gray.svg', tone: 'gray', label: '设置', desc: '消息通知、隐私与安全、关于平台', go: goSettings },
])

const onAccountSelect = (i) => {
  const item = accountItems.value[i]
  if (item && item.go) item.go()
}

const certTailClass = (status, okStatus) => {
  if (status === okStatus) return 'ok'
  if (status === 'rejected') return 'danger'
  if (status === 'pending' || status === 'submitted' || status === 'supplement_required') return 'wait'
  return ''
}

// ── 数据加载 ──
const fetchData = async () => {
  const fx = getMineFixture()

  if (fx.active) {
    // 开发态视觉 fixture：仅展示用，不写 storage、无权限变化
    user.value = fx.userOverride || getStoredUser()
    enterpriseStatus.value = fx.data.enterpriseStatus || ''
    pilotStatus.value = fx.data.pilotStatus || ''
    overviewCounts.value = fx.data.overview || {}
    device.value = fx.data.device ? { ...fx.data.device } : null
    overviewLoading.value = false
    // 未读消息仍走真实接口（不强求）
    try {
      const msgRes = await request({ url: '/api/v1/messages/unread-count' })
      unreadCount.value = msgRes?.data?.count || msgRes?.count || 0
    } catch (e) { unreadCount.value = 0 }
    return
  }

  const currentUser = getStoredUser()
  user.value = currentUser

  if (currentUser) {
    // 刷新服务端用户信息（只覆盖非空字段，避免 me 的空 name 覆盖本地已存的昵称）
    try {
      const meRes = await request({ url: '/api/auth/me' })
      if (meRes?.user) {
        const fresh = meRes.user || {}
        const merged = { ...currentUser }
        for (const k of Object.keys(fresh)) {
          if (fresh[k] !== '' && fresh[k] != null) merged[k] = fresh[k]
        }
        user.value = merged
        uni.setStorageSync('user', JSON.stringify(merged))
      }
    } catch (e) { /* fallback to cache */ }

    // 未读消息
    try {
      const msgRes = await request({ url: '/api/v1/messages/unread-count' })
      unreadCount.value = msgRes?.data?.count || msgRes?.count || 0
    } catch (e) { unreadCount.value = 0 }

    // 并行读取认证状态（企业 / 飞手），互不连坐；实名认证为演示写死状态
    await Promise.allSettled([fetchEnterpriseStatus(), fetchPilotStatus()])
    // 概览统计：真实可用则取，失败回退 0/暂无数据
    await fetchOverviewCounts()
  } else {
    userInitialReset()
  }
}

const fetchEnterpriseStatus = async () => {
  try {
    const res = await request({ url: '/api/v1/enterprises' })
    const data = (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    enterpriseStatus.value = items.length ? items[0].status || '' : ''
  } catch (e) {
    // 请求失败：显示受控空态，不降级为普通个人、不渲染假已通过
    enterpriseStatus.value = 'sync'
  }
}

const fetchPilotStatus = async () => {
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    pilotStatus.value = (res && res.status) || (res && res.data && res.data.status) || ''
    pilotId.value = (res && res.id) || (res && res.data && res.data.id) || ''
  } catch (e) {
    pilotStatus.value = 'sync'
  }
}

const fetchOverviewCounts = async () => {
  overviewLoading.value = true
  const counts = {}
  // 我的发布数：真实接口可读 total
  try {
    const res = await request({ url: '/api/v1/demands', data: { mine: 1, page_size: 1 } })
    const list = Array.isArray(res) ? res : ((res && res.data) || [])
    counts.publish = String(res?.total ?? list.length ?? 0)
  } catch (e) {
    counts.publish = '0'
  }
  // 洽谈会话数：当前无统一计数接口，回退 0（不用 "—"）
  counts.talk = '0'
  overviewCounts.value = counts
  overviewLoading.value = false
}

const userInitialReset = () => {
  unreadCount.value = 0
  pilotStatus.value = ''
  enterpriseStatus.value = ''
  overviewCounts.value = {}
  device.value = null
  overviewLoading.value = false
}

onShow(() => {
  try {
    const info = uni.getSystemInfoSync()
    statusBarH.value = info.statusBarHeight || 20
    let mr = null
    if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
      mr = uni.getMenuButtonBoundingClientRect()
    }
    // 只避让到微信胶囊左缘，让右上按钮尽量靠右
    capsuleGap.value = mr ? Math.max(info.windowWidth - mr.right + 4, 8) : 8
  } catch (e) { /* keep defaults */ }
  fetchData()
})

// ── 导航 ──
const goLogin = () => uni.navigateTo({ url: '/pages/login/index' })

const avatarSrc = (u) => {
  if (!u) return ''
  return u.startsWith('http') ? u : BASE_URL + u
}

const requireLogin = () => {
  if (user.value) return true
  goLogin()
  return false
}

const handleUserClick = () => {
  if (!user.value) return goLogin()
  uni.navigateTo({ url: '/pages/mine/profile' })
}

const goMessages = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pages/messages/index' })
}

const goSettings = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pages/mine/profile' })
}

// 认证 / 资料聚合入口
const goPrimaryCert = () => {
  if (!requireLogin()) return
  if (identity.value === 'enterprise') return goEnterpriseCert()
  if (identity.value === 'pilot') return goPilotCert()
  if (identity.value === 'individual') {
    // 飞手申请中/被驳回：先进飞手入口（审核中提示、驳回重提），无申请才走实名认证
    if (pilotStatus.value && pilotStatus.value !== 'sync') return goPilotCert()
    return goAuth()
  }
  return goComingSoon()
}

const goAuth = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pages/mine/auth' })
}

const goPilotCert = () => {
  if (!requireLogin()) return
  // 已认证 → 查看我的档案；未认证/待审/驳回 → 飞手申请页（不能导到普通实名认证）
  if (pilotStatus.value === 'approved' && pilotId.value) {
    uni.removeStorageSync('pilot_detail')
    uni.navigateTo({ url: '/pkg-talent/pages/pilots/detail?id=' + encodeURIComponent(pilotId.value) })
    return
  }
  uni.navigateTo({ url: '/pkg-talent/pages/pilots/apply' })
}

const goEnterpriseCert = () => {
  if (!requireLogin()) return
  if (enterpriseStatus.value === '' || enterpriseStatus.value === 'sync') {
    uni.navigateTo({ url: '/pkg-eco/pages/enterprise/register' })
    return
  }
  uni.navigateTo({ url: '/pkg-eco/pages/enterprise/status' })
}

const goMyDemands = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pkg-demand/pages/demands/mine' })
}

const goIntents = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pages/intents/mine' })
}

const goCertificates = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pkg-talent/pages/training/certificates' })
}

const goCourses = () => {
  uni.navigateTo({ url: '/pkg-talent/pages/training/courses' })
}

const goOrders = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pages/orders/index' })
}

const goProfile = () => {
  if (!requireLogin()) return
  uni.navigateTo({ url: '/pages/mine/profile' })
}

const goDeviceManage = () => {
  uni.showToast({ title: '设备管理即将开放', icon: 'none' })
}

const goComingSoon = () => {
  uni.showToast({ title: '功能即将开放', icon: 'none' })
}

const doLogout = () => {
  uni.showModal({
    title: '提示',
    content: '确定退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        authStorage.clearTokens()
        uni.removeStorageSync('user')
        user.value = null
        userInitialReset()
        uni.showToast({ title: '已退出登录', icon: 'none' })
      }
    },
  })
}
</script>

<style scoped>
.mine-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(24px + env(safe-area-inset-bottom));
}

/* ===== 板块 ===== */
.page-section {
  margin: 20rpx 24rpx 0;
}
.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 32px;
  padding: 0 8rpx 16rpx;
}
.section-title-text {
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
}
.section-link {
  font-size: 22rpx;
  color: #667085;
  padding: 4rpx 0;
}

.card {
  background: #fff;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  box-shadow: 0 8rpx 32rpx rgba(16,24,40,.06);
  overflow: hidden;
}

/* ===== 概览卡 ===== */
.overview-card {
  display: flex;
  flex-wrap: wrap;
  padding-bottom: 20rpx;
}
.ov-cell {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
  min-height: 144rpx;
  padding: 20rpx 8rpx;
}
.ov-value {
  font-size: 40rpx;
  font-weight: 700;
  color: var(--color-primary);
  line-height: 1.2;
}
.ov-label {
  font-size: 24rpx;
  color: #344054;
}
.ov-note {
  flex-basis: 100%;
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  padding: 22rpx 28rpx 20rpx;
  border-top: 1rpx solid #EEF1F4;
}
.ov-note-lead {
  font-size: 22rpx;
  color: #168A55;
  font-weight: 600;
}
.ov-note-rest {
  font-size: 22rpx;
  color: #667085;
}
/* 骨架 */
.ov-skel {
  width: 72rpx;
  height: 40rpx;
  border-radius: 8rpx;
  background: var(--color-divider);
}
.ov-skel-label {
  width: 96rpx;
  height: 24rpx;
  border-radius: 8rpx;
  background: var(--color-divider);
}

/* ===== 设备与飞行卡 ===== */
.device-card {
  display: flex;
  align-items: stretch;
  min-height: 168rpx;
}
.device-main {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 28rpx;
}
.device-icon {
  width: 96rpx;
  height: 96rpx;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: #E8F2FC;
}
.device-icon-img {
  width: 58rpx;
  height: 58rpx;
}
.device-copy {
  min-width: 0;
}
.device-title {
  display: block;
  font-size: 32rpx;
  font-weight: 700;
  color: #17212B;
}
.device-sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: #667085;
}
.device-online {
  color: #168A55;
  font-weight: 600;
}
.device-sub-rest {
  color: #667085;
}
.device-more {
  flex-shrink: 0;
  width: 178rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
  background: #F7FBFF;
  border-left: 1rpx solid #E1EEFB;
  font-size: 24rpx;
  font-weight: 600;
  color: var(--color-primary);
}
.device-more-arrow {
  font-size: 28rpx;
  font-weight: 300;
}

/* ===== 退出登录 ===== */
.logout-card {
  background: #fff;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  padding: 28rpx 0;
  text-align: center;
  margin: 32rpx 24rpx 0;
  box-shadow: 0 8rpx 32rpx rgba(16,24,40,.06);
}
.logout-text {
  font-size: 28rpx;
  color: #D92D20;
  font-weight: 500;
}

.bottom-spacer { height: 8px; }

.tap-fade { opacity: .8; }
</style>
