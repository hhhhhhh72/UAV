<template>
  <el-dialog
    :model-value="show"
    :title="title"
    width="720px"
    top="6vh"
    append-to-body
    destroy-on-close
    @update:model-value="v => emit('update:show', v)"
  >
    <div class="cropper-container">
      <vue-cropper
        ref="cropperRef"
        :img="imageUrl"
        :output-size="1"
        :output-type="'png'"
        :info="false"
        :can-scale="true"
        :auto-crop="true"
        :auto-crop-width="cropWidth"
        :auto-crop-height="cropHeight"
        :fixed="fixed"
        :fixed-number="fixedNumber"
        :center-box="true"
        :full="false"
        :mode="'contain'"
        :can-move="true"
        :can-move-box="false"
        :fixed-box="true"
      />
    </div>
    <div class="cropper-tools">
      <el-button size="small" @click="zoomIn">放大</el-button>
      <el-button size="small" @click="zoomOut">缩小</el-button>
      <el-button size="small" @click="reset">重置</el-button>
    </div>
    <div class="cropper-tips">
      提示：拖动图片调整位置，双指缩放调整大小
    </div>
    <template #footer>
      <el-button @click="close">取消</el-button>
      <el-button type="primary" @click="confirm">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { VueCropper } from 'vue-cropper'
import 'vue-cropper/dist/index.css'
import { showLoadingToast, closeToast, showFailToast } from '@/utils/feedback'

const props = defineProps({
  show: Boolean,
  imageUrl: String,
  title: { type: String, default: '图片裁剪' },
  // 裁剪比例: '16:9', '4:3', '1:1', 'free'
  aspectRatio: { type: String, default: '16:9' }
})

const emit = defineEmits(['update:show', 'confirm'])

const cropperRef = ref(null)

// 根据比例设置裁剪框尺寸
const fixed = ref(true)
const fixedNumber = ref([16, 9])
const cropWidth = ref(300)
const cropHeight = ref(169)

watch(() => props.aspectRatio, (ratio) => {
  switch (ratio) {
    case '16:9':
      fixedNumber.value = [16, 9]
      cropWidth.value = 320
      cropHeight.value = 180
      break
    case '4:3':
      fixedNumber.value = [4, 3]
      cropWidth.value = 240
      cropHeight.value = 180
      break
    case '1:1':
      fixedNumber.value = [1, 1]
      cropWidth.value = 200
      cropHeight.value = 200
      break
    case 'free':
    default:
      fixed.value = false
      cropWidth.value = 200
      cropHeight.value = 200
  }
}, { immediate: true })

const zoomIn = () => cropperRef.value?.changeScale(1)
const zoomOut = () => cropperRef.value?.changeScale(-1)
const reset = () => cropperRef.value?.refresh()

const close = () => {
  emit('update:show', false)
}

const confirm = () => {
  showLoadingToast({ message: '处理中...', forbidClick: true })
  cropperRef.value.getCropBlob((blob) => {
    closeToast()
    if (blob) {
      // 将 blob 转换为 File 对象
      const file = new File([blob], 'cropped.png', { type: 'image/png' })
      emit('confirm', file)
      emit('update:show', false)
    } else {
      showFailToast('裁剪失败')
    }
  })
}
</script>

<style scoped>
.cropper-container {
  height: 420px;
  overflow: hidden;
  background: #1a1a1a;
  border-radius: 6px;
}

/* 自定义裁剪框样式 - 类似微信头像选择 */
:deep(.vue-cropper) {
  background: #1a1a1a;
}

:deep(.cropper-view-box) {
  outline: 2px solid #fff;
  outline-color: rgba(255, 255, 255, 0.8);
}

:deep(.cropper-face) {
  background-color: transparent !important;
}

:deep(.cropper-line) {
  background-color: #0071e3;
}

:deep(.cropper-point) {
  background-color: #0071e3;
  width: 8px;
  height: 8px;
}

:deep(.cropper-bg) {
  background-image: none;
  background-color: #1a1a1a;
}

:deep(.cropper-modal) {
  background-color: rgba(0, 0, 0, 0.6);
}

.cropper-tools {
  display: flex;
  justify-content: center;
  gap: 12px;
  padding: 12px 16px 0;
}

.cropper-tips {
  text-align: center;
  padding: 8px 16px 0;
  font-size: 12px;
  color: #86868b;
}
</style>
