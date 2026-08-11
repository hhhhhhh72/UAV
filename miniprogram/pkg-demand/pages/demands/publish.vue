<template>
  <view class="publish-page">
    <u-nav-bar title="发布需求" show-back @back="goBack" />

    <!-- 基本信息 -->
    <view class="section-title">基本信息</view>
    <u-cell-group inset>
      <u-field
        v-model="form.title"
        label="需求标题"
        placeholder="一句话说清需求，如：光伏电站红外巡检"
      />
      <view class="field-row" @tap="showDistrictPicker = true">
        <u-field
          :model-value="districtText"
          label="所在地区"
          placeholder="请选择重庆区县"
          disabled
        />
        <text class="field-arrow">›</text>
      </view>
    </u-cell-group>

    <!-- 业务类型（chips 单选，与业务术语标准一致） -->
    <view class="section-title">业务类型</view>
    <view class="chips-card">
      <view
        v-for="opt in bizTypeOptions"
        :key="opt.value"
        class="chip"
        :class="{ 'chip--active': form.biz_type === opt.value }"
        @tap="pickBizType(opt.value)"
      >
        {{ opt.text }}
      </view>
    </view>

    <!-- 预算与联系 -->
    <view class="section-title">预算与联系</view>
    <u-cell-group inset>
      <view class="field-row">
        <u-field
          v-model="form.budget"
          label="预算"
          placeholder="项目预算金额（选填）"
          type="digit"
        />
        <text class="unit">元</text>
      </view>
      <view class="field-row">
        <u-field
          v-model="form.contact"
          label="联系电话"
          placeholder="对接人电话，公开可见"
          type="number"
        />
        <text class="required">*</text>
      </view>
    </u-cell-group>

    <!-- 需求详情 -->
    <view class="section-title">需求详情</view>
    <u-cell-group inset>
      <u-field
        v-model="form.description"
        label="描述"
        type="textarea"
        auto-height
        placeholder="请详细描述需求内容、交付要求、作业工期等"
      />
    </u-cell-group>

    <!-- 图片上传（选填，最多 9 张） -->
    <view class="section-title">现场图 / 资料图（选填）</view>
    <view class="upload-card">
      <view class="upload-grid">
        <view
          v-for="(img, i) in images"
          :key="i"
          class="upload-item"
          @tap="previewImage(i)"
        >
          <image :src="img" mode="aspectFill" class="upload-thumb" />
          <view class="upload-remove" @tap.stop="removeImage(i)">
            <u-icon name="close" size="24rpx" color="var(--color-danger)" />
          </view>
        </view>
        <view v-if="images.length < 9" class="upload-add" @tap="chooseImages">
          <u-icon name="plus" size="32rpx" color="var(--color-text-secondary)" />
          <text class="upload-hint">{{ images.length }}/9</text>
        </view>
      </view>
      <text class="upload-tip">支持从相册选择或拍照，将自动裁剪为 16:9 比例，审核通过后在需求大厅公开展示</text>
    </view>

    <!-- 底部操作栏 -->
    <view class="action-bar">
      <u-button
        type="primary"
        round
        block
        :loading="submitting"
        @click="handleSubmit"
      >
        发布需求
      </u-button>
    </view>

    <!-- 区县选择器 -->
    <u-picker
      :show="showDistrictPicker"
      title="请选择重庆区县"
      :columns="districtOptions"
      @confirm="onDistrictConfirm"
      @update:show="showDistrictPicker = $event"
    />

    <!-- 图片裁剪（16:9 统一比例，保证首页/大厅卡片整齐） -->
    <crop-image
      :visible="showCrop"
      :src="cropSrc"
      @confirm="onCropConfirm"
      @cancel="onCropCancel"
    />
  </view>
</template>

<script>
import { request, authStorage, getStoredUser, BASE_URL } from '../../../utils/request'

// 业务类型与 utils/enums.js BIZ_TYPE_LABEL 保持一致（对应后端 biz_standard.go）
var BIZ_TYPE_OPTIONS = [
  { text: '巡检', value: 'cable_inspection' },
  { text: '植保', value: 'plant_transport' },
  { text: '农药', value: 'spray_pesticide' },
  { text: '租赁', value: 'trade_lease' },
  { text: '清洗', value: 'clean_paint' },
  { text: '其他', value: 'other' },
]

