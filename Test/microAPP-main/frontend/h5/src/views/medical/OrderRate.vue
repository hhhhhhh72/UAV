<template>
  <div class="order-rate-page">
    <van-nav-bar title="订单评价" left-arrow @click-left="$router.back()" fixed placeholder />

    <div class="rate-content">
      <div class="rate-header">
        <p class="rate-title">配送服务满意吗？</p>
        <van-rate v-model="rating" size="30" allow-half :count="5" />
        <p class="rate-desc">{{ ratingDesc }}</p>
      </div>

      <!-- 快速标签 -->
      <div class="tag-section">
        <p class="section-title">选择标签</p>
        <div class="tags">
          <van-tag
            v-for="tag in currentTags"
            :key="tag"
            :plain="!selectedTags.includes(tag)"
            :type="selectedTags.includes(tag) ? 'primary' : 'default'"
            size="large"
            round
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </van-tag>
        </div>
      </div>

      <!-- 文字评价 -->
      <div class="comment-section">
        <van-field
          v-model="comment"
          type="textarea"
          placeholder="请输入您的评价（选填，最多200字）"
          rows="4"
          maxlength="200"
          show-word-limit
        />
      </div>

      <div class="submit-section">
        <van-button type="primary" block round :loading="submitting" @click="onSubmit" :disabled="!rating">提交评价</van-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showSuccessToast, showFailToast } from 'vant'
import { useMedicalStore } from '@/stores/medical'

const router = useRouter()
const route = useRoute()
const store = useMedicalStore()

const rating = ref(5)
const selectedTags = ref([])
const comment = ref('')
const submitting = ref(false)

const goodTags = ['配送迅速', '包装完好', '服务态度好', '准时送达']
const midTags = ['一般般', '等待较久']
const badTags = ['配送超时', '包装破损', '服务态度差', '没有通知']

const currentTags = computed(() => {
  if (rating.value >= 4) return goodTags
  if (rating.value === 3) return midTags
  return badTags
})

const ratingDesc = computed(() => {
  if (rating.value >= 5) return '非常满意'
  if (rating.value >= 4) return '满意'
  if (rating.value >= 3) return '一般'
  if (rating.value >= 2) return '不满意'
  return '非常不满意'
})

function toggleTag(tag) {
  const index = selectedTags.value.indexOf(tag)
  if (index > -1) {
    selectedTags.value.splice(index, 1)
  } else {
    selectedTags.value.push(tag)
  }
}

async function onSubmit() {
  if (!rating.value) {
    showFailToast('请选择评分')
    return
  }

  submitting.value = true
  try {
    const res = await store.rateOrder(route.params.id, {
      rating: rating.value,
      tags: selectedTags.value,
      comment: comment.value
    })
    if (res?.success) {
      showSuccessToast('评价成功')
      router.back()
    } else {
      showFailToast(res?.message || '评价失败')
    }
  } catch (e) {
    showFailToast(e.response?.data?.message || '评价失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.order-rate-page {
  min-height: 100vh;
  background: #f5f5f5;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

/* 固定导航栏居中约束 */
.order-rate-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

.rate-content {
  padding: 20px 16px;
}
.rate-header {
  text-align: center;
  padding: 24px 0;
  background: #fff;
  border-radius: 10px;
  margin-bottom: 16px;
}
.rate-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}
.rate-desc {
  margin-top: 8px;
  font-size: 14px;
  color: #ff976a;
}
.tag-section {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
}
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.comment-section {
  background: #fff;
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 24px;
}
.submit-section {
  padding: 0 16px;
}
</style>
