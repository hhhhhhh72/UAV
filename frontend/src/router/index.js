import { createRouter, createWebHistory } from 'vue-router'
import { showFailToast } from '@/utils/feedback'
import axios, { authStorage } from '@/utils/http'

const routes = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/layout/Index.vue'),
    redirect: '/home',
    children: [
      {
        path: '/home',
        name: 'Home',
        component: () => import('@/views/home/Index.vue'),
        meta: { title: '服务大厅', showTabbar: true }
      },
      {
        path: '/services',
        name: 'Services',
        component: () => import('@/views/services/Index.vue'),
        meta: { title: '全部服务', showTabbar: true }
      },
      {
        path: '/messages',
        name: 'Messages',
        component: () => import('@/views/messages/Index.vue'),
        meta: { title: '消息', showTabbar: true }
      },
      {
        path: '/applications',
        name: 'Applications',
        component: () => import('@/views/applications/Index.vue'),
        meta: { title: '我的申请', showTabbar: true }
      },
      {
        path: '/mine',
        name: 'Mine',
        component: () => import('@/views/mine/Index.vue'),
        meta: { title: '个人中心', showTabbar: true }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Index.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/register/Index.vue'),
    meta: { title: '注册账号' }
  },
  {
    path: '/study',
    name: 'StudyIndex',
    component: () => import('@/views/study/Index.vue'),
    meta: { title: '低空研学' }
  },
  {
    path: '/study/:id',
    name: 'StudyDetail',
    component: () => import('@/views/study/Detail.vue'),
    meta: { title: '课程详情' }
  },
  {
    path: '/service-detail/:id',
    name: 'ServiceDetail',
    component: () => import('@/views/services/Detail.vue'),
    meta: { title: '服务详情' }
  },
  {
    path: '/service-apply/:id',
    name: 'ServiceApply',
    component: () => import('@/views/services/Apply.vue'),
    meta: { title: '服务申请' }
  },
  {
    path: '/cases',
    name: 'Cases',
    component: () => import('@/views/cases/Index.vue'),
    meta: { title: '案例展示' }
  },
  {
    path: '/reviews',
    name: 'Reviews',
    component: () => import('@/views/reviews/Index.vue'),
    meta: { title: '服务评价' }
  },
  {
    path: '/demand/publish',
    name: 'DemandPublish',
    component: () => import('@/views/demand/Publish.vue'),
    meta: { title: '发布需求', requiresAuth: true }
  },
  {
    path: '/demand/:id',
    name: 'DemandDetail',
    component: () => import('@/views/demand/Detail.vue'),
    meta: { title: '需求详情' }
  },
  {
    path: '/admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    redirect: '/admin/dashboard',
    children: [
      { path: 'dashboard', component: () => import('@/views/admin/Dashboard.vue') },
      { path: 'cases', component: () => import('@/views/admin/cases/CaseList.vue') },
      { path: 'users', component: () => import('@/views/admin/users/UserList.vue') },
      { path: 'competition', component: () => import('@/views/admin/competition/CompetitionList.vue') },
      { path: 'config', component: () => import('@/views/admin/config/ServiceConfigList.vue') },
      { path: 'reviews', component: () => import('@/views/admin/reviews/ReviewList.vue') },
      { path: 'orders', component: () => import('@/views/admin/orders/OrderList.vue') },
      { path: 'enterprises', component: () => import('@/views/admin/enterprises/EnterpriseList.vue') },
      { path: 'demands', component: () => import('@/views/admin/demands/DemandList.vue') },
      // --- Sprint 0: 20 new modules ---
      { path: 'shops', component: () => import('@/views/admin/shops/ShopList.vue') },
      { path: 'experts', component: () => import('@/views/admin/experts/ExpertList.vue') },
      { path: 'resources', component: () => import('@/views/admin/resources/ResourceList.vue') },
      { path: 'compliance', component: () => import('@/views/admin/compliance/ComplianceList.vue') },
      { path: 'training', component: () => import('@/views/admin/training/CourseList.vue') },
      { path: 'certs', component: () => import('@/views/admin/training/CertList.vue') },
      { path: 'jobs', component: () => import('@/views/admin/jobs/JobList.vue') },
      { path: 'colleges', component: () => import('@/views/admin/colleges/CollegeList.vue') },
      { path: 'admin-study', component: () => import('@/views/admin/study/StudyList.vue') },
      { path: 'achievements', component: () => import('@/views/admin/achievements/AchievementList.vue') },
      { path: 'challenges', component: () => import('@/views/admin/challenges/ChallengeList.vue') },
      { path: 'projects', component: () => import('@/views/admin/projects/ProjectList.vue') },
      { path: 'testsites', component: () => import('@/views/admin/testsites/TestSiteList.vue') },
      { path: 'transformations', component: () => import('@/views/admin/transformations/TransList.vue') },
      { path: 'events', component: () => import('@/views/admin/events/EventList.vue') },
      { path: 'portfolios', component: () => import('@/views/admin/portfolios/PortfolioList.vue') },
      { path: 'exhibitions', component: () => import('@/views/admin/exhibitions/ExhibitionList.vue') },
      { path: 'reports', component: () => import('@/views/admin/reports/ReportList.vue') },
      { path: 'emergency-resources', component: () => import('@/views/admin/emergency/ResourceList.vue') },
      { path: 'emergency-dispatches', component: () => import('@/views/admin/emergency/DispatchList.vue') },
      { path: 'messages', component: () => import('@/views/admin/messages/NotifyList.vue') },
      // --- Consolidated (9-module) ---
      { path: 'members', component: () => import('@/views/admin/consolidated/MembersPage.vue') },
      { path: 'trading', component: () => import('@/views/admin/consolidated/TradingPage.vue') },
      { path: 'content', component: () => import('@/views/admin/consolidated/ContentPage.vue') },
      { path: 'talent', component: () => import('@/views/admin/consolidated/TalentPage.vue') },
      { path: 'innovation', component: () => import('@/views/admin/consolidated/InnovationPage.vue') },
      { path: 'promotion', component: () => import('@/views/admin/consolidated/PromotionPage.vue') },
      { path: 'emergency', component: () => import('@/views/admin/consolidated/EmergencyPage.vue') },
      { path: 'settings', component: () => import('@/views/admin/consolidated/SettingsPage.vue') },
    ]
  },
]

