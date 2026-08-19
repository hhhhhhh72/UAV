<template>
  <!-- v20250324-1: 课程包独立字段管理（Arco Design Vue 版，脱离 Vant） -->
  <div class="config-page">
    <DataToolbar>
      <template #filters>
        <span class="toolbar-label">服务配置</span>
      </template>
      <template #actions>
        <a-button size="small" @click="fetchAllServiceConfigs">
          <template #icon><icon-refresh /></template>刷新
        </a-button>
      </template>
    </DataToolbar>

    <!-- 首页配置（仅管理员可见） -->
    <a-card v-if="isPlatformAdmin" :bordered="false" class="config-card">
      <template #title><span class="card-title">首页配置</span></template>
      <div class="link-row" @click="editHomeConfig">
        <div class="link-row-main">
          <div class="link-row-title">首页背景图 &amp; 轮播配置</div>
          <div class="link-row-label">{{ (homeConfig?.headerImage ? '背景图已配置' : '背景图未配置') + '  ·  ' + (homeConfig?.banners?.length || 0) + ' 张Banner  ·  ' + (homeConfig?.notices?.length || 0) + ' 条轮播消息' }}</div>
        </div>
        <icon-arrow-right class="link-row-arrow" />
      </div>
    </a-card>

    <!-- 商业化费率（功能方案修订版 v2 第八章，仅管理员可见） -->
    <a-card v-if="isPlatformAdmin" :bordered="false" class="config-card">
      <template #title><span class="card-title">商业化费率</span></template>
      <div class="fee-row">
        <div class="fee-field">
          <span class="fee-label">撮合服务费率（%）</span>
          <a-input-number v-model="platformConfig.match_fee_rate" :min="0" :max="100" :precision="1" :step="0.5" size="small" hide-button />
        </div>
        <div class="fee-field">
          <span class="fee-label">费率说明（对会员展示）</span>
          <a-input v-model="platformConfig.match_fee_note" size="small" placeholder="如：供需对接成功后按 2% 收取撮合服务费" />
        </div>
        <a-button size="small" type="primary" @click="savePlatformConfig">保存费率</a-button>
      </div>
      <div class="fee-hint">0 表示未启用收费；保存后写入平台配置（/api/v1/admin/config）。</div>
    </a-card>

    <!-- 服务列表（按分区） -->
    <a-card
      v-for="group in groupedServiceEntries"
      :key="group.title"
      :bordered="false"
      class="config-card"
    >
      <template #title><span class="card-title">{{ group.title }}</span></template>
      <div
        v-for="[id, cfg] in group.items"
        :key="id"
        class="link-row"
        @click="editServiceConfig(id)"
      >
        <div class="link-row-main">
          <div class="link-row-title">{{ cfg.name }}</div>
        </div>
        <icon-arrow-right class="link-row-arrow" />
      </div>
    </a-card>

    <!-- ========== 服务编辑弹窗 ========== -->
    <a-modal
      v-model:visible="showServiceEditPopup"
      :title="editingService?.name || '服务配置'"
      :width="860"
      :top="'4vh'"
      :unmount-on-close="true"
      :mask-closable="false"
      :on-before-cancel="guardServiceEditClose"
    >
      <div class="dialog-body" v-if="editingService">
        <div class="config-section">
          <div class="section-title">服务介绍</div>
          <a-input
            v-model="editingService.intro"
            type="textarea"
            :auto-size="{ minRows: 3 }"
            placeholder="请输入服务介绍，建议50-100字"
            :max-length="200"
            show-word-limit
          />
        </div>

        <div class="config-section">
          <div class="section-title">联系方式</div>
          <div class="field-row">
            <span class="field-label">电话</span>
            <a-input v-model="editingService.contactPhone" placeholder="联系电话" />
          </div>
          <div class="field-row">
            <span class="field-label">热线</span>
            <a-input v-model="editingService.contactPhone2" placeholder="选填，第二个电话" />
          </div>
          <div class="field-row">
            <span class="field-label">地址</span>
            <a-input v-model="editingService.address" placeholder="选填" type="textarea" :auto-size="{ minRows: 1 }" />
          </div>
        </div>

        <!-- 服务项目（非研学服务显示） -->
        <div class="config-section" v-if="editingServiceId !== '9'">
          <div class="section-title">服务项目（{{ editingService.projects?.length || 0 }}）</div>
          <div v-for="(p, idx) in editingService.projects" :key="idx" class="list-item">
            <a-input v-model="p.name" placeholder="项目名称" />
            <a-button type="text" status="danger" size="small" @click="editingService.projects.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
          </div>
          <div class="list-add">
            <a-button type="outline" size="small" @click="editingService.projects.push({ name: '', icon: 'star-o' })"><template #icon><icon-plus /></template>添加项目</a-button>
          </div>
        </div>

        <!-- 服务优势（非研学服务显示） -->
        <div class="config-section" v-if="editingServiceId !== '9'">
          <div class="section-title">服务优势（{{ editingService.advantages?.length || 0 }}）</div>
          <div v-for="(adv, idx) in editingService.advantages" :key="idx" class="list-item">
            <a-input v-model="editingService.advantages[idx]" placeholder="优势描述" />
            <a-button type="text" status="danger" size="small" @click="editingService.advantages.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
          </div>
          <div class="list-add">
            <a-button type="outline" size="small" @click="editingService.advantages.push('')"><template #icon><icon-plus /></template>添加优势</a-button>
          </div>
        </div>

        <!-- 背景图（仅培训显示） -->
        <div class="config-section" v-if="editingServiceId === '6'">
          <div class="section-title">页面背景图</div>
          <div class="field-row">
            <span class="field-label">上传图片</span>
            <a-upload :show-file-list="false" :custom-request="wrapUpload(file => onReadServiceFile(file, 'headerImage'))">
              <a-button size="small"><template #icon><icon-plus /></template>选择图片</a-button>
            </a-upload>
          </div>
          <div v-if="editingService.headerImage" class="img-preview">
            <img :src="normalizeMediaUrl(editingService.headerImage)" />
          </div>
        </div>

        <!-- 培训亮点（仅培训） -->
        <div class="config-section" v-if="editingServiceId === '6'">
          <div class="section-title">培训亮点（{{ editingService.highlights?.length || 0 }}）</div>
          <div v-for="(hl, idx) in editingService.highlights" :key="idx" class="list-item-block">
            <div class="list-item-head">
              <span class="item-num">{{ idx + 1 }}</span>
              <a-button type="text" status="danger" size="small" @click="editingService.highlights.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
            </div>
            <div class="field-row">
              <span class="field-label">标题</span>
              <a-input v-model="hl.title" placeholder="如：官方认证" />
            </div>
            <div class="field-row">
              <span class="field-label">描述</span>
              <a-input v-model="hl.desc" placeholder="如：CAAC民航局授权" />
            </div>
          </div>
          <div class="list-add">
            <a-button type="outline" size="small" @click="editingService.highlights.push({ title: '', desc: '', icon: 'star-o' })"><template #icon><icon-plus /></template>添加亮点</a-button>
          </div>
        </div>

        <!-- 图文展示（仅培训显示，研学在课程包内管理） -->
        <div class="config-section" v-if="editingServiceId === '6'">
          <div class="section-title">图文展示（{{ editingService.studyShowcase?.length || 0 }}）</div>
          <div v-for="(item, idx) in editingService.studyShowcase" :key="idx" class="showcase-item">
            <div class="showcase-left">
              <img v-if="item.image" :src="normalizeMediaUrl(item.image)" class="showcase-img" />
              <icon-image v-else :size="24" :style="{ color: 'var(--color-text-3)' }" />
            </div>
            <div class="showcase-mid">{{ item.title || '未命名' }}</div>
            <a-button size="small" type="outline" @click="innerEditStudyItem(idx)">编辑</a-button>
            <a-button type="text" status="danger" size="small" @click="editingService.studyShowcase.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
          </div>
          <div class="list-add">
            <a-button type="outline" size="small" @click="innerAddStudyItem"><template #icon><icon-plus /></template>添加展示</a-button>
          </div>
        </div>

        <!-- ===== 研学课程包管理（仅研学） ===== -->
        <template v-if="editingServiceId === '9'">
          <div class="config-section">
            <div class="section-title">课程包管理</div>
            <div class="pkg-tabs">
              <div
                v-for="pkgId in studyPackageIds"
                :key="pkgId"
                class="pkg-tab"
                :class="{ active: activeStudyPkgId === pkgId }"
                @click="activeStudyPkgId = pkgId"
              >
                {{ studyPackages[pkgId]?.tag || pkgId }} (¥{{ studyPackages[pkgId]?.price || '' }})
                <icon-close class="pkg-tab-close" @click.stop="removeStudyPackage(pkgId)" />
              </div>
              <div class="pkg-tab pkg-tab-add" @click="showAddPackageDialog">
                <icon-plus />
              </div>
            </div>
          </div>

          <template v-if="activeStudyPkg">
            <!-- 基本信息 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 基本信息</div>
              <div class="field-row">
                <span class="field-label">课程名称</span>
                <a-input v-model="activeStudyPkg.name" placeholder="如：无人机研学实践中心半日营" />
              </div>
              <div class="field-row">
                <span class="field-label">标签</span>
                <a-input v-model="activeStudyPkg.tag" placeholder="如：半日营" />
              </div>
              <div class="field-row">
                <span class="field-label">票价</span>
                <a-input v-model="activeStudyPkg.price" placeholder="198" />
              </div>

              <!-- 头部背景 - 支持渐变或图片 -->
              <div class="field-row">
                <span class="field-label">头部背景</span>
                <div class="bg-input-wrap">
                  <a-input v-model="activeStudyPkg.headerBg" placeholder="渐变色或图片URL" />
                  <a-upload :show-file-list="false" :custom-request="wrapUpload(file => onReadPackageImage(file, 'headerBg'))" accept="image/*">
                    <a-button size="small"><template #icon><icon-image /></template>上传</a-button>
                  </a-upload>
                </div>
              </div>
              <div v-if="activeStudyPkg.headerBg" class="img-preview">
                <img :src="normalizeMediaUrl(activeStudyPkg.headerBg)" @click="previewCrop(normalizeMediaUrl(activeStudyPkg.headerBg), 'headerBg')" />
                <div class="img-actions">
                  <a-button size="small" type="outline" @click="previewCrop(normalizeMediaUrl(activeStudyPkg.headerBg), 'headerBg')">裁剪</a-button>
                </div>
              </div>

              <div class="field-row">
                <span class="field-label">介绍</span>
                <a-input v-model="activeStudyPkg.intro" type="textarea" :auto-size="{ minRows: 3 }" placeholder="课程介绍" />
              </div>
            </div>

            <!-- 服务项目（每个课程包独立） -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 服务项目（{{ activeStudyPkg.projects?.length || 0 }}）</div>
              <div v-for="(p, idx) in activeStudyPkg.projects" :key="idx" class="list-item-block">
                <div class="list-item-head">
                  <span class="item-num">{{ idx + 1 }}</span>
                  <a-button type="text" status="danger" size="small" @click="activeStudyPkg.projects.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
                </div>
                <div class="field-row">
                  <span class="field-label">名称</span>
                  <a-input v-model="p.name" placeholder="如：展厅参观" />
                </div>
                <div class="field-row">
                  <span class="field-label">图标</span>
                  <a-input v-model="p.icon" placeholder="Vant图标名或图片URL" />
                </div>
                <div v-if="p.icon && p.icon.startsWith('http')" class="icon-preview">
                  <img :src="p.icon" style="width:40px;height:40px;border-radius:8px;object-fit:cover;" />
                </div>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.projects.push({ name: '', icon: 'star-o' })"><template #icon><icon-plus /></template>添加项目</a-button>
              </div>
            </div>

            <!-- 服务优势（每个课程包独立） -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 服务优势（{{ activeStudyPkg.advantages?.length || 0 }}）</div>
              <div v-for="(adv, idx) in activeStudyPkg.advantages" :key="idx" class="list-item">
                <a-input v-model="activeStudyPkg.advantages[idx]" placeholder="优势描述，如：专业导师全程指导" />
                <a-button type="text" status="danger" size="small" @click="activeStudyPkg.advantages.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.advantages.push('')"><template #icon><icon-plus /></template>添加优势</a-button>
              </div>
            </div>

            <!-- 精彩回顾（每个课程包独立） -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 精彩回顾（{{ activeStudyPkg.showcase?.length || 0 }}）</div>
              <div v-for="(item, idx) in activeStudyPkg.showcase" :key="idx" class="showcase-item">
                <div class="showcase-left">
                  <img v-if="item.image" :src="item.image" class="showcase-img" @click="previewCrop(item.image, 'showcase', idx)" />
                  <icon-image v-else :size="24" :style="{ color: 'var(--color-text-3)' }" />
                </div>
                <div class="showcase-mid">{{ item.title || '未命名' }}</div>
                <a-button size="small" type="outline" @click="editShowcaseItem(idx)">编辑</a-button>
                <a-button type="text" status="danger" size="small" @click="activeStudyPkg.showcase.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="addShowcaseItem"><template #icon><icon-plus /></template>添加展示</a-button>
              </div>
            </div>

            <!-- 课程安排 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 课程安排（{{ activeStudyPkg.schedule?.length || 0 }}）</div>
              <div v-for="(step, idx) in activeStudyPkg.schedule" :key="idx" class="list-item-block">
                <div class="list-item-head">
                  <span class="item-num">{{ idx + 1 }}</span>
                  <a-button type="text" status="danger" size="small" @click="activeStudyPkg.schedule.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
                </div>
                <div class="field-row">
                  <span class="field-label">上午</span>
                  <a-input v-model="step.amTime" placeholder="08:50" />
                </div>
                <div class="field-row">
                  <span class="field-label">下午</span>
                  <a-input v-model="step.pmTime" placeholder="13:50" />
                </div>
                <div class="field-row">
                  <span class="field-label">环节</span>
                  <a-input v-model="step.name" placeholder="集合签到" />
                </div>
                <div class="field-row">
                  <span class="field-label">说明</span>
                  <a-input v-model="step.desc" placeholder="研学中心集合，签到报到" />
                </div>
                <div class="field-row">
                  <span class="field-label">地点</span>
                  <a-input v-model="step.location" placeholder="研学中心A教室" />
                </div>
                <div class="field-row">
                  <span class="field-label">课程目的</span>
                  <a-input v-model="step.purpose" type="textarea" :auto-size="{ minRows: 2 }" placeholder="了解无人机基本原理" />
                </div>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.schedule.push({ amTime: '', pmTime: '', name: '', desc: '', location: '', purpose: '' })"><template #icon><icon-plus /></template>添加步骤</a-button>
              </div>
            </div>

            <!-- 研学目标 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 研学目标（{{ activeStudyPkg.studyGoals?.length || 0 }}）</div>
              <div v-for="(goal, idx) in activeStudyPkg.studyGoals" :key="idx" class="list-item-block">
                <div class="list-item-head">
                  <span class="item-num">{{ idx + 1 }}</span>
                  <a-button type="text" status="danger" size="small" @click="activeStudyPkg.studyGoals.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
                </div>
                <div class="field-row">
                  <span class="field-label">目标类型</span>
                  <a-input v-model="goal.label" placeholder="如：知识目标、能力目标" />
                </div>
                <div class="field-row">
                  <span class="field-label">具体内容</span>
                  <a-input v-model="goal.content" type="textarea" :auto-size="{ minRows: 3 }" placeholder="掌握无人机基本原理" />
                </div>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.studyGoals.push({ label: '', content: '' })"><template #icon><icon-plus /></template>添加目标</a-button>
              </div>
            </div>

            <!-- 安全宣讲 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 安全宣讲（{{ activeStudyPkg.safetyBriefing?.length || 0 }}）</div>
              <div v-for="(item, idx) in activeStudyPkg.safetyBriefing" :key="idx" class="list-item">
                <a-input v-model="activeStudyPkg.safetyBriefing[idx]" placeholder="如：操作前检查设备电量" />
                <a-button type="text" status="danger" size="small" @click="activeStudyPkg.safetyBriefing.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.safetyBriefing.push('')"><template #icon><icon-plus /></template>添加安全提示</a-button>
              </div>
            </div>

            <!-- 研学总结 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 研学总结</div>
              <a-input
                v-model="activeStudyPkg.studySummary"
                type="textarea"
                :auto-size="{ minRows: 4 }"
                placeholder="课程总结内容..."
                :max-length="500"
                show-word-limit
              />
            </div>

            <!-- 适合人群 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 适合人群（{{ activeStudyPkg.audience?.length || 0 }}）</div>
              <div v-for="(a, idx) in activeStudyPkg.audience" :key="idx" class="list-item">
                <a-input v-model="activeStudyPkg.audience[idx]" placeholder="如：6-16岁青少年" />
                <a-button type="text" status="danger" size="small" @click="activeStudyPkg.audience.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.audience.push('')"><template #icon><icon-plus /></template>添加</a-button>
              </div>
            </div>

            <!-- 费用说明 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 费用说明（{{ activeStudyPkg.feeInfo?.length || 0 }}）</div>
              <div v-for="(f, idx) in activeStudyPkg.feeInfo" :key="idx" class="list-item-block">
                <div class="list-item-head">
                  <span class="item-num">{{ idx + 1 }}</span>
                  <a-button type="text" status="danger" size="small" @click="activeStudyPkg.feeInfo.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
                </div>
                <div class="field-row">
                  <span class="field-label">项目</span>
                  <a-input v-model="f.label" placeholder="课程价格" />
                </div>
                <div class="field-row">
                  <span class="field-label">内容</span>
                  <a-input v-model="f.value" placeholder="¥198/人" />
                </div>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.feeInfo.push({ label: '', value: '' })"><template #icon><icon-plus /></template>添加</a-button>
              </div>
            </div>

            <!-- 温馨提示 -->
            <div class="config-section">
              <div class="section-title">{{ activeStudyPkg.tag }} - 温馨提示（{{ activeStudyPkg.tips?.length || 0 }}）</div>
              <div v-for="(t, idx) in activeStudyPkg.tips" :key="idx" class="list-item">
                <a-input v-model="activeStudyPkg.tips[idx]" placeholder="提示内容" />
                <a-button type="text" status="danger" size="small" @click="activeStudyPkg.tips.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
              </div>
              <div class="list-add">
                <a-button type="outline" size="small" @click="activeStudyPkg.tips.push('')"><template #icon><icon-plus /></template>添加</a-button>
              </div>
            </div>
          </template>
        </template>

        <!-- ===== 飞手培训专属 ===== -->
        <template v-if="editingServiceId === '6'">
          <div class="config-section">
            <div class="section-title">报名条件</div>
            <div v-for="(cond, idx) in trainingConditions" :key="idx" class="list-item">
              <a-input v-model="trainingConditions[idx]" placeholder="如：年满16周岁" />
              <a-button type="text" status="danger" size="small" @click="trainingConditions.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
            </div>
            <div class="list-add">
              <a-button type="outline" size="small" @click="trainingConditions.push('')"><template #icon><icon-plus /></template>添加条件</a-button>
            </div>
          </div>

          <div class="config-section">
            <div class="section-title">培训费用</div>
            <div v-for="(p, idx) in trainingPrices" :key="idx" class="list-item-block">
              <div class="list-item-head">
                <span class="item-num">{{ idx + 1 }}</span>
                <a-button type="text" status="danger" size="small" @click="trainingPrices.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
              </div>
              <div class="field-row">
                <span class="field-label">项目</span>
                <a-input v-model="p.label" placeholder="视距内驾驶员" />
              </div>
              <div class="field-row">
                <span class="field-label">价格</span>
                <a-input v-model="p.price" placeholder="¥4,800起" />
              </div>
            </div>
            <div class="list-add">
              <a-button type="outline" size="small" @click="trainingPrices.push({ label: '', price: '' })"><template #icon><icon-plus /></template>添加</a-button>
            </div>
          </div>

          <div class="config-section">
            <div class="section-title">教学特色</div>
            <div v-for="(f, idx) in trainingFeatures" :key="idx" class="list-item-block">
              <div class="list-item-head">
                <span class="item-num">{{ idx + 1 }}</span>
                <a-button type="text" status="danger" size="small" @click="trainingFeatures.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
              </div>
              <div class="field-row">
                <span class="field-label">标题</span>
                <a-input v-model="f.title" placeholder="小班制教学" />
              </div>
              <div class="field-row">
                <span class="field-label">描述</span>
                <a-input v-model="f.desc" type="textarea" :auto-size="{ minRows: 2 }" placeholder="详细说明" />
              </div>
            </div>
            <div class="list-add">
              <a-button type="outline" size="small" @click="trainingFeatures.push({ title: '', desc: '' })"><template #icon><icon-plus /></template>添加</a-button>
            </div>
          </div>

          <div class="config-section">
            <div class="section-title">公司简介</div>
            <div class="field-row">
              <span class="field-label">公司名</span>
              <a-input v-model="trainingCompanyTitle" placeholder="温州低空科技集团" />
            </div>
            <div class="field-row">
              <span class="field-label">简介</span>
              <a-input v-model="trainingCompanyContent" type="textarea" :auto-size="{ minRows: 3 }" placeholder="公司介绍" />
            </div>
          </div>

          <div class="config-section">
            <div class="section-title">执照说明</div>
            <div class="field-row">
              <span class="field-label">说明</span>
              <a-input v-model="trainingLicenseContent" type="textarea" :auto-size="{ minRows: 2 }" placeholder="执照功能介绍" />
            </div>
            <div class="field-row">
              <span class="field-label">法规</span>
              <a-input v-model="trainingLicenseQuote" type="textarea" :auto-size="{ minRows: 2 }" placeholder="法规条文引用" />
            </div>
          </div>
        </template>
      </div>
      <template #footer>
        <a-button @click="guardServiceEditClose">取消</a-button>
        <a-button type="primary" @click="saveServiceConfig">保存</a-button>
      </template>
    </a-modal>

    <!-- 图文展示编辑子弹窗 -->
    <a-modal
      v-model:visible="showStudyItemEditPopup"
      title="编辑图文"
      :width="560"
      :unmount-on-close="true"
      :on-before-cancel="guardStudyItemClose"
    >
      <div class="dialog-body" v-if="studyEditingItem">
        <div class="field-row">
          <span class="field-label">标题</span>
          <a-input v-model="studyEditingItem.title" placeholder="请输入标题" />
        </div>
        <div class="field-row">
          <span class="field-label">描述</span>
          <a-input v-model="studyEditingItem.desc" type="textarea" :auto-size="{ minRows: 2 }" placeholder="请输入描述" />
        </div>
        <div class="field-row">
          <span class="field-label">图片</span>
          <a-upload :show-file-list="false" :custom-request="wrapUpload(file => onReadStudyImage(file))" accept="image/*">
            <a-button size="small"><template #icon><icon-plus /></template>上传图片</a-button>
          </a-upload>
        </div>
        <div v-if="studyEditingItem.image" style="margin-top: 8px;">
          <img :src="normalizeMediaUrl(studyEditingItem.image)" style="width:100%; border-radius: 8px; display:block;" />
        </div>
      </div>
      <template #footer>
        <a-button @click="guardStudyItemClose">取消</a-button>
        <a-button type="primary" @click="confirmStudyItemEdit">确定</a-button>
      </template>
    </a-modal>

    <!-- 图片裁剪弹窗 -->
    <ImageCropper
      v-model:show="showCropper"
      :image-url="cropperImageUrl"
      :aspect-ratio="cropperAspectRatio"
      :title="cropperTitle"
      @confirm="onCropConfirm"
    />

    <!-- 精彩回顾编辑弹窗 -->
    <a-modal
      v-model:visible="showShowcaseEditPopup"
      title="编辑精彩回顾"
      :width="560"
      :unmount-on-close="true"
      :on-before-cancel="guardShowcaseClose"
    >
      <div class="dialog-body" v-if="showcaseEditingItem">
        <div class="field-row">
          <span class="field-label">标题</span>
          <a-input v-model="showcaseEditingItem.title" placeholder="请输入标题" />
        </div>
        <div class="field-row">
          <span class="field-label">描述</span>
          <a-input v-model="showcaseEditingItem.desc" type="textarea" :auto-size="{ minRows: 2 }" placeholder="请输入描述" />
        </div>
        <div class="field-row">
          <span class="field-label">图片</span>
          <a-upload :show-file-list="false" :custom-request="wrapUpload(file => onReadShowcaseImage(file))" accept="image/*">
            <a-button size="small"><template #icon><icon-plus /></template>上传图片</a-button>
          </a-upload>
        </div>
        <div v-if="showcaseEditingItem.image" style="margin-top: 8px;">
          <img :src="showcaseEditingItem.image" style="width:100%; border-radius: 8px; display:block;" @click="previewCrop(showcaseEditingItem.image, 'showcaseEditing')" />
          <div style="margin-top:8px; text-align:center;">
            <a-button size="small" type="outline" @click="previewCrop(showcaseEditingItem.image, 'showcaseEditing')">裁剪图片</a-button>
          </div>
        </div>
      </div>
      <template #footer>
        <a-button @click="guardShowcaseClose">取消</a-button>
        <a-button type="primary" @click="confirmShowcaseEdit">确定</a-button>
      </template>
    </a-modal>

    <!-- 首页配置弹窗 -->
    <a-modal
      v-model:visible="showHomeConfigPopup"
      title="首页配置"
      :width="640"
      :unmount-on-close="true"
      :on-before-cancel="guardHomeConfigClose"
    >
      <div class="dialog-body" v-if="editingHomeConfig">
        <div class="config-section">
          <div class="section-title">背景图</div>
          <div class="field-row">
            <span class="field-label">上传</span>
            <a-upload :show-file-list="false" :custom-request="wrapUpload(file => onReadHomeHeaderImage(file))" accept="image/*">
              <a-button size="small"><template #icon><icon-plus /></template>选择图片</a-button>
            </a-upload>
          </div>
          <div v-if="editingHomeConfig.headerImage" style="margin-top: 8px;">
            <img :src="normalizeMediaUrl(editingHomeConfig.headerImage)" style="width:100%; border-radius: 8px; display:block;" />
          </div>
        </div>
        <div class="config-section">
          <div class="section-title">轮播消息</div>
          <div v-for="(msg, idx) in editingHomeConfig.notices" :key="idx" class="list-item" style="padding: 4px 0;">
            <a-input
              :model-value="msg"
              @update:model-value="v => editingHomeConfig.notices[idx] = v"
              placeholder="请输入通知消息"
            />
            <a-button type="text" status="danger" size="small" @click="editingHomeConfig.notices.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
          </div>
          <div class="list-add">
            <a-button type="outline" size="small" @click="editingHomeConfig.notices.push('')"><template #icon><icon-plus /></template>添加消息</a-button>
          </div>
        </div>
        <div class="config-section">
          <div class="section-title">轮播 Banner</div>
          <div v-for="(banner, idx) in editingHomeConfig.banners" :key="idx" class="banner-item">
            <div class="list-item-head">
              <span class="item-num">#{{ idx + 1 }}</span>
              <a-button type="text" status="danger" size="small" @click="editingHomeConfig.banners.splice(idx, 1)"><template #icon><icon-delete /></template></a-button>
            </div>
            <div v-if="banner.image" style="margin-bottom: 8px;">
              <img :src="normalizeMediaUrl(banner.image)" style="width:100%; height: 80px; object-fit: cover; border-radius: 6px; display:block;" />
            </div>
            <a-upload :show-file-list="false" :custom-request="wrapUpload(file => onReadBannerImage(file, idx))" accept="image/*">
              <a-button size="small"><template #icon><icon-image /></template>{{ banner.image ? '更换图片' : '上传图片' }}</a-button>
            </a-upload>
            <a-input
              v-model="banner.link"
              placeholder="如 delivery 或 /cases/1"
              style="margin-top: 4px;"
            />
          </div>
          <div class="list-add">
            <a-button type="outline" size="small" @click="addBanner"><template #icon><icon-plus /></template>添加 Banner</a-button>
          </div>
        </div>
      </div>
      <template #footer>
        <a-button @click="guardHomeConfigClose">取消</a-button>
        <a-button type="primary" @click="saveHomeConfig">保存</a-button>
      </template>
    </a-modal>

    <!-- 新增课程包弹窗 -->
    <a-modal
      v-model:visible="showAddPackagePopup"
      title="新增课程包"
      :width="480"
      :unmount-on-close="true"
      :on-before-cancel="guardAddPackageClose"
    >
      <div class="field-row">
        <span class="field-label">标识</span>
        <a-input v-model="newPackage.id" placeholder="如：study-fullday（英文，唯一）" />
      </div>
      <div class="field-row">
        <span class="field-label">标签</span>
        <a-input v-model="newPackage.tag" placeholder="如：全日营" />
      </div>
      <div class="field-row">
        <span class="field-label">价格</span>
        <a-input v-model="newPackage.price" placeholder="298" />
      </div>
      <div class="add-pkg-tip">
        <p>提示：标识建议使用 study- 前缀，如 study-fullday、study-summer 等</p>
      </div>
      <template #footer>
        <a-button @click="guardAddPackageClose">取消</a-button>
        <a-button type="primary" @click="confirmAddPackage">确定</a-button>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted, nextTick } from 'vue'
