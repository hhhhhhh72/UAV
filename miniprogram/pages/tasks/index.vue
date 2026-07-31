<template>
<view class="page">
  <view class="header">
    <text class="title">需求大厅</text>
    <text class="city">{{ city }}</text>
  </view>

  <view class="cat-row">
    <scroll-view scroll-x :show-scrollbar="false">
      <text v-for="c in cats" :key="c.id" class="cat" :class="{on:activeCat===c.id}" @click="switchCat(c.id)">{{ c.name }}</text>
    </scroll-view>
  </view>

  <scroll-view scroll-y class="list" @scrolltolower="loadMore">
    <view v-for="dm in demands" :key="dm.id" class="card" @click="openDetail(dm)">
      <view class="card-hd">
        <text class="tag" :class="tagClass(dm.biz_type)">{{ dm.biz_type||'其他' }}</text>
        <text class="status" v-if="dm.status">{{ statusMap[dm.status]||dm.status }}</text>
        <text class="district">{{ dm.district }}</text>
      </view>
      <text class="card-title">{{ dm.title }}</text>
      <view class="card-info">
        <text>💰 ¥{{ ((dm.budget_fen||0)/100).toFixed(2) }}</text>
        <text>👤 {{ dm.publisher_name||'匿名' }}</text>
      </view>
      <view class="card-foot">
        <text class="price">¥{{ ((dm.budget_fen||0)/100).toFixed(2) }}</text>
        <view class="foot-right">
          <text class="bids">{{ dm.bid_count||0 }}报价</text>
          <view class="btn" @click.stop="grab(dm)">抢单</view>
        </view>
      </view>
    </view>
    <view v-if="!demands.length" class="empty">暂无需求</view>
  </scroll-view>

  <view class="float-btn" @click="goMap">🔍 附近飞手</view>
</view>
</template>

<script setup>
import { ref } from 'vue'
import { request } from '@/utils/request'

const city = ref('全国')
const cats = ref([
  { id:'', name:'全部' }, { id:'吊运', name:'吊运' }, { id:'航拍', name:'航拍' },
  { id:'植保', name:'植保' }, { id:'巡检', name:'巡检' }, { id:'测绘', name:'测绘' },
])
const activeCat = ref('')
const demands = ref([])
let p = 1
const statusMap = { published:'待接单', bidding:'竞价中', assigned:'已接单', completed:'已完成' }

const fetchDemands = async (reset) => {
  if (reset) { p = 1; demands.value = [] }
  try {
    const res = await request({ url:'/api/v1/demands', data:{ page:p, page_size:20, biz_type:activeCat.value||undefined } })
    const arr = Array.isArray(res) ? res : (res.data||[])
    demands.value = p===1 ? arr : demands.value.concat(arr)
  } catch {}
}
fetchDemands(true)

const switchCat = (id) => { activeCat.value = id; fetchDemands(true) }
const loadMore = () => { p++; fetchDemands(false) }
const openDetail = (dm) => uni.navigateTo({ url:'/pages/tasks/detail?id='+dm.id })
const grab = (dm) => uni.navigateTo({ url:'/pages/tasks/detail?id='+dm.id })
const goMap = () => uni.showToast({ title:'附近飞手开发中', icon:'none' })
const tagClass = (t) => { const m={吊运:'dy',航拍:'hp',植保:'zb',巡检:'xj',测绘:'ch'}; return 'tag-'+ (m[t]||'other') }
</script>

<style scoped>
.page { min-height:100vh; background:#f5f6f7; }
.header { display:flex; justify-content:space-between; padding:14px 16px; background:#fff; }
.title { font-size:18px; font-weight:700; color:#1a1a1a; }
.city { font-size:14px; color:#0A66C2; }
.cat-row { background:#fff; padding:0 12px 8px; }
.cat-row scroll-view { display:flex; gap:6px; white-space:nowrap; }
.cat { font-size:13px; padding:6px 14px; border-radius:16px; background:#f5f5f5; color:#666; margin-right:8px; display:inline-block; }
.cat.on { background:#0A66C2; color:#fff; }
.list { padding:6px 12px 100px; height: calc(100vh - 90px); }
.card { background:#fff; border-radius:10px; padding:14px; margin-bottom:8px; }
.card-hd { display:flex; align-items:center; gap:8px; margin-bottom:8px; }
.tag { font-size:11px; padding:2px 8px; border-radius:3px; color:#fff; }
.tag-dy { background:#ff6b35; } .tag-hp { background:#4caf50; } .tag-zb { background:#2196f3; } .tag-xj { background:#9c27b0; } .tag-ch { background:#607d8b; } .tag-other { background:#999; }
.status { font-size:12px; color:#07c160; flex:1; }
.district { font-size:12px; color:#999; }
.card-title { font-size:16px; font-weight:600; color:#111; display:block; margin-bottom:8px; }
.card-info { display:flex; gap:12px; font-size:13px; color:#666; margin-bottom:8px; }
.card-foot { display:flex; justify-content:space-between; align-items:center; border-top:1px solid #f0f0f0; padding-top:10px; }
.price { font-size:20px; font-weight:700; color:#ff4d4f; }
.foot-right { display:flex; align-items:center; gap:10px; }
.bids { font-size:12px; color:#999; }
.btn { background:#07c160; color:#fff; font-size:13px; padding:6px 14px; border-radius:16px; }
.float-btn { position:fixed; right:16px; bottom:80px; background:#07c160; color:#fff; font-size:13px; padding:10px 14px; border-radius:20px; box-shadow:0 2px 8px rgba(7,193,96,.3); z-index:10; }
.empty { text-align:center; padding:60px 0; color:#999; font-size:14px; }
</style>
