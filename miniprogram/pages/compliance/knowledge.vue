<template>
  <view class="knowledge-page">
    <u-nav-bar
      title="合规知识库"
      show-back
      @back="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="state-view">
      <view class="loading-inline">
        <u-loading size="24rpx" />
        <text>加载中...</text>
      </view>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="state-view">
      <u-empty description="加载失败" />
      <view class="retry-btn" @tap="fetchAll">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && allEmpty" class="state-view">
      <u-empty description="暂无合规知识文档" />
    </view>

    <!-- Normal: collapse accordion (CSS 实现) -->
    <view v-else class="collapse-list">
      <view
        v-for="section in sections"
        :key="section.key"
        class="collapse-item"
      >
        <view class="collapse-header" @tap="onCollapseChange(section.key)">
          <text class="collapse-title">{{ section.title }}</text>
          <text class="collapse-arrow" :class="{ expanded: activeCollapse === section.key }">›</text>
        </view>

        <view v-if="activeCollapse === section.key" class="collapse-panel">
          <view v-if="section.loading" class="section-loading">
            <view class="loading-inline">
              <u-loading size="20rpx" />
              <text>加载中...</text>
            </view>
          </view>

          <view v-else-if="section.error" class="section-error">
            <text class="section-error-text">加载失败</text>
            <text class="section-retry" @tap="fetchSection(section)">重试</text>
          </view>

          <view v-else-if="section.list.length === 0" class="section-empty">
            <u-empty description="暂无文档" />
          </view>

          <u-cell-group v-else inset>
            <u-cell
              v-for="doc in section.list"
              :key="doc.id"
              :label="doc.description || ''"
              is-link
              @click="openDoc(doc)"
            >
              <template #title>
                <view class="doc-title-row">
                  <text class="doc-icon">文</text>
                  <text class="doc-title-text">{{ doc.title || '--' }}</text>
                </view>
              </template>
            </u-cell>
          </u-cell-group>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { request } from '../../utils/request'

var sectionConfigs = [
  { key: 'registration', title: '实名登记', value: 'real_name' },
  { key: 'flight_report', title: '飞行报备', value: 'flight_report' },
  { key: 'operation_license', title: '经营资质', value: 'operation_license' },
  { key: 'airworthiness', title: '适航指南', value: 'airworthiness' },
  { key: 'no_fly', title: '禁飞区域', value: 'no_fly_zone' },
]

export default {
  data() {
    var sections = sectionConfigs.map(function (cfg) {
      return {
        key: cfg.key,
        title: cfg.title,
        value: cfg.value,
        list: [],
        loading: false,
        error: false,
        loaded: false,
      }
    })

    return {
      loading: false,
      errorMsg: '',
      sections: sections,
      activeCollapse: '',
    }
  },
  computed: {
    allEmpty() {
      return this.sections.every(function (s) {
        return s.loaded && s.list.length === 0 && !s.error
      })
    },
  },
  onLoad() {
    this.fetchAll()
  },
  methods: {
    async fetchAll() {
      this.loading = true
      this.errorMsg = ''
      try {
        // Fetch all sections in parallel
        var promises = this.sections.map(function (section) {
          return this.fetchSection(section)
        }.bind(this))
        await Promise.all(promises)
      } catch (e) {
        this.errorMsg = '网络异常，请稍后重试'
      } finally {
        this.loading = false
      }
    },
    async fetchSection(section) {
      if (section.loaded && section.list.length > 0) return

      section.loading = true
      section.error = false

      try {
        var res = await request({
          url: '/api/v1/compliance-docs',
          data: { category: section.value },
        })
        var data = Array.isArray(res) ? res : (res && res.items) || []
        section.list = Array.isArray(data) ? data : (data && data.items) || []
        section.loaded = true
      } catch (e) {
        section.error = true
      } finally {
        section.loading = false
      }
    },
    onCollapseChange(name) {
      this.activeCollapse = name
      // Fetch data when section is expanded
      if (name && name.length > 0) {
        var section = this.sections.find(function (s) {
          return s.key === name
        })
        if (section && !section.loaded && !section.loading) {
          this.fetchSection(section)
        }
      }
    },
    openDoc(doc) {
      uni.showModal({
        title: doc.title || '文档内容',
        content: doc.content || doc.description || '暂无详细内容',
        showCancel: false,
        confirmText: '知道了',
      })
    },
    goBack() {
      uni.navigateBack()
    },
  },
}
</script>

<style scoped>
.knowledge-page {
  min-height: 100vh;
  background: var(--color-bg);
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.loading-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Collapse accordion (CSS 实现) */
.collapse-list {
  padding: 12px 16px 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.collapse-item {
  background: var(--color-bg-card);
  border-radius: 16px;
  overflow: hidden;
}

.collapse-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 16px;
}

.collapse-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
}

.collapse-arrow {
  font-size: 20px;
  color: var(--color-text-placeholder);
  transition: transform 0.2s;
}

.collapse-arrow.expanded {
  transform: rotate(90deg);
  color: var(--color-primary);
}

.collapse-panel {
  border-top: 1rpx solid var(--color-divider);
  padding: 16px 16px 8px;
}

/* Section states */
.section-loading {
  display: flex;
  justify-content: center;
  padding: 24px 0;
}

.section-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0;
}

.section-error-text {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-bottom: 4px;
}

.section-retry {
  font-size: 13px;
  color: var(--color-primary);
}

/* Doc list */
.doc-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.doc-icon {
  width: 40rpx;
  height: 40rpx;
  line-height: 40rpx;
  text-align: center;
  font-size: 24rpx;
  color: var(--color-primary);
  background: var(--color-primary-light);
  border-radius: 8rpx;
  flex-shrink: 0;
}

.doc-title-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
