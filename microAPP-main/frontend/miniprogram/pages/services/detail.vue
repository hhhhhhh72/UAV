<template>
  <view class="service-detail-page" v-if="service">
    <view class="detail-content">
      <!-- 1. 顶部基础信息 (立即渲染) -->
      <view class="service-header">
        <view class="service-icon-big" :style="{ background: service.color }">
          <image :src="service.icon" mode="aspectFit" class="service-icon-img" />
        </view>
        <view class="service-name">{{ service.name }}</view>
        <view class="service-slogan">{{ service.slogan }}</view>
      </view>

      <!-- 2. 延迟渲染区 (确保跳转流畅) -->
      <view v-if="contentReady">
        <!-- 通用服务介绍 -->
        <view class="section-card" v-if="service.id !== '6'">
          <view class="section-title" :style="{ borderLeftColor: service.mainColor }">服务介绍</view>
          <text class="section-text">{{ service.intro }}</text>

          <!-- 研学（ID: 9）固定图文展示：嵌入“服务介绍”内，不接入案例系统 -->
          <view v-if="service.id === '9'" class="study-showcase">
            <view class="study-subtitle">往期活动展示</view>
            <view class="study-grid">
              <view
                v-for="item in studyShowcase"
                :key="item.title"
                class="study-item"
                @tap="previewStudy(item)"
              >
                <image :src="item.image" mode="aspectFill" class="study-img" />
                <view class="study-info">
                  <view class="study-title">{{ item.title }}</view>
                  <view class="study-desc">{{ item.desc }}</view>
                </view>
              </view>
            </view>
            <view class="study-tip">说明：当前为固定展示内容，后续可升级为可配置/可运营。</view>
          </view>
        </view>

        <!-- 通用服务项目 -->
        <view class="section-card" v-if="service.id !== '6'">
          <view class="section-title" :style="{ borderLeftColor: service.mainColor }">服务项目</view>
          <view class="project-grid">
            <view v-for="(item, index) in service.projects" :key="index" class="project-item">
              <text class="project-icon" :style="{ color: service.mainColor }">◈</text>
              <text>{{ item.name }}</text>
            </view>
          </view>
        </view>

        <!-- 通用服务优势 -->
        <view class="section-card" v-if="service.id !== '6'">
          <view class="section-title" :style="{ borderLeftColor: service.mainColor }">服务优势</view>
          <view class="advantage-list">
            <view v-for="(adv, index) in service.advantages" :key="index" class="advantage-item">
              <text class="check-icon">✓</text>
              <text>{{ adv }}</text>
            </view>
          </view>
        </view>

        <!-- 飞手培训 (ID: 6) 专属内容 -->
        <template v-if="service.id === '6'">
          <view class="section-card">
            <view class="section-title" :style="{ borderLeftColor: service.mainColor }">报名条件</view>
            <view class="training-list">
              <view class="training-item">(一)、中华人民共和国公民;</view>
              <view class="training-item">(二)、年满16周岁以上，70周岁以下;</view>
              <view class="training-item">(三)、初中以上文化程度;</view>
              <view class="training-item">(四)、遵纪守法，无不良行为，五年内无犯罪记录;</view>
              <view class="training-item">(五)、身体健康;矫正视力1.0以上，无色盲、色弱，肢体无残疾;</view>
              <view class="training-item">(六)、具有适应无人机操控需要的基本知识和操作能力。</view>
            </view>
          </view>

          <view class="section-card">
            <view class="section-title" :style="{ borderLeftColor: service.mainColor }">培训费用</view>
            <view class="price-list">
              <view class="price-item"><text class="label">小型无人机-多旋翼-视距内</text><text class="price">8800元/人</text></view>
              <view class="price-item"><text class="label">小型无人机-多旋翼-超视距</text><text class="price">12800元/人</text></view>
              <view class="price-item"><text class="label">中型无人机-多旋翼-视距内</text><text class="price">10800元/人</text></view>
              <view class="price-item"><text class="label">中型无人机-多旋翼-超视距</text><text class="price">15800元/人</text></view>
              <view class="price-item"><text class="label">U-BOX 3.0 套装</text><view class="price">490元/套 <text class="tip">(自愿)</text></view></view>
            </view>
          </view>

          <view class="section-card">
            <view class="section-title" :style="{ borderLeftColor: service.mainColor }">教学特色</view>
            <view class="feature-list">
              <view class="feature-item">
                <view class="feature-title">权威认证</view>
                <view class="feature-desc">御风航空多次荣获中国AOPA年度优秀训练机构称号学员，通过培训可获得中国民用航空局发放的无人机操控员执照。</view>
              </view>
              <view class="feature-item">
                <view class="feature-title">全面课程</view>
                <view class="feature-desc">涵盖无人机基础知识、飞行操作、维护保养、法律法规等，提供丰富的实操机会。</view>
              </view>
              <view class="feature-item">
                <view class="feature-title">资深老牌</view>
                <view class="feature-desc">深耕无人机培训7年，积累丰富经验，专业教员团队，个性化教程，精准施教。</view>
              </view>
            </view>
          </view>

          <view class="section-card">
            <view class="section-title" :style="{ borderLeftColor: service.mainColor }">公司简介</view>
            <view class="company-intro">
              <view class="intro-title">浙江御风航空科技有限公司</view>
              <text class="section-text">系温州交运集团所属低空公司的控股子公司，成立于2018年，致力于无人机专业培训，为行业客户提供专业的解决方案和人才培养。作为国内早期开展无人机驾驶员资格培训的机构之一，是浙南闽北地区最早一家具备民航局认定的CAAC执照培训资质的机构。</text>
            </view>
          </view>

          <view class="section-card">
            <view class="section-title" :style="{ borderLeftColor: service.mainColor }">执照功能</view>
            <view class="license-intro">
              <text class="section-text">CAAC 是中国民用航空局签发的操控员执照，含金量极高，是无人机行业入行必备的敲门砖，具有权威法律效力。取得该执照可申报空域、申请航线、从事无人机相关的商业活动等。</text>
            </view>
          </view>
        </template>

        <!-- 联系客服 -->
        <view class="section-card contact-card">
          <view class="section-title" :style="{ borderLeftColor: service.mainColor }">联系客服</view>
          <view class="contact-info">
            <view class="contact-row">如有疑问，请咨询客服热线：</view>
            <view class="phone-link" :style="{ color: service.mainColor }" @tap="makeCall('0577-55558188')">0577-55558188</view>
            <view v-if="service.id === '6'" class="phone-link" :style="{ color: service.mainColor }" @tap="makeCall('0577-88360168')">0577-88360168</view>
            <view class="work-time">工作时间：工作日 8:30-17:30</view>
          </view>
        </view>
      </view>

      <!-- 骨架屏占位 -->
      <view v-else class="skeleton-wrap">
        <view class="skeleton-block"></view>
        <view class="skeleton-block"></view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view class="action-bar" v-if="service">
      <button class="apply-btn" type="primary" :style="{ background: service.color }" @tap="onApply">
        {{ actionButtonText }}
      </button>
    </view>

    <HomeFloatButton />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReady } from '@dcloudio/uni-app'
