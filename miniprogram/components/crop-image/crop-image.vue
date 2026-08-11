<template>
  <view v-if="visible" class="crop-mask" @touchmove.stop.prevent>
    <view class="crop-panel">
      <view class="crop-head">
        <text class="crop-title">调整图片位置与大小</text>
        <text class="crop-ratio">4:3 比例</text>
      </view>
      <view class="crop-stage">
        <canvas
          type="2d"
          id="cropCanvas"
          class="crop-canvas"
          @touchstart="onTouchStart"
          @touchmove="onTouchMove"
          @touchend="onTouchEnd"
        />
      </view>
      <view class="crop-tip">单指拖动图片，双指缩放大小</view>
      <view class="crop-foot">
        <view class="crop-btn" hover-class="tap-fade" hover-stay-time="120" @tap="onCancel">取消</view>
        <view class="crop-btn primary" hover-class="tap-fade" hover-stay-time="120" @tap="onConfirm">完成</view>
      </view>
    </view>
  </view>
</template>

<script>
// 4:3 固定比例图片裁剪器（canvas 2d，无第三方依赖）
// 流程：visible=true 且 src 变化 → 初始化 canvas → 加载图片 → 可拖动/缩放 →
// 确认后导出裁剪框区域为 1200×900 的 jpg 临时文件，emit confirm(tempFilePath)
// 注意：导出前重绘一次纯图片（去掉遮罩/框线），避免框线混入成图
export default {
  name: 'CropImage',
  props: {
    visible: { type: Boolean, default: false },
    src: { type: String, default: '' },
  },
  data() {
    return {
      canvas: null,
      ctx: null,
      cw: 0,
      ch: 0,
      canvasImg: null,
      imgW: 0,
      imgH: 0,
      baseW: 0,
      baseH: 0,
      scale: 1,
      minScale: 1,
      maxScale: 4,
      offsetX: 0,
      offsetY: 0,
      box: null, // { x, y, w, h } 裁剪框（CSS px，居中）
      touch: null,
      inited: false,
    }
  },
  watch: {
    visible(v) {
      if (v) {
        this.$nextTick(() => this.initCanvas())
      } else {
        this.reset()
      }
    },
  },
  methods: {
    reset() {
      this.inited = false
      this.canvasImg = null
      this.offsetX = 0
      this.offsetY = 0
      this.touch = null
    },
    initCanvas() {
      if (!this.src) {
        this.$emit('cancel')
        return
      }
      const query = uni.createSelectorQuery().in(this)
      query
        .select('#cropCanvas')
        .fields({ node: true, size: true })
        .exec((res) => {
          if (!res || !res[0] || !res[0].node) {
            this.$emit('cancel')
            return
          }
          const { node, width, height } = res[0]
          this.canvas = node
          this.ctx = node.getContext('2d')
          const dpr = uni.getSystemInfoSync().pixelRatio || 2
          node.width = width * dpr
          node.height = height * dpr
          this.ctx.scale(dpr, dpr)
          this.cw = width
          this.ch = height
          this.loadImage()
        })
    },
    loadImage() {
      uni.getImageInfo({
        src: this.src,
        success: (info) => {
          if (info.width && info.height) {
            this.imgW = info.width
            this.imgH = info.height
          }
          this.prepare()
        },
        fail: () => this.$emit('cancel'),
      })
    },
    prepare() {
      const margin = 24
      const availW = this.cw - margin * 2
      const availH = this.ch - margin * 2
      const fit = Math.min(availW / this.imgW, availH / this.imgH)
      this.baseW = this.imgW * fit
      this.baseH = this.imgH * fit
      // 裁剪框：宽为画布 88%，高按 4:3
      const boxW = this.cw * 0.88
      const boxH = (boxW * 3) / 4
      this.box = { x: (this.cw - boxW) / 2, y: (this.ch - boxH) / 2, w: boxW, h: boxH }
      // 最小缩放：图片必须完全覆盖裁剪框（框内不出现空白）
      this.minScale = Math.max(boxW / this.baseW, boxH / this.baseH, 1)
      this.scale = this.minScale
      this.maxScale = this.minScale * 4
      this.offsetX = 0
      this.offsetY = 0
      // canvas 2d 的 Image 对象（仅微信小程序可用）
      if (this.canvas.createImage) {
        const img = this.canvas.createImage()
        img.onload = () => {
          this.canvasImg = img
          this.render()
        }
        img.onerror = () => this.$emit('cancel')
        img.src = this.src
      } else {
        this.$emit('cancel')
      }
    },
    // 图片绘制位置（CSS px）
    imgRect() {
      const iw = this.baseW * this.scale
      const ih = this.baseH * this.scale
      return {
        ix: this.cw / 2 - iw / 2 + this.offsetX,
        iy: this.ch / 2 - ih / 2 + this.offsetY,
        iw,
        ih,
      }
    },
    render() {
      const ctx = this.ctx
      const { box } = this
      if (!ctx || !box) return
      const { ix, iy, iw, ih } = this.imgRect()
      ctx.clearRect(0, 0, this.cw, this.ch)
      ctx.drawImage(this.canvasImg, ix, iy, iw, ih)
      // 裁剪框外遮罩
      ctx.fillStyle = 'rgba(0,0,0,0.62)'
      ctx.fillRect(0, 0, this.cw, box.y)
      ctx.fillRect(0, box.y + box.h, this.cw, this.ch - box.y - box.h)
      ctx.fillRect(0, box.y, box.x, box.h)
      ctx.fillRect(box.x + box.w, box.y, this.cw - box.x - box.w, box.h)
      // 框线 + 三等分辅助线
      ctx.strokeStyle = 'rgba(255,255,255,0.9)'
      ctx.lineWidth = 1
      ctx.strokeRect(box.x + 0.5, box.y + 0.5, box.w - 1, box.h - 1)
      ctx.beginPath()
      ctx.strokeStyle = 'rgba(255,255,255,0.35)'
      for (let i = 1; i < 3; i++) {
        const gx = box.x + (box.w * i) / 3
        ctx.moveTo(gx, box.y)
        ctx.lineTo(gx, box.y + box.h)
        const gy = box.y + (box.h * i) / 3
        ctx.moveTo(box.x, gy)
        ctx.lineTo(box.x + box.w, gy)
      }
      ctx.stroke()
    },
    // 导出前重绘：只画图片，无遮罩/框线（避免框线混入成图）
    renderPlain() {
      const ctx = this.ctx
      if (!ctx) return
      const { ix, iy, iw, ih } = this.imgRect()
      ctx.clearRect(0, 0, this.cw, this.ch)
      ctx.drawImage(this.canvasImg, ix, iy, iw, ih)
    },
    // 图片显示区域必须覆盖裁剪框
    clamp() {
      const { ix, iw } = this.imgRect()
      const box = this.box
      this.offsetX = Math.min(this.offsetX, box.x - ix)
      this.offsetX = Math.max(this.offsetX, box.x + box.w - (ix + iw))
      const rect = this.imgRect()
      this.offsetY = Math.min(this.offsetY, box.y - rect.iy)
      this.offsetY = Math.max(this.offsetY, box.y + box.h - (rect.iy + rect.ih))
    },
    onTouchStart(e) {
      const t = e.touches || []
      if (t.length === 1) {
        this.touch = {
          mode: 'move',
          x: t[0].clientX,
          y: t[0].clientY,
          sx: this.offsetX,
          sy: this.offsetY,
        }
      } else if (t.length >= 2) {
        const dx = t[0].clientX - t[1].clientX
        const dy = t[0].clientY - t[1].clientY
        this.touch = {
          mode: 'pinch',
          dist: Math.sqrt(dx * dx + dy * dy),
          startScale: this.scale,
        }
      }
    },
    onTouchMove(e) {
      const t = e.touches || []
      if (!this.touch) return
      if (this.touch.mode === 'move' && t.length === 1) {
        this.offsetX = this.touch.sx + (t[0].clientX - this.touch.x)
        this.offsetY = this.touch.sy + (t[0].clientY - this.touch.y)
        this.clamp()
        this.render()
      } else if (this.touch.mode === 'pinch' && t.length >= 2) {
        const dx = t[0].clientX - t[1].clientX
        const dy = t[0].clientY - t[1].clientY
        const dist = Math.sqrt(dx * dx + dy * dy)
        if (dist > 0) {
          this.scale = Math.min(Math.max(this.touch.startScale * (dist / this.touch.dist), this.minScale), this.maxScale)
          this.clamp()
          this.render()
        }
      }
    },
    onTouchEnd() {
      this.touch = null
    },
    onCancel() {
      this.$emit('cancel')
    },
    onConfirm() {
      if (!this.canvas || !this.box) return
      const { box } = this
      this.renderPlain()
      uni.canvasToTempFilePath(
        {
          canvas: this.canvas,
          x: box.x,
          y: box.y,
          width: box.w,
          height: box.h,
          destWidth: 1200,
          destHeight: 900,
          fileType: 'jpg',
          quality: 0.9,
          success: (res) => {
            this.render() // 恢复带遮罩的编辑态
            this.$emit('confirm', res.tempFilePath)
          },
          fail: () => {
            this.render()
            uni.showToast({ title: '裁剪失败，请重试', icon: 'none' })
          },
        },
        this
      )
    },
  },
}
</script>

<style scoped>
.crop-mask {
  position: fixed;
  left: 0;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 999;
  background: rgba(0, 0, 0, 0.72);
  display: flex;
  align-items: center;
  justify-content: center;
}
.crop-panel {
  width: 88%;
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
}
.crop-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px 10px;
}
.crop-title {
  font-size: 15px;
  font-weight: 600;
  color: #17212b;
}
.crop-ratio {
  font-size: 11px;
  color: #0a66c2;
  background: #eaf3fb;
  padding: 2px 8px;
  border-radius: 9px;
}
.crop-stage {
  padding: 0 12px;
  position: relative;
}
.crop-canvas {
  width: 100%;
  height: 420px;
  display: block;
  border-radius: 8px;
  background: #111;
}
.crop-tip {
  text-align: center;
  font-size: 11px;
  color: #98a2b3;
  padding: 10px 0 4px;
}
.crop-foot {
  display: flex;
  gap: 12px;
  padding: 12px 18px 18px;
}
.crop-btn {
  flex: 1;
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 21px;
  font-size: 15px;
  color: #344054;
  background: #f2f4f7;
}
.crop-btn.primary {
  color: #fff;
  background: #0a66c2;
}
</style>
