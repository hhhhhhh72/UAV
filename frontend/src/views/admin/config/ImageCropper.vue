<template>
  <van-popup :show="show" @update:show="v => $emit('update:show', v)" position="bottom" :style="{ height: '80%' }" round>
    <div class="cropper-popup">
      <van-nav-bar
        :title="title"
        left-text="取消"
        right-text="确认"
        @click-left="close"
        @click-right="confirm"
      />
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
        <van-button size="small" icon="enlarge-o" @click="zoomIn">放大</van-button>
        <van-button size="small" icon="shrink-o" @click="zoomOut">缩小</van-button>
        <van-button size="small" icon="aim" @click="reset">重置</van-button>
      </div>
      <div class="cropper-tips">
        提示：拖动图片调整位置，双指缩放调整大小
      </div>
    </div>
  </van-popup>
</template>

<script setup>
import { ref, watch } from 'vue'
import { VueCropper } from 'vue-cropper'
import 'vue-cropper/dist/index.css'
import { showLoadingToast, closeToast, showFailToast } from 'vant'

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

const rotateLeft = () => cropperRef.value?.rotateLeft()
const rotateRight = () => cropperRef.value?.rotateRight()
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
.cropper-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #000;
}

.cropper-container {
  flex: 1;
  overflow: hidden;
  background: #1a1a1a;
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
  padding: 12px 16px;
  background: #fff;
  border-top: 1px solid #f5f5f7;
}

.cropper-tips {
  text-align: center;
  padding: 8px 16px 16px;
  font-size: 12px;
  color: #86868b;
  background: #fff;
}
</style>