import axios from '@/utils/http'
import Modal from '@arco-design/web-vue/es/modal'
import '@arco-design/web-vue/es/modal/style/css'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast } from '@/utils/feedback'
import ImageCropper from './ImageCropper.vue'
import DataToolbar from '../components/DataToolbar.vue'
import { normalizeMediaUrl, uploadFile } from '../composables/useMedia'
import { useAuth } from '../composables/useAuth'

const { userRole, isPlatformAdmin, isAssociationAdmin } = useAuth()

// el-upload on-change → a-upload custom-request 适配：保持原有上传业务逻辑不变
const wrapUpload = (handler) => async ({ fileItem, onSuccess, onError }) => {
  try {
    await handler(fileItem.file)
    onSuccess && onSuccess()
  } catch (err) {
    onError && onError(err)
  }
}

// ===== 未保存守卫 =====
// 本仓库 Arco 2.58 的 a-modal 仅支持 on-before-cancel 同步钩子（before-close 属性不生效），
// 有改动时先弹确认，确认后才真正关闭；返回 false 阻止 Arco 自动关闭。
const makeCloseGuard = (isDirty, doClose) => () => {
  if (!isDirty()) { doClose(); return true }
  Modal.confirm({
    title: '放弃修改',
    content: '当前内容有未保存的修改，确定放弃吗？',
    okText: '放弃修改',
    cancelText: '继续编辑',
    onOk: () => doClose(),
  })
  return false
}
// 服务编辑弹窗：快照覆盖 服务本体 + 研学课程包 + 培训拆分字段
let serviceEditSnapshot = ''
const serviceEditSnapshotOf = () => JSON.stringify({
  svc: editingService.value,
  pkgs: studyPackages.value,
  conditions: trainingConditions.value,
  prices: trainingPrices.value,
  features: trainingFeatures.value,
  companyTitle: trainingCompanyTitle.value,
  companyContent: trainingCompanyContent.value,
  licenseContent: trainingLicenseContent.value,
  licenseQuote: trainingLicenseQuote.value,
})
const guardServiceEditClose = makeCloseGuard(
  () => serviceEditSnapshot !== serviceEditSnapshotOf(),
  () => { showServiceEditPopup.value = false }
)
// 首页配置弹窗
let homeConfigSnapshot = ''
const guardHomeConfigClose = makeCloseGuard(
  () => homeConfigSnapshot !== JSON.stringify(editingHomeConfig.value),
  () => { showHomeConfigPopup.value = false }
)
// 图文展示子弹窗
let studyItemSnapshot = ''
const guardStudyItemClose = makeCloseGuard(
  () => studyItemSnapshot !== JSON.stringify(studyEditingItem.value),
  () => { showStudyItemEditPopup.value = false }
)
// 精彩回顾子弹窗
let showcaseSnapshot = ''
const guardShowcaseClose = makeCloseGuard(
  () => showcaseSnapshot !== JSON.stringify(showcaseEditingItem.value),
  () => { showShowcaseEditPopup.value = false }
)
// 新增课程包弹窗
let addPackageSnapshot = ''
const guardAddPackageClose = makeCloseGuard(
  () => addPackageSnapshot !== JSON.stringify(newPackage.value),
  () => { showAddPackagePopup.value = false }
)

