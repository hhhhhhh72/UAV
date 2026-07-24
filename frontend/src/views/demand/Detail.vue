<template>
  <div class="page">
    <van-nav-bar title="需求详情" left-arrow @click-left="$router.back()" fixed placeholder />
    <van-loading v-if="!detail" />
    <template v-else>
      <van-cell-group>
        <van-cell :title="detail.title" :label="detail.description" />
        <van-cell title="类型" :value="detail.biz_type" />
        <van-cell title="预算" :value="'¥' + (detail.budget_fen / 100).toFixed(2)" />
        <van-cell title="城市" :value="detail.city" />
        <van-cell title="状态" :value="detail.status" />
      </van-cell-group>
      <div class="bids">
        <h3>竞标列表</h3>
        <van-cell v-for="bid in bids" :key="bid.id"
          :title="bid.bidder_name" :label="bid.proposal"
          :value="'¥' + (bid.amount_fen / 100).toFixed(2)" />
        <van-empty v-if="bids.length === 0" description="暂无竞标" />
      </div>
      <van-button type="primary" block @click="showBidForm = true">参与竞标</van-button>
      <van-action-sheet v-model:show="showBidForm" title="提交竞标" close-on-click-action>
        <van-form @submit="onBidSubmit" style="padding: 16px">
          <van-field
            v-model="bidForm.bidder_name"
            name="bidder_name"
            label="竞标人"
            placeholder="请输入姓名"
            :rules="[{ required: true, message: '请输入竞标人' }]"
          />
          <van-field
            v-model="bidForm.proposal"
            name="proposal"
            label="方案说明"
            type="textarea"
            rows="3"
            placeholder="请输入方案说明"
            :rules="[{ required: true, message: '请输入方案说明' }]"
          />
          <van-field
            v-model="bidForm.amount_fen"
            name="amount_fen"
            label="报价（分）"
            type="number"
            placeholder="请输入报价（单位：分）"
            :rules="[{ required: true, message: '请输入报价' }]"
          />
          <div style="margin: 16px 0">
            <van-button round block type="primary" native-type="submit">提交竞标</van-button>
          </div>
        </van-form>
      </van-action-sheet>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { showToast, showLoadingToast, closeToast, showFailToast } from 'vant'
import http from '@/utils/http'

const route = useRoute()
const detail = ref(null)
const bids = ref([])
const showBidForm = ref(false)
const bidForm = ref({
  bidder_name: '',
  proposal: '',
  amount_fen: ''
})

onMounted(async () => {
  try {
    const { data } = await http.get(`/api/v1/demands/${route.params.id}`)
    detail.value = data
    const bres = await http.get(`/api/v1/demands/${route.params.id}/applications`)
    bids.value = bres.data || []
  } catch (e) {
    showFailToast('加载需求详情失败')
  }
})

const onBidSubmit = async () => {
  showLoadingToast({ message: '提交中...', forbidClick: true, duration: 0 })
  try {
    await http.post(`/api/v1/demands/${route.params.id}/applications`, bidForm.value)
    closeToast()
    showToast('竞标成功')
    showBidForm.value = false
    // 刷新竞标列表
    const bres = await http.get(`/api/v1/demands/${route.params.id}/applications`)
    bids.value = bres.data || []
  } catch (e) {
    closeToast()
    showFailToast('提交失败，请重试')
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #f7f8fa;
  padding-bottom: 80px;
}
.bids {
  margin-top: 16px;
}
.bids h3 {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
  padding: 0 16px;
  margin-bottom: 8px;
}
.van-button {
  margin: 16px;
}
</style>
