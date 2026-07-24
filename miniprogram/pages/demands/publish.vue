<template>
  <view class="publish-page">
    <van-nav-bar
      title="发布需求"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <van-cell-group inset>
      <van-form @submit="handleSubmit">
        <van-field
          v-model="form.title"
          name="title"
          label="标题"
          placeholder="请输入需求标题"
          :rules="[{ required: true, message: '请填写需求标题' }]"
        />

        <van-field
          v-model="bizTypeText"
          is-link
          readonly
          name="biz_type"
          label="业务类型"
          placeholder="请选择业务类型"
          :rules="[{ required: true, message: '请选择业务类型' }]"
          @tap="showBizTypePicker = true"
        />

        <van-field
          v-model="form.budget"
          name="budget"
          label="预算"
          placeholder="请输入预算金额"
          type="digit"
          suffix-icon="label"
        >
          <template #extra>
            <text class="unit">元</text>
          </template>
        </van-field>

        <van-field
          v-model="districtText"
          is-link
          readonly
          name="district"
          label="地区"
          placeholder="请选择重庆区县"
          :rules="[{ required: true, message: '请选择地区' }]"
          @tap="showDistrictPicker = true"
        />

        <van-field
          v-model="form.description"
          name="description"
          label="描述"
          type="textarea"
          placeholder="请详细描述需求内容"
          autosize
          maxlength="500"
          show-word-limit
        />

        <view class="submit-wrap">
          <van-button
            round
            block
            type="primary"
            native-type="submit"
            :loading="submitting"
          >
            发布需求
          </van-button>
        </view>
      </van-form>
    </van-cell-group>

    <!-- Biz type picker -->
    <van-popup
      :show="showBizTypePicker"
      position="bottom"
      round
      @close="showBizTypePicker = false"
    >
      <van-picker
        :columns="bizTypeOptions"
        @confirm="onBizTypeConfirm"
        @cancel="showBizTypePicker = false"
      />
    </van-popup>

    <!-- District picker -->
    <van-popup
      :show="showDistrictPicker"
      position="bottom"
      round
      @close="showDistrictPicker = false"
    >
      <van-picker
        :columns="districtOptions"
        @confirm="onDistrictConfirm"
        @cancel="showDistrictPicker = false"
      />
    </van-popup>
  </view>
</template>

<script>
import { request, authStorage } from '../../utils/request'

var BIZ_TYPE_MAP = {
  cable_inspection: '巡检',
  plant_transport: '植保',
  spray_pesticide: '农药',
  trade_lease: '租赁',
  clean_paint: '清洗',
  other: '其他',
}

var BIZ_TYPE_OPTIONS = Object.keys(BIZ_TYPE_MAP).map(function (k) {
  return { text: BIZ_TYPE_MAP[k], value: k }
})

var DISTRICT_OPTIONS = [
  '万州区', '涪陵区', '渝中区', '大渡口区', '江北区',
  '沙坪坝区', '九龙坡区', '南岸区', '北碚区', '綦江区',
  '大足区', '渝北区', '巴南区', '黔江区', '长寿区',
  '江津区', '合川区', '永川区', '南川区', '璧山区',
  '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区',
  '武隆区', '城口县', '丰都县', '垫江县', '忠县',
  '云阳县', '奉节县', '巫山县', '巫溪县', '石柱县',
  '秀山县', '酉阳县', '彭水县',
]

export default {
  data() {
    return {
      form: {
        title: '',
        biz_type: '',
        budget: '',
        district: '',
        description: '',
      },
      bizTypeText: '',
      districtText: '',
      showBizTypePicker: false,
      showDistrictPicker: false,
      submitting: false,
    }
  },
  onLoad() {
    var token = authStorage.getAccessToken()
    if (!token) {
      uni.showToast({ title: '请先登录', icon: 'none' })
      setTimeout(function () {
        uni.navigateTo({ url: '/pages/login/index' })
      }, 500)
    }
  },
  methods: {
    onBizTypeConfirm(e) {
      var selected = e.detail.value
      // map single-column picker value back to key
      for (var i = 0; i < BIZ_TYPE_OPTIONS.length; i++) {
        if (BIZ_TYPE_OPTIONS[i].text === selected) {
          this.form.biz_type = BIZ_TYPE_OPTIONS[i].value
          this.bizTypeText = selected
          break
        }
      }
      this.showBizTypePicker = false
    },
    onDistrictConfirm(e) {
      this.form.district = e.detail.value
      this.districtText = e.detail.value
      this.showDistrictPicker = false
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

      this.submitting = true
      uni.showLoading({ title: '发布中...', mask: true })

      try {
        await request({
          url: '/api/v1/demands',
          method: 'POST',
          data: {
            title: this.form.title,
            biz_type: this.form.biz_type,
            budget: this.form.budget ? parseFloat(this.form.budget) : 0,
            district: this.form.district,
            description: this.form.description,
          },
        })
        uni.hideLoading()
        uni.showToast({ title: '发布成功', icon: 'success' })
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
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.publish-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 40px;
}

.unit {
  font-size: 14px;
  color: #969799;
}

.submit-wrap {
  padding: 24px 16px 16px;
}
</style>