const DEFAULT_HOME_CONFIG = {
  headerImage: '',
  headerImagePosition: 'center',
  banners: [],
  notices: ['交享点无人机外卖配送正式上线', '新开通江心屿无人机外卖配送']
}

const allServiceConfigs = ref({})
const editingServiceId = ref(null)
const editingService = ref(null)
const showServiceEditPopup = ref(false)
const showStudyItemEditPopup = ref(false)
const studyEditingIndex = ref(-1)
const studyEditingItem = ref(null)

const homeConfig = ref(JSON.parse(JSON.stringify(DEFAULT_HOME_CONFIG)))
const editingHomeConfig = ref(JSON.parse(JSON.stringify(DEFAULT_HOME_CONFIG)))
const showHomeConfigPopup = ref(false)

// trainingInfo 拆分字段
const trainingConditions = ref([])
const trainingPrices = ref([])
const trainingFeatures = ref([])
const trainingCompanyTitle = ref('')
const trainingCompanyContent = ref('')
const trainingLicenseContent = ref('')
const trainingLicenseQuote = ref('')

// 研学课程包
const studyPackages = ref({})
const studyPackageIds = ref(['study-halfday', 'study-family'])
const activeStudyPkgId = ref('study-halfday')
const activeStudyPkg = computed(() => studyPackages.value[activeStudyPkgId.value] || null)

