<template>
  <view class="chat-page">
    <!-- 头部 -->
    <view class="page-header">
      <view class="back-btn" @tap="goBack"><text class="back-sym">‹</text></view>
      <text class="page-title">与 {{ partner.name }} 洽谈</text>
      <view class="head-action" @tap="confirmDone">
        <text class="head-action-text">合作确认</text>
      </view>
    </view>

    <!-- 提醒条 -->
    <view class="chat-notice">
      <text>本会话仅用于供需对接。请勿在平台外进行提前付款，成交结果可在合作确认中登记。</text>
    </view>

    <!-- 消息区 -->
    <scroll-view scroll-y class="chat-body" :scroll-into-view="scrollInto" scroll-with-animation>
      <view class="chat-inner">
        <view
          v-for="msg in messages"
          :key="msg.id"
          :id="'msg-' + msg.id"
          class="bubble-row"
          :class="{ mine: msg.mine }"
        >
          <view class="bubble-avatar"><text>{{ msg.avatar }}</text></view>
          <view class="bubble-wrap">
            <view class="bubble"><text>{{ msg.text }}</text></view>
            <view v-if="msg.attach" class="attachment">
              <text class="attach-sym">▣</text>
              <text class="attach-name">{{ msg.attach }}</text>
            </view>
          </view>
        </view>
      </view>
    </scroll-view>

    <!-- 输入区 -->
    <view class="chat-input">
      <view class="tool-btn" @tap="toastImage"><text class="tool-sym">▧</text></view>
      <view class="tool-btn" @tap="toastFile"><text class="tool-sym">＋</text></view>
      <input
        v-model="draft"
        class="input"
        placeholder="输入消息"
        confirm-type="send"
        @confirm="sendMessage"
      />
      <view class="send-btn" @tap="sendMessage"><text>发送</text></view>
    </view>

    <!-- 合作确认弹层 -->
    <u-popup :show="showDone" position="bottom" round @close="showDone = false">
      <view class="sheet">
        <view class="sheet-head">
          <text class="sheet-title">确认合作结果</text>
          <view class="sheet-close" @tap="showDone = false"><text class="sheet-x">×</text></view>
        </view>
        <view class="sheet-body">
          <text class="sheet-desc">平台仅记录撮合结果，不涉及支付与合同。请如实选择当前合作进展。</text>
          <view class="done-list">
            <view class="done-choice" @tap="finishDone('已合作')">
              <view class="done-copy">
                <text class="done-name">已合作</text>
                <text class="done-desc">双方已达成线下合作</text>
              </view>
              <text class="done-arrow">›</text>
            </view>
            <view class="done-choice" @tap="finishDone('未合作')">
              <view class="done-copy">
                <text class="done-name">未合作</text>
                <text class="done-desc">本次未达成合作</text>
              </view>
              <text class="done-arrow">›</text>
            </view>
            <view class="done-choice" @tap="finishDone('洽谈中')">
              <view class="done-copy">
                <text class="done-name">继续洽谈</text>
                <text class="done-desc">保留会话，后续更新结果</text>
              </view>
              <text class="done-arrow">›</text>
            </view>
          </view>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { SEED_CHAT, getReceivedIntents, saveReceivedIntents } from '../../utils/hallData'

const partner = ref({ name: '重庆翼航科技', initial: '翼' })

try {
  const raw = uni.getStorageSync('hall_chat_partner')
  if (raw) partner.value = JSON.parse(raw)
} catch (e) { /* 保持默认 */ }

const messages = ref(SEED_CHAT.map((m) => ({ ...m })))
const draft = ref('')
const scrollInto = ref('')
const showDone = ref(false)

function sendMessage() {
  const text = draft.value.trim()
  if (!text) return
  messages.value.push({
    id: Date.now(),
    mine: true,
    avatar: '我',
    text,
  })
  draft.value = ''
  scrollToBottom()
}

function scrollToBottom() {
  nextTick(() => {
    const last = messages.value[messages.value.length - 1]
    if (last) scrollInto.value = 'msg-' + last.id
  })
}

const confirmDone = () => { showDone.value = true }

function finishDone(status) {
  showDone.value = false
  // 同步给收到的意向
  const intents = getReceivedIntents()
  if (intents.length) {
    intents[0].status = status
    saveReceivedIntents(intents)
  }
  uni.showToast({ title: status === '洽谈中' ? '继续洽谈' : '合作状态已更新', icon: 'success' })
}

