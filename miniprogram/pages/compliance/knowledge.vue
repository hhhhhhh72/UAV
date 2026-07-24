<template>
  <view class="knowledge-page">
    <van-nav-bar
      title="合规知识库"
      fixed
      placeholder
      left-arrow
      @click-left="goBack"
    />

    <!-- Loading state -->
    <view v-if="loading" class="state-view">
      <van-loading size="24">加载中...</van-loading>
    </view>

    <!-- Error state -->
    <view v-else-if="errorMsg" class="state-view">
      <van-empty description="加载失败" image="error" />
      <view class="retry-btn" @tap="fetchAll">
        <text>重新加载</text>
      </view>
    </view>

    <!-- Empty state -->
    <view v-else-if="!loading && allEmpty" class="state-view">
      <van-empty image="search" description="暂无合规知识文档" />
    </view>

    <!-- Normal: collapse accordion -->
    <view v-else class="collapse-body">
      <van-collapse
        accordion
        :value="activeCollapse"
        @change="onCollapseChange"
      >
        <van-collapse-item
          v-for="section in sections"
          :key="section.key"
          :title="section.title"
          :name="section.key"
        >
          <view v-if="section.loading" class="section-loading">
            <van-loading size="20">加载中...</van-loading>
          </view>

          <view v-else-if="section.error" class="section-error">
            <text class="section-error-text">加载失败</text>
            <text class="section-retry" @tap="fetchSection(section)">重试</text>
          </view>

          <view v-else-if="section.list.length === 0" class="section-empty">
            <van-empty description="暂无文档" image="search" />
          </view>

          <van-cell-group v-else inset>
            <van-cell
              v-for="doc in section.list"
              :key="doc.id"
              :title="doc.title || '--'"
              :label="doc.description || ''"
              is-link
              @tap="openDoc(doc)"
            >
              <template #title>
                <view class="doc-title-row">
                  <text class="doc-icon">📄</text>
                  <text class="doc-title-text">{{ doc.title || '--' }}</text>
                </view>
              </template>
            </van-cell>
          </van-cell-group>
        </van-collapse-item>
      </van-collapse>
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
    onCollapseChange(e) {
      var name = e.detail
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
      if (doc.content) {
        uni.showModal({
          title: doc.title || '文档内容',
          content: doc.content,
          showCancel: false,
          confirmText: '知道了',
        })
      } else {
        uni.showToast({ title: '即将上线', icon: 'none' })
      }
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
  background: #f7f8fa;
  padding-bottom: env(safe-area-inset-bottom);
}

/* State views */
.state-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 120px;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #1989fa;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
}

/* Collapse body */
.collapse-body {
  padding: 12px 0 24px;
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
  color: #969799;
  margin-bottom: 4px;
}

.section-retry {
  font-size: 13px;
  color: #1989fa;
}

/* Doc list */
.doc-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.doc-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.doc-title-text {
  font-size: 14px;
  font-weight: 500;
  color: #323233;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
