<template>
  <view class="publish-page">
    <u-nav-bar title="发布需求" show-back @back="goBack" />

    <u-cell-group inset>
      <view class="form-wrap">
        <u-field
          v-model="form.title"
          label="标题"
          placeholder="请输入需求标题"
        />

        <view class="field-row" @tap="showBizTypePicker = true">
          <u-field
            v-model="bizTypeText"
            label="业务类型"
            placeholder="请选择业务类型"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <view class="field-row">
          <u-field
            v-model="form.budget"
            label="预算"
            placeholder="请输入预算金额"
            type="digit"
          />
          <text class="unit">元</text>
        </view>

        <view class="field-row" @tap="showDistrictPicker = true">
          <u-field
            v-model="districtText"
            label="地区"
            placeholder="请选择重庆区县"
            disabled
          />
          <text class="field-arrow">›</text>
        </view>

        <u-field
          v-model="form.description"
          label="描述"
          type="textarea"
          placeholder="请详细描述需求内容"
          auto-height
        />

        <view class="submit-wrap">
          <u-button
            round
            block
            type="primary"
            :loading="submitting"
            @click="handleSubmit"
          >
            发布需求
          </u-button>
        </view>
      </view>
    </u-cell-group>

    <!-- Biz type picker -->
    <u-picker
      :show="showBizTypePicker"
      title="请选择业务类型"
      :columns="bizTypeNames"
      @confirm="onBizTypeConfirm"
      @update:show="showBizTypePicker = $event"
    />

    <!-- District picker -->
    <u-picker
      :show="showDistrictPicker"
      title="请选择重庆区县"
      :columns="districtOptions"
      @confirm="onDistrictConfirm"
      @update:show="showDistrictPicker = $event"
    />
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
      // u-picker 的 columns 只接受字符串数组，业务类型选项先映射为名称列表
      bizTypeNames: BIZ_TYPE_OPTIONS.map(function (o) { return o.text }),
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
    onBizTypeConfirm(selected) {
      // u-picker confirm 直接回传选中的字符串（单列）
      for (var i = 0; i < BIZ_TYPE_OPTIONS.length; i++) {
        if (BIZ_TYPE_OPTIONS[i].text === selected) {
          this.form.biz_type = BIZ_TYPE_OPTIONS[i].value
          this.bizTypeText = selected
          break
        }
      }
      this.showBizTypePicker = false
    },
    onDistrictConfirm(selected) {
      this.form.district = selected
      this.districtText = selected
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
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.publish-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: 40px;
}

.form-wrap {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.field-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.field-row .u-field {
  flex: 1;
}

.field-arrow {
  font-size: 20px;
  color: var(--color-text-placeholder);
}

.unit {
  font-size: 14px;
  color: var(--color-text-secondary);
}

.submit-wrap {
  padding: 8px 0 0;
}
</style>