// 新增课程包
const showAddPackagePopup = ref(false)
const newPackage = ref({ id: '', tag: '', price: '' })

// 图片裁剪
const showCropper = ref(false)
const cropperImageUrl = ref('')
const cropperAspectRatio = ref('16:9')
const cropperTitle = ref('图片裁剪')
const cropperTarget = ref({ type: '', field: '', index: -1 })

// 精彩回顾编辑
const showShowcaseEditPopup = ref(false)
const showcaseEditingItem = ref(null)
const showcaseEditingIndex = ref(-1)

const serviceConfigEntries = computed(() => {
  const entries = Object.entries(allServiceConfigs.value || {}).filter(([id]) => /^\d+$/.test(String(id)))
  entries.sort((a, b) => Number(a[0]) - Number(b[0]))
  return entries
})

// 按前端服务分类进行分组
const serviceGroupsConfig = [
  { title: '核心服务', ids: ['2', '8', '1'], roles: ['platform_admin'] },
  { title: '商业应用', ids: ['4', '5', '3', '7', '13'], roles: ['platform_admin', 'association_admin'] },
  { title: '研学教育', ids: ['9'], roles: ['platform_admin', 'association_admin'] },
  { title: '增值服务', ids: ['10', '11', '12'], roles: ['platform_admin'] }
]

