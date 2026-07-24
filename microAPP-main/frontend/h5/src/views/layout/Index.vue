<template>
  <div class="layout">
    <router-view />
    <van-tabbar
      v-if="showTabbar"
      route
      active-color="#000000"
      inactive-color="#969799"
    >
      <van-tabbar-item to="/home">
        <span>服务大厅</span>
        <template #icon="props">
          <van-icon :name="props.active ? 'wap-home' : 'wap-home-o'" />
        </template>
      </van-tabbar-item>

      <van-tabbar-item to="/services">
        <span>全部服务</span>
        <template #icon="props">
          <van-icon :name="props.active ? 'shop' : 'shop-o'" />
        </template>
      </van-tabbar-item>

      <van-tabbar-item to="/applications">
        <span>我的申请</span>
        <template #icon="props">
          <van-icon :name="props.active ? 'todo-list' : 'todo-list-o'" />
        </template>
      </van-tabbar-item>

      <van-tabbar-item to="/mine">
        <span>个人中心</span>
        <template #icon="props">
          <van-icon :name="props.active ? 'user' : 'user-o'" />
        </template>
      </van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const showTabbar = computed(() => route.meta.showTabbar)
</script>

<style scoped>
.layout {
  min-height: 100vh;
  background-color: #f5f5f5;
}

/* 全局底部栏：更高、更大点击热区，且在大屏下保持“手机宽度居中” */
:deep(.van-tabbar) {
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: var(--page-max-width);

  height: 74px;
  padding-top: 6px;
  padding-bottom: calc(10px + env(safe-area-inset-bottom));
  box-sizing: content-box;
}

:deep(.van-tabbar-item) {
  padding: 10px 0;
}

:deep(.van-tabbar-item__icon) {
  margin-bottom: 6px;
  font-size: 22px;
}

:deep(.van-tabbar-item__text) {
  font-size: 12px;
}
</style>


