<template>
  <div class="page">
    <van-nav-bar title="发布需求" left-arrow @click-left="$router.back()" fixed placeholder />
    <van-form @submit="onSubmit" style="padding: 16px">
      <van-field
        v-model="form.title"
        name="title"
        label="标题"
        placeholder="请输入需求标题"
        :rules="[{ required: true, message: '请输入标题' }]"
      />
      <van-field
        v-model="form.description"
        name="description"
        label="描述"
        type="textarea"
        rows="3"
        placeholder="请输入需求描述"
        :rules="[{ required: true, message: '请输入描述' }]"
      />
      <van-field
        v-model="form.biz_type_text"
        is-link
        readonly
        name="biz_type"
        label="业务类型"
        placeholder="请选择业务类型"
        :rules="[{ required: true, message: '请选择业务类型' }]"
        @click="showBizTypeSheet = true"
      />
      <van-action-sheet
        v-model:show="showBizTypeSheet"
        :actions="bizTypeOptions"
        @select="onBizTypeSelect"
        close-on-select
        title="选择业务类型"
      />
      <van-field
        v-model="form.budget"
        name="budget"
        label="预算（元）"
        type="number"
        placeholder="请输入预算金额"
        :rules="[{ required: true, message: '请输入预算' }]"
      />
      <van-field
        v-model="form.city"
        name="city"
        label="城市"
        placeholder="请输入城市"
        :rules="[{ required: true, message: '请输入城市' }]"
      />
      <div style="margin: 24px 16px">
        <van-button round block type="primary" native-type="submit">发布需求</van-button>
      </div>
    </van-form>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast, showFailToast } from 'vant'
import http from '@/utils/http'

const router = useRouter()
const showBizTypeSheet = ref(false)

const bizTypeOptions = [
  { name: '巡检', value: '巡检' },
  { name: '植保', value: '植保' },
  { name: '农药', value: '农药' },
  { name: '租赁', value: '租赁' },
  { name: '清洗', value: '清洗' },
  { name: '其他', value: '其他' }
]

const form = reactive({
  title: '',
  description: '',
  biz_type: '',
  biz_type_text: '',
  budget: '',
  city: ''
})

const onBizTypeSelect = (action) => {
  form.biz_type = action.value
  form.biz_type_text = action.name
}

const onSubmit = async () => {
  showLoadingToast({ message: '发布中...', forbidClick: true, duration: 0 })
  try {
    const payload = {
      title: form.title,
      description: form.description,
      biz_type: form.biz_type,
      budget_fen: Math.round(parseFloat(form.budget) * 100),
      city: form.city
    }
    await http.post('/api/v1/demands', payload)
    closeToast()
    showToast('发布成功')
    router.back()
  } catch (e) {
    closeToast()
    showFailToast('发布失败，请重试')
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #f7f8fa;
}
</style>