const groupedServiceEntries = computed(() => {
  const allEntries = serviceConfigEntries.value
  const entryMap = Object.fromEntries(allEntries)
  const usedIds = new Set()
  const role = userRole.value

  const groups = serviceGroupsConfig
    .filter(group => group.roles.includes(role))
    .map(group => {
      const items = group.ids
        .filter(id => entryMap[id])
        .map(id => {
          usedIds.add(id)
          return [id, entryMap[id]]
        })
      return items.length > 0 ? { title: group.title, items } : null
    })
    .filter(Boolean)

  // 未归类的服务放到最后（仅 admin 可见）
  if (isPlatformAdmin.value) {
    const ungrouped = allEntries.filter(([id]) => !usedIds.has(id))
    if (ungrouped.length > 0) {
      groups.push({ title: '其他服务', items: ungrouped })
    }
  }

  return groups
})

const fetchAllServiceConfigs = async () => {
  try {
    const res = await axios.get('/api/services/config')
    allServiceConfigs.value = res.data || {}
    homeConfig.value = JSON.parse(JSON.stringify(allServiceConfigs.value._home || DEFAULT_HOME_CONFIG))
  } catch (error) {
    console.error('[ServiceConfig] 获取配置失败', error)
    showFailToast('获取服务配置失败')
  }
}

