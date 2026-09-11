<template>
  <div class="page">
    <a-card :bordered="false" class="card">
      <template #title>账号信息</template>
      <a-descriptions :column="2" size="medium">
        <a-descriptions-item label="账号">{{ me.id || '-' }}</a-descriptions-item>
        <a-descriptions-item label="角色">{{ roleLabel }}</a-descriptions-item>
        <a-descriptions-item label="昵称">{{ form.name || '-' }}</a-descriptions-item>
        <a-descriptions-item label="手机号">{{ maskedPhone }}</a-descriptions-item>
      </a-descriptions>
    </a-card>

    <a-card :bordered="false" class="card">
      <template #title>资料</template>
      <a-form :model="form" layout="vertical" class="form">
        <a-form-item label="昵称">
          <a-input v-model="form.name" :max-length="20" placeholder="用于后台显示的名字" />
        </a-form-item>
        <a-form-item label="头像地址">
          <a-input v-model="form.avatar_url" placeholder="https://…（可留空）" />
        </a-form-item>
        <a-button type="primary" :loading="savingProfile" @click="saveProfile">保存资料</a-button>
      </a-form>
    </a-card>

    <a-card :bordered="false" class="card">
      <template #title>修改密码</template>
      <a-alert v-if="!hasPassword" type="warning" class="alert">
        该账号还没设置密码（只能用微信/验证码登录），请联系平台管理员重置后再来修改。
      </a-alert>
      <a-form :model="pwd" layout="vertical" class="form">
        <a-form-item label="原密码">
          <a-input-password v-model="pwd.old" placeholder="当前密码" />
        </a-form-item>
        <a-form-item label="新密码">
          <a-input-password v-model="pwd.next" placeholder="至少 8 位" />
        </a-form-item>
        <a-form-item label="确认新密码">
          <a-input-password v-model="pwd.confirm" placeholder="再输入一次" />
        </a-form-item>
        <a-button type="primary" status="warning" :loading="savingPwd" @click="changePassword">修改密码</a-button>
        <span class="tip">改完密码后所有设备都需要重新登录</span>
      </a-form>
    </a-card>

    <a-card :bordered="false" class="card">
      <template #title>我的权限</template>
      <a-descriptions :column="1" size="medium">
        <a-descriptions-item label="当前角色">{{ roleLabel }}</a-descriptions-item>
        <a-descriptions-item label="可管理的模块">
          <span v-if="isPlatformAdmin">全部模块（含用户管理、系统配置、操作审计、数据导出）</span>
          <span v-else>业务运营模块（用户管理、系统配置、操作审计、数据导出仅平台管理员可用）</span>
        </a-descriptions-item>
      </a-descriptions>
      <div class="tip">按模块的细粒度授权（能看 / 能改 / 能审 / 能删）正在开发，上线后这里会列出你的具体权限。</div>
    </a-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Message from '@arco-design/web-vue/es/message'
import '@arco-design/web-vue/es/message/style/css'
import axios, { authStorage } from '@/utils/http'
import { useAuth } from '../composables/useAuth'

const router = useRouter()
const { isPlatformAdmin } = useAuth()

const me = ref({})
const form = reactive({ name: '', avatar_url: '' })
const pwd = reactive({ old: '', next: '', confirm: '' })
const savingProfile = ref(false)
const savingPwd = ref(false)

const ROLE_LABEL = { platform_admin: '终端管理员', association_admin: '协会管理员', enterprise: '企业用户', individual: '个人用户' }
const roleLabel = computed(() => ROLE_LABEL[me.value.role] || me.value.role || '-')
const hasPassword = ref(true)

// 手机号脱敏：只留前 3 后 4
const maskedPhone = computed(() => {
  const p = me.value.phone || ''
  if (p.length < 7) return p || '未绑定'
  return p.slice(0, 3) + '****' + p.slice(-4)
})

const errMsg = (e) => e?.response?.data?.error?.message || e?.response?.data?.message || e?.message || '操作失败'

const load = async () => {
  try {
    const res = await axios.get('/api/v1/me')
    const data = res.data?.data || res.data || {}
    me.value = data
    form.name = data.name || ''
    form.avatar_url = data.avatar_url || ''
    hasPassword.value = data.has_password !== false // 后端 /me 直接返回是否设置过密码
  } catch (e) {
    Message.error(errMsg(e))
  }
}

const saveProfile = async () => {
  savingProfile.value = true
  try {
    await axios.patch('/api/v1/me', { name: form.name, avatar_url: form.avatar_url })
    Message.success('资料已保存')
    const cached = JSON.parse(localStorage.getItem('user') || '{}')
    localStorage.setItem('user', JSON.stringify({ ...cached, name: form.name, avatar_url: form.avatar_url }))
  } catch (e) {
    Message.error(errMsg(e))
  } finally {
    savingProfile.value = false
  }
}

const changePassword = async () => {
  if (!pwd.old || !pwd.next) { Message.warning('请填写原密码和新密码'); return }
  if (pwd.next !== pwd.confirm) { Message.warning('两次输入的新密码不一致'); return }
  savingPwd.value = true
  try {
    await axios.post('/api/v1/auth/password', { old_password: pwd.old, new_password: pwd.next })
    pwd.old = pwd.next = pwd.confirm = ''
    Message.success('密码已修改，请用新密码重新登录')
    authStorage.clearTokens()
    setTimeout(() => router.push('/login'), 800)
  } catch (e) {
    Message.error(errMsg(e))
  } finally {
    savingPwd.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.page { max-width: 900px; margin: 0 auto; display: flex; flex-direction: column; gap: 16px; }
.card { border-radius: 8px; }
.form { max-width: 420px; }
.alert { margin-bottom: 12px; }
.tip { margin-left: 12px; color: var(--color-text-3); font-size: 12px; }
</style>
