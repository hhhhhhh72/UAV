<template>
  <view class="tabbar-root">
    <view class="tabbar-placeholder"></view>
    <view class="tabbar">
      <view 
        v-for="(item, index) in tabList" 
        :key="index"
        class="tabbar-item"
        :class="{ active: current === index }"
        @tap="handleTabClick(index)"
      >
        <view class="tabbar-icon-wrap">
          <view 
            class="tab-icon" 
            :class="'icon-' + index"
            :style="{ backgroundColor: current === index ? '#1a1a1a' : '#969799' }"
          ></view>
        </view>
        <text class="tabbar-text">{{ item.text }}</text>
      </view>
    </view>
  </view>
</template>

<script setup>
const props = defineProps(['current'])

const tabList = [
  { text: '服务大厅', pagePath: '/pages/home/index' },
  { text: '全部服务', pagePath: '/pages/services/index' },
  { text: '我的申请', pagePath: '/pages/applications/index' },
  { text: '个人中心', pagePath: '/pages/mine/index' }
]

const handleTabClick = (index) => {
  if (props.current === index) return
  uni.switchTab({ url: tabList[index].pagePath })
}
</script>

<style scoped>
.tabbar-placeholder {
  height: calc(50px + env(safe-area-inset-bottom));
  width: 100%;
}
.tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 50px;
  background: #ffffff;
  display: flex;
  padding-bottom: env(safe-area-inset-bottom);
  border-top: 1px solid #f2f3f5;
  z-index: 999;
}
.tabbar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.tabbar-icon-wrap {
  width: 22px;
  height: 22px;
  margin-bottom: 4px;
}
.tab-icon {
  width: 100%;
  height: 100%;
  mask-size: contain;
  mask-repeat: no-repeat;
  mask-position: center;
  -webkit-mask-size: contain;
  -webkit-mask-repeat: no-repeat;
  -webkit-mask-position: center;
}
.icon-0 {
  mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik05MjIuMyA0MzAuM0w1NDIuNCA5MC43YTQyLjk0IDQyLjk0IDAgMCAwLTU4LjggMEwxMDMuNyA0MzAuM2MtMjMuNyAyMS4yLTguNiA2MC4xIDIzLjMgNjAuMWg1My4xdjM4Mi43YzAgMzAuNiAyNC44IDU1LjQgNTUuNCA1NS40aDE1OS4zVjY3Mi40aDI1Ni42VjkyOC41aDE1OS4zYzMwLjYgMCA1NS40LTI0LjggNTUuNC01NS40VjQ5MC40aDUxYzMxLjggMCA0Ni45LTM4LjkgMjMuMi02MC4xeiIvPjwvc3ZnPg==");
  -webkit-mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik05MjIuMyA0MzAuM0w1NDIuNCA5MC43YTQyLjk0IDQyLjk0IDAgMCAwLTU4LjggMEwxMDMuNyA0MzAuM2MtMjMuNyAyMS4yLTguNiA2MC4xIDIzLjMgNjAuMWg1My4xdjM4Mi43YzAgMzAuNiAyNC44IDU1LjQgNTUuNCA1NS40aDE1OS4zVjY3Mi40aDI1Ni42VjkyOC41aDE1OS4zYzMwLjYgMCA1NS40LTI0LjggNTUuNC01NS40VjQ5MC40aDUxYzMxLjggMCA0Ni45LTM4LjkgMjMuMi02MC4xeiIvPjwvc3ZnPg==");
}
.icon-1 {
  mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik05MTIgMTkySDExMmMtMjYuNSAwLTQ4IDIxLjUtNDggNDh2NTQ0YzAgMjYuNSAyMS41IDQ4IDQ4IDQ4aDgwMGMyNi41IDAgNDgtMjEuNSA0OC00OFYyNDBjMC0yNi41LTIxLjUtNDgtNDgtNDh6bS02NCA1NDRIMTc2VjMyMGg2NzJ2NDE2ek0zMjAgNDE2SDIyNHY5Nmg5NnYtOTZ6bTAgMTYwSDIyNHY5Nmg5NnYtOTZ6bTE5Mi0xNjBoLTk2djk2aDk2di05NnptMCAxNjBoLTk2djk2aDk2di05NnptMTkyLTE2MGgtOTZ2OTZoOTZ2LTk2em0wIDE2MGgtOTZ2OTZoOTZ2LTk2eiIvPjwvc3ZnPg==");
  -webkit-mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik05MTIgMTkySDExMmMtMjYuNSAwLTQ4IDIxLjUtNDggNDh2NTQ0YzAgMjYuNSAyMS41IDQ4IDQ4IDQ4aDgwMGMyNi41IDAgNDgtMjEuNSA0OC00OFYyNDBjMC0yNi41LTIxLjUtNDgtNDgtNDh6bS02NCA1NDRIMTc2VjMyMGg2NzJ2NDE2ek0zMjAgNDE2SDIyNHY5Nmg5NnYtOTZ6bTAgMTYwSDIyNHY5Nmg5NnYtOTZ6bTE5Mi0xNjBoLTk2djk2aDk2di05NnptMCAxNjBoLTk2djk2aDk2di05NnptMTkyLTE2MGgtOTZ2OTZoOTZ2LTk2em0wIDE2MGgtOTZ2OTZoOTZ2LTk2eiIvPjwvc3ZnPg==");
}
.icon-2 {
  mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik04MzIgMTI4SDE5MmMtMzUuMyAwLTY0IDI4LjctNjQgNjR2NjQwYzAgMzUuMyAyOC43IDY0IDY0IDY0aDY0MGMzNS4zIDAgNjQtMjguNyA2NC02NFYxOTJjMC0zNS4zLTI4LjctNjQtNjQtNjR6TTM1MiA3MzZoLTY0di02NGg2NHY2NHptMC0xNjBoLTY0di02NGg2NHY2NHptMC0xNjBoLTY0di02NGg2NHY2NHptNDE2IDMyMEg0MTZ2LTY0aDM1MnY2NHptMC0xNjBINDE2di02NGgzNTJ2NjR6bTAtMTYwSDQxNnYtNjRoMzUydjY0eiIvPjwvc3ZnPg==");
  -webkit-mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik04MzIgMTI4SDE5MmMtMzUuMyAwLTY0IDI4LjctNjQgNjR2NjQwYzAgMzUuMyAyOC43IDY0IDY0IDY0aDY0MGMzNS4zIDAgNjQtMjguNyA2NC02NFYxOTJjMC0zNS4zLTI4LjctNjQtNjQtNjR6TTM1MiA3MzZoLTY0di02NGg2NHY2NHptMC0xNjBoLTY0di02NGg2NHY2NHptMC0xNjBoLTY0di02NGg2NHY2NHptNDE2IDMyMEg0MTZ2LTY0aDM1MnY2NHptMC0xNjBINDE2di02NGgzNTJ2NjR6bTAtMTYwSDQxNnYtNjRoMzUydjY0eiIvPjwvc3ZnPg==");
}
.icon-3 {
  mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik01MTIgNTEyYTE5MiAxOTIgMCAxIDAgMC0zODQgMTkyIDE5MiAwIDAgMCAwIDM4NHptMCA2NGMtMTMxLjMgMC0zMjAgNjUuOC0zMjAgMTkydjY0aDY0MHYtNjRjMC0xMjYuMi0xODguNy0xOTItMzIwLTE5MnoiLz48L3N2Zz4=");
  -webkit-mask-image: url("data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjxwYXRoIGQ9Ik01MTIgNTEyYTE5MiAxOTIgMCAxIDAgMC0zODQgMTkyIDE5MiAwIDAgMCAwIDM4NHptMCA2NGMtMTMxLjMgMC0zMjAgNjUuOC0zMjAgMTkydjY0aDY0MHYtNjRjMC0xMjYuMi0xODguNy0xOTItMzIwLTE5MnoiLz48L3N2Zz4=");
}
.tabbar-text { font-size: 10px; color: #969799; }
.active .tabbar-text { color: #1a1a1a; font-weight: bold; }
</style>