// --- Home config ---
const editHomeConfig = () => {
  editingHomeConfig.value = JSON.parse(JSON.stringify(homeConfig.value || DEFAULT_HOME_CONFIG))
  if (!Array.isArray(editingHomeConfig.value.notices)) {
    editingHomeConfig.value.notices = [...DEFAULT_HOME_CONFIG.notices]
  }
  if (!Array.isArray(editingHomeConfig.value.banners)) {
    editingHomeConfig.value.banners = []
  }
  homeConfigSnapshot = JSON.stringify(editingHomeConfig.value)
  showHomeConfigPopup.value = true
}

const onReadHomeHeaderImage = async (file) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(file)
  closeToast()
  if (url && editingHomeConfig.value) {
    editingHomeConfig.value = { ...editingHomeConfig.value, headerImage: normalizeMediaUrl(url) }
    showSuccessToast('上传成功')
  }
}

const onReadBannerImage = async (file, idx) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(file)
  closeToast()
  if (url && editingHomeConfig.value?.banners?.[idx]) {
    editingHomeConfig.value.banners[idx].image = normalizeMediaUrl(url)
    showSuccessToast('上传成功')
  }
}

const addBanner = () => {
  if (!editingHomeConfig.value.banners) {
    editingHomeConfig.value.banners = []
  }
  editingHomeConfig.value.banners.push({ image: '', link: '' })
}

const saveHomeConfig = async () => {
  try {
    showLoadingToast({ message: '保存中...', forbidClick: true })
    const newConfigs = { ...allServiceConfigs.value }
    newConfigs._home = editingHomeConfig.value
    await axios.post('/api/services/config', { config: newConfigs })
    closeToast()
    showSuccessToast('保存成功')
    showHomeConfigPopup.value = false
    // Re-fetch to ensure label updates from server data
    await fetchAllServiceConfigs()
  } catch (error) {
    closeToast()
    showFailToast(error?.response?.data?.message || '保存失败')
  }
}

// --- Service config ---
const editServiceConfig = (id) => {
  editingServiceId.value = id
  const raw = JSON.parse(JSON.stringify(allServiceConfigs.value[id]))
  if (!raw.projects) raw.projects = []
  if (!raw.advantages) raw.advantages = []
  if (id === '6') {
    if (!raw.highlights) raw.highlights = []
    if (!raw.studyShowcase) raw.studyShowcase = []
  }
  editingService.value = raw

  // 初始化研学课程包
  if (id === '9') {
    const pkgs = raw.packages || {}
    studyPackages.value = JSON.parse(JSON.stringify(pkgs))
    // 动态获取所有课程包ID（兼容旧数据）
    const existingIds = Object.keys(studyPackages.value)
    // 合并默认ID和已有ID
    studyPackageIds.value = [...new Set([...studyPackageIds.value, ...existingIds])]
    // 确保每个包有必需的字段
    for (const pkgId of studyPackageIds.value) {
      if (!studyPackages.value[pkgId]) {
        studyPackages.value[pkgId] = createEmptyPackage(pkgId)
      }
      const p = studyPackages.value[pkgId]
      if (!p.schedule) p.schedule = []
      if (!p.studyGoals) p.studyGoals = []
      if (!p.audience) p.audience = []
      if (!p.feeInfo) p.feeInfo = []
      if (!p.tips) p.tips = []
      if (!p.projects) p.projects = []
      if (!p.advantages) p.advantages = []
      if (!p.showcase) p.showcase = []
      if (!p.safetyBriefing) p.safetyBriefing = []
      if (!p.studySummary) p.studySummary = ''
    }
    activeStudyPkgId.value = studyPackageIds.value[0] || 'study-halfday'
  }

  // 初始化 trainingInfo
  if (id === '6') {
    const ti = raw.trainingInfo || {}
    trainingConditions.value = [...(ti.conditions || [])]
    trainingPrices.value = (ti.prices || []).map(p => ({ ...p }))
    trainingFeatures.value = (ti.features || []).map(f => ({ ...f }))
    trainingCompanyTitle.value = ti.companyIntro?.title || ''
    trainingCompanyContent.value = ti.companyIntro?.content || ''
    trainingLicenseContent.value = ti.licenseFunction?.content || ''
    trainingLicenseQuote.value = ti.licenseFunction?.quote || ''
  }

  serviceEditSnapshot = serviceEditSnapshotOf()
  showServiceEditPopup.value = true
}

