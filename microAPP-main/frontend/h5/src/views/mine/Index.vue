<template>
  <div class="mine-page page-container">
    <van-nav-bar title="个人中心" fixed placeholder />

    <!-- 用户信息卡片 -->
    <div class="user-card">
      <div class="user-header" @click="handleUserClick">
        <van-image
          v-if="user?.avatar"
          round
          width="60"
          height="60"
          :src="user.avatar"
        />
        <div v-else class="default-avatar">
          <van-icon name="contact" size="36" color="#bdc3c7" />
        </div>
        <div class="user-info">
          <h2 class="user-name">{{ user?.name || '点击登录' }}</h2>
          <p class="user-phone">{{ user?.phone || '登录后享受更多服务' }}</p>
        </div>
        <van-icon name="arrow" size="16" color="#8e8e93" v-if="!user" />
      </div>

      <!-- 统计数据 -->
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-value">{{ totalCount }}</div>
          <div class="stat-label">总申请</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ processingCount }}</div>
          <div class="stat-label">处理中</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ completedCount }}</div>
          <div class="stat-label">已完成</div>
        </div>
      </div>
    </div>

    <!-- 功能菜单 -->
    <div class="menu-section">
      <van-cell-group inset>
        <van-cell
          v-if="user && (user.role === 'admin' || user.role === 'dsl_admin' || user.role === 'study_admin')"
          title="后台管理"
          icon="setting-o"
          is-link
          @click="$router.push('/admin')"
        />
        <van-cell
          title="我的申请"
          icon="records"
          is-link
          @click="$router.push('/applications')"
        />
        <van-cell
          title="案例展示"
          icon="video-o"
          is-link
          @click="$router.push('/cases')"
        />
        <van-cell
          title="个人信息"
          icon="contact"
          is-link
          @click="showToast('功能开发中')"
        />
        <van-cell
          title="实名认证"
          icon="shield-o"
          is-link
          label="未认证"
          @click="showToast('功能开发中')"
        />
      </van-cell-group>
    </div>

    <div class="menu-section">
      <van-cell-group inset>
        <van-cell
          title="服务指南"
          icon="guide-o"
          is-link
          @click="showGuide"
        />
        <van-cell
          title="常见问题"
          icon="question-o"
          is-link
          @click="showFAQ"
        />
        <van-cell
          title="联系客服"
          icon="service-o"
          is-link
          @click="showContact"
        />
        <van-cell
          title="关于我们"
          icon="info-o"
          is-link
          @click="showAbout"
        />
      </van-cell-group>
    </div>

    <div class="menu-section" v-if="user">
      <van-cell-group inset>
        <van-cell
          title="退出登录"
          icon="close"
          is-link
          @click="handleLogout"
        />
      </van-cell-group>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showDialog, showToast, showFailToast, showConfirmDialog } from 'vant'
import axios, { authStorage } from '@/utils/http'
import { useRouter } from 'vue-router'

const router = useRouter()
const totalCount = ref(0)
const processingCount = ref(0)
const completedCount = ref(0)
const user = ref(null)
const SERVICE_PHONE = '0577-55558188'

const fetchStats = async (userId) => {
  if (!userId) {
      totalCount.value = 0;
      processingCount.value = 0;
      completedCount.value = 0;
      return;
  }
  try {
    const res = await axios.get('/api/list', { params: { userId } })
    const list = res.data || []
    
    totalCount.value = list.length
    processingCount.value = list.filter(item => item.status === '处理中').length
    completedCount.value = list.filter(item => item.status === '已完成').length
  } catch (error) {
    console.error('Failed to fetch stats:', error)
    // showFailToast('获取数据失败') // Optional: suppress error toast on mine page to avoid annoyance
  }
}

const handleUserClick = () => {
  if (!user.value) {
    router.push('/login')
  }
}

const handleLogout = () => {
  showConfirmDialog({
    title: '提示',
    message: '确定要退出登录吗？',
  })
    .then(() => {
      axios.post('/api/auth/logout').catch(() => {})
      localStorage.removeItem('user')
      authStorage.clearTokens()
      user.value = null
      fetchStats(null) // Reset stats
      showToast('已退出登录')
    })
    .catch(() => {
      // on cancel
    })
}

onMounted(() => {
  const accessToken = authStorage.getAccessToken()
  const userStr = localStorage.getItem('user')
  if (accessToken) {
    axios
      .get('/api/auth/me')
      .then((res) => {
        if (res.data?.success) {
          user.value = res.data.user
          localStorage.setItem('user', JSON.stringify(res.data.user))
          fetchStats(res.data.user.id)
          return
        }
        fetchStats(null)
      })
      .catch(() => {
        authStorage.clearTokens()
        localStorage.removeItem('user')
        fetchStats(null)
      })
    return
  }

  if (userStr) {
    try {
      user.value = JSON.parse(userStr)
      fetchStats(user.value.id)
    } catch (e) {
      console.error(e)
      fetchStats(null)
    }
  } else {
    fetchStats(null)
  }
})

const showGuide = () => {
  showDialog({
    title: '服务指南',
    message: '1. 选择所需服务\n2. 填写申请表单\n3. 提交申请\n4. 等待客服联系\n5. 确认服务详情\n6. 服务执行'
  })
}

const showFAQ = () => {
  showDialog({
    title: '常见问题',
    message: 'Q: 申请后多久联系？\nA: 2小时内会有客服联系您\n\nQ: 如何修改申请？\nA: 请联系客服进行修改'
  })
}

const showContact = () => {
  showConfirmDialog({
    title: '联系客服',
    message: `客服电话：${SERVICE_PHONE}\n工作时间：工作日 9:00-18:00`,
    confirmButtonText: '拨打电话',
    cancelButtonText: '取消'
  })
    .then(() => {
      window.location.href = `tel:${SERVICE_PHONE}`
    })
    .catch(() => {
      // on cancel
    })
}

const showAbout = () => {
  showDialog({
    title: '关于我们',
    message: '低空综合服务平台\n开发主体：温州低空经济发展有限公司\n版本：v1.1.0\n\n专注于提供专业、高效、安全的低空服务'
  })
}
</script>

<style scoped>
.mine-page {
  background: #f5f5f7;
}

.mine-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

.user-card {
  background: #ffffff;
  margin: 12px;
  padding: 16px;
  border-radius: 18px;
  color: #1d1d1f;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}

.default-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: #f2f2f7;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(0, 0, 0, 0.06);
}

.user-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  cursor: pointer;
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 6px;
}

.user-phone {
  font-size: 14px;
  color: #86868b;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 13px;
  color: #86868b;
}

.menu-section {
  margin: 12px 0;
}
</style>