var DISTRICT_OPTIONS = [
  '渝中区', '江北区', '南岸区', '渝北区', '沙坪坝区', '九龙坡区', '大渡口区', '北碚区', '巴南区',
  '两江新区', '高新区', '万州区', '涪陵区', '黔江区', '长寿区', '江津区', '合川区', '永川区',
  '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区',
  '武隆区', '城口县', '丰都县', '垫江县', '忠县', '云阳县', '奉节县', '巫山县', '巫溪县',
  '石柱县', '秀山县', '酉阳县', '彭水县',
]

var MAX_IMAGES = 9

export default {
  data() {
    return {
      form: {
        title: '',
        biz_type: '',
        budget: '',
        district: '',
        contact: '',
        description: '',
      },
      bizTypeOptions: BIZ_TYPE_OPTIONS,
      districtOptions: DISTRICT_OPTIONS,
      districtText: '',
      // 预览态图片数组：完整 URL（BASE_URL + 相对路径），提交时转回相对路径
      images: [],
      showDistrictPicker: false,
      submitting: false,
      // 裁剪流程状态：选图后逐张裁剪（16:9）再上传
      showCrop: false,
      cropSrc: '',
      cropQueue: [],
    }
  },
  onLoad() {
    var token = authStorage.getAccessToken()
    if (!token) {
      uni.showToast({ title: '请先登录', icon: 'none' })
      setTimeout(function () {
        uni.navigateTo({ url: '/pages/login/index' })
      }, 500)
      return
    }
    // 预填联系电话（微信登录用户可能无手机号，留空需手动填写）
    var u = getStoredUser()
    if (u && u.phone) {
      this.form.contact = u.phone
    }
  },
  methods: {
    goBack() {
      uni.navigateBack()
    },
    pickBizType(value) {
      this.form.biz_type = value
    },
    onDistrictConfirm(selected) {
      this.form.district = selected
      this.districtText = selected
      this.showDistrictPicker = false
    },
    // 相对路径（存库格式）→ 完整 URL（预览格式）
    resolveUrl(u) {
      if (!u) return ''
      if (u.indexOf('http') === 0) return u
      return BASE_URL + u
    },
    // 完整 URL（预览格式）→ 相对路径（存库格式）
    storageUrl(u) {
      if (!u) return ''
      if (u.indexOf(BASE_URL) === 0) return u.slice(BASE_URL.length)
      return u
    },
    // ---- 多图上传（选图 → 逐张 16:9 裁剪 → 上传） ----
    chooseImages() {
      var self = this
      var remaining = MAX_IMAGES - self.images.length
      uni.chooseImage({
        count: remaining,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: function (res) {
          var paths = res.tempFilePaths || []
          if (!paths.length) return
          self.cropQueue = paths.slice()
          self.openNextCrop()
        },
        fail: function () {
          uni.showToast({ title: '选择图片失败', icon: 'none' })
        },
      })
    },
    // 打开下一张图的裁剪弹层
    openNextCrop() {
      var self = this
      if (!self.cropQueue.length) return
      self.cropSrc = self.cropQueue.shift()
      self.showCrop = true
    },
    // 裁剪完成：上传裁剪结果，继续下一张
    onCropConfirm(path) {
      var self = this
      self.showCrop = false
      if (!path) {
        self.openNextCrop()
        return
      }
      uni.showLoading({ title: '上传中...' })
      self.uploadOne(path, function () {
        uni.hideLoading()
        self.openNextCrop()
      })
    },
    onCropCancel() {
      this.showCrop = false
      this.cropQueue = []
    },
    // 单张上传（注意：预览必须用完整 URL，小程序 image 的 src 若为 /uploads/xxx
    // 会被当作本地包内资源 → 白图，与上传企业 logo 的问题同源）
    uploadOne(filePath, onDone) {
      var self = this
      uni.uploadFile({
        url: BASE_URL + '/api/v1/files/upload',
        filePath: filePath,
        name: 'file',
        header: { Authorization: 'Bearer ' + authStorage.getAccessToken() },
        success: function (r) {
          var data = null
          try { data = JSON.parse(r.data) } catch (e) {}
          if (r.statusCode >= 400 || !data || (!data.file_id && !(data.data && data.data.file_id))) {
            var msg = ''
            if (data && data.error) msg = data.error.message || data.error.code || ''
            if (data && data.message) msg = data.message
            var reason = msg || ('HTTP ' + r.statusCode)
            var tip = reason.indexOf('401') >= 0 || reason.indexOf('登录') >= 0 || reason.indexOf('token') >= 0
              ? '登录已过期，请重新登录后重试'
              : ('上传失败：' + reason)
            uni.showToast({ title: tip, icon: 'none', duration: 2500 })
            if (onDone) onDone()
            return
          }
          var fid = data.file_id || (data.data && data.data.file_id)
          if (fid) {
            self.images.push(self.resolveUrl('/uploads/' + fid))
          }
          if (onDone) onDone()
        },
        fail: function () {
          uni.showToast({ title: '上传失败，请重试', icon: 'none' })
          if (onDone) onDone()
        },
      })
    },
    previewImage(i) {
      uni.previewImage({ urls: this.images, current: this.images[i] })
    },
    removeImage(i) {
      this.images.splice(i, 1)
    },
    async handleSubmit() {
      var token = authStorage.getAccessToken()
      if (!token) {
        uni.showToast({ title: '请先登录', icon: 'none' })
        uni.navigateTo({ url: '/pages/login/index' })
        return
      }

      if (!this.form.title) {
        uni.showToast({ title: '请填写需求标题', icon: 'none' })
        return
      }
      if (!this.form.biz_type) {
        uni.showToast({ title: '请选择业务类型', icon: 'none' })
        return
      }
      if (!this.form.district) {
        uni.showToast({ title: '请选择地区', icon: 'none' })
        return
      }
      if (!this.form.contact) {
        uni.showToast({ title: '请填写联系电话', icon: 'none' })
        return
      }

      this.submitting = true
      uni.showLoading({ title: '发布中...', mask: true })

      try {
        // 公告模式联系方式必填：表单字段（预填手机号，可编辑）
        var currentUser = getStoredUser()
        await request({
          url: '/api/v1/demands',
          method: 'POST',
          data: {
            title: this.form.title,
            biz_type: this.form.biz_type,
            budget_fen: this.form.budget ? Math.round(parseFloat(this.form.budget) * 100) : 0,
            district: this.form.district,
            description: this.form.description,
            contact: this.form.contact,
            publisher_name: (currentUser && currentUser.name) || '',
            images: this.images.map(function (u) { return this.storageUrl(u) }.bind(this)),
          },
        })
        uni.hideLoading()
        uni.showToast({ title: '已提交审核，请等待管理员审核', icon: 'none' })
        setTimeout(function () {
          uni.navigateBack()
        }, 500)
      } catch (e) {
        uni.hideLoading()
        uni.showToast({ title: '发布失败，请稍后重试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
  },
}
</script>

<style scoped>
.publish-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: calc(120rpx + env(safe-area-inset-bottom));
}

/* 区块标题（与入驻页一致） */
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  padding: 8px 20px 4px;
}

/* 带箭头 / 单位 / 必填 的行 */
.field-row {
  display: flex;
  align-items: center;
  background: #fff;
  padding-right: 16px;
}

.field-row .u-field {
  flex: 1;
}

.field-arrow {
  font-size: 20px;
  color: var(--color-text-placeholder);
  flex-shrink: 0;
}

.unit {
  font-size: 14px;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.required {
  color: var(--color-danger);
  font-size: 28rpx;
  flex-shrink: 0;
}

/* 业务类型 chips */
.chips-card {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px;
  background: var(--color-bg-card);
  border-radius: 12px;
  margin: 0 12px 12px;
}

.chip {
  padding: 10px 18px;
  border-radius: 24rpx;
  background: var(--color-bg);
  border: 1px solid var(--color-divider);
  font-size: 13px;
  color: var(--color-text-secondary);
  transition: all 0.2s;
}

.chip--active {
  background: rgba(10, 102, 194, 0.08);
  border-color: var(--color-primary);
  color: var(--color-primary);
  font-weight: 500;
}

/* 图片上传 */
.upload-card {
  margin: 0 12px 12px;
  background: var(--color-bg-card);
  border-radius: 12px;
  padding: 16px;
}

.upload-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.upload-item {
  position: relative;
  width: 150rpx;
  height: 150rpx;
  border-radius: 8px;
  overflow: hidden;
}

.upload-thumb {
  width: 100%;
  height: 100%;
}

.upload-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  padding: 2px;
}

.upload-add {
  width: 150rpx;
  height: 150rpx;
  border: 1px dashed var(--color-text-placeholder);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.upload-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.upload-tip {
  display: block;
  margin-top: 12px;
  font-size: 12px;
  color: var(--color-text-placeholder);
}

/* 底部操作栏 */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px 16px;
  background: var(--color-bg-card);
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  z-index: 100;
}
</style>
