<template>
  <view class="page">
    <view class="search-box"><input v-model="searchText" placeholder="搜索业务分类" class="search-input" /></view>
    <view v-if="filteredCats.length === 0" class="empty">未找到匹配的业务</view>
    <view v-for="cat in filteredCats" :key="cat.id" class="cat-card">
      <view class="cat-header" @click="goCat(cat)">
        <view class="cat-icon" :style="{background: cat.gradient}">{{ cat.icon }}</view>
        <view class="cat-info"><text class="cat-title">{{ cat.title }}</text><text class="cat-sub">{{ cat.subtitle }}</text></view>
        <text class="arrow">&gt;</text>
      </view>
      <view class="sub-list">
        <view v-for="sub in cat.subItems" :key="sub.id" class="sub-item" @click="goSub(sub)">
          <text>{{ sub.icon }} {{ sub.name }}</text><text class="arrow">&gt;</text>
        </view>
      </view>
    </view>
  </view>
</template>
<script>
export default {
  data() {
    return {
      searchText: '',
      categories: [
        {id:'supply',title:'产业供需对接',subtitle:'资源匹配 高效对接',icon:'🔁',gradient:'linear-gradient(135deg,#6366f1,#8b5cf6)',subItems:[{id:'hall',name:'需求大厅',icon:'📢'},{id:'show',name:'供应展示',icon:'🏪'},{id:'bid',name:'竞标报价',icon:'📋'}]},
        {id:'training',title:'培训认证',subtitle:'专业考证 技能提升',icon:'🎖️',gradient:'linear-gradient(135deg,#06b6d4,#0891b2)',subItems:[{id:'caac',name:'CAAC执照',icon:'📜'},{id:'utc',name:'UTC认证',icon:'🏅'},{id:'hr',name:'人社认证',icon:'📋'},{id:'pilot',name:'飞手培训',icon:'👤'}]},
        {id:'trade',title:'无人机交易',subtitle:'整机配件 一站购齐',icon:'🛒',gradient:'linear-gradient(135deg,#f59e0b,#d97706)',subItems:[{id:'unit',name:'整机购买',icon:'💎'},{id:'repair',name:'维修服务',icon:'🔧'},{id:'parts',name:'配件商城',icon:'📦'}]},
        {id:'contract',title:'合同签约',subtitle:'电子签章 安全合规',icon:'✏️',gradient:'linear-gradient(135deg,#10b981,#059669)',subItems:[{id:'tpl',name:'合同模板',icon:'📄'},{id:'sign',name:'在线签章',icon:'✍️'},{id:'void',name:'合同作废',icon:'🗑️'}]},
        {id:'insurance',title:'保险金融',subtitle:'全面保障 资金支持',icon:'🛡️',gradient:'linear-gradient(135deg,#ef4444,#dc2626)',subItems:[{id:'policy',name:'无人机保单',icon:'🛡️'},{id:'annual',name:'年审服务',icon:'⏰'},{id:'loan',name:'金融贷款',icon:'🪙'}]},
        {id:'emergency',title:'应急资源协同',subtitle:'快速响应 资源调度',icon:'🔥',gradient:'linear-gradient(135deg,#f97316,#ea580c)',subItems:[{id:'case',name:'救援案例',icon:'ℹ️'},{id:'dispatch',name:'资源调度',icon:'📦'}]},
      ]
    }
  },
  computed: {
    filteredCats() { const q=this.searchText.toLowerCase(); if(!q) return this.categories; return this.categories.filter(c=>c.title.includes(q)||c.subItems.some(s=>s.name.includes(q))) }
  },
  methods: {
    goCat(cat) { uni.navigateTo({url:'/pages/services/detail?id='+cat.id+'&title='+cat.title}) },
    goSub(sub) { uni.showToast({title:sub.name+' — 即将上线',icon:'none'}) }
  }
}
</script>
<style scoped>
.page{padding:12px;background:#f5f5f5;min-height:100vh}
.search-input{background:#fff;padding:10px 14px;border-radius:20px;font-size:14px;margin-bottom:12px}
.empty{text-align:center;color:#999;padding:40px 0}
.cat-card{background:#fff;border-radius:12px;margin-bottom:12px;overflow:hidden}
.cat-header{display:flex;align-items:center;padding:14px;gap:12px}
.cat-icon{width:44px;height:44px;border-radius:12px;display:flex;align-items:center;justify-content:center;font-size:22px;color:#fff;flex-shrink:0}
.cat-info{flex:1}
.cat-title{font-size:16px;font-weight:600;display:block}
.cat-sub{font-size:12px;color:#999;margin-top:2px;display:block}
.arrow{color:#c8c9cc;font-size:14px}
.sub-list{padding:0 14px 14px}
.sub-item{display:flex;justify-content:space-between;padding:10px 12px;background:#f8f9fa;border-radius:8px;margin-top:6px;font-size:14px}
</style>