const router = createRouter({
  history: createWebHistory('/'),
  routes
})

router.beforeEach(async (to, from, next) => {
  document.title = to.meta.title || '无人机产业综合服务平台'

  // 处理微信OAuth回调（可能落在任意页面）
  const wechatAuth = to.query.wechat_auth
  const userData = to.query.user
  const tokensData = to.query.tokens
  if (wechatAuth === '1' && userData && tokensData) {
    try {
      const user = JSON.parse(atob(userData))
      const tokens = JSON.parse(atob(tokensData))
      localStorage.setItem('user', JSON.stringify(user))
      authStorage.setTokens(tokens.accessToken, tokens.refreshToken)
      // 清除URL中的认证参数，跳转到目标页面
      const { wechat_auth: _w, user: _u, tokens: _t, ...rest } = to.query
      const targetPath = to.path
      next({ path: targetPath, query: rest, replace: true })
      return
    } catch (e) {
      console.error('解析微信登录数据失败:', e)
    }
  }
  
  // 支持 jyauthcode 和 authcode 两种参数名
  const authcode = (typeof to.query.jyauthcode === 'string' ? to.query.jyauthcode.trim() : '') ||
                   (typeof to.query.authcode === 'string' ? to.query.authcode.trim() : '')
  if (authcode) {
    const lastAuthcode = sessionStorage.getItem('ssoAuthcode')
    if (lastAuthcode !== authcode) {
      try {
        const res = await axios.post('/api/sso/login', { authcode })
        if (!res.data?.success) {
          throw new Error(res.data?.message || '授权登录失败')
        }
        localStorage.setItem('user', JSON.stringify(res.data.user))
        authStorage.setTokens(res.data.accessToken, res.data.refreshToken)
        sessionStorage.setItem('ssoAuthcode', authcode)
        const { authcode: _ignored, jyauthcode: _ignored2, ...rest } = to.query
        next({ path: to.path, query: rest, replace: true })
        return
      } catch (error) {
        console.error(error)
        showFailToast(error?.message || '授权登录失败')
        next('/login')
        return
      }
    }
  }

  // RequiresAuth route protection
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('accessToken')
    if (!token) {
      next('/login')
      return
    }
  }

  // Admin route protection
  if (to.path.startsWith('/admin')) {
    let userStr = localStorage.getItem('user')
    let token = localStorage.getItem('accessToken')

    // Auto dev-login: get token from /api/v1/admin/token if missing
    if (!token || !userStr) {
      try {
        const res = await axios.post('/api/v1/admin/token', { role: 'platform_admin' })
        const data = res.data?.data || res.data
        const accessToken = data?.access_token || data?.accessToken
        const refreshToken = data?.refresh_token || data?.refreshToken
        const userInfo = data?.user || {}
        if (accessToken) {
          localStorage.setItem('accessToken', accessToken)
          if (refreshToken) localStorage.setItem('refreshToken', refreshToken)
          localStorage.setItem('user', JSON.stringify({
            id: userInfo.id, role: userInfo.role || 'platform_admin',
            phone: userInfo.id
          }))
          userStr = localStorage.getItem('user')
          token = accessToken
        }
      } catch (e) {
        console.error('dev auto-login failed', e)
        next('/login')
        return
      }
    }

    if (!userStr) { next('/login'); return }
    let user
    try { user = JSON.parse(userStr) } catch (e) { next('/login'); return }
    if (!['platform_admin', 'association_admin'].includes(user.role)) {
      showFailToast('无管理权限，请使用管理员账号登录')
      next('/login')
      return
    }
  }

  next()
})

export default router