import HomeFloatButton from '@/components/HomeFloatButton.vue'
import { request } from '../../utils/request'

const contentReady = ref(false)
const service = ref(null)
const serviceConfig = ref({})

const studyShowcase = ref([
  {
    title: '低空科普课堂',
    desc: '从基础原理到安全规范，互动式讲解让孩子更易理解。',
    image: '/static/images/study/science-class.svg'
  },
  {
    title: '真机飞行体验',
    desc: '在专业导师指导下完成基础操控与任务闯关，提升动手能力。',
    image: '/static/images/study/flight-experience.svg'
  },
  {
    title: '成果与纪念',
    desc: '完成学习任务与展示，记录成长瞬间，获得满满成就感。',
    image: '/static/images/study/achievement.svg'
  }
])

const fetchServiceConfig = async (id) => {
  try {
    const res = await request({ url: '/api/services/config' })
    const allConfigs = res?.data || res || {}
    const config = allConfigs[id] || {}
    serviceConfig.value = config

    if (id === '9') {
      if (config.studyShowcase && config.studyShowcase.length > 0) {
        studyShowcase.value = config.studyShowcase
      } else {
        try {
          const showcaseRes = await request({ url: '/api/study/showcase' })
          const items = showcaseRes?.data || showcaseRes || []
          if (Array.isArray(items) && items.length > 0) {
            studyShowcase.value = items
          }
        } catch (e) { /* use default */ }
      }
    }
  } catch (e) {
    console.warn('Failed to load service config:', e)
  }
}

