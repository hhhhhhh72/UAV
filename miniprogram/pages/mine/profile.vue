<template>
  <view class="profile-page">
    <u-nav-bar title="个人信息" show-back @back="goBack" />

    <!-- ═══════ 资料 ═══════ -->
    <view class="card">
      <view class="row" hover-class="row-fade" @tap="chooseAvatar">
        <view class="row-label-wrap">
          <text class="row-label">头像</text>
        </view>
        <view class="row-right">
          <image v-if="form.avatar" :src="avatarFull(form.avatar)" mode="aspectFill" class="row-avatar" />
          <view v-else class="row-avatar row-avatar-ph">
            <text>{{ initial }}</text>
          </view>
          <text class="row-chev">›</text>
        </view>
      </view>

      <view class="row">
        <view class="row-label-wrap">
          <text class="row-label">昵称</text>
        </view>
        <view class="row-right row-right--input">
          <input
            class="row-input"
            v-model="form.name"
            placeholder="请输入昵称"
            placeholder-class="row-input-ph"
            maxlength="20"
          />
        </view>
      </view>
    </view>

    <!-- ═══════ 基础信息 ═══════ -->
    <view class="card">
      <view class="row">
        <view class="row-label-wrap">
          <text class="row-label">手机号</text>
        </view>
        <view class="row-right row-right--input">
          <input
            class="row-input"
            v-model="form.phone"
            type="number"
            maxlength="11"
            placeholder="请输入手机号"
            placeholder-class="row-input-ph"
          />
        </view>
      </view>

      <picker mode="selector" :range="genderOptions" @change="onGenderChange">
        <view class="row" hover-class="row-fade">
          <view class="row-label-wrap">
            <text class="row-label">性别</text>
          </view>
          <view class="row-right">
            <text class="row-tail" :class="form.gender ? 'val' : 'dim'">{{ form.gender || '请选择' }}</text>
            <text class="row-chev">›</text>
          </view>
        </view>
      </picker>

      <picker mode="date" :start="'1950-01-01'" :end="todayStr" @change="onBirthdayChange">
        <view class="row" hover-class="row-fade">
          <view class="row-label-wrap">
            <text class="row-label">生日</text>
          </view>
          <view class="row-right">
            <text class="row-tail" :class="form.birthday ? 'val' : 'dim'">{{ form.birthday || '请选择' }}</text>
            <text class="row-chev">›</text>
          </view>
        </view>
      </picker>

      <view class="row">
        <view class="row-label-wrap">
          <text class="row-label">所在地区</text>
        </view>
        <view class="row-right row-right--input">
          <input
            class="row-input"
            v-model="form.region"
            maxlength="30"
            placeholder="如：重庆市江北区"
            placeholder-class="row-input-ph"
          />
        </view>
      </view>
    </view>

    <!-- ═══════ 个人简介 ═══════ -->
    <view class="card">
      <view class="row row--textarea">
        <view class="row-label-wrap">
          <text class="row-label">个人简介</text>
        </view>
        <textarea
          class="row-textarea"
          v-model="form.bio"
          maxlength="120"
          placeholder="介绍一下自己，如从业经历、擅长领域（可选）"
          placeholder-class="row-input-ph"
        />
      </view>
    </view>

    <!-- ═══════ 账号 ═══════ -->
    <view class="card">
      <view class="row">
        <view class="row-label-wrap">
          <text class="row-label">微信账号</text>
        </view>
        <view class="row-right">
          <text class="row-tail" :class="wechatBound ? 'ok' : 'dim'">{{ wechatBound ? '已绑定' : '未绑定' }}</text>
          <text class="row-chev">›</text>
        </view>
      </view>

      <view class="row" hover-class="row-fade" @tap="goAuth">
        <view class="row-label-wrap">
          <text class="row-label">实名认证</text>
        </view>
        <view class="row-right">
          <text class="row-tail" :class="form.isAuth ? 'ok' : 'wait'">{{ form.isAuth ? '已认证' : '未认证' }}</text>
          <text class="row-chev">›</text>
        </view>
      </view>
    </view>

    <!-- ═══════ 保存 ═══════ -->
    <view class="save-wrap">
      <view class="save-btn" hover-class="save-btn-hover" @tap="handleSave">
        <text class="save-btn-text">保存修改</text>
      </view>
    </view>

    <view class="bottom-spacer"></view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { request, BASE_URL, authStorage } from '../../utils/request'

const form = ref({ name: '', avatar: '', isAuth: false, phone: '', gender: '', birthday: '', region: '', bio: '' })
const wechatBound = ref(false)
const genderOptions = ['男', '女']
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
            const body = JSON.parse(upRes.data || '{}')
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
    setTimeout(() => uni.navigateBack(), 1200)
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '保存失败，请重试', icon: 'none' })
  }
}
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
}

/* ===== 卡片 ===== */
.card {
  background: #fff;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  box-shadow: 0 8rpx 32rpx rgba(16, 24, 40, 0.06);
  margin: 24rpx 24rpx 0;
  overflow: hidden;
}

/* ===== 行 ===== */
.row {
  display: flex;
  align-items: center;
  min-height: 104rpx;
  padding: 0 28rpx;
  border-top: 1rpx solid #EEF1F4;
  box-sizing: border-box;
}
.row:first-child {
  border-top: none;
}
.row-label-wrap {
  min-width: 0;
  margin-right: 24rpx;
}
.row-label {
  font-size: 26rpx;
  font-weight: 600;
  color: #17212B;
}
.row-right {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12rpx;
}
.row-right--input {
  height: 100%;
}
.row-input {
  flex: 1;
  min-width: 0;
  text-align: right;
  font-size: 28rpx;
  color: #344054;
  padding: 8rpx 0;
}
.row-input-ph {
  color: #98A2B3;
}

/* 头像 */
.row-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  flex-shrink: 0;
  display: block;
  box-sizing: border-box;
  background: #F0F2F5;
}
.row-avatar-ph {
  background: linear-gradient(145deg, #3A8BDD, #0B579F);
  border: 2rpx solid rgba(10, 102, 194, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
}
.row-avatar-ph text {
  font-size: 40rpx;
  font-weight: 700;
  color: #DDECFF;
}

/* 状态文字 */
.row-tail {
  flex-shrink: 0;
  font-size: 24rpx;
  color: #98A2B3;
}
.row-tail.ok { color: #168A55; }
.row-tail.wait { color: #B54708; }
.row-tail.dim { color: #C0C8D2; }
.row-tail.val { color: #344054; }

/* 个人简介（多行） */
.row--textarea {
  flex-direction: column;
  align-items: stretch;
  min-height: 0;
  padding: 24rpx 28rpx;
}
.row--textarea .row-label-wrap {
  margin-right: 0;
  margin-bottom: 12rpx;
}
.row-textarea {
  width: 100%;
  min-height: 132rpx;
  font-size: 26rpx;
  line-height: 1.6;
  color: #344054;
}

.row-chev {
  flex-shrink: 0;
  color: #98A2B3;
  font-size: 30rpx;
  font-weight: 300;
  margin-left: 4rpx;
}

.row-fade { opacity: 0.8; }

/* ===== 保存按钮（低圆角主按钮） ===== */
.save-wrap {
  margin: 40rpx 24rpx 0;
}
.save-btn {
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12rpx;
  background: var(--color-primary);
  box-shadow: 0 6rpx 16rpx rgba(10, 102, 194, 0.22);
}
.save-btn-text {
  font-size: 30rpx;
  font-weight: 700;
  color: #fff;
}
.save-btn-hover { opacity: 0.85; }

.bottom-spacer { height: 24rpx; }
</style>
