<template>
  <view class="page">
    <StateView
      :loading="loading"
      :error="!!errorMsg"
      :empty="!loading && !errorMsg && !detail"
      empty-text="赛事不存在"
      @retry="loadDetail"
    >
      <template v-if="detail">
        <!-- ① 海军蓝 Banner -->
        <view class="banner">
          <view class="back-btn" @click="goBack"><text class="back-icon">‹</text></view>
          <text class="banner-title">赛事详情</text>
        </view>

        <!-- ② 赛事封面图 -->
        <view class="cover-wrapper">
          <image
            v-if="detail.cover || detail.cover_image || detail.image"
            :src="detail.cover || detail.cover_image || detail.image"
            class="cover-img"
            mode="aspectFill"
          />
          <view v-else class="cover-placeholder">
            <text class="cover-emoji">赛</text>
          </view>
          <view
            class="status-badge"
            :style="{ background: statusColor[detail.status] || '#969799' }"
          >{{ statusText[detail.status] || '未知' }}</view>
        </view>

        <!-- ③ 主卡片 -->
        <view class="main-card">
          <text class="comp-name">{{ detail.title || detail.name || '未知赛事' }}</text>

          <view class="tag-row">
            <text
              v-for="tag in compTags(detail)"
              :key="tag"
              class="comp-tag"
              :style="{ background: tagBgColor(tag), color: tagTc(tag) }"
            >{{ tag }}</text>
          </view>

          <!-- ④ 基本信息 -->
          <view class="info-card">
            <view class="info-row">
              <view class="info-icon" style="background:var(--color-primary-light);"><text class="info-icon-text">期</text></view>
              <view class="info-text">
                <text class="info-label">比赛时间</text>
                <text class="info-value">{{ detail.start_date || '2026年9月15日' }} - {{ detail.end_date || '9月18日' }}</text>
              </view>
            </view>
            <view class="info-row">
              <view class="info-icon" style="background:#fff4e6;"><text class="info-icon-text loc">址</text></view>
              <view class="info-text">
                <text class="info-label">比赛地点</text>
                <text class="info-value">{{ detail.location || '深圳宝安区国际会展中心' }}</text>
              </view>
            </view>
            <view class="info-row info-row-last">
              <view class="info-icon" style="background:var(--color-primary-light);"><text class="info-icon-text">止</text></view>
              <view class="info-text">
                <text class="info-label">报名截止</text>
                <text class="info-value deadline">{{ detail.deadline || '2026年9月1日' }}</text>
              </view>
            </view>
          </view>

          <!-- ⑤ 赛事简介 -->
          <view v-if="detail.intro || detail.description" class="section-block">
            <view class="section-title">赛事简介</view>
            <view class="intro-text">{{ detail.intro || detail.description }}</view>
          </view>

          <!-- ⑥ 报名条件 -->
          <view class="section-block">
            <view class="section-title">报名条件</view>
            <view class="requirements-card">
              <view v-for="req in requirements(detail)" :key="req.name" class="req-item">
                <view class="req-icon" :style="{ background: reqBgColor(req.level) }">
                  <text style="font-size:28rpx;">{{ req.icon }}</text>
                </view>
                <view class="req-body">
                  <text class="req-name">{{ req.name }}</text>
                  <text class="req-desc">{{ req.desc }}</text>
                </view>
                <view class="req-badge" :style="reqBadgeStyle(req.level)">{{ req.level }}</view>
              </view>
            </view>
          </view>

          <!-- ⑦ 参赛项目 -->
          <view class="section-block">
            <view class="section-title">参赛项目</view>
            <view class="event-list">
              <view v-for="ev in eventList(detail)" :key="ev.name" class="event-item">
                <view class="event-info">
                  <text class="event-name">{{ ev.name }}</text>
                  <text class="event-meta">{{ ev.type }} · {{ ev.format }}</text>
                </view>
                <text class="event-price">¥{{ ev.fee.toLocaleString() }}</text>
              </view>
            </view>
          </view>

          <!-- ⑧ 奖项 -->
          <view v-if="prizes(detail).length > 0" class="section-block">
            <view class="section-title">奖项设置</view>
            <view class="prize-row">
              <view v-for="p in prizes(detail)" :key="p.level" class="prize-card"
                :style="{ background: p.bg || 'linear-gradient(135deg, #f8f9fc, #eee)' }">
                <text class="prize-emoji">{{ p.emoji }}</text>
                <text class="prize-level">{{ p.level }}</text>
                <text class="prize-amount">¥{{ p.amount.toLocaleString() }}</text>
              </view>
            </view>
          </view>

          <!-- 主办单位 -->
          <view class="section-block">
            <view class="section-title">主办单位</view>
            <view class="organizer-row">
              <view class="org-avatar">{{ orgInitial(detail) }}</view>
              <view class="org-info">
                <text class="org-name">{{ detail.organizer || '中国航空器拥有者及驾驶员协会' }}</text>
                <text class="org-sub">{{ detail.organizer_sub || '国家级行业协会' }}</text>
              </view>
            </view>
          </view>

          <!-- ⑨ 底部按钮 -->
          <view class="bottom-bar">
            <view class="bottom-left">
              <text class="fee-label">报名费</text>
              <text class="fee-value">¥{{ compMinFee(detail) }}起</text>
              <text class="fee-unit">/人</text>
            </view>
            <view class="bottom-actions">
              <view class="btn-outline" @click="handleConsult">咨询</view>
              <view class="btn-primary" :class="{ disabled: detail.status === 'closed' }"
                @click="goRegister">{{ detail.status === 'closed' ? '已截止' : '立即报名' }}</view>
            </view>
          </view>
          <view class="bottom-spacer" />
        </view>
      </template>
    </StateView>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { request } from '../../utils/request'
