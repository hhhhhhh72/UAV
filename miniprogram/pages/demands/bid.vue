<template>
  <view class="bid-page">
    <van-nav-bar
      title="竞标报价"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading -->
    <view v-if="loading" class="state-view">
      <van-loading size="24" vertical>加载中...</van-loading>
    </view>

    <!-- Error -->
    <view v-else-if="errorMsg" class="state-view">
      <van-empty description="加载失败" image="error" />
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
          <van-tag :type="bizTypeTagType(demandSummary.biz_type)" size="small">
            {{ bizTypeLabel(demandSummary.biz_type) }}
          </van-tag>
          <text class="summary-budget">{{ formatBudget(demandSummary.budget_fen) }}</text>
          <text class="summary-district">{{ demandSummary.district || '' }}</text>
        </view>
      </view>

      <!-- Bid form -->
      <view class="form-card">
        <van-cell-group inset>
          <van-field
            v-model="amount"
            label="报价金额"
            type="number"
            placeholder="请输入报价金额"
            :border="true"
          >
            <template #right-icon>
              <text class="unit-text">元</text>
            </template>
          </van-field>

          <van-field
            v-model="proposal"
            label="方案说明"
            type="textarea"
            placeholder="请描述您的服务方案（选填）"
            rows="4"
            autosize
          />

          <van-field
            v-model="contactPhone"
            label="联系电话"
            type="number"
            placeholder="请输入联系电话"
            maxlength="11"
          />
        </van-cell-group>

        <view class="submit-wrap">
          <van-button
            type="primary"
            block
            round
            :loading="submitting"
            loading-text="提交中..."
            @tap="submitBid"
          >
            提交报价
          </van-button>
        </view>
      </view>
    </template>
  </view>
</template>

<script>
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
      const accessToken = uni.getStorageSync('accessToken') || ''
      const url = 'http://localhost:8080/api/v1/demands/' + encodeURIComponent(this.id)
      try {
        const [err, resp] = await uniRequest(url, {
          header: accessToken ? { Authorization: 'Bearer ' + accessToken } : {},
        })
        if (err) {
          this.errorMsg = err.message || '加载失败'
          return
        }
        this.demandSummary = resp.data || resp
      } catch (e) {
        this.errorMsg = '网络异常'
      } finally {
        this.loading = false
      }
    },
    async submitBid() {
      // Validate
      const amountNum = parseFloat(this.amount)
      if (!this.amount || isNaN(amountNum) || amountNum <= 0) {
        uni.showToast({ title: '请输入有效的报价金额', icon: 'none' })
        return
      }
      if (!this.contactPhone) {
        uni.showToast({ title: '请输入联系电话', icon: 'none' })
        return
      }
      if (!/^\d{11}$/.test(this.contactPhone)) {
        uni.showToast({ title: '请输入正确的11位手机号', icon: 'none' })
        return
      }

      this.submitting = true
      const accessToken = uni.getStorageSync('accessToken') || ''
      const url = 'http://localhost:8080/api/v1/demands/' + encodeURIComponent(this.id) + '/applications'
      const body = {
        amount_fen: Math.round(amountNum * 100),
        proposal: this.proposal || '',
      }

      try {
        const [err] = await uniPost(url, body, {
          header: {
            'Content-Type': 'application/json',
            'Authorization': accessToken ? 'Bearer ' + accessToken : '',
          },
        })
        if (err) {
          uni.showToast({ title: err.message || '提交失败', icon: 'none' })
          return
        }
        uni.showToast({ title: '报价提交成功', icon: 'success', duration: 1500 })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        uni.showToast({ title: '网络异常', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
    goBack() {
      uni.navigateBack()
    },
    bizTypeLabel(type) {
      const map = {
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
      const map = {
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
      const yuan = (fen / 100).toFixed(2)
      return yuan.replace(/\.00$/, '') + ' 元'
    },
  },
}

function uniRequest(url, options) {
  return new Promise((resolve) => {
    uni.request({
      url,
      method: 'GET',
      header: options.header || {},
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve([null, res.data])
        } else {
          const msg =
            (res.data && res.data.error && res.data.error.message) ||
            '请求失败 (' + res.statusCode + ')'
          resolve([new Error(msg), null])
        }
      },
      fail: (err) => {
        resolve([err || new Error('网络异常'), null])
      },
    })
  })
}

function uniPost(url, data, options) {
  return new Promise((resolve) => {
    uni.request({
      url,
      method: 'POST',
      header: options.header || {},
      data,
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve([null, res.data])
        } else {
          const msg =
            (res.data && res.data.error && res.data.error.message) ||
            '提交失败 (' + res.statusCode + ')'
          resolve([new Error(msg), null])
        }
      },
      fail: (err) => {
        resolve([err || new Error('网络异常'), null])
      },
    })
  })
}
</script>

<style scoped>
.bid-page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

.state-view {
  padding-top: 120px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #1989fa;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

.summary-card {
  background: #fff;
  padding: 20px 16px;
  margin: 12px 12px 0;
  border-radius: 12px;
}

.summary-title {
  font-size: 17px;
  font-weight: 700;
  color: #323233;
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
  color: #ee0a24;
}

.summary-district {
  font-size: 13px;
  color: #969799;
}

.form-card {
  margin: 12px 12px 0;
}

.unit-text {
  font-size: 14px;
  color: #969799;
}

.submit-wrap {
  padding: 20px 16px;
}
</style>
