<template>
  <!-- v20250324-1: 课程包独立字段管理 -->
  <div class="config-page">
    <DataToolbar>
      <template #filters>
        <span class="toolbar-label">服务配置</span>
      </template>
      <template #actions>
        <van-button type="default" size="small" icon="replay" @click="fetchAllServiceConfigs">刷新</van-button>
      </template>
    </DataToolbar>

    <!-- 首页配置（仅管理员可见） -->
    <van-cell-group v-if="isAdmin" inset title="首页配置" style="margin-bottom: 12px; border-radius: var(--card-radius);">
      <van-cell
        title="首页背景图 & 轮播配置"
        :label="(homeConfig?.headerImage ? '背景图已配置' : '背景图未配置') + '  ·  ' + (homeConfig?.banners?.length || 0) + ' 张Banner  ·  ' + (homeConfig?.notices?.length || 0) + ' 条轮播消息'"
        is-link
        @click="editHomeConfig"
      />
    </van-cell-group>

    <!-- 服务列表（按分区） -->
    <van-cell-group
      v-for="group in groupedServiceEntries"
      :key="group.title"
      inset
      :title="group.title"
      style="margin-bottom: 12px; border-radius: var(--card-radius);"
    >
      <van-cell
        v-for="[id, cfg] in group.items"
        :key="id"
        :title="cfg.name"
        is-link
        @click="editServiceConfig(id)"
      />
    </van-cell-group>

    <!-- ========== 服务编辑弹窗 ========== -->
    <van-popup :show="showServiceEditPopup" @update:show="v => showServiceEditPopup = v" position="bottom" :style="{ height: '90%' }" round>
      <div class="edit-popup" v-if="editingService">
        <van-nav-bar
          :title="editingService.name"
          left-text="取消"
          right-text="保存"
          @click-left="showServiceEditPopup = false"
          @click-right="saveServiceConfig"
        />

        <div class="edit-body">
          <!-- 服务介绍 -->
          <van-cell-group inset title="服务介绍">
            <van-field v-model="editingService.intro" type="textarea" rows="3" placeholder="请输入服务介绍，建议50-100字" autosize show-word-limit maxlength="200" />
          </van-cell-group>

          <!-- 联系方式 -->
          <van-cell-group inset title="联系方式">
            <van-field v-model="editingService.contactPhone" label="电话" placeholder="联系电话" type="tel" />
            <van-field v-model="editingService.contactPhone2" label="热线" placeholder="选填，第二个电话" type="tel" />
            <van-field v-model="editingService.address" label="地址" placeholder="选填" type="textarea" rows="1" autosize />
          </van-cell-group>

          <!-- 服务项目（非研学服务显示） -->
          <van-cell-group inset :title="'服务项目（' + (editingService.projects?.length || 0) + '）'" v-if="editingServiceId !== '9'">
            <div v-for="(p, idx) in editingService.projects" :key="idx" class="list-item">
              <van-field v-model="p.name" placeholder="项目名称" dense />
              <van-button size="mini" type="danger" plain icon="cross" @click="editingService.projects.splice(idx, 1)" />
            </div>
            <div class="list-add">
              <van-button size="small" type="primary" block plain icon="plus" @click="editingService.projects.push({ name: '', icon: 'star-o' })">添加项目</van-button>
            </div>
          </van-cell-group>

          <!-- 服务优势（非研学服务显示） -->
          <van-cell-group inset :title="'服务优势（' + (editingService.advantages?.length || 0) + '）'" v-if="editingServiceId !== '9'">
            <div v-for="(adv, idx) in editingService.advantages" :key="idx" class="list-item">
              <van-field v-model="editingService.advantages[idx]" placeholder="优势描述" dense />
              <van-button size="mini" type="danger" plain icon="cross" @click="editingService.advantages.splice(idx, 1)" />
            </div>
            <div class="list-add">
              <van-button size="small" type="primary" block plain icon="plus" @click="editingService.advantages.push('')">添加优势</van-button>
            </div>
          </van-cell-group>

          <!-- 背景图（仅培训显示） -->
          <van-cell-group inset title="页面背景图" v-if="editingServiceId === '6'">
            <van-field label="上传图片">
              <template #input>
                <van-uploader :after-read="file => onReadServiceFile(file, 'headerImage')" max-count="1">
                  <van-button icon="plus" size="small" type="primary" plain>选择图片</van-button>
                </van-uploader>
              </template>
            </van-field>
            <div v-if="editingService.headerImage" class="img-preview">
              <img :src="normalizeMediaUrl(editingService.headerImage)" />
            </div>
          </van-cell-group>

          <!-- 培训亮点（仅培训） -->
          <van-cell-group inset :title="'培训亮点（' + (editingService.highlights?.length || 0) + '）'" v-if="editingServiceId === '6'">
            <div v-for="(hl, idx) in editingService.highlights" :key="idx" class="list-item-block">
              <div class="list-item-head">
                <span class="item-num">{{ idx + 1 }}</span>
                <van-button size="mini" type="danger" plain icon="cross" @click="editingService.highlights.splice(idx, 1)" />
              </div>
              <van-field v-model="hl.title" label="标题" placeholder="如：官方认证" dense />
              <van-field v-model="hl.desc" label="描述" placeholder="如：CAAC民航局授权" dense />
            </div>
            <div class="list-add">
              <van-button size="small" type="primary" block plain icon="plus" @click="editingService.highlights.push({ title: '', desc: '', icon: 'star-o' })">添加亮点</van-button>
            </div>
          </van-cell-group>

          <!-- 图文展示（仅培训显示，研学在课程包内管理） -->
          <van-cell-group inset :title="'图文展示（' + (editingService.studyShowcase?.length || 0) + '）'" v-if="editingServiceId === '6'">
            <div v-for="(item, idx) in editingService.studyShowcase" :key="idx" class="showcase-item">
              <div class="showcase-left">
                <img v-if="item.image" :src="normalizeMediaUrl(item.image)" class="showcase-img" />
                <van-icon v-else name="photo-o" size="24" color="#ddd" />
              </div>
              <div class="showcase-mid">{{ item.title || '未命名' }}</div>
              <van-button size="mini" type="primary" plain @click="innerEditStudyItem(idx)">编辑</van-button>
              <van-button size="mini" type="danger" plain icon="cross" @click="editingService.studyShowcase.splice(idx, 1)" />
            </div>
            <div class="list-add">
              <van-button size="small" type="primary" block plain icon="plus" @click="innerAddStudyItem">添加展示</van-button>
            </div>
          </van-cell-group>

          <!-- ===== 研学课程包管理（仅研学） ===== -->
          <template v-if="editingServiceId === '9'">
            <van-cell-group inset title="课程包管理" style="margin-top: 12px;">
              <div class="pkg-tabs">
                <div
                  v-for="pkgId in studyPackageIds"
                  :key="pkgId"
                  class="pkg-tab"
                  :class="{ active: activeStudyPkgId === pkgId }"
                  @click="activeStudyPkgId = pkgId"
                >
                  {{ studyPackages[pkgId]?.tag || pkgId }} (¥{{ studyPackages[pkgId]?.price || '' }})
                  <van-icon name="cross" class="pkg-tab-close" @click.stop="removeStudyPackage(pkgId)" />
                </div>
                <div class="pkg-tab pkg-tab-add" @click="showAddPackageDialog">
                  <van-icon name="plus" />
                </div>
              </div>
            </van-cell-group>

            <template v-if="activeStudyPkg">
              <!-- 基本信息 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 基本信息'">
                <van-field v-model="activeStudyPkg.name" label="课程名称" placeholder="如：无人机研学实践中心半日营" />
                <van-field v-model="activeStudyPkg.tag" label="标签" placeholder="如：半日营" />
                <van-field v-model.number="activeStudyPkg.price" label="票价" placeholder="198" type="number" />
                
                <!-- 头部背景 - 支持渐变或图片 -->
                <van-field label="头部背景">
                  <template #input>
                    <div class="bg-input-wrap">
                      <van-field v-model="activeStudyPkg.headerBg" placeholder="渐变色或图片URL" style="flex:1;" />
                      <van-uploader :after-read="f => onReadPackageImage(f, 'headerBg')" max-count="1" accept="image/*">
                        <van-button icon="photo-o" size="small" type="primary" plain>上传</van-button>
                      </van-uploader>
                    </div>
                  </template>
                </van-field>
                <div v-if="activeStudyPkg.headerBg" class="img-preview">
                  <img :src="normalizeMediaUrl(activeStudyPkg.headerBg)" @click="previewCrop(normalizeMediaUrl(activeStudyPkg.headerBg), 'headerBg')" />
                  <div class="img-actions">
                    <van-button size="mini" type="primary" plain @click="previewCrop(normalizeMediaUrl(activeStudyPkg.headerBg), 'headerBg')">裁剪</van-button>
                  </div>
                </div>
                
                <van-field v-model="activeStudyPkg.intro" label="介绍" type="textarea" rows="3" placeholder="课程介绍" autosize />
              </van-cell-group>

              <!-- 服务项目（每个课程包独立） -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 服务项目（' + (activeStudyPkg.projects?.length || 0) + '）'">
                <div v-for="(p, idx) in activeStudyPkg.projects" :key="idx" class="list-item-block">
                  <div class="list-item-head">
                    <span class="item-num">{{ idx + 1 }}</span>
                    <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.projects.splice(idx, 1)" />
                  </div>
                  <van-field v-model="p.name" label="名称" placeholder="如：展厅参观" dense />
                  <van-field v-model="p.icon" label="图标" placeholder="Vant图标名或图片URL" dense />
                  <div v-if="p.icon && p.icon.startsWith('http')" class="icon-preview">
                    <img :src="p.icon" style="width:40px;height:40px;border-radius:8px;object-fit:cover;" />
                  </div>
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.projects.push({ name: '', icon: 'star-o' })">添加项目</van-button>
                </div>
              </van-cell-group>

              <!-- 服务优势（每个课程包独立） -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 服务优势（' + (activeStudyPkg.advantages?.length || 0) + '）'">
                <div v-for="(adv, idx) in activeStudyPkg.advantages" :key="idx" class="list-item">
                  <van-field v-model="activeStudyPkg.advantages[idx]" placeholder="优势描述，如：专业导师全程指导" dense />
                  <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.advantages.splice(idx, 1)" />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.advantages.push('')">添加优势</van-button>
                </div>
              </van-cell-group>

              <!-- 精彩回顾（每个课程包独立） -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 精彩回顾（' + (activeStudyPkg.showcase?.length || 0) + '）'">
                <div v-for="(item, idx) in activeStudyPkg.showcase" :key="idx" class="showcase-item">
                  <div class="showcase-left">
                    <img v-if="item.image" :src="item.image" class="showcase-img" @click="previewCrop(item.image, 'showcase', idx)" />
                    <van-icon v-else name="photo-o" size="24" color="#ddd" />
                  </div>
                  <div class="showcase-mid">{{ item.title || '未命名' }}</div>
                  <van-button size="mini" type="primary" plain @click="editShowcaseItem(idx)">编辑</van-button>
                  <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.showcase.splice(idx, 1)" />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="addShowcaseItem">添加展示</van-button>
                </div>
              </van-cell-group>

              <!-- 课程安排 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 课程安排（' + (activeStudyPkg.schedule?.length || 0) + '）'">
                <div v-for="(step, idx) in activeStudyPkg.schedule" :key="idx" class="list-item-block">
                  <div class="list-item-head">
                    <span class="item-num">{{ idx + 1 }}</span>
                    <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.schedule.splice(idx, 1)" />
                  </div>
                  <van-field v-model="step.amTime" label="上午" placeholder="08:50" dense />
                  <van-field v-model="step.pmTime" label="下午" placeholder="13:50" dense />
                  <van-field v-model="step.name" label="环节" placeholder="集合签到" dense />
                  <van-field v-model="step.desc" label="说明" placeholder="研学中心集合，签到报到" dense />
                  <van-field v-model="step.location" label="地点" placeholder="研学中心A教室" dense />
                  <van-field v-model="step.purpose" label="课程目的" placeholder="了解无人机基本原理" dense type="textarea" rows="2" autosize />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.schedule.push({ amTime: '', pmTime: '', name: '', desc: '', location: '', purpose: '' })">添加步骤</van-button>
                </div>
              </van-cell-group>

              <!-- 研学目标 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 研学目标（' + (activeStudyPkg.studyGoals?.length || 0) + '）'">
                <div v-for="(goal, idx) in activeStudyPkg.studyGoals" :key="idx" class="list-item-block">
                  <div class="list-item-head">
                    <span class="item-num">{{ idx + 1 }}</span>
                    <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.studyGoals.splice(idx, 1)" />
                  </div>
                  <van-field v-model="goal.label" label="目标类型" placeholder="如：知识目标、能力目标" dense />
                  <van-field v-model="goal.content" label="具体内容" placeholder="掌握无人机基本原理" dense type="textarea" rows="3" autosize />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.studyGoals.push({ label: '', content: '' })">添加目标</van-button>
                </div>
              </van-cell-group>

              <!-- 安全宣讲 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 安全宣讲（' + (activeStudyPkg.safetyBriefing?.length || 0) + '）'">
                <div v-for="(item, idx) in activeStudyPkg.safetyBriefing" :key="idx" class="list-item">
                  <van-field v-model="activeStudyPkg.safetyBriefing[idx]" placeholder="如：操作前检查设备电量" dense />
                  <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.safetyBriefing.splice(idx, 1)" />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.safetyBriefing.push('')">添加安全提示</van-button>
                </div>
              </van-cell-group>

              <!-- 研学总结 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 研学总结'">
                <van-field v-model="activeStudyPkg.studySummary" label="总结内容" type="textarea" rows="4" placeholder="课程总结内容..." autosize show-word-limit maxlength="500" />
              </van-cell-group>

              <!-- 适合人群 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 适合人群（' + (activeStudyPkg.audience?.length || 0) + '）'">
                <div v-for="(a, idx) in activeStudyPkg.audience" :key="idx" class="list-item">
                  <van-field v-model="activeStudyPkg.audience[idx]" placeholder="如：6-16岁青少年" dense />
                  <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.audience.splice(idx, 1)" />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.audience.push('')">添加</van-button>
                </div>
              </van-cell-group>

              <!-- 费用说明 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 费用说明（' + (activeStudyPkg.feeInfo?.length || 0) + '）'">
                <div v-for="(f, idx) in activeStudyPkg.feeInfo" :key="idx" class="list-item-block">
                  <div class="list-item-head">
                    <span class="item-num">{{ idx + 1 }}</span>
                    <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.feeInfo.splice(idx, 1)" />
                  </div>
                  <van-field v-model="f.label" label="项目" placeholder="课程价格" dense />
                  <van-field v-model="f.value" label="内容" placeholder="¥198/人" dense />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.feeInfo.push({ label: '', value: '' })">添加</van-button>
                </div>
              </van-cell-group>

              <!-- 温馨提示 -->
              <van-cell-group inset :title="activeStudyPkg.tag + ' - 温馨提示（' + (activeStudyPkg.tips?.length || 0) + '）'">
                <div v-for="(t, idx) in activeStudyPkg.tips" :key="idx" class="list-item">
                  <van-field v-model="activeStudyPkg.tips[idx]" placeholder="提示内容" dense />
                  <van-button size="mini" type="danger" plain icon="cross" @click="activeStudyPkg.tips.splice(idx, 1)" />
                </div>
                <div class="list-add">
                  <van-button size="small" type="primary" block plain icon="plus" @click="activeStudyPkg.tips.push('')">添加</van-button>
                </div>
              </van-cell-group>
            </template>
          </template>

          <!-- ===== 飞手培训专属 ===== -->
          <template v-if="editingServiceId === '6'">
            <van-cell-group inset title="报名条件">
              <div v-for="(cond, idx) in trainingConditions" :key="idx" class="list-item">
                <van-field v-model="trainingConditions[idx]" placeholder="如：年满16周岁" dense />
                <van-button size="mini" type="danger" plain icon="cross" @click="trainingConditions.splice(idx, 1)" />
              </div>
              <div class="list-add">
                <van-button size="small" type="primary" block plain icon="plus" @click="trainingConditions.push('')">添加条件</van-button>
              </div>
            </van-cell-group>

            <van-cell-group inset title="培训费用">
              <div v-for="(p, idx) in trainingPrices" :key="idx" class="list-item-block">
                <div class="list-item-head">
                  <span class="item-num">{{ idx + 1 }}</span>
                  <van-button size="mini" type="danger" plain icon="cross" @click="trainingPrices.splice(idx, 1)" />
                </div>
                <van-field v-model="p.label" label="项目" placeholder="视距内驾驶员" dense />
                <van-field v-model="p.price" label="价格" placeholder="¥4,800起" dense />
              </div>
              <div class="list-add">
                <van-button size="small" type="primary" block plain icon="plus" @click="trainingPrices.push({ label: '', price: '' })">添加</van-button>
              </div>
            </van-cell-group>

            <van-cell-group inset title="教学特色">
              <div v-for="(f, idx) in trainingFeatures" :key="idx" class="list-item-block">
                <div class="list-item-head">
                  <span class="item-num">{{ idx + 1 }}</span>
                  <van-button size="mini" type="danger" plain icon="cross" @click="trainingFeatures.splice(idx, 1)" />
                </div>
                <van-field v-model="f.title" label="标题" placeholder="小班制教学" dense />
                <van-field v-model="f.desc" label="描述" type="textarea" rows="2" placeholder="详细说明" dense autosize />
              </div>
              <div class="list-add">
                <van-button size="small" type="primary" block plain icon="plus" @click="trainingFeatures.push({ title: '', desc: '' })">添加</van-button>
              </div>
            </van-cell-group>

            <van-cell-group inset title="公司简介">
              <van-field v-model="trainingCompanyTitle" label="公司名" placeholder="温州低空科技集团" />
              <van-field v-model="trainingCompanyContent" label="简介" type="textarea" rows="3" placeholder="公司介绍" autosize />
            </van-cell-group>

            <van-cell-group inset title="执照说明">
              <van-field v-model="trainingLicenseContent" label="说明" type="textarea" rows="2" placeholder="执照功能介绍" autosize />
              <van-field v-model="trainingLicenseQuote" label="法规" type="textarea" rows="2" placeholder="法规条文引用" autosize />
            </van-cell-group>
          </template>
        </div>
      </div>
    </van-popup>

    <!-- 图文展示编辑子弹窗 -->
    <van-popup :show="showStudyItemEditPopup" @update:show="v => showStudyItemEditPopup = v" position="bottom" :style="{ height: '65%' }" round>
      <div style="padding: 16px 0;" v-if="studyEditingItem">
        <van-cell-group title="编辑图文">
          <van-field v-model="studyEditingItem.title" label="标题" placeholder="请输入标题" />
          <van-field v-model="studyEditingItem.desc" label="描述" type="textarea" rows="2" placeholder="请输入描述" />
          <van-field label="图片">
            <template #input>
              <van-uploader :after-read="onReadStudyImage" max-count="1" accept="image/*">
                <van-button icon="plus" type="primary" size="small" plain>上传图片</van-button>
              </van-uploader>
            </template>
          </van-field>
          <div v-if="studyEditingItem.image" style="padding: 0 16px 16px;">
            <img :src="normalizeMediaUrl(studyEditingItem.image)" style="width:100%; border-radius: 8px; display:block;" />
          </div>
        </van-cell-group>
        <div style="margin: 16px;">
          <van-button round block type="primary" @click="confirmStudyItemEdit">确定</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 图片裁剪弹窗 -->
    <ImageCropper
      v-model:show="showCropper"
      :image-url="cropperImageUrl"
      :aspect-ratio="cropperAspectRatio"
      :title="cropperTitle"
      @confirm="onCropConfirm"
    />

    <!-- 精彩回顾编辑弹窗 -->
    <van-popup :show="showShowcaseEditPopup" @update:show="v => showShowcaseEditPopup = v" position="bottom" :style="{ height: '65%' }" round>
      <div style="padding: 16px 0;" v-if="showcaseEditingItem">
        <van-cell-group title="编辑精彩回顾">
          <van-field v-model="showcaseEditingItem.title" label="标题" placeholder="请输入标题" />
          <van-field v-model="showcaseEditingItem.desc" label="描述" type="textarea" rows="2" placeholder="请输入描述" />
          <van-field label="图片">
            <template #input>
              <van-uploader :after-read="onReadShowcaseImage" max-count="1" accept="image/*">
                <van-button icon="plus" type="primary" size="small" plain>上传图片</van-button>
              </van-uploader>
            </template>
          </van-field>
          <div v-if="showcaseEditingItem.image" style="padding: 0 16px 16px;">
            <img :src="showcaseEditingItem.image" style="width:100%; border-radius: 8px; display:block;" @click="previewCrop(showcaseEditingItem.image, 'showcaseEditing')" />
            <div style="margin-top:8px; text-align:center;">
              <van-button size="small" type="primary" plain @click="previewCrop(showcaseEditingItem.image, 'showcaseEditing')">裁剪图片</van-button>
            </div>
          </div>
        </van-cell-group>
        <div style="margin: 16px;">
          <van-button round block type="primary" @click="confirmShowcaseEdit">确定</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 首页配置弹窗 -->
    <van-popup :show="showHomeConfigPopup" @update:show="v => showHomeConfigPopup = v" position="bottom" :style="{ height: '70%' }" round>
      <div style="padding: 16px 0;" v-if="editingHomeConfig">
        <van-nav-bar title="首页配置" left-text="取消" right-text="保存" @click-left="showHomeConfigPopup = false" @click-right="saveHomeConfig" />
        <div style="padding-bottom: 40px;">
          <van-cell-group title="背景图">
            <van-field label="上传">
              <template #input>
                <van-uploader :after-read="onReadHomeHeaderImage" max-count="1" accept="image/*">
                  <van-button icon="plus" type="primary" size="small" plain>选择图片</van-button>
                </van-uploader>
              </template>
            </van-field>
            <div v-if="editingHomeConfig.headerImage" style="padding: 0 16px 16px;">
              <img :src="normalizeMediaUrl(editingHomeConfig.headerImage)" style="width:100%; border-radius: 8px; display:block;" />
            </div>
          </van-cell-group>
          <van-cell-group title="轮播消息">
            <div v-for="(msg, idx) in editingHomeConfig.notices" :key="idx" class="list-item" style="padding: 4px 16px;">
              <van-field
                :model-value="msg"
                @update:model-value="v => editingHomeConfig.notices[idx] = v"
                placeholder="请输入通知消息"
                dense
              />
              <van-button size="mini" type="danger" plain icon="cross" @click="editingHomeConfig.notices.splice(idx, 1)" />
            </div>
            <div class="list-add">
              <van-button size="small" type="primary" block plain icon="plus" @click="editingHomeConfig.notices.push('')">添加消息</van-button>
            </div>
          </van-cell-group>
          <van-cell-group title="轮播 Banner">
            <div v-for="(banner, idx) in editingHomeConfig.banners" :key="idx" style="padding: 8px 16px; border-bottom: 1px solid #f5f5f5;">
              <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
                <span style="font-size: 12px; color: #999;">#{{ idx + 1 }}</span>
                <van-button size="mini" type="danger" plain icon="cross" @click="editingHomeConfig.banners.splice(idx, 1)" />
              </div>
              <div v-if="banner.image" style="margin-bottom: 8px;">
                <img :src="normalizeMediaUrl(banner.image)" style="width:100%; height: 80px; object-fit: cover; border-radius: 6px; display:block;" />
              </div>
              <van-uploader :after-read="file => onReadBannerImage(file, idx)" max-count="1" accept="image/*">
                <van-button icon="photo-o" type="primary" size="mini" plain>{{ banner.image ? '更换图片' : '上传图片' }}</van-button>
              </van-uploader>
              <van-field
                v-model="banner.link"
                label="链接"
                placeholder="如 delivery 或 /cases/1"
                dense
                style="margin-top: 4px;"
              />
            </div>
            <div class="list-add">
              <van-button size="small" type="primary" block plain icon="plus" @click="addBanner">添加 Banner</van-button>
            </div>
          </van-cell-group>
        </div>
      </div>
    </van-popup>

    <!-- 新增课程包弹窗 -->
    <van-popup :show="showAddPackagePopup" @update:show="v => showAddPackagePopup = v" position="bottom" :style="{ height: '50%' }" round>
      <div style="padding: 16px 0;">
        <van-nav-bar title="新增课程包" left-text="取消" right-text="确定" @click-left="showAddPackagePopup = false" @click-right="confirmAddPackage" />
        <div style="padding: 16px;">
          <van-cell-group>
            <van-field v-model="newPackage.id" label="标识" placeholder="如：study-fullday（英文，唯一）" />
            <van-field v-model="newPackage.tag" label="标签" placeholder="如：全日营" />
            <van-field v-model.number="newPackage.price" label="价格" placeholder="298" type="number" />
          </van-cell-group>
          <div style="margin-top: 16px; padding: 12px; background: #f5f5f7; border-radius: 8px; font-size: 12px; color: #666;">
            <p>提示：标识建议使用 study- 前缀，如 study-fullday、study-summer 等</p>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted } from 'vue'
