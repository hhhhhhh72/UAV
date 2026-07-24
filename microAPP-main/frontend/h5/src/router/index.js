import { createRouter, createWebHistory } from 'vue-router'
import { showFailToast } from 'vant'
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
  // 医疗配送模块路由
  {
    path: '/medical/certification',
    name: 'MedicalCertification',
    component: () => import('@/views/medical/Certification.vue'),
    meta: { title: '寄件人认证' }
  },
  {
    path: '/medical/certification/status',
    name: 'MedicalCertificationStatus',
    component: () => import('@/views/medical/CertificationStatus.vue'),
    meta: { title: '认证状态' }
  },
  {
    path: '/medical/order/create',
    name: 'MedicalOrderCreate',
    component: () => import('@/views/medical/OrderCreate.vue'),
    meta: { title: '医疗配送下单' }
  },
  {
    path: '/medical/order/map-select',
    name: 'MedicalMapSelect',
    component: () => import('@/views/medical/MapSelect.vue'),
    meta: { title: '选择起降场' }
  },
  {
    path: '/medical/contacts',
    name: 'MedicalContacts',
    component: () => import('@/views/medical/Contacts.vue'),
    meta: { title: '常用联系人' }
  },
  {
    path: '/medical/orders',
    name: 'MedicalOrders',
    component: () => import('@/views/medical/Orders.vue'),
    meta: { title: '我的配送订单' }
  },
  {
    path: '/medical/orders/:id',
    name: 'MedicalOrderDetail',
    component: () => import('@/views/medical/OrderDetail.vue'),
    meta: { title: '订单详情' }
  },
  {
    path: '/medical/orders/:id/rate',
    name: 'MedicalOrderRate',
    component: () => import('@/views/medical/OrderRate.vue'),
    meta: { title: '订单评价' }
  },
  {
    path: '/medical/received',
    name: 'MedicalReceivedOrders',
    component: () => import('@/views/medical/ReceivedOrders.vue'),
    meta: { title: '寄给我的' }
  },
  {
    path: '/admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    meta: { title: '后台管理' },
    redirect: '/admin/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: { title: '数据概览' }
      },
      {
        path: 'orders',
        name: 'AdminOrders',
        component: () => import('@/views/admin/orders/OrderList.vue'),
        meta: { title: '订单管理' }
      },
      {
        path: 'cases',
        name: 'AdminCases',
        component: () => import('@/views/admin/cases/CaseList.vue'),
        meta: { title: '案例管理' }
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/users/UserList.vue'),
        meta: { title: '用户管理' }
      },
      {
        path: 'competition',
        name: 'AdminCompetition',
        component: () => import('@/views/admin/competition/CompetitionList.vue'),
        meta: { title: '赛事管理' }
      },
      {
        path: 'config',
        name: 'AdminConfig',
        component: () => import('@/views/admin/config/ServiceConfigList.vue'),
        meta: { title: '服务配置' }
      },
      {
        path: 'reviews',
        name: 'AdminReviews',
        component: () => import('@/views/admin/reviews/ReviewList.vue'),
        meta: { title: '评价管理' }
      },
      // 医疗配送管理端
      {
        path: 'medical/orders',
        name: 'AdminMedicalOrders',
        component: () => import('@/views/admin/medical/OrderList.vue'),
        meta: { title: '配送订单管理' }
      },
      {
        path: 'medical/certifications',
        name: 'AdminMedicalCertifications',
        component: () => import('@/views/admin/medical/CertificationList.vue'),
        meta: { title: '认证审核' }
      },
      {
        path: 'medical/pads',
        name: 'AdminMedicalPads',
        component: () => import('@/views/admin/medical/PadList.vue'),
        meta: { title: '起降场管理' }
      }
    ]
  },
  {
    path: '/games',
    name: 'Games',
    component: () => import('@/views/games/Lobby.vue'),
    meta: { title: 'Fruit Box', showTabbar: false }
  },
  {
    path: '/games/play',
    name: 'GamesPlay',
    component: () => import('@/views/games/Play.vue'),
    meta: { title: 'Fruit Box', showTabbar: false }
  }
]

const router = createRouter({
  history: createWebHistory('/'),
  routes
})

router.beforeEach(async (to, from, next) => {
  document.title = to.meta.title || '低空综合服务平台'

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
      const targetPath = ['admin', 'dsl_admin', 'study_admin'].includes(user.role) ? '/admin' : to.path
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

  // Medical module route protection - require login
  if (to.path.startsWith('/medical')) {
    const accessToken = authStorage.getAccessToken()
    if (!accessToken) {
      showFailToast('请先登录后再使用医疗配送功能')
      next('/login')
      return
    }
  }

  // Admin route protection
  if (to.path.startsWith('/admin')) {
    const accessToken = authStorage.getAccessToken()
    const userStr = localStorage.getItem('user')
    if (!accessToken && !userStr) {
      next('/login')
      return
    }
    let user = null
    if (userStr) {
      try {
        user = JSON.parse(userStr)
      } catch (e) {
        user = null
      }
    }
    if (!user && accessToken) {
      try {
        const res = await axios.get('/api/auth/me')
        if (res.data?.success) {
          user = res.data.user
          localStorage.setItem('user', JSON.stringify(user))
        }
      } catch (e) {
        user = null
      }
    }
    if (!user) {
      next('/login')
      return
    }
    if (!['admin', 'dsl_admin', 'study_admin'].includes(user.role)) {
      showFailToast('无管理权限，请使用管理员账号登录')
      next('/login')
      return
    }
  }
  
  next()
})

export default router

