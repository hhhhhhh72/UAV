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
      { path: 'dashboard', component: () => import('@/views/admin/Dashboard.vue'), meta: { title: '数据看板' } },
      { path: 'cases', component: () => import('@/views/admin/cases/CaseList.vue'), meta: { title: '案例管理' } },
      { path: 'users', component: () => import('@/views/admin/users/UserList.vue'), meta: { title: '用户管理' } },
      { path: 'competition', component: () => import('@/views/admin/competition/CompetitionList.vue'), meta: { title: '赛事管理' } },
      { path: 'config', component: () => import('@/views/admin/config/ServiceConfigList.vue'), meta: { title: '系统配置' } },
      { path: 'reviews', component: () => import('@/views/admin/reviews/ReviewList.vue'), meta: { title: '评价管理' } },
      { path: 'orders', component: () => import('@/views/admin/orders/OrderList.vue'), meta: { title: '订单管理' } },
      { path: 'products', component: () => import('@/views/admin/products/ProductList.vue'), meta: { title: '商品管理' } },
      { path: 'enterprises', component: () => import('@/views/admin/enterprises/EnterpriseList.vue'), meta: { title: '企业管理' } },
      { path: 'demands', component: () => import('@/views/admin/demands/DemandList.vue'), meta: { title: '需求管理' } },
      // --- Sprint 0: 20 new modules ---
      { path: 'shops', component: () => import('@/views/admin/shops/ShopList.vue'), meta: { title: '商家管理' } },
      { path: 'experts', component: () => import('@/views/admin/experts/ExpertList.vue'), meta: { title: '专家管理' } },
      { path: 'resources', component: () => import('@/views/admin/resources/ResourceList.vue'), meta: { title: '产业资源' } },
      { path: 'compliance', component: () => import('@/views/admin/compliance/ComplianceList.vue'), meta: { title: '合规管理' } },
      { path: 'training', component: () => import('@/views/admin/training/CourseList.vue'), meta: { title: '培训课程' } },
      { path: 'certs', component: () => import('@/views/admin/training/CertList.vue'), meta: { title: '证书管理' } },
      { path: 'jobs', component: () => import('@/views/admin/jobs/JobList.vue'), meta: { title: '职位管理' } },
      { path: 'colleges', component: () => import('@/views/admin/colleges/CollegeList.vue'), meta: { title: '院校管理' } },
      { path: 'admin-study', component: () => import('@/views/admin/study/StudyList.vue'), meta: { title: '研学管理' } },
      { path: 'achievements', component: () => import('@/views/admin/achievements/AchievementList.vue'), meta: { title: '科技成果' } },
      { path: 'challenges', component: () => import('@/views/admin/challenges/ChallengeList.vue'), meta: { title: '研发难题' } },
      { path: 'projects', component: () => import('@/views/admin/projects/ProjectList.vue'), meta: { title: '课题项目' } },
      { path: 'testsites', component: () => import('@/views/admin/testsites/TestSiteList.vue'), meta: { title: '测试场地' } },
      { path: 'transformations', component: () => import('@/views/admin/transformations/TransList.vue'), meta: { title: '成果转化' } },
      { path: 'events', component: () => import('@/views/admin/events/EventList.vue'), meta: { title: '活动管理' } },
      { path: 'portfolios', component: () => import('@/views/admin/portfolios/PortfolioList.vue'), meta: { title: '企业案例' } },
      { path: 'exhibitions', component: () => import('@/views/admin/exhibitions/ExhibitionList.vue'), meta: { title: '展会排期' } },
      { path: 'reports', component: () => import('@/views/admin/reports/ReportList.vue'), meta: { title: '行业报告' } },
      { path: 'emergency-resources', component: () => import('@/views/admin/emergency/ResourceList.vue'), meta: { title: '应急资源' } },
      { path: 'emergency-dispatches', component: () => import('@/views/admin/emergency/DispatchList.vue'), meta: { title: '应急调度' } },
      { path: 'messages', component: () => import('@/views/admin/messages/NotifyList.vue'), meta: { title: '消息通知' } },
      // --- Consolidated (9-module) ---
      { path: 'members', component: () => import('@/views/admin/consolidated/MembersPage.vue'), meta: { title: '会员管理' } },
      { path: 'trading', component: () => import('@/views/admin/consolidated/TradingPage.vue'), meta: { title: '交易管理' } },
      { path: 'content', component: () => import('@/views/admin/consolidated/ContentPage.vue'), meta: { title: '内容管理' } },
      { path: 'talent', component: () => import('@/views/admin/consolidated/TalentPage.vue'), meta: { title: '人才教育' } },
      { path: 'innovation', component: () => import('@/views/admin/consolidated/InnovationPage.vue'), meta: { title: '产学研协同' } },
      { path: 'promotion', component: () => import('@/views/admin/consolidated/PromotionPage.vue'), meta: { title: '运营推广' } },
      { path: 'emergency', component: () => import('@/views/admin/consolidated/EmergencyPage.vue'), meta: { title: '应急协同' } },
      { path: 'settings', component: () => import('@/views/admin/consolidated/SettingsPage.vue'), meta: { title: '系统设置' } },
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
