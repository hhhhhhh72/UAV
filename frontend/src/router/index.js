import { createRouter, createWebHistory } from 'vue-router'
import { showFailToast } from '@/utils/feedback'
import axios from '@/utils/http'

const routes = [
  {
    path: '/',
    redirect: '/admin'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Index.vue'),
    meta: { title: '登录' }
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