import axios from '@/utils/http'
import { showFailToast, showSuccessToast, showLoadingToast, closeToast } from 'vant'
import ImageCropper from './ImageCropper.vue'
import DataToolbar from '../components/DataToolbar.vue'
import { normalizeMediaUrl, uploadFile } from '../composables/useMedia'
import { useAuth } from '../composables/useAuth'

const { userRole, isAdmin, isStudyAdmin } = useAuth()

const DEFAULT_HOME_CONFIG = {
  headerImage: '',
  headerImagePosition: 'center',
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
  { title: '核心服务', ids: ['2', '8', '1'], roles: ['admin'] },
  { title: '商业应用', ids: ['4', '5', '3', '7', '13'], roles: ['admin', 'dsl_admin'] },
  { title: '研学教育', ids: ['9'], roles: ['admin', 'study_admin'] },
  { title: '增值服务', ids: ['10', '11', '12'], roles: ['admin'] }
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
  if (isAdmin.value) {
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
    allServiceConfigs.value = res.data.data || {}
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
    allServiceConfigs.value = newConfigs
    homeConfig.value = JSON.parse(JSON.stringify(editingHomeConfig.value))
    closeToast()
    showSuccessToast('保存成功')
    showHomeConfigPopup.value = false
  } catch (error) {
    closeToast()
    showFailToast('保存失败')
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
    price: 0,
    recommended: false,
    desc: '',
    cardHighlights: [],
    headerBg: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)',
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
  showShowcaseEditPopup.value = true
}

