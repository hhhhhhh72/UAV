<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">个人信息</view>
    </view>

    <!-- 表单头部 -->
    <view class="pub-form-intro">
      <view class="pub-form-intro-h2">个人信息</view>
      <view class="pub-form-intro-p">完善个人资料，让行业伙伴更好地了解你</view>
    </view>

    <!-- 资料 -->
    <view class="pub-section">
      <view class="pub-section-title">资料</view>
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">头像</view>
          <view class="pub-upload-row pub-avatar-row">
            <view v-if="form.avatar" class="pub-photo" @tap="chooseAvatar">
              <image :src="avatarFull(form.avatar)" mode="aspectFill" class="pub-photo-img" />
            </view>
            <view v-else class="pub-photo pub-avatar-initial" @tap="chooseAvatar">
              <text>{{ initial }}</text>
            </view>
            <view class="pub-add-photo" hover-class="pub-fade" @tap="chooseAvatar">＋</view>
          </view>
          <view class="pub-upload-tip">点击头像可更换，选择后自动上传</view>
        </view>
        <view class="pub-field">
          <view class="pub-field-label">昵称<text class="pub-required">*</text></view>
          <input
            class="pub-input"
            v-model="form.name"
            placeholder="请输入昵称"
            placeholder-class="pub-placeholder"
            maxlength="20"
          />
        </view>
      </view>
    </view>

    <!-- 基础信息 -->
    <view class="pub-section">
      <view class="pub-section-title">基础信息</view>
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">手机号</view>
          <input
            class="pub-input"
            v-model="form.phone"
            type="number"
            maxlength="11"
            placeholder="请输入手机号"
            placeholder-class="pub-placeholder"
          />
        </view>

        <picker mode="selector" :range="genderOptions" @change="onGenderChange">
          <view class="pub-field" hover-class="pub-fade">
            <view class="pub-field-label">性别</view>
            <view class="pub-select-field">
              <text :class="form.gender ? 'pub-select-value' : 'pub-placeholder'">{{ form.gender || '请选择' }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </view>
        </picker>

        <picker mode="date" :start="'1950-01-01'" :end="todayStr" @change="onBirthdayChange">
          <view class="pub-field" hover-class="pub-fade">
            <view class="pub-field-label">生日</view>
            <view class="pub-select-field">
              <text :class="form.birthday ? 'pub-select-value' : 'pub-placeholder'">{{ form.birthday || '请选择' }}</text>
              <text class="pub-arrow">›</text>
            </view>
          </view>
        </picker>

        <view class="pub-field">
          <view class="pub-field-label">所在地区</view>
          <input
            class="pub-input"
            v-model="form.region"
            maxlength="30"
            placeholder="如：重庆市江北区"
            placeholder-class="pub-placeholder"
          />
        </view>
      </view>
    </view>

    <!-- 个人简介 -->
    <view class="pub-section">
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">个人简介</view>
          <textarea
            class="pub-input pub-input--textarea"
            v-model="form.bio"
            maxlength="120"
            placeholder="介绍一下自己，如从业经历、擅长领域（可选）"
            placeholder-class="pub-placeholder"
          />
          <text class="pub-field-hint">{{ (form.bio || '').length }}/120</text>
        </view>
      </view>
    </view>

    <!-- 账号 -->
    <view class="pub-section">
      <view class="pub-section-title">账号</view>
      <view class="pub-form-card">
        <view class="pub-field">
          <view class="pub-field-label">微信账号</view>
          <view class="pub-select-field">
            <text :class="wechatBound ? 'status-live' : 'status-draft'">{{ wechatBound ? '已绑定' : '未绑定' }}</text>
            <text class="pub-arrow">›</text>
          </view>
        </view>
        <view class="pub-field" hover-class="pub-fade" @tap="goAuth">
          <view class="pub-field-label">实名认证</view>
          <view class="pub-select-field">
            <text :class="form.isAuth ? 'status-live' : 'status-pending'">{{ form.isAuth ? '已认证' : '未认证' }}</text>
            <text class="pub-arrow">›</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 固定底部操作区（与发布页同款） -->
    <view class="pub-sticky">
      <view class="pub-btn pub-btn--primary" hover-class="pub-btn--active" @tap="handleSave">保存修改</view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onLoad, onUnload } from '@dcloudio/uni-app'
import { request, BASE_URL, authStorage } from '../../utils/request'
import { useSafeTop } from '../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

const form = ref({ name: '', avatar: '', isAuth: false, phone: '', gender: '', birthday: '', region: '', bio: '' })
const wechatBound = ref(false)
const genderOptions = ['男', '女']
let backTimer = null
const pad2 = (n) => String(n).padStart(2, '0')
const todayStr = `${new Date().getFullYear()}-${pad2(new Date().getMonth() + 1)}-${pad2(new Date().getDate())}`

const avatarFull = (u) => {
  if (!u) return ''
  return u.startsWith('http') ? u : BASE_URL + u
}

const initial = computed(() => {
  const c = (form.value.name || '我').charAt(0)
  return c.toUpperCase()
})

onLoad(() => {
  initSafeTop()
})

