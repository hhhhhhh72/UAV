<template>
  <view class="page">
    <view class="header">我的业务</view>
    <view class="tabs">
      <view v-for="t in tabs" :key="t.key" :class="['tab', {active:activeTab===t.key}]" @click="activeTab=t.key">{{ t.label }}</view>
    </view>
    <view v-if="items.length===0" class="empty">暂无记录</view>
    <view v-for="item in items" :key="item.id" class="card">
      <view class="card-title">{{ item.title || item.name || '—' }}</view>
      <view class="card-status">{{ item.status }}</view>
    </view>
  </view>
</template>
<script>
export default {
  data() { return { activeTab: 'demands', tabs: [{key:'demands',label:'我的需求'},{key:'bids',label:'我的竞标'},{key:'contracts',label:'我的合同'},{key:'orders',label:'我的订单'}], items: [] } },
  onShow() { this.fetchItems() },
  methods: {
    async fetchItems() {
      try {
        const token = uni.getStorageSync('accessToken')
        const endpoints = {demands:'/api/v1/demands?mine=1',bids:'/api/v1/demands/bids/mine',contracts:'/api/v1/contracts',orders:'/api/v1/trade-orders/mine'}
        const res = await uni.request({url:'http://localhost:8080'+endpoints[this.activeTab],header:{Authorization:'Bearer '+token}})
        this.items = res.data?.data?.items || res.data?.items || []
      } catch(e) { this.items = [] }
    }
  },
  watch: { activeTab() { this.fetchItems() } }
}
</script>
<style scoped>
.page{padding:12px;background:#f5f5f5;min-height:100vh}
.header{font-size:20px;font-weight:bold;margin-bottom:12px}
.tabs{display:flex;gap:8px;margin-bottom:12px}
.tab{padding:8px 16px;background:#fff;border-radius:20px;font-size:13px}
.tab.active{background:#1565C0;color:#fff}
.empty{text-align:center;color:#999;padding:60px 0}
.card{background:#fff;padding:14px;border-radius:8px;margin-bottom:8px}
.card-title{font-size:15px;font-weight:600}
.card-status{font-size:12px;color:#999;margin-top:4px}
</style>