// 编辑精彩回顾
const editShowcaseItem = (idx) => {
  if (!activeStudyPkg.value?.showcase) return
  showcaseEditingIndex.value = idx
  showcaseEditingItem.value = { ...activeStudyPkg.value.showcase[idx] }
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
  studyPackages.value[id].price = Number(price) || 0
  studyPackageIds.value.push(id)
  activeStudyPkgId.value = id
  showAddPackagePopup.value = false
  showSuccessToast('添加成功')
}

// 删除课程包
const removeStudyPackage = (pkgId) => {
  if (studyPackageIds.value.length <= 1) {
    showFailToast('至少保留一个课程包')
    return
  }
  const idx = studyPackageIds.value.indexOf(pkgId)
  if (idx > -1) {
    studyPackageIds.value.splice(idx, 1)
    delete studyPackages.value[pkgId]
    // 切换到第一个课程包
    activeStudyPkgId.value = studyPackageIds.value[0]
    showSuccessToast('删除成功')
  }
}

const addCourseStep = () => {
  if (!editingService.value.courseSchedule) editingService.value.courseSchedule = []
  editingService.value.courseSchedule.push({ time: '', content: '', remark: '' })
}

// --- Study showcase ---
const innerAddStudyItem = () => {
  studyEditingIndex.value = -1
  studyEditingItem.value = { title: '', desc: '', image: '' }
  showStudyItemEditPopup.value = true
}