import StateView from '../../components/StateView.vue'

const id = ref('')
const loading = ref(false)
const errorMsg = ref('')
const detail = ref(null)

const statusColor = { enrolling: 'var(--color-warning)', open: 'var(--color-warning)', ongoing: 'var(--color-primary)', closed: 'var(--color-text-secondary)', full: 'var(--color-text-secondary)' }
const statusText = { enrolling: '报名中', open: '报名中', ongoing: '进行中', closed: '已结束', full: '已满额' }

function compTags(item) {
  if (Array.isArray(item.tags) && item.tags.length > 0) return item.tags
  if (item.category) return [item.category]
  return ['多旋翼', '国家级']
}

function tagBgColor(tag) {
  if (['固定翼', '竞速FPV', '航拍', '多旋翼'].indexOf(tag) >= 0) return 'var(--color-primary-light)'
  if (['国家级', '国际赛'].indexOf(tag) >= 0) return '#fff4e6'
  return 'var(--color-primary-light)'
}

function tagTc(tag) {
  if (['固定翼', '竞速FPV', '航拍', '多旋翼'].indexOf(tag) >= 0) return 'var(--color-primary)'
  if (['国家级', '国际赛'].indexOf(tag) >= 0) return 'var(--color-warning)'
  return 'var(--color-primary)'
}

function requirements(item) {
  if (Array.isArray(item.requirements) && item.requirements.length > 0) return item.requirements
  return [
    { icon: '证', name: '持证要求', desc: '须持有CAAC/AOPA/UTC任一类无人机执照', level: '必满足' },
    { icon: '龄', name: '年龄限制', desc: '年满16周岁，未满18周岁需监护人签字同意', level: '必满足' },
    { icon: '时', name: '飞行时长', desc: '累计飞行时长不低于20小时', level: '建议满足' },
    { icon: '康', name: '健康要求', desc: '身体健康，无色盲色弱', level: '必满足' },
    { icon: '险', name: '保险要求', desc: '须自行购买比赛期间的第三方责任险', level: '建议满足' },
  ]
}

function eventList(item) {
  if (Array.isArray(item.events) && item.events.length > 0) return item.events
  return [
    { name: '多旋翼竞速赛', type: '个人赛', format: '计时排名', fee: 380 },
    { name: '固定翼编队赛', type: '团体赛', format: '3人一队', fee: 680 },
    { name: '航拍创作赛', type: '个人赛', format: '主题创作', fee: 280 },
  ]
}

function prizes(item) {
  if (Array.isArray(item.prizes) && item.prizes.length > 0) return item.prizes
  return [
    { level: '一等奖', amount: 10000, emoji: '冠', bg: 'linear-gradient(135deg, #fff8e1, #fff3c4)' },
    { level: '二等奖', amount: 5000, emoji: '亚', bg: 'linear-gradient(135deg, #f5f5f5, #eeeeee)' },
    { level: '三等奖', amount: 2000, emoji: '季', bg: 'linear-gradient(135deg, #fdf5e6, #fae5c3)' },
  ]
}

function compMinFee(item) {
  if (item.minFee != null) return item.minFee
  var evts = eventList(item)
  if (evts.length > 0) return Math.min.apply(null, evts.map(function (e) { return e.fee }))
  return 280
}

function orgInitial(item) {
  var name = item.organizer || '中'
  return name.charAt(0)
}