const previewStudy = (item) => {
  if (!item?.image) return
  uni.previewImage({ urls: [item.image] })
}

// 补全完整服务数据映射 (1:1 同步自 H5)
const serviceData = {
  '1': { id: '1', name: '无人机物流服务', slogan: '快速配送 · 安全可靠', icon: '/static/icons/logistics-drone.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)', mainColor: '#1677ff', intro: '利用先进的无人机技术，为城市和偏远地区提供快速、高效的物资配送服务。', projects: [{name: '城市配送'}, {name: '紧急物资'}, {name: '医疗运输'}], advantages: ['2小时快速响应', '全程GPS跟踪', '专业团队操作', '全程保险覆盖'] },
  '2': { id: '2', name: '政务服务', slogan: '智能巡检 · 降本增效', icon: '/static/icons/government.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)', mainColor: '#722ed1', intro: '提供环保监测、安全巡查、设施检查等，助力智慧城市建设。', projects: [{name: '环保监测'}, {name: '安全巡查'}, {name: '设施检查'}], advantages: ['高清数据采集', '实时监控反馈', 'AI智能分析'] },
  '3': { id: '3', name: '无人机托管服务', slogan: '专业托管 · 安全放心', icon: '/static/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)', mainColor: '#52c41a', intro: '提供保养维护、安全存储、代飞服务等一站式解决方案。', projects: [{name: '专业维护'}, {name: '安全存储'}, {name: '保险服务'}], advantages: ['专业维保团队', '安全存储环境', '完善保险保障'] },
  '4': { id: '4', name: '无人机吊运服务', slogan: '高空作业 · 精准操控', icon: '/static/icons/lifting.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)', mainColor: '#faad14', intro: '适用于高空作业、建筑施工、设备安装等场景。', projects: [{name: '高空吊运'}, {name: '设备安装'}, {name: '建筑施工'}], advantages: ['专业吊运设备', '严格安全规范', '经验丰富团队'] },
  '5': { id: '5', name: '无人机表演服务', slogan: '震撼视觉 · 创意编排', icon: '/static/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)', mainColor: '#eb2f96', intro: '包括无人机编队飞行、灯光秀、创意表演等。', projects: [{name: '编队飞行'}, {name: '灯光秀'}, {name: '活动定制'}], advantages: ['创意编排', '震撼效果', '安全可控'] },
  '6': { id: '6', name: '飞手培训服务', slogan: '专业培训 · 证书认证', icon: '/static/icons/training-v2.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)', mainColor: '#faad14', intro: '提供 CAAC 执照培训、技能提升等。', projects: [{name: 'CAAC执照'}, {name: '技能培训'}, {name: '实操教学'}], advantages: ['资质齐全', '经验丰富', '通过率高'] },
  '7': { id: '7', name: '无人机租赁服务', slogan: '灵活租赁 · 多种机型', icon: '/static/icons/rent.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)', mainColor: '#13c2c2', intro: '多种机型可选，满足不同场景的使用需求。', projects: [{name: '设备租赁'}, {name: '配件租赁'}], advantages: ['机型丰富', '价格优惠', '技术支持'] },
  '8': { id: '8', name: '无人机外卖配送', slogan: '即时配送 · 快速送达', icon: '/static/icons/delivery.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)', mainColor: '#f5222d', intro: '实现城市即时配送，带来全新的外卖体验。', projects: [{name: '即时配送'}, {name: '在线下单'}], advantages: ['30分钟送达', '保温保鲜', '无接触配送'] },
  '9': { id: '9', name: '低空研学服务', slogan: '科普教育 · 实践体验', icon: '/static/icons/study.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)', mainColor: '#722ed1', intro: '面向青少年开展无人机科普、飞行体验活动。', projects: [{name: '科普讲座'}, {name: '飞行体验'}], advantages: ['专业导师', '安全场地', '完整体系'] },
  '10': { id: '10', name: '无人机销售', slogan: '品质保障 · 专业服务', icon: '/static/icons/shop.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)', mainColor: '#fa8c16', intro: '支持设备买卖、以旧换新、专业检测。', projects: [{name: '设备买卖'}, {name: '以旧换新'}], advantages: ['交易安全', '置换优惠', '检测报告'] },
  '11': { id: '11', name: '金融服务', slogan: '设备保险 · 飞行护航', icon: '/static/icons/finance.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)', mainColor: '#1677ff', intro: '涵盖设备险、责任险、飞手险等。', projects: [{name: '设备保险'}, {name: '快速理赔'}], advantages: ['全面保障', '快速理赔', '风险评估'] },
  '12': { id: '12', name: '维修服务', slogan: '专业维修 · 原厂配件', icon: '/static/icons/wrench.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)', mainColor: '#2f54eb', intro: '解决各类硬件故障与软件问题。', projects: [{name: '故障维修'}, {name: '定期保养'}], advantages: ['官方授权', '正品配件', '质保承诺'] },
  '13': { id: '13', name: '无人机赛事', slogan: '竞技比赛 · 精彩纷呈', icon: '/static/icons/competition.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)', mainColor: '#e11d48', intro: '提供无人机竞技赛事组织、报名、赛事执行等全流程服务。涵盖竞速赛、花飞赛、编程赛等多种赛事类型。', projects: [{name: '赛事组织'}, {name: '选手报名'}, {name: '裁判服务'}, {name: '赛事执行'}], advantages: ['专业赛事团队', '标准化赛事流程', '全程安全保障', '多赛事类型支持'] }
}