const innerEditStudyItem = (idx) => {
  studyEditingIndex.value = idx
  studyEditingItem.value = { ...editingService.value.studyShowcase[idx] }
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
  try {
    const formData = new FormData()
    formData.append('file', file.file)
    const res = await axios.post('/api/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    if (res.data.success && studyEditingItem.value) {
      studyEditingItem.value = { ...studyEditingItem.value, image: res.data.url }
    } else {
      showFailToast('上传失败')
    }
  } catch (error) {
    showFailToast('上传失败')
  }
}

onMounted(fetchAllServiceConfigs)
</script>

<style scoped>
.toolbar-label { font-size: 14px; font-weight: 500; color: var(--text-color); }

/* 编辑弹窗 */
.edit-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--background-color, #f5f5f7);
}
.edit-popup :deep(.van-nav-bar) {
  flex-shrink: 0;
}
.edit-body {
  flex: 1;
  overflow-y: auto;
  padding: 0 0 80px;
}
.edit-body :deep(.van-cell-group--inset) {
  margin: 12px 12px 0;
}

/* 列表项：单行 */
.list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 16px;
}
.list-item :deep(.van-field) {
  flex: 1;
  padding: 6px 0;
}

/* 列表项：多行 */
.list-item-block {
  padding: 8px 16px;
  border-bottom: 1px solid #f5f5f7;
}
.list-item-block:last-of-type {
  border-bottom: none;
}
.list-item-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2px;
}
.item-num {
  font-size: 12px;
  font-weight: 600;
  color: var(--accent-color, #0071e3);
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
  padding: 8px 16px;
}
.list-add .van-button {
  border-style: dashed;
}

/* 图文展示行 */
.showcase-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  border-bottom: 1px solid #f5f5f7;
}
.showcase-left {
  width: 48px;
  height: 36px;
  border-radius: 6px;
  overflow: hidden;
  background: #f5f5f7;
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
  color: var(--text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 课程包 Tab */
.pkg-tabs {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
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
  color: #86868b;
  background: #f5f5f7;
  cursor: pointer;
  transition: all 0.25s;
}
.pkg-tab.active {
  background: var(--accent-color, #0071e3);
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
  color: var(--accent-color, #0071e3);
}

/* 背景输入 */
.bg-input-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

/* 图片预览 */
.img-preview {
  padding: 0 16px 12px;
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
  padding: 8px 16px;
}

/* 背景图预览 */
.img-preview {
  padding: 0 16px 12px;
}
.img-preview img {
  width: 100%;
  border-radius: 8px;
  display: block;
}
</style>