const toastImage = () => uni.showToast({ title: '图片选择器将在正式版本接入', icon: 'none' })
const toastFile = () => uni.showToast({ title: '方案附件选择器将在正式版本接入', icon: 'none' })

const goBack = () => uni.navigateBack()
</script>

<style scoped>
.chat-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #F4F6F8;
  box-sizing: border-box;
}

.page-header {
  height: 56px;
  padding: 0 28rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  background: #fff;
  border-bottom: 1px solid #EEF1F4;
  flex-shrink: 0;
}
.back-btn { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; }
.back-sym { font-size: 52rpx; color: #17212B; line-height: 1; }
.page-title {
  flex: 1;
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.head-action { padding: 14rpx; }
.head-action-text { color: #0A66C2; font-size: 26rpx; font-weight: 600; white-space: nowrap; }

.chat-notice {
  padding: 18rpx 32rpx;
  background: #FFF5EA;
  color: #A94C10;
  font-size: 22rpx;
  line-height: 1.5;
  flex-shrink: 0;
}

.chat-body {
  flex: 1;
  min-height: 0;
}
.chat-inner {
  padding: 30rpx 32rpx;
  display: flex;
  flex-direction: column;
  gap: 26rpx;
}

.bubble-row { display: flex; gap: 16rpx; align-items: flex-start; }
.bubble-row.mine { justify-content: flex-end; }
.bubble-avatar {
  width: 58rpx;
  height: 58rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #EAF3FB;
  color: #0A66C2;
  font-size: 20rpx;
  font-weight: 700;
  flex-shrink: 0;
}
.bubble-wrap { max-width: 73%; display: flex; flex-direction: column; align-items: flex-start; }
.mine .bubble-wrap { align-items: flex-end; }
.bubble {
  padding: 18rpx 22rpx;
  border-radius: 4px 16rpx 16rpx 16rpx;
  background: #fff;
  color: #344054;
  font-size: 24rpx;
  line-height: 1.65;
  box-shadow: 0 2px 7px rgba(16, 24, 40, 0.04);
}
.mine .bubble {
  border-radius: 16rpx 4px 16rpx 16rpx;
  color: #fff;
  background: #0A66C2;
}
.attachment {
  display: flex;
  align-items: center;
  gap: 10rpx;
  margin-top: 12rpx;
  padding: 12rpx;
  border-radius: 10rpx;
  background: rgba(255, 255, 255, 0.13);
  font-size: 22rpx;
}
.mine .attachment { background: rgba(255, 255, 255, 0.16); }
.attach-sym { color: currentColor; }
.attach-name { color: currentColor; }

/* 输入区 */
.chat-input {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 18rpx 24rpx calc(18rpx + env(safe-area-inset-bottom));
  background: #fff;
  border-top: 1px solid #EEF1F4;
  flex-shrink: 0;
}
.tool-btn { padding: 8rpx; }
.tool-sym { font-size: 34rpx; color: #667085; line-height: 1; }
.input {
  flex: 1;
  min-width: 0;
  height: 68rpx;
  border: 0;
  border-radius: 12rpx;
  background: #F3F5F7;
  padding: 0 20rpx;
  font-size: 24rpx;
  color: #17212B;
}
.send-btn { padding: 0 12rpx; }
.send-btn text { color: #0A66C2; font-size: 26rpx; font-weight: 700; }

/* 弹层 */
.sheet { padding-bottom: 20rpx; }
.sheet-head { display: flex; align-items: center; padding: 28rpx 32rpx 20rpx; }
.sheet-title { flex: 1; font-size: 32rpx; font-weight: 700; color: #17212B; }
.sheet-close { width: 56rpx; height: 56rpx; display: flex; align-items: center; justify-content: center; }
.sheet-x { font-size: 40rpx; color: #98A2B3; line-height: 1; }
.sheet-body { padding: 0 32rpx 24rpx; }
.sheet-desc {
  display: block;
  font-size: 26rpx;
  color: #667085;
  line-height: 1.7;
}
.done-list { margin-top: 24rpx; display: flex; flex-direction: column; gap: 20rpx; }
.done-choice {
  display: flex;
  align-items: center;
  gap: 20rpx;
  padding: 26rpx;
  border: 1px solid #E4E7EC;
  border-radius: 16rpx;
  background: #fff;
}
.done-copy { flex: 1; }
.done-name { display: block; font-size: 28rpx; font-weight: 700; color: #17212B; }
.done-desc { display: block; font-size: 22rpx; color: #667085; margin-top: 6rpx; }
.done-arrow { font-size: 40rpx; color: #98A2B3; }
</style>