onLoad((options) => {
  const id = String(options.id || '1')
  service.value = serviceData[id] || serviceData['1']
  if (service.value) {
    uni.setNavigationBarTitle({ title: service.value.name })
  }
  fetchServiceConfig(id)
})

onReady(() => {
  setTimeout(() => { contentReady.value = true }, 150)
})

const actionButtonText = computed(() => {
  if (!service.value) return '立即办理'
  const id = service.value.id
  if (['1', '4', '8'].includes(id)) return '立即下单'
  if (['6', '9', '13'].includes(id)) return '立即报名'
  return '立即办理'
})

const onApply = () => {
  if (service.value.id === '8') {
    uni.navigateTo({ url: `/pages/webview/index?src=${encodeURIComponent('https://app.wzsjy.com:8446/h5/#/pages/diy/diy?pageId=130')}` })
  } else {
    uni.navigateTo({ url: `/pages/services/apply?id=${service.value.id}` })
  }
}

const makeCall = (phone) => { uni.makePhoneCall({ phoneNumber: phone }) }
</script>

<style scoped>
.service-detail-page { min-height: 100vh; background: #f7f8fa; padding-bottom: 100px; }
.service-header { background: #fff; padding: 32px 20px; text-align: center; }
.service-icon-big { width: 88px; height: 88px; border-radius: 20px; display: flex; align-items: center; justify-content: center; margin: 0 auto 16px; box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1); }
.service-icon-img { width: 48px; height: 48px; filter: brightness(0) invert(1); }
.service-name { font-size: 22px; font-weight: bold; color: #323233; margin-bottom: 8px; }
.service-slogan { font-size: 14px; color: #969799; }
.section-card { background: #fff; margin: 12px 16px; padding: 16px; border-radius: 12px; }
.section-title { font-size: 16px; font-weight: bold; color: #323233; margin-bottom: 16px; padding-left: 12px; border-left: 4px solid #1677ff; }
.section-text { font-size: 14px; color: #646566; line-height: 1.8; display: block; }
.project-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.project-item { display: flex; align-items: center; gap: 8px; padding: 12px; background: #f7f8fa; border-radius: 8px; font-size: 13px; color: #323233; }
.advantage-list { display: flex; flex-direction: column; gap: 10px; }
.advantage-item { display: flex; align-items: center; gap: 8px; font-size: 14px; color: #646566; }
.check-icon { color: #07c160; font-weight: bold; }
.training-list { font-size: 14px; color: #646566; line-height: 1.6; }
.training-item { margin-bottom: 8px; }
.price-list { display: flex; flex-direction: column; }
.price-item { display: flex; justify-content: space-between; align-items: center; padding: 12px 0; border-bottom: 1px dashed #ebedf0; }
.price-item:last-child { border-bottom: none; }
.price-item .label { font-size: 14px; color: #323233; }
.price-item .price { font-size: 16px; font-weight: bold; color: #ee0a24; }
.price-item .tip { font-size: 12px; color: #969799; font-weight: normal; }
.feature-item { margin-bottom: 16px; }
.feature-title { font-size: 15px; font-weight: bold; color: #323233; margin-bottom: 4px; position: relative; padding-left: 10px; }
.feature-title::before { content: ''; position: absolute; left: 0; top: 50%; transform: translateY(-50%); width: 4px; height: 4px; border-radius: 50%; background-color: #323233; }
.feature-desc { font-size: 13px; color: #646566; line-height: 1.6; padding-left: 10px; }
.company-intro .intro-title { font-size: 15px; font-weight: bold; margin-bottom: 8px; }
.contact-card { text-align: center; }
.contact-row { font-size: 14px; color: #646566; margin-bottom: 8px; }
.phone-link { font-size: 24px; font-weight: bold; margin: 12px 0; }
.work-time { font-size: 12px; color: #969799; margin-top: 8px; }
.action-bar { position: fixed; bottom: 0; left: 0; right: 0; padding: 16px; background: #fff; border-top: 1px solid #eee; padding-bottom: calc(16px + env(safe-area-inset-bottom)); z-index: 100; }
.apply-btn { width: 100%; border-radius: 999px; font-weight: bold; }
.skeleton-wrap { padding: 20px; }
.skeleton-block { height: 120px; background: #eee; border-radius: 12px; margin-bottom: 16px; animation: blink 1.5s infinite; }
@keyframes blink { 0% { opacity: 0.5; } 50% { opacity: 1; } 100% { opacity: 0.5; } }

.study-showcase { margin-top: 20rpx; }
.study-subtitle { font-size: 30rpx; font-weight: 700; color: #323233; margin: 12rpx 0; }
.study-grid { display: flex; flex-direction: column; gap: 20rpx; }
.study-item { background: #fff; border-radius: 24rpx; overflow: hidden; border: 1rpx solid rgba(0, 0, 0, 0.04); }
.study-img { width: 100%; height: 240rpx; }
.study-info { padding: 18rpx 20rpx 20rpx; }
.study-title { font-size: 28rpx; font-weight: 700; color: #1a1a1a; margin-bottom: 10rpx; }
.study-desc { font-size: 24rpx; color: #646566; line-height: 1.6; }
.study-tip { margin-top: 14rpx; font-size: 22rpx; color: #969799; }
</style>