function reqBgColor(level) {
  if (level === '必满足') return '#fff4e6'
  if (level === '建议满足') return 'var(--color-primary-light)'
  return '#f5f5f5'
}

function reqBadgeStyle(level) {
  return {
    color: level === '必满足' ? 'var(--color-warning)' : 'var(--color-primary)',
    background: level === '必满足' ? '#fff4e6' : 'var(--color-primary-light)',
  }
}

async function loadDetail() {
  loading.value = true
  errorMsg.value = ''
  try {
    var res = await request({ url: '/api/v1/competitions' })
    var data = Array.isArray(res) ? res : (res && res.data) || res || {}
    var items = Array.isArray(data) ? data : (data && data.items) || data || []
    var found = null
    for (var i = 0; i < items.length; i++) {
      if (String(items[i].id) === String(id.value)) { found = items[i]; break }
    }
    detail.value = found
    if (!found) detail.value = getMockDetail()
  } catch (e) {
    detail.value = getMockDetail()
  } finally {
    loading.value = false
  }
}

function getMockDetail() {
  var mockMap = {
    'comp-1': {
      id: 'comp-1',
      name: '2026全国无人机职业技能大赛',
      title: '2026全国无人机职业技能大赛',
      status: 'enrolling',
      tags: ['多旋翼', '固定翼', '国家级'],
      start_date: '2026年9月15日',
      end_date: '9月18日',
      location: '深圳宝安区国际会展中心',
      deadline: '2026年9月1日',
      intro: '2026全国无人机职业技能大赛是由中国航空器拥有者及驾驶员协会主办的国家级专业赛事，旨在推动无人机技术的应用与发展。\n\n本赛事设有多个竞赛项目，涵盖多旋翼、固定翼等多种机型，欢迎广大飞手踊跃报名参加。',
      organizer: '中国航空器拥有者及驾驶员协会',
      organizer_sub: '简称中国AOPA · 国家级行业协会',
      minFee: 280,
    },
    'comp-2': {
      id: 'comp-2',
      name: '首届西南无人机FPV竞速挑战赛',
      title: '首届西南无人机FPV竞速挑战赛',
      status: 'enrolling',
      tags: ['竞速FPV', '多旋翼'],
      start_date: '2026年10月1日',
      end_date: '10月3日',
      location: '成都天府新区无人机竞速基地',
      deadline: '2026年9月20日',
      intro: '首届西南无人机FPV竞速挑战赛汇聚全国顶尖FPV飞手，在成都天府新区专业竞速赛道上展开速度与技巧的终极对决。\n\n赛道全长1.2公里，设有12个障碍门，最高时速可达120km/h。',
      organizer: '四川省航空运动协会',
      organizer_sub: '省级行业协会 · 专业竞速赛事',
      minFee: 280,
    },
    'comp-3': {
      id: 'comp-3',
      name: '2026无人机创新应用大赛',
      title: '2026无人机创新应用大赛',
      status: 'ongoing',
      tags: ['航拍', '固定翼', '国家级'],
      start_date: '2026年8月1日',
      end_date: '8月15日',
      location: '北京亦庄经济技术开发区',
      deadline: '2026年7月20日',
      intro: '聚焦无人机在航拍、应急救援、农业植保、物流配送等领域的创新应用方案评选。\n\n参赛者需提交完整的项目方案和飞行演示视频，由行业专家组成的评审团进行综合评分。',
      organizer: '工信部人才交流中心',
      organizer_sub: '国家级人才评测机构',
      minFee: 0,
    },
    'comp-4': {
      id: 'comp-4',
      name: '青少年无人机编程挑战赛',
      title: '青少年无人机编程挑战赛',
      status: 'enrolling',
      tags: ['多旋翼', '航拍'],
      start_date: '2026年11月1日',
      end_date: '11月2日',
      location: '上海市浦东新区青少年活动中心',
      deadline: '2026年10月25日',
      intro: '面向8-16岁青少年的无人机编程挑战赛，通过Python/Scratch编程控制无人机完成闯关任务。\n\n比赛分为初级组（8-12岁）和高级组（13-16岁），每组设置不同的难度关卡。',
      organizer: '上海市教育委员会',
      organizer_sub: '市级教育主管部门',
      minFee: 120,
    },
    'comp-5': {
      id: 'comp-5',
      name: '国际无人机系统博览会竞技赛',
      title: '国际无人机系统博览会竞技赛',
      status: 'enrolling',
      tags: ['多旋翼', '固定翼', '国际赛'],
      start_date: '2026年12月5日',
      end_date: '12月7日',
      location: '广州琶洲国际会展中心',
      deadline: '2026年11月20日',
      intro: '全球无人机竞速爱好者的年度盛会，设有专业组和公开组两个级别。\n\n同期举办无人机系统博览会，参展企业超过500家，涵盖整机厂商、零部件供应商、行业解决方案提供商等。',
      organizer: '广州市低空经济产业协会',
      organizer_sub: '市级低空经济行业组织',
      minFee: 580,
    },
    'comp-6': {
      id: 'comp-6',
      name: '2026贵州无人机应急救援演练赛',
      title: '2026贵州无人机应急救援演练赛',
      status: 'closed',
      tags: ['多旋翼', '航拍'],
      start_date: '2026年6月10日',
      end_date: '6月12日',
      location: '贵阳市观山湖区应急指挥中心',
      deadline: '2026年5月30日',
      intro: '模拟山地救援、森林火灾监测、灾害应急物资投送等场景，考察无人机在应急救援中的协同作战能力。\n\n参赛队伍需在指定时间内完成搜索定位、物资投送和灾情评估三项任务。',
      organizer: '贵州省应急管理厅',
      organizer_sub: '省级应急管理部门',
      minFee: 0,
    },
  }
  return mockMap[id.value] || mockMap['comp-1']
}