onMounted(async () => {
  const user = JSON.parse(uni.getStorageSync('user') || '{}')
  form.value = {
    name: user.name || '',
    avatar: user.avatar || user.avatar_url || '',
    isAuth: !!user.isAuth,
    phone: user.phone || '',
    gender: user.gender || '',
    birthday: user.birthday || '',
    region: user.region || '',
    bio: user.bio || '',
  }
  wechatBound.value = !!(user.has_wechat || user.wechat_openid || user.openid)
  // 从后端拉最新资料（手机号/性别/生日等可能在其他设备上改过）
  try {
    const res = await request({ url: '/api/v1/me', method: 'GET' })
    const d = res && res.data ? res.data : res
    if (d && typeof d === 'object') {
      form.value.phone = d.phone || form.value.phone
      form.value.gender = d.gender || form.value.gender
      form.value.birthday = d.birthday || form.value.birthday
      form.value.region = d.region || form.value.region
      form.value.bio = d.bio || form.value.bio
    }
  } catch (e) {
    // 拉取失败时沿用本地缓存
  }
})

const onGenderChange = (e) => {
  form.value.gender = genderOptions[Number(e.detail.value)] || ''
}

const onBirthdayChange = (e) => {
  form.value.birthday = e.detail.value
}

// 选择头像后立即上传 /api/v1/upload，成功后保存服务器 URL
const chooseAvatar = () => {
  uni.chooseImage({
    count: 1,
    success: (res) => {
      const tempPath = res.tempFilePaths[0]
      uni.showLoading({ title: '上传中...', mask: true })
      const token = authStorage.getAccessToken()
      uni.uploadFile({
        url: BASE_URL + '/api/v1/upload',
        filePath: tempPath,
        name: 'file',
        header: token ? { Authorization: `Bearer ${token}` } : {},
        success: (upRes) => {
          uni.hideLoading()
          if (upRes.statusCode >= 200 && upRes.statusCode < 300) {
            let body = null
            try {
              body = JSON.parse(upRes.data || '{}')
            } catch (e) {
              uni.showToast({ title: '上传失败', icon: 'none' })
              return
            }
            const url = body.data?.url || body.url
            if (url) {
              form.value.avatar = url
              uni.showToast({ title: '头像已上传' })
            } else {
              uni.showToast({ title: '上传失败', icon: 'none' })
            }
          } else {
            uni.showToast({ title: '上传失败', icon: 'none' })
          }
        },
        fail: () => {
          uni.hideLoading()
          uni.showToast({ title: '上传失败', icon: 'none' })
        },
      })
    },
  })
}

const goAuth = () => {
  uni.navigateTo({ url: '/pages/mine/auth' })
}

const goBack = () => {
  uni.navigateBack()
}

const handleSave = async () => {
  if (!form.value.name.trim()) return uni.showToast({ title: '请输入昵称', icon: 'none' })
  const phone = form.value.phone.trim()
  if (phone && !/^1[3-9]\d{9}$/.test(phone)) {
    return uni.showToast({ title: '手机号格式不正确', icon: 'none' })
  }

  uni.showLoading({ title: '保存中...', mask: true })
  try {
    // 昵称/头像/手机号/性别/生日/地区/简介保存到服务端
    await request({
      url: '/api/v1/me',
      method: 'PATCH',
      data: {
        name: form.value.name.trim(),
        avatar_url: form.value.avatar,
        phone,
        gender: form.value.gender,
        birthday: form.value.birthday,
        region: form.value.region.trim(),
        bio: form.value.bio.trim(),
      },
    })
    // 同步本地缓存
    const user = JSON.parse(uni.getStorageSync('user') || '{}')
    const updatedUser = {
      ...user,
      name: form.value.name.trim(),
      avatar: form.value.avatar,
      phone,
      gender: form.value.gender,
      birthday: form.value.birthday,
      region: form.value.region.trim(),
      bio: form.value.bio.trim(),
    }
    uni.setStorageSync('user', JSON.stringify(updatedUser))
    uni.hideLoading()
    uni.showToast({ title: '保存成功' })
    backTimer = setTimeout(() => uni.navigateBack(), 1200)
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '保存失败，请重试', icon: 'none' })
  }
}

onUnload(() => {
  if (backTimer) clearTimeout(backTimer)
})
</script>

<style scoped>
@import '../../pages/publish/pub-style.css';

.pub-fade { opacity: 0.6; }
.pub-form-intro-h2 {
  font-size: 20px;
  margin: 0 0 4px;
  color: #17212B;
}
.pub-form-intro-p {
  font-size: 12px;
  color: #667085;
  margin: 0;
  line-height: 1.5;
}
.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
}

/* 头像上传行：贴齐字段内边距 */
.pub-avatar-row { padding: 0; }

/* 无头像时的首字占位方块（替代原圆形首字头像） */
.pub-avatar-initial {
  display: flex;
  align-items: center;
  justify-content: center;
}
.pub-avatar-initial text {
  font-size: 22px;
  font-weight: 700;
  color: #0A66C2;
  line-height: 1;
}
</style>
