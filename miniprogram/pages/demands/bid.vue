<template>
  <view class="bid-page">
    <u-nav-bar
      title="竞标报价"
      show-back
      @back="goBack"
    />

    <!-- Loading -->
    <view v-if="loading" class="loading-state">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchDemandSummary">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Normal -->
    <template v-else-if="demandSummary">
      <!-- Demand summary -->
      <view class="summary-card">
        <text class="summary-title">{{ demandSummary.title }}</text>
        <view class="summary-meta">
          <u-tag :type="bizTypeTagType(demandSummary.biz_type)" size="mini">
            {{ bizTypeLabel(demandSummary.biz_type) }}
          </u-tag>
          <text class="summary-budget">{{ formatBudget(demandSummary.budget_fen) }}</text>
          <text v-if="demandSummary.district" class="summary-district">{{ demandSummary.district }}</text>
        </view>
      </view>

      <!-- Bid form -->
      <view class="form-section">
        <u-cell-group inset>
          <u-field
            v-model="amount"
            label="报价金额"
            type="digit"
            placeholder="请输入报价金额(元)"
          />

          <u-field
            v-model="proposal"
            label="方案描述"
            type="textarea"
            auto-height
            placeholder="描述您的方案和优势"
          />

          <u-field
            v-model="contactPhone"
            label="联系电话"
            type="number"
            placeholder="请输入联系电话"
          />
        </u-cell-group>

        <view class="submit-wrap">
          <u-button
            type="primary"
            block
            round
            :loading="submitting"
            @click="submitBid"
          >
            提交报价
          </u-button>
        </view>
      </view>
    </template>
  </view>
</template>

<script>
import { request, getStoredUser } from '../../utils/request'

export default {
  data() {
    return {
      id: '',
      loading: false,
      errorMsg: '',
      demandSummary: null,
      amount: '',
      proposal: '',
      contactPhone: '',
      submitting: false,
    }
  },
  onLoad(options) {
    this.id = options.id || ''
    if (!this.id) {
      this.errorMsg = '缺少需求ID'
      return
    }
    this.fetchDemandSummary()
  },
  methods: {
    async fetchDemandSummary() {
      this.loading = true
      this.errorMsg = ''

      try {
        const res = await request({ url: '/api/v1/demands/' + encodeURIComponent(this.id) })
        this.demandSummary = (res && res.data) || res || null
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    async submitBid() {
      // Validate
      var amountNum = parseFloat(this.amount)
      if (!this.amount || isNaN(amountNum) || amountNum <= 0) {
        uni.showToast({ title: '请输入有效的报价金额', icon: 'none' })
        return
      }
      if (!this.contactPhone) {
        uni.showToast({ title: '请输入联系电话', icon: 'none' })
        return
      }
      if (!/^1\d{10}$/.test(this.contactPhone)) {
        uni.showToast({ title: '请输入正确的11位手机号', icon: 'none' })
        return
      }

      this.submitting = true

      try {
        await request({
          url: '/api/v1/demands/' + encodeURIComponent(this.id) + '/applications',
          method: 'POST',
          data: {
            amount_fen: Math.round(amountNum * 100),
            proposal: this.proposal || '',
          },
        })
        uni.showToast({ title: '报价提交成功', icon: 'success', duration: 1500 })
        setTimeout(function () {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        uni.showToast({ title: '提交失败，请稍后重试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
    goBack() {
      uni.navigateBack()
    },
    bizTypeLabel(type) {
      var map = {
        cable_inspection: '巡检',
        plant_transport: '植保',
        spray_pesticide: '农药',
        trade_lease: '租赁',
        clean_paint: '清洗',
        other: '其他',
      }
      return map[type] || type || '其他'
    },
    bizTypeTagType(type) {
      var map = {
        cable_inspection: 'primary',
        plant_transport: 'success',
        spray_pesticide: 'warning',
        trade_lease: 'danger',
        clean_paint: 'primary',
      }
      return map[type] || 'default'
    },
    formatBudget(fen) {
      if (fen == null || fen === 0) return '面议'
      var yuan = (fen / 100).toFixed(2)
      return '¥' + yuan.replace(/\.00$/, '')
    },
  },
}
</script>

<style scoped>
.bid-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Summary card */
.summary-card {
  background: var(--color-bg-card);
  padding: 14px 16px;
  margin: 12px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.02);
}

.summary-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  display: block;
  margin-bottom: 10px;
  line-height: 1.4;
}

.summary-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.summary-budget {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-danger);
}

.summary-district {
  font-size: 13px;
  color: var(--color-text-secondary);
}

/* Form */
.form-section {
  margin: 0 12px;
}

.submit-wrap {
  padding: 20px 16px;
}
</style>