function goBack() { uni.navigateBack({ delta: 1 }) }

function goRegister() {
  if (detail.value && detail.value.status === 'closed') return
  uni.navigateTo({ url: '/pages/events/register?id=' + encodeURIComponent(id.value) })
}

function handleConsult() {
  uni.showToast({ title: '咨询功能开发中', icon: 'none' })
}

onLoad(function (options) {
  id.value = options.id || ''
  loadDetail()
})
</script>

<style scoped>
.page { min-height: 100vh; background: var(--color-bg); padding-bottom: env(safe-area-inset-bottom); }

/* ① Banner */
.banner {
  background: linear-gradient(135deg, #1a365d, #2a4a7f);
  padding: 80rpx 32rpx 160rpx;
}

.back-btn {
  width: 64rpx; height: 64rpx; background: rgba(255,255,255,0.15);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
  margin-bottom: 24rpx;
}

.back-icon { color: #ffffff; font-size: 40rpx; font-weight: 300; }
.banner-title { color: rgba(255,255,255,0.9); font-size: 28rpx; font-weight: 500; }

/* ② 封面 */
.cover-wrapper {
  margin: -112rpx 24rpx 0; height: 440rpx; border-radius: 32rpx;
  overflow: hidden; position: relative; z-index: 2;
  box-shadow: 0 16rpx 48rpx rgba(26,54,93,0.3);
}

.cover-img { width: 100%; height: 100%; }

.cover-placeholder {
  width: 100%; height: 100%;
  background: linear-gradient(135deg, #0d2137, #1a3a5c);
  display: flex; align-items: center; justify-content: center;
}

.cover-emoji { font-size: 160rpx; opacity: 0.1; }

.status-badge {
  position: absolute; top: 20rpx; right: 20rpx;
  padding: 8rpx 24rpx; border-radius: 24rpx;
  color: #ffffff; font-size: 24rpx; font-weight: 600; letter-spacing: 1rpx;
}

/* ③ 主卡片 */
.main-card {
  background: #ffffff; border-radius: 48rpx 48rpx 0 0; margin-top: -48rpx;
  padding: 48rpx 32rpx 32rpx; position: relative; z-index: 1;
}

.comp-name { font-size: 44rpx; font-weight: 700; color: var(--color-text); line-height: 1.3; margin-bottom: 16rpx; }

.tag-row { display: flex; flex-wrap: wrap; gap: 12rpx; margin-bottom: 28rpx; }
.comp-tag { padding: 6rpx 18rpx; border-radius: 12rpx; font-size: 22rpx; font-weight: 500; }

/* ④ 基本信息 */
.info-card { background: #f8fafc; border-radius: 24rpx; padding: 28rpx; margin-bottom: 32rpx; }

.info-row {
  display: flex; align-items: flex-start; gap: 20rpx;
  padding-bottom: 20rpx; margin-bottom: 20rpx; border-bottom: 1rpx solid #e8ecf0;
}

.info-row-last { border-bottom: none; padding-bottom: 0; margin-bottom: 0; }

.info-icon {
  width: 80rpx; height: 80rpx; border-radius: 20rpx;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}

.info-icon-text { font-size: 28rpx; color: var(--color-primary); font-weight: 600; }
.info-icon-text.loc { color: var(--color-warning); }

.info-text { flex: 1; }
.info-label { font-size: 24rpx; color: var(--color-text-secondary); display: block; margin-bottom: 6rpx; }
.info-value { font-size: 28rpx; color: var(--color-text); font-weight: 500; }
.info-value.deadline { color: var(--color-warning); font-weight: 500; }

/* Section */
.section-block { margin-top: 36rpx; }

.section-title {
  font-size: 30rpx; font-weight: 700; color: var(--color-text); padding-left: 20rpx;
  border-left: 6rpx solid var(--color-primary); line-height: 1.3; margin-bottom: 20rpx;
}

/* ⑤ 简介 */
.intro-text { font-size: 28rpx; color: #4a4a4a; line-height: 1.8; white-space: pre-line; margin-bottom: 36rpx; }

/* ⑥ 报名条件 */
.requirements-card {
  background: #fefaf3; border-radius: 24rpx; border: 1rpx solid #ffe8c0;
  padding: 28rpx; margin-bottom: 36rpx;
}

.req-item {
  display: flex; align-items: flex-start; gap: 16rpx;
  padding-bottom: 20rpx; margin-bottom: 20rpx; border-bottom: 1rpx solid #f5e6c8;
}

.req-item:last-child { border-bottom: none; padding-bottom: 0; margin-bottom: 0; }

.req-icon {
  width: 56rpx; height: 56rpx; border-radius: 14rpx;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}

.req-body { flex: 1; }
.req-name { font-size: 28rpx; font-weight: 500; color: #1a1a1a; display: block; margin-bottom: 4rpx; }
.req-desc { font-size: 24rpx; color: #969799; line-height: 1.5; }

.req-badge { padding: 4rpx 16rpx; border-radius: 10rpx; font-size: 22rpx; font-weight: 600; flex-shrink: 0; }

/* ⑦ 参赛项目 */
.event-list { display: flex; flex-direction: column; gap: 16rpx; margin-bottom: 36rpx; }

.event-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 24rpx; background: #f8fafc; border-radius: 18rpx;
}

.event-name { font-size: 28rpx; font-weight: 500; color: #1a1a1a; display: block; }
.event-meta { font-size: 24rpx; color: #969799; margin-top: 6rpx; display: block; }
.event-price { font-size: 32rpx; font-weight: 600; color: var(--color-warning); }

/* ⑧ 奖项 */
.prize-row { display: flex; gap: 16rpx; margin-bottom: 36rpx; }
.prize-card { flex: 1; padding: 24rpx 12rpx; border-radius: 18rpx; text-align: center; }
.prize-emoji { font-size: 48rpx; display: block; margin-bottom: 6rpx; }
.prize-level { font-size: 26rpx; font-weight: 500; display: block; margin-bottom: 4rpx; }
.prize-amount { font-size: 30rpx; font-weight: 700; display: block; }

/* 主办单位 */
.organizer-row { display: flex; align-items: flex-start; gap: 20rpx; margin-bottom: 40rpx; }

.org-avatar {
  width: 80rpx; height: 80rpx; background: var(--color-primary); border-radius: 20rpx;
  display: flex; align-items: center; justify-content: center;
  color: #ffffff; font-size: 36rpx; font-weight: 600; flex-shrink: 0;
}

.org-name { font-size: 28rpx; font-weight: 500; color: var(--color-text); display: block; margin-bottom: 4rpx; }
.org-sub { font-size: 24rpx; color: var(--color-text-secondary); }

/* ⑨ 底部按钮 */
.bottom-bar {
  display: flex; justify-content: space-between; align-items: center;
  border-top: 1rpx solid #ebedf0; padding-top: 24rpx;
}

.fee-label { font-size: 24rpx; color: var(--color-text-secondary); }
.fee-value { font-size: 44rpx; font-weight: 700; color: var(--color-warning); margin: 0 8rpx; }
.fee-unit { font-size: 24rpx; color: var(--color-text-secondary); }

.bottom-actions { display: flex; gap: 20rpx; }

.btn-outline {
  padding: 18rpx 36rpx; border-radius: 30rpx;
  border: 2rpx solid var(--color-primary); color: var(--color-primary);
  font-size: 28rpx; font-weight: 500;
}

.btn-primary {
  padding: 18rpx 40rpx; border-radius: 30rpx;
  background: var(--color-primary); color: #ffffff;
  font-size: 28rpx; font-weight: 600;
}

.btn-primary.disabled { background: var(--color-text-placeholder); }
.bottom-spacer { height: calc(40rpx + env(safe-area-inset-bottom)); }
</style>