const saveServiceConfig = async () => {
  try {
    showLoadingToast({ message: '保存中...', forbidClick: true })

    // 合并研学课程包
    if (editingServiceId.value === '9') {
      editingService.value.packages = JSON.parse(JSON.stringify(studyPackages.value))
    }

    // 合并 trainingInfo
    if (editingServiceId.value === '6') {
      editingService.value.trainingInfo = {
        conditions: [...trainingConditions.value],
        prices: trainingPrices.value.map(p => ({ ...p })),
        features: trainingFeatures.value.map(f => ({ ...f })),
        companyIntro: { title: trainingCompanyTitle.value, content: trainingCompanyContent.value },
        licenseFunction: { content: trainingLicenseContent.value, quote: trainingLicenseQuote.value }
      }
    }

    const newConfigs = { ...allServiceConfigs.value }
    newConfigs[editingServiceId.value] = editingService.value
    await axios.post('/api/services/config', { config: newConfigs })
    allServiceConfigs.value = newConfigs
    closeToast()
    showSuccessToast('保存成功')
    serviceEditSnapshot = serviceEditSnapshotOf()
    showServiceEditPopup.value = false
  } catch (error) {
    closeToast()
    const status = error?.response?.status
    const msg = error?.response?.data?.message
    showFailToast(status === 403 ? '无权限' : (msg || '保存失败'))
  }
}

const onReadServiceFile = async (file, field) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(file)
  closeToast()
  if (url && editingService.value) {
    editingService.value = { ...editingService.value, [field]: normalizeMediaUrl(url) }
    showSuccessToast('上传成功')
  }
}

// 创建空课程包模板
const createEmptyPackage = (pkgId) => {
  return {
    id: pkgId,
    name: '新课程',
    tag: '新课程',
    price: null,
    recommended: false,
    desc: '',
    cardHighlights: [],
    headerBg: 'linear-gradient(135deg, #06b6d4 0%, #165DFF 100%)',
    intro: '',
    schedule: [],
    studyGoals: [],
    audience: [],
    feeInfo: [],
    tips: [],
    projects: [],
    advantages: [],
    showcase: [],
    safetyBriefing: [],
    studySummary: ''
  }
}

// 上传课程包图片
const onReadPackageImage = async (file, field) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(file)
  closeToast()
  if (url && activeStudyPkg.value) {
    activeStudyPkg.value[field] = normalizeMediaUrl(url)
    showSuccessToast('上传成功')
  }
}

// 预览裁剪
const previewCrop = (imageUrl, type, index = -1) => {
  cropperImageUrl.value = imageUrl
  cropperTarget.value = { type, field: type, index }

  // 根据类型设置裁剪比例
  if (type === 'headerBg') {
    cropperAspectRatio.value = '16:9'
    cropperTitle.value = '裁剪头部背景'
  } else if (type === 'showcase' || type === 'showcaseEditing') {
    cropperAspectRatio.value = '16:9'
    cropperTitle.value = '裁剪展示图片'
  } else {
    cropperAspectRatio.value = 'free'
    cropperTitle.value = '裁剪图片'
  }

  showCropper.value = true
}

// 裁剪确认
const onCropConfirm = async (croppedFile) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(croppedFile)
  closeToast()

  if (url) {
    const normalizedUrl = normalizeMediaUrl(url)
    const { type, index } = cropperTarget.value

    if (type === 'headerBg' && activeStudyPkg.value) {
      activeStudyPkg.value.headerBg = normalizedUrl
    } else if (type === 'showcase' && activeStudyPkg.value && index >= 0) {
      activeStudyPkg.value.showcase[index].image = normalizedUrl
    } else if (type === 'showcaseEditing' && showcaseEditingItem.value) {
      showcaseEditingItem.value.image = normalizedUrl
    }

    showSuccessToast('裁剪上传成功')
  }
}

// 添加精彩回顾
const addShowcaseItem = () => {
  showcaseEditingIndex.value = -1
  showcaseEditingItem.value = { title: '', desc: '', image: '' }
  showcaseSnapshot = JSON.stringify(showcaseEditingItem.value)
  showShowcaseEditPopup.value = true
}

// 编辑精彩回顾
const editShowcaseItem = (idx) => {
  if (!activeStudyPkg.value?.showcase) return
  showcaseEditingIndex.value = idx
  showcaseEditingItem.value = { ...activeStudyPkg.value.showcase[idx] }
  showcaseSnapshot = JSON.stringify(showcaseEditingItem.value)
  showShowcaseEditPopup.value = true
}

// 确认精彩回顾编辑
const confirmShowcaseEdit = () => {
  if (!showcaseEditingItem.value || !activeStudyPkg.value) return
  if (!activeStudyPkg.value.showcase) activeStudyPkg.value.showcase = []

  if (showcaseEditingIndex.value >= 0) {
    activeStudyPkg.value.showcase[showcaseEditingIndex.value] = { ...showcaseEditingItem.value }
  } else {
    activeStudyPkg.value.showcase.push({ ...showcaseEditingItem.value })
  }

  showShowcaseEditPopup.value = false
}

// 上传精彩回顾图片 - 上传后自动打开裁剪
const onReadShowcaseImage = async (file) => {
  showLoadingToast({ message: '上传中...', forbidClick: true })
  const url = await uploadFile(file)
  closeToast()
  if (url && showcaseEditingItem.value) {
    showcaseEditingItem.value.image = normalizeMediaUrl(url)
    showSuccessToast('上传成功，请裁剪图片')
    // 自动打开裁剪工具
    setTimeout(() => {
      previewCrop(showcaseEditingItem.value.image, 'showcaseEditing')
    }, 300)
  }
}

// 显示新增课程包弹窗
const showAddPackageDialog = () => {
  newPackage.value = { id: '', tag: '', price: '' }
  addPackageSnapshot = JSON.stringify(newPackage.value)
  showAddPackagePopup.value = true
}

// 确认新增课程包
const confirmAddPackage = () => {
  const { id, tag, price } = newPackage.value
  if (!id || !tag) {
    showFailToast('请填写完整信息')
    return
  }
  if (studyPackageIds.value.includes(id)) {
    showFailToast('课程包标识已存在')
    return
  }
  // 创建新课程包
  studyPackages.value[id] = createEmptyPackage(id)
  studyPackages.value[id].tag = tag
  // 空值不写 0：价格未填时存 null（避免前台展示虚假的 ¥0）
  studyPackages.value[id].price = price === '' || price == null ? null : Number(price)
  studyPackageIds.value.push(id)
  activeStudyPkgId.value = id
  showAddPackagePopup.value = false
  showSuccessToast('添加成功')
}

// 删除课程包（整包数据删除，保存后不可恢复，先确认）
const removeStudyPackage = (pkgId) => {
  if (studyPackageIds.value.length <= 1) {
    showFailToast('至少保留一个课程包')
    return
  }
  Modal.confirm({
    title: '删除课程包',
    content: `确定删除课程包「${studyPackages.value[pkgId]?.tag || pkgId}」吗？其全部配置（课程安排/精彩回顾等）将一并删除`,
    okText: '删除',
    cancelText: '取消',
    onOk: () => {
      const idx = studyPackageIds.value.indexOf(pkgId)
      if (idx > -1) {
        studyPackageIds.value.splice(idx, 1)
        delete studyPackages.value[pkgId]
        // 切换到第一个课程包
        activeStudyPkgId.value = studyPackageIds.value[0]
        showSuccessToast('删除成功')
      }
    }
  })
}

// --- Study showcase ---
const innerAddStudyItem = () => {
  studyEditingIndex.value = -1
  studyEditingItem.value = { title: '', desc: '', image: '' }
  studyItemSnapshot = JSON.stringify(studyEditingItem.value)
  showStudyItemEditPopup.value = true
}

const innerEditStudyItem = (idx) => {
  studyEditingIndex.value = idx
  studyEditingItem.value = { ...editingService.value.studyShowcase[idx] }
  studyItemSnapshot = JSON.stringify(studyEditingItem.value)
  showStudyItemEditPopup.value = true
}

const confirmStudyItemEdit = () => {
  const it = studyEditingItem.value
  if (!it || !editingService.value) return
  if (!editingService.value.studyShowcase) editingService.value.studyShowcase = []
  if (studyEditingIndex.value >= 0) {
    editingService.value.studyShowcase[studyEditingIndex.value] = { ...it }
  } else {
    editingService.value.studyShowcase.push({ ...it })
  }
  showStudyItemEditPopup.value = false
}

const onReadStudyImage = async (file) => {
  // /api/upload 是 dev-only 路由，生产 404——统一走 uploadFile（POST /api/v1/upload）
  const url = await uploadFile(file)
  if (url && studyEditingItem.value) {
    studyEditingItem.value = { ...studyEditingItem.value, image: url }
  }
}

// --- 商业化费率（/api/v1/admin/config，与 H5 services config 独立） ---
const platformConfig = ref({ match_fee_rate: 0, match_fee_note: '' })

const fetchPlatformConfig = async () => {
  try {
    const res = await axios.get('/api/v1/admin/config')
    platformConfig.value = {
      match_fee_rate: Number(res.data?.match_fee_rate || 0),
      match_fee_note: res.data?.match_fee_note || '',
    }
  } catch (error) {
    console.error('[ServiceConfig] 获取平台配置失败', error)
  }
}

const savePlatformConfig = async () => {
  try {
    showLoadingToast({ message: '保存中...', forbidClick: true })
    // /api/v1/admin/config 为整包替换语义：先取全量，再改费率字段后整体写回
    const res = await axios.get('/api/v1/admin/config')
    const full = {
      ...res.data,
      match_fee_rate: platformConfig.value.match_fee_rate,
      match_fee_note: platformConfig.value.match_fee_note,
    }
    await axios.post('/api/v1/admin/config', full)
    closeToast()
    showSuccessToast('费率已保存')
  } catch (error) {
    closeToast()
    showFailToast(error?.response?.data?.message || '保存失败')
  }
}

onMounted(() => {
  fetchAllServiceConfigs()
  fetchPlatformConfig()
})
</script>

<style scoped>
.toolbar-label { font-size: 14px; font-weight: 500; color: var(--color-text-1); }

/* 配置卡片 */
.config-card {
  margin-bottom: 12px;
  border-radius: var(--card-radius, 10px);
}
.config-card :deep(.arco-card-header) {
  padding: 10px 16px;
}
.card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-1);
}

/* 商业化费率 */
.fee-row {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
  padding: 4px;
}
.fee-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 220px;
}
.fee-label {
  font-size: 12px;
  color: var(--color-text-2);
}
.fee-hint {
  margin-top: 8px;
  padding: 0 4px;
  font-size: 12px;
  color: var(--color-text-3);
}

/* 可点击行 */
.link-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 4px;
  cursor: pointer;
  border-radius: 6px;
  transition: background 0.2s;
}
.link-row:hover {
  background: var(--color-fill-2);
}
.link-row-main { min-width: 0; }
.link-row-title {
  font-size: 14px;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.link-row-label {
  font-size: 12px;
  color: var(--color-text-3);
  margin-top: 2px;
}
.link-row-arrow {
  color: var(--color-text-3);
  flex-shrink: 0;
}

/* 弹窗内容 */
.dialog-body {
  max-height: 68vh;
  overflow-y: auto;
  padding-right: 4px;
}
.config-section {
  background: var(--color-fill-2);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 12px;
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-1);
  margin-bottom: 10px;
}

/* 带标签的字段行 */
.field-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 10px;
}
.field-row:last-child {
  margin-bottom: 0;
}
.field-label {
  width: 76px;
  flex-shrink: 0;
  font-size: 13px;
  color: var(--color-text-2);
  line-height: 32px;
  text-align: right;
}
.field-row .arco-input {
  flex: 1;
}

/* 列表项：单行 */
.list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}
.list-item .arco-input {
  flex: 1;
}

/* 列表项：多行 */
.list-item-block {
  padding: 8px 0;
  border-bottom: 1px solid var(--color-fill-2);
}
.list-item-block:last-of-type {
  border-bottom: none;
}
.list-item-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.item-num {
  font-size: 12px;
  font-weight: 600;
  color: var(--accent-color, #165DFF);
  background: var(--accent-light, #e8f2fc);
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

/* 添加按钮 */
.list-add {
  padding: 8px 0 0;
}
.list-add :deep(.arco-btn) {
  border-style: dashed;
}

/* 图文展示行 */
.showcase-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid var(--color-fill-2);
}
.showcase-item:last-of-type {
  border-bottom: none;
}
.showcase-left {
  width: 48px;
  height: 36px;
  border-radius: 6px;
  overflow: hidden;
  background: var(--color-fill-2);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.showcase-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.showcase-mid {
  flex: 1;
  font-size: 14px;
  color: var(--color-text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 课程包 Tab */
.pkg-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.pkg-tab {
  position: relative;
  flex: 1;
  min-width: 100px;
  text-align: center;
  padding: 10px 8px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-3);
  background: var(--color-fill-2);
  cursor: pointer;
  transition: all 0.25s;
}
.pkg-tab.active {
  background: var(--accent-color, #165DFF);
  color: #fff;
  font-weight: 600;
}
.pkg-tab-close {
  position: absolute;
  top: 2px;
  right: 2px;
  font-size: 10px;
  padding: 2px;
  opacity: 0.6;
}
.pkg-tab-close:hover {
  opacity: 1;
}
.pkg-tab-add {
  flex: 0 0 auto;
  min-width: 40px;
  background: #e8f2fc;
  color: var(--accent-color, #165DFF);
}

/* 背景输入 */
.bg-input-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

/* 图片预览 */
.img-preview {
  padding: 4px 0 0;
}
.img-preview img {
  width: 100%;
  max-height: 150px;
  object-fit: cover;
  border-radius: 8px;
  cursor: pointer;
}
.img-actions {
  margin-top: 8px;
  text-align: center;
}

/* 图标预览 */
.icon-preview {
  padding: 4px 0;
}

/* Banner 项 */
.banner-item {
  padding: 8px 0;
  border-bottom: 1px solid var(--color-fill-2);
}
.banner-item:last-of-type {
  border-bottom: none;
}

/* 新增课程包提示 */
.add-pkg-tip {
  margin-top: 12px;
  padding: 12px;
  background: var(--color-fill-2);
  border-radius: 8px;
  font-size: 12px;
  color: var(--color-text-2);
}
.add-pkg-tip p {
  margin: 0;
}
</style>
