<template>
  <div class="admin-page">
    <van-nav-bar
      title="后台数据管理"
      left-arrow
      @click-left="$router.push('/')"
      fixed
      placeholder
    >
        <template #right>
            <van-icon name="wap-home-o" size="18" @click="$router.push('/')"/>
        </template>
    </van-nav-bar>

    <div class="filter-section" v-if="activeTab === 0 || activeTab === 3">
      <van-cell title="选择日期范围" :value="dateRange" is-link @click="showCalendar = true" />
      <van-calendar 
        :show="showCalendar"
        @update:show="(v) => (showCalendar = v)"
        type="range" 
        @confirm="onConfirmDate" 
        :min-date="new Date(2024, 0, 1)"
      />
      
      <div class="action-buttons">
        <van-button type="primary" size="small" block @click="fetchData">查询</van-button>
        <van-button type="success" size="small" block @click="handleExport" style="margin-top: 8px;">
          导出{{ activeTab === 3 ? '赛事' : '' }}Excel
        </van-button>
      </div>
    </div>

    <van-tabs
      :active="activeTab"
      @update:active="(v) => (activeTab = v)"
      sticky
      offset-top="46"
    >
        <van-tab title="订单管理" v-if="userRole === 'admin'">
            <div class="data-list">
                <div class="selection-actions" v-if="list.length > 0" style="padding: 10px 16px; background: #fff; border-bottom: 1px solid #ebedf0; display: flex; justify-content: space-between; align-items: center;">
                    <van-checkbox v-model="allSelected" @click="toggleSelectAll">全选 ({{ selectedIds.length }})</van-checkbox>
                    <van-button type="success" size="mini" :disabled="selectedIds.length === 0" @click="handleSelectiveExport">导出所选</van-button>
                </div>
                <van-empty v-if="list.length === 0" description="暂无数据" />
                <van-cell-group v-else inset>
                    <van-cell 
                        v-for="item in list" 
                        :key="item.id" 
                        :label="formatDate(item.createTime)"
                        is-link
                        @click="showDetail(item)"
                    >
                        <template #title>
                            <div style="display: flex; align-items: center;">
                                <van-checkbox v-model="item.selected" @click.stop style="margin-right: 8px;" />
                                <span>{{ item.serviceName || '未知服务' }}</span>
                            </div>
                        </template>
                        <template #value>
                            <span :class="getStatusClass(item.status)">{{ item.status || '待处理' }}</span>
                        </template>
                    </van-cell>
                </van-cell-group>
            </div>
        </van-tab>
        <van-tab title="案例管理" v-if="userRole === 'admin'">
            <div style="padding: 12px;">
                <div style="display: flex; gap: 10px; margin-bottom: 12px;">
                    <van-button type="primary" block icon="plus" @click="createCase">新增案例</van-button>
                    <van-button type="default" block icon="replay" @click="fetchCases">刷新</van-button>
                </div>

                <van-empty v-if="cases.length === 0" description="暂无案例数据" />

                <van-cell-group v-else inset>
                    <van-cell
                        v-for="caseItem in cases"
                        :key="caseItem.id"
                        is-link
                        @click="editCase(caseItem)"
                    >
                        <template #title>
                            <div style="display:flex; align-items:center; gap:10px;">
                                <div style="width: 56px; height: 56px; border-radius: 8px; overflow: hidden; background: #f7f8fa; flex: 0 0 56px;">
                                    <img
                                        v-if="caseItem.coverType !== 'video'"
                                        :src="caseItem.cover"
                                        :alt="caseItem.title"
                                        style="width: 100%; height: 100%; object-fit: cover; display:block;"
                                    />
                                    <video
                                        v-else
                                        :src="caseItem.cover"
                                        muted
                                        playsinline
                                        preload="metadata"
                                        style="width: 100%; height: 100%; object-fit: cover; display:block;"
                                    ></video>
                                </div>
                                <div style="min-width: 0;">
                                    <div style="font-weight: 600; color: #1a1a1a; line-height: 1.2; margin-bottom: 4px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">
                                        {{ caseItem.title || '未命名案例' }}
                                    </div>
                                    <div style="font-size: 12px; color: #969799; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">
                                        分类：{{ caseItem.categoryId }} · {{ caseItem.location || '-' }} · {{ caseItem.date || '-' }}
                                    </div>
                                </div>
                            </div>
                        </template>
                        <template #value>
                            <div style="display:flex; flex-direction: column; align-items: flex-end; gap: 6px;">
                                <van-tag type="primary" size="medium">{{ caseItem.coverType === 'video' ? '视频' : '图片' }}</van-tag>
                                <span style="font-size: 12px; color: #969799;">{{ caseItem.views || '-' }}</span>
                            </div>
                        </template>
                    </van-cell>
                </van-cell-group>
            </div>
        </van-tab>
        <van-tab title="用户管理" v-if="userRole === 'admin'">
            <div style="padding: 12px;">
                <div style="display: flex; gap: 10px; margin-bottom: 12px;">
                    <van-button type="default" block icon="replay" @click="fetchUsers">刷新</van-button>
                </div>

                <van-empty v-if="users.length === 0" description="暂无用户数据" />

                <van-cell-group v-else inset>
                    <van-cell v-for="u in users" :key="u.id">
                        <template #title>
                            <div style="display:flex; flex-direction: column; gap: 4px;">
                                <div style="font-weight: 600; color: #1a1a1a;">
                                    {{ u.name || '-' }}
                                </div>
                                <div style="font-size: 12px; color: #969799;">
                                    {{ u.phone || '-' }}
                                </div>
                            </div>
                        </template>
                        <template #value>
                            <div style="display:flex; flex-direction: column; align-items: flex-end; gap: 6px;">
                                <van-tag
                                    :type="u.role === 'admin' ? 'success' : (u.role === 'dsl_admin' ? 'primary' : 'default')"
                                    size="medium"
                                >
                                    {{ u.role || '-' }}
                                </van-tag>
                                <van-button
                                    v-if="isSuperAdmin && u.role !== 'dsl_admin' && u.phone !== SUPER_ADMIN_PHONE"
                                    size="mini"
                                    type="primary"
                                    plain
                                    @click="toggleUserRole(u)"
                                >
                                    切换权限
                                </van-button>
                            </div>
                        </template>
                    </van-cell>
                </van-cell-group>
            </div>
        </van-tab>
        <van-tab title="赛事管理" v-if="userRole === 'admin' || userRole === 'dsl_admin'">
            <div class="competition-admin">
                <!-- 搜索和筛选 -->
                <div class="search-filter-area" style="background: #fff; padding-bottom: 10px; margin-bottom: 10px;">
                    <van-search
                        v-model="searchText"
                        placeholder="搜索姓名、单位或手机号"
                        @search="onFilterChange"
                        @clear="onFilterChange"
                    />
                    <div class="role-filter" style="padding: 0 12px;">
                        <van-dropdown-menu>
                            <van-dropdown-item v-model="selectedRole" :options="roleFilterOptions" @change="onFilterChange" />
                            <van-dropdown-item v-model="selectedStatus" :options="statusFilterOptions" @change="onFilterChange" />
                        </van-dropdown-menu>
                    </div>
                </div>

                <!-- 赛事统计卡片 -->
                <div class="stats-grid" style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; padding: 12px;">
                    <div class="stat-card" style="background: #fff; padding: 12px; border-radius: 8px; text-align: center; grid-column: span 3;">
                        <div style="font-size: 12px; color: #969799;">总报名人数</div>
                        <div style="font-size: 24px; font-weight: bold; color: #1677ff;">{{ competitionStats.total }}</div>
                    </div>
                    <div class="stat-card" style="background: #fff; padding: 12px; border-radius: 8px; text-align: center;">
                        <div style="font-size: 11px; color: #969799;">运动员</div>
                        <div style="font-size: 18px; font-weight: bold; color: #ee0a24;">{{ competitionStats.athlete }}</div>
                    </div>
                    <div class="stat-card" style="background: #fff; padding: 12px; border-radius: 8px; text-align: center;">
                        <div style="font-size: 11px; color: #969799;">教练员</div>
                        <div style="font-size: 18px; font-weight: bold; color: #ff976a;">{{ competitionStats.coach }}</div>
                    </div>
                    <div class="stat-card" style="background: #fff; padding: 12px; border-radius: 8px; text-align: center;">
                        <div style="font-size: 11px; color: #969799;">裁判员</div>
                        <div style="font-size: 18px; font-weight: bold; color: #07c160;">{{ competitionStats.referee }}</div>
                    </div>
                    <div class="stat-card" style="background: #fff; padding: 12px; border-radius: 8px; text-align: center; grid-column: span 3;">
                        <div style="font-size: 11px; color: #969799;">俱乐部</div>
                        <div style="font-size: 18px; font-weight: bold; color: #722ed1;">{{ competitionStats.club }}</div>
                    </div>
                </div>

                <div class="selection-actions" v-if="competitionList.length > 0" style="padding: 10px 16px; background: #fff; margin: 0 16px; border-radius: 8px 8px 0 0; display: flex; justify-content: space-between; align-items: center;">
                    <van-checkbox v-model="allSelected" @click="toggleSelectAll">全选 ({{ selectedIds.length }})</van-checkbox>
                    <van-button type="success" size="mini" :disabled="selectedIds.length === 0" @click="handleSelectiveExport">导出所选</van-button>
                </div>

                <van-cell-group inset :title="competitionList.length > 0 ? '' : '报名列表'" :style="{ borderRadius: competitionList.length > 0 ? '0 0 8px 8px' : '8px' }">
                    <van-cell 
                        v-for="item in competitionList" 
                        :key="item.id" 
                        :label="formatDate(item.createTime)"
                        is-link
                        @click="showDetail(item)"
                    >
                        <template #title>
                            <div style="display: flex; align-items: center;">
                                <van-checkbox v-model="item.selected" @click.stop style="margin-right: 8px;" />
                                <span>{{ (item.name || item.companyName || '未知姓名') + ' - ' + (item.competitionRoleText || '未知角色') }}</span>
                            </div>
                        </template>
                        <template #value>
                            <span :class="getStatusClass(item.status)">{{ item.status || '待处理' }}</span>
                        </template>
                    </van-cell>
                </van-cell-group>
            </div>
        </van-tab>
        <van-tab title="服务配置" v-if="userRole === 'admin'">
            <div style="padding: 12px;">
                <div style="margin-bottom: 12px; display: flex; gap: 10px;">
                    <van-button type="default" block icon="replay" @click="fetchAllServiceConfigs">刷新配置</van-button>
                </div>
                
                <van-cell-group inset title="首页配置" style="margin-bottom: 12px;">
                    <van-cell
                        title="首页背景图 & 轮播消息"
                        :label="(homeConfig?.headerImage ? '背景图已配置' : '背景图未配置') + '  ·  ' + (homeConfig?.notices?.length || 0) + ' 条轮播消息'"
                        is-link
                        @click="editHomeConfig"
                    />
                </van-cell-group>

                <van-cell-group inset title="选择要编辑的服务">
                    <van-cell
                        v-for="[id, cfg] in serviceConfigEntries"
                        :key="id"
                        :title="cfg.name"
                        :label="cfg.slogan"
                        is-link
                        @click="editServiceConfig(id)"
                    />
                </van-cell-group>
            </div>
        </van-tab>
    </van-tabs>

    <!-- Detail Popup -->
    <van-popup
      :show="showDetailPopup"
      @update:show="(v) => (showDetailPopup = v)"
      position="bottom"
      :style="{ height: '70%' }"
      round
    >
        <div class="detail-content" v-if="currentItem">
            <van-cell-group title="基本信息">
                <van-cell title="申请单号" :value="itemValue(currentItem.id)" />
                <van-cell title="申请时间" :value="formatDate(currentItem.createTime)" />
                <van-cell title="当前状态" 
                    :value="currentItem.status" 
                    is-link 
                    @click="showStatusPicker = true"
                />
                <template v-if="currentItem.serviceId === '6'">
                  <van-cell title="姓名" :value="itemValue(currentItem.traineeName)" />
                  <van-cell title="联系电话" :value="itemValue(currentItem.traineePhone)" />
                </template>
                <template v-else-if="currentItem.serviceId === '9'">
                  <van-cell title="学校/机构" :value="itemValue(currentItem.studyOrg)" />
                  <van-cell v-if="currentItem.studyGrade" title="年级/年龄段" :value="itemValue(currentItem.studyGrade)" />
                  <van-cell title="参与人数" :value="itemValue(currentItem.studyParticipants)" />
                  <van-cell title="期望日期" :value="itemValue(currentItem.studyDate)" />
                  <van-cell title="场次" :value="itemValue(currentItem.studySessionText || currentItem.studySession)" />
                </template>
                <template v-else-if="currentItem.serviceId === '13'">
                  <van-cell v-if="currentItem.name || currentItem.manager" :title="currentItem.competitionRole === 'club' ? '负责人' : '姓名'" :value="itemValue(currentItem.name || currentItem.manager)" />
                  <van-cell v-if="currentItem.companyName" title="单位名称" :value="itemValue(currentItem.companyName)" />
                  <van-cell title="联系电话" :value="itemValue(currentItem.phone || currentItem.managerPhone || currentItem.contactPhone)" />
                </template>
                <template v-else>
                  <van-cell title="联系人" :value="itemValue(currentItem.contactName)" />
                  <van-cell title="联系电话" :value="itemValue(currentItem.contactPhone)" />
                </template>
            </van-cell-group>

            <van-cell-group title="服务详情">
                <van-cell title="服务类型" :value="itemValue(currentItem.serviceName)" />
                
                <!-- 无人机赛事报名详情 -->
                <template v-if="currentItem.serviceId === '13'">
                    <van-cell title="注册号" :value="itemValue(currentItem.regNo)" />
                    <van-cell title="报名角色" :value="itemValue(currentItem.competitionRoleText)" />
                    <template v-if="currentItem.competitionRole === 'club'">
                        <van-cell title="单位简称" :value="itemValue(currentItem.companyShortName)" />
                        <van-cell title="所在地" :value="itemValue(currentItem.location)" />
                        <van-cell title="负责人" :value="itemValue(currentItem.manager)" />
                        <van-cell title="负责人电话" :value="itemValue(currentItem.managerPhone)" />
                        <van-cell title="主要对接人" :value="itemValue(currentItem.contactPerson)" />
                        <van-cell title="对接人电话" :value="itemValue(currentItem.contactPhone)" />
                    </template>
                    <template v-else>
                        <van-cell title="性别" :value="currentItem.gender === 'male' ? '男' : '女'" />
                        <van-cell title="证件号" :value="itemValue(currentItem.idCard)" />
                        <van-cell title="组别" :value="itemValue(currentItem.competitionGroup || currentItem.athleteGroup)" />
                        <van-cell v-if="currentItem.competitionProject" title="参赛项目" :value="itemValue(currentItem.competitionProject)" />
                        <van-cell 
                            :title="currentItem.competitionRole === 'referee' ? '裁判员等级' : (currentItem.competitionRole === 'coach' ? '教练员等级' : '等级')" 
                            :value="itemValue(currentItem.level)" 
                        />
                        <van-cell v-if="currentItem.validDate" title="有效期" :value="itemValue(currentItem.validDate)" />
                        <van-cell title="电子邮箱" :value="itemValue(currentItem.email)" />
                    </template>
                </template>

                <!-- 飞手培训服务详情 -->
                <template v-else-if="currentItem.serviceId === '6'">
                    <van-cell title="性别" :value="currentItem.traineeGender === 'male' ? '男' : '女'" />
                    <van-cell title="出生日期" :value="itemValue(currentItem.traineeBirthday)" />
                    <van-cell title="身份证号" :value="itemValue(currentItem.traineeIdCard)" />
                    <van-cell title="执照种类" :value="itemValue(currentItem.examModel)" />
                    <van-cell title="证照级别" :value="itemValue(currentItem.licenseLevel)" />
                    <van-cell title="有无基础" :value="currentItem.hasExperience === 'yes' ? '有' : '无'" />
                </template>

                <!-- 物流服务详情 -->
                <template v-else>
                    <van-cell title="客户类型" :value="currentItem.customerType === 'enterprise' ? '企业' : '个人'" />
                    <van-cell v-if="currentItem.companyName" title="企业名称" :value="currentItem.companyName" />
                    <van-cell v-if="currentItem.cargoType" title="货物类型" :value="currentItem.cargoType" />
                    <van-cell v-if="currentItem.startAddress" title="起运地" :value="currentItem.startAddress" />
                    <van-cell v-if="currentItem.endAddress" title="目的地" :value="currentItem.endAddress" />
                </template>
                
                <van-cell title="备注" :label="itemValue(currentItem.remark)" />
            </van-cell-group>
        </div>
    </van-popup>
    <!-- Status Picker -->
    <van-popup
      :show="showStatusPicker"
      @update:show="(v) => (showStatusPicker = v)"
      position="bottom"
    >
        <van-picker
            :columns="statusOptions"
            @confirm="onUpdateStatus"
            @cancel="showStatusPicker = false"
            title="修改订单状态"
        />
    </van-popup>

    <!-- Case Edit Popup -->
    <van-popup
      :show="showCaseEditPopup"
      @update:show="(v) => (showCaseEditPopup = v)"
      position="bottom"
      :style="{ height: '90%' }"
      round
    >
        <div class="detail-content" v-if="currentCase">
            <van-form @submit="onSaveCase">
                <van-cell-group title="基本信息">
                    <van-field name="categoryId" label="所属分类" required>
                        <template #input>
                            <van-radio-group v-model="currentCase.categoryId" direction="horizontal">
                                <van-radio :name="1">无人机物流</van-radio>
                                <van-radio :name="4">无人机吊运</van-radio>
                                <van-radio :name="5">无人机表演</van-radio>
                            </van-radio-group>
                        </template>
                    </van-field>
                    <van-field v-model="currentCase.title" label="标题" placeholder="请输入标题" required />
                    <van-field v-model="currentCase.description" label="简介" type="textarea" rows="2" placeholder="请输入简介" />
                    <van-field v-model="currentCase.location" label="地点" placeholder="请输入地点" />
                    <van-field v-model="currentCase.date" label="时间" placeholder="请输入时间" />
                </van-cell-group>
                
                <van-cell-group title="封面设置">
                    <van-field name="coverType" label="封面类型">
                        <template #input>
                            <van-radio-group v-model="currentCase.coverType" direction="horizontal">
                                <van-radio name="image">图片</van-radio>
                                <van-radio name="video">视频</van-radio>
                            </van-radio-group>
                        </template>
                    </van-field>
                    
                    <van-field v-model="currentCase.cover" label="封面地址" placeholder="输入URL或上传" />
                    
                    <van-field label="上传封面">
                        <template #input>
                            <van-uploader 
                                :after-read="file => onReadCover(file)" 
                                max-count="1"
                                :accept="currentCase.coverType === 'video' ? 'video/*' : 'image/*'"
                            >
                                <van-button icon="plus" type="primary" size="small" plain>
                                    上传{{ currentCase.coverType === 'video' ? '视频' : '图片' }}
                                </van-button>
                            </van-uploader>
                        </template>
                    </van-field>
                </van-cell-group>

                <van-cell-group title="媒体资源 (图片/视频)">
                    <div v-for="(media, index) in currentCase.media" :key="index" style="padding: 10px; border-bottom: 1px dashed #eee;">
                        <div style="display: flex; justify-content: space-between; margin-bottom: 8px;">
                            <span style="font-weight: bold;">资源 #{{ index + 1 }}</span>
                            <van-button size="mini" type="danger" icon="cross" @click="removeMedia(index)"></van-button>
                        </div>
                        <van-radio-group v-model="media.type" direction="horizontal" style="margin-bottom: 8px;">
                            <van-radio name="image">图片</van-radio>
                            <van-radio name="video">视频</van-radio>
                        </van-radio-group>
                        <van-field v-model="media.url" label="地址" placeholder="URL" />
                        <div style="margin-top: 8px;">
                            <van-uploader 
                                :after-read="file => onReadMedia(file, index)"
                                :accept="media.type === 'video' ? 'video/*' : 'image/*'"
                            >
                                <van-button icon="plus" size="mini" type="default">
                                    上传{{ media.type === 'video' ? '视频' : '图片' }}
                                </van-button>
                            </van-uploader>
                        </div>
                    </div>
                    <div style="padding: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="addMedia">添加资源</van-button>
                    </div>
                </van-cell-group>

                <van-cell-group title="项目亮点 (标签)">
                    <div v-for="(tag, index) in currentCase.highlights" :key="index" style="display: flex; align-items: center; padding: 0 16px;">
                        <van-field v-model="currentCase.highlights[index]" :label="'标签 ' + (index + 1)" placeholder="输入标签内容" />
                        <van-button size="mini" type="danger" icon="cross" @click="removeHighlight(index)" style="margin-left: 8px;"></van-button>
                    </div>
                    <div style="padding: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="addHighlight">添加标签</van-button>
                    </div>
                </van-cell-group>

                <van-cell-group title="详细内容">
                    <van-field 
                        v-model="currentCase.fullDescription" 
                        label="详细描述" 
                        type="textarea" 
                        rows="4" 
                        placeholder="请输入详细描述" 
                        autosize
                    />
                </van-cell-group>

                <div style="margin: 16px; padding-bottom: 30px;">
                    <van-button round block type="primary" native-type="submit" style="margin-bottom: 12px;">保存修改</van-button>
                    <van-button v-if="currentCase.id" round block type="danger" native-type="button" @click="onDeleteCase">删除案例</van-button>
                </div>
            </van-form>
        </div>
    </van-popup>

    <!-- Study Item Edit Popup -->
    <van-popup
      :show="showStudyItemEditPopup"
      @update:show="(v) => (showStudyItemEditPopup = v)"
      position="bottom"
      :style="{ height: '70%' }"
      round
    >
        <div class="detail-content" v-if="studyEditingItem">
            <van-cell-group title="往期回顾内容">
                <van-field v-model="studyEditingItem.title" label="标题" placeholder="请输入标题" />
                <van-field v-model="studyEditingItem.desc" label="描述" type="textarea" rows="2" placeholder="请输入描述" />
                <van-field v-model="studyEditingItem.image" label="图片地址" placeholder="输入URL或上传" />
                <van-field label="上传图片">
                    <template #input>
                        <van-uploader :after-read="onReadStudyImage" max-count="1" accept="image/*">
                            <van-button icon="plus" type="primary" size="small" plain>上传图片</van-button>
                        </van-uploader>
                    </template>
                </van-field>
                <div v-if="studyEditingItem.image" style="padding: 0 16px 16px;">
                    <img :src="normalizeMediaUrl(studyEditingItem.image)" style="width:100%; border-radius: 10px; display:block;" />
                </div>
            </van-cell-group>

            <div style="margin: 16px; padding-bottom: 30px;">
                <van-button round block type="primary" @click="confirmStudyItemEdit">确定</van-button>
            </div>
        </div>
    </van-popup>

    <!-- Service Config Edit Popup -->
    <van-popup
      :show="showServiceEditPopup"
      @update:show="(v) => (showServiceEditPopup = v)"
      position="bottom"
      :style="{ height: '90%' }"
      round
    >
        <div class="detail-content" v-if="editingService">
            <van-nav-bar
                :title="'编辑 - ' + editingService.name"
                left-text="取消"
                right-text="保存"
                @click-left="showServiceEditPopup = false"
                @click-right="saveServiceConfig"
            />
            
            <div style="padding-bottom: 40px;">
                <van-cell-group title="基本信息">
                    <van-field v-model="editingService.name" label="服务名称" placeholder="请输入名称" />
                    <van-field v-model="editingService.slogan" label="口号/标语" placeholder="请输入标语" />
                    <van-field v-model="editingService.intro" label="服务介绍" type="textarea" rows="3" placeholder="请输入介绍" />
                    <van-field v-model="editingService.mainColor" label="主题色" placeholder="例如 #1677ff" />
                    <van-field v-model="editingService.contactPhone" label="联系电话" placeholder="主要联系电话" />
                    <van-field v-model="editingService.contactPhone2" label="咨询热线" placeholder="第二个联系电话（可选）" />
                    <van-field v-model="editingService.address" label="联系地址" type="textarea" rows="2" placeholder="请输入公司地址" />
                </van-cell-group>

                <van-cell-group title="背景图/图标">
                    <van-field v-model="editingService.headerImage" label="背景图URL" placeholder="研学展示用" />
                    <van-field name="headerImagePosition" label="图片对齐">
                        <template #input>
                            <van-radio-group v-model="editingService.headerImagePosition" direction="horizontal">
                                <van-radio name="top">顶部</van-radio>
                                <van-radio name="center">居中</van-radio>
                                <van-radio name="bottom">底部</van-radio>
                            </van-radio-group>
                        </template>
                    </van-field>
                    <van-field label="选区预览">
                        <template #input>
                            <div v-if="editingService.headerImage" class="aspect-preview-container">
                                <div class="aspect-label">Banner 选区 (5:3)</div>
                                <div class="preview-box banner-box" :class="'pos-' + (editingService.headerImagePosition || 'center')">
                                    <img :src="normalizeMediaUrl(editingService.headerImage)" />
                                    <div class="safe-area-marker">页面文字显示区</div>
                                </div>
                            </div>
                            <div v-else style="color: #969799; font-size: 12px;">上传后可查看页面实际选区</div>
                        </template>
                    </van-field>
                    <van-field label="上传背景图">
                        <template #input>
                            <van-uploader :after-read="file => onReadServiceFile(file, 'headerImage')" max-count="1">
                                <van-button icon="plus" size="small" type="primary" plain>点击上传海报图</van-button>
                            </van-uploader>
                        </template>
                    </van-field>
                    <van-field v-model="editingService.icon" label="图标名称/URL" placeholder="Vant图标或SVG路径" />
                </van-cell-group>

                <van-cell-group title="服务项目">
                    <div v-for="(p, idx) in editingService.projects" :key="idx" style="padding: 10px; border-bottom: 1px solid #f7f8fa; display: flex; align-items: center; gap: 10px;">
                        <van-field v-model="p.name" label="项目名" dense style="flex: 1;" />
                        <van-field v-model="p.icon" label="图标" dense style="width: 100px;" />
                        <van-button size="mini" type="danger" icon="cross" @click="editingService.projects.splice(idx, 1)" />
                    </div>
                    <div style="padding: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="editingService.projects.push({name:'', icon:''})">添加项目</van-button>
                    </div>
                </van-cell-group>

                <van-cell-group title="服务优势">
                    <div v-for="(adv, idx) in editingService.advantages" :key="idx" style="padding: 5px 10px; display: flex; align-items: center; gap: 10px;">
                        <van-field v-model="editingService.advantages[idx]" dense style="flex: 1;" />
                        <van-button size="mini" type="danger" icon="cross" @click="editingService.advantages.splice(idx, 1)" />
                    </div>
                    <div style="padding: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="editingService.advantages.push('')">添加优势</van-button>
                    </div>
                </van-cell-group>

                <!-- 研学专属：课程安排 -->
                <van-cell-group title="课程安排 (仅研学有效)" v-if="editingServiceId === '9'">
                    <van-field v-model="editingService.studyPrice" label="统一票价" placeholder="例如 198元/人" />
                    <div v-for="(step, idx) in editingService.courseSchedule" :key="idx" style="padding: 10px; border-bottom: 1px solid #f7f8fa;">
                        <van-field v-model="step.time" label="时间段" placeholder="8:50-9:10" dense />
                        <van-field v-model="step.content" label="项目" placeholder="集合签到" dense />
                        <van-field v-model="step.remark" label="备注" placeholder="选填" dense />
                        <div style="text-align: right; margin-top: 5px;">
                            <van-button size="mini" type="danger" @click="editingService.courseSchedule.splice(idx, 1)">删除步骤</van-button>
                        </div>
                    </div>
                    <div style="padding: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="editingService.courseSchedule ? editingService.courseSchedule.push({time:'', content:'', remark:''}) : editingService.courseSchedule = [{time:'', content:'', remark:''}]">添加课程步骤</van-button>
                    </div>
                </van-cell-group>

                <!-- 专属配置：亮点卡片 -->
                <van-cell-group :title="(editingServiceId === '9' ? '研学' : '培训') + '亮点'" v-if="['9', '6'].includes(editingServiceId)">
                    <div v-for="(hl, idx) in editingService.highlights" :key="idx" style="padding: 10px; border-bottom: 1px solid #f7f8fa;">
                        <van-field v-model="hl.title" label="标题" dense />
                        <van-field v-model="hl.desc" label="描述" dense />
                        <van-field v-model="hl.icon" label="图标" dense />
                        <div style="text-align: right; margin-top: 5px;">
                            <van-button size="mini" type="danger" @click="editingService.highlights.splice(idx, 1)">删除</van-button>
                        </div>
                    </div>
                    <div style="padding: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="editingService.highlights.push({title:'', desc:'', icon:''})">添加亮点</van-button>
                    </div>
                </van-cell-group>

                <!-- 专属配置：精彩回顾 -->
                <van-cell-group :title="(editingServiceId === '9' ? '研学' : '培训') + '展示'" v-if="['9', '6'].includes(editingServiceId)">
                    <div v-for="(item, idx) in editingService.studyShowcase" :key="idx" style="padding: 10px; border-bottom: 1px solid #f7f8fa;">
                        <div style="display:flex; gap:10px; align-items:center; margin-bottom: 10px;">
                            <div style="width: 60px; height: 45px; border-radius: 4px; overflow: hidden; background: #f7f8fa; flex: 0 0 60px;">
                                <img v-if="item.image" :src="normalizeMediaUrl(item.image)" style="width: 100%; height: 100%; object-fit: cover;" />
                            </div>
                            <div style="flex: 1; font-weight: 600; font-size: 14px;">{{ item.title || '未命名回顾' }}</div>
                            <van-button size="mini" type="primary" plain @click="innerEditStudyItem(idx)">编辑</van-button>
                            <van-button size="mini" type="danger" plain @click="editingService.studyShowcase.splice(idx, 1)">删除</van-button>
                        </div>
                    </div>
                    <div style="padding: 10px; display: flex; gap: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="innerAddStudyItem">添加展示</van-button>
                    </div>
                </van-cell-group>
            </div>
        </div>
    </van-popup>

    <!-- Home Config Popup -->
    <van-popup
      :show="showHomeConfigPopup"
      @update:show="(v) => (showHomeConfigPopup = v)"
      position="bottom"
      :style="{ height: '70%' }"
      round
    >
        <div class="detail-content" v-if="editingHomeConfig">
            <van-nav-bar
                title="首页配置"
                left-text="取消"
                right-text="保存"
                @click-left="showHomeConfigPopup = false"
                @click-right="saveHomeConfig"
            />

            <div style="padding-bottom: 40px;">
                <van-cell-group title="首页背景图">
                    <van-field v-model="editingHomeConfig.headerImage" label="图片地址" placeholder="输入URL或上传" />
                    <van-field label="上传图片">
                        <template #input>
                            <van-uploader :after-read="onReadHomeHeaderImage" max-count="1" accept="image/*">
                                <van-button icon="plus" type="primary" size="small" plain>上传图片</van-button>
                            </van-uploader>
                        </template>
                    </van-field>

                    <van-field name="headerImagePosition" label="图片焦点">
                        <template #input>
                            <van-radio-group v-model="editingHomeConfig.headerImagePosition" direction="horizontal">
                                <van-radio name="top">上</van-radio>
                                <van-radio name="center">中</van-radio>
                                <van-radio name="bottom">下</van-radio>
                            </van-radio-group>
                        </template>
                    </van-field>

                    <div v-if="editingHomeConfig.headerImage" style="padding: 0 16px 16px;">
                        <div class="aspect-preview-container">
                            <div class="aspect-label">背景预览（固定高度，不受金刚区调参影响）</div>
                            <div
                                class="preview-box"
                                :class="editingHomeConfig.headerImagePosition === 'top' ? 'pos-top' : (editingHomeConfig.headerImagePosition === 'bottom' ? 'pos-bottom' : 'pos-center')"
                                style="height: 200px;"
                            >
                                <img :src="normalizeMediaUrl(editingHomeConfig.headerImage)" alt="preview" />
                            </div>
                        </div>
                    </div>
                </van-cell-group>

                <van-cell-group title="轮播消息（首页通知栏）">
                    <div v-for="(msg, idx) in editingHomeConfig.notices" :key="idx" style="display: flex; align-items: center; padding: 0 16px;">
                        <van-field
                            :model-value="msg"
                            @update:model-value="(v) => editingHomeConfig.notices[idx] = v"
                            :label="'消息 ' + (idx + 1)"
                            placeholder="请输入通知消息"
                        />
                        <van-button size="mini" type="danger" icon="cross" @click="editingHomeConfig.notices.splice(idx, 1)" style="margin-left: 8px;" />
                    </div>
                    <div style="padding: 10px;">
                        <van-button size="small" type="primary" block plain icon="plus" @click="editingHomeConfig.notices.push('')">添加消息</van-button>
                    </div>
                </van-cell-group>
            </div>
        </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import axios, { authStorage } from '@/utils/http';
import { showToast, showFailToast, showSuccessToast, showLoadingToast, closeToast, showConfirmDialog } from 'vant';

// 归一化上传/存量媒体地址，避免出现：
// - http://127.0.0.1:8090/uploads/...（线上 https 混合内容/不可达）
// - uploads/xxx（相对路径被拼到 /admin/uploads/xxx）
const normalizeMediaUrl = (raw) => {
    if (!raw || typeof raw !== 'string') return raw;
    const url = raw.trim();
    if (!url) return url;
    if (url.startsWith('data:') || url.startsWith('blob:')) return url;
    if (url.startsWith('uploads/')) return `/${url}`;
    if (url.startsWith('/')) return url;
    if (url.startsWith('http://') || url.startsWith('https://')) {
        try {
            const u = new URL(url);
            const host = u.hostname;
            const port = u.port;
            const isLocalish =
                host === 'localhost' ||
                host === '127.0.0.1' ||
                host === '0.0.0.0' ||
                host === '172.17.0.1' ||
                port === '8090';
            if (isLocalish) {
                return `${u.pathname}${u.search}${u.hash}`;
            }
        } catch (e) {
            // ignore
        }
        return url;
    }
    return url;
};

const list = ref([]);
const cases = ref([]);
const users = ref([]);
const showCalendar = ref(false);
const startDate = ref('');
const endDate = ref('');
const showDetailPopup = ref(false);
const showCaseEditPopup = ref(false);
const showServiceEditPopup = ref(false);
const showStudyItemEditPopup = ref(false);
const currentItem = ref(null);
const currentCase = ref(null);
const studyEditingIndex = ref(-1);
const studyEditingItem = ref(null);

// 服务配置相关
const allServiceConfigs = ref({});
const editingServiceId = ref(null);
const editingService = ref(null);
const showStatusPicker = ref(false);
const activeTab = ref(0);

// User info and permissions
const userStr = localStorage.getItem('user');
const user = userStr ? JSON.parse(userStr) : null;
const userRole = ref(user ? user.role : 'user');
const SUPER_ADMIN_PHONE = 'wzdkjjfzyxgs';
const isSuperAdmin = ref(user ? user.phone === SUPER_ADMIN_PHONE : false);

const refreshCurrentUser = async () => {
    const accessToken = authStorage.getAccessToken();
    if (!accessToken) return;
    try {
        const res = await axios.get('/api/auth/me');
        if (res.data?.success) {
            const current = res.data.user || {};
            localStorage.setItem('user', JSON.stringify(current));
            userRole.value = current.role || 'user';
            isSuperAdmin.value = current.phone === SUPER_ADMIN_PHONE;
        }
    } catch (error) {
        authStorage.clearTokens();
        localStorage.removeItem('user');
    }
};

// Selection logic
const allSelected = ref(false);
const selectedIds = computed(() => {
    const currentList = activeTab.value === 3 ? competitionList.value : list.value;
    // 防御性检查确保是数组
    const listData = Array.isArray(currentList) ? currentList : [];
    return listData.filter(item => item.selected).map(item => item.id);
});

const toggleSelectAll = () => {
    const currentList = activeTab.value === 3 ? competitionList.value : list.value;
    // 防御性检查确保是数组
    const listData = Array.isArray(currentList) ? currentList : [];
    listData.forEach(item => {
        item.selected = allSelected.value;
    });
};

const selectedRole = ref('all');
const selectedStatus = ref('all');
const searchText = ref('');

const roleFilterOptions = [
    { text: '全部角色', value: 'all' },
    { text: '运动员', value: 'athlete' },
    { text: '教练员', value: 'coach' },
    { text: '裁判员', value: 'referee' },
    { text: '俱乐部', value: 'club' }
];

const statusFilterOptions = [
    { text: '全部状态', value: 'all' },
    { text: '待处理', value: '待处理' },
    { text: '处理中', value: '处理中' },
    { text: '已完成', value: '已完成' },
    { text: '已取消', value: '已取消' }
];

const onFilterChange = () => {
    // 触发重新计算
};

const competitionList = computed(() => {
    // 防御性检查确保 list.value 是数组
    const listData = Array.isArray(list.value) ? list.value : [];
    return listData.filter(item => {
        const isCompetition = item.serviceId === '13';
        const matchesRole = selectedRole.value === 'all' || item.competitionRole === selectedRole.value;
        const matchesStatus = selectedStatus.value === 'all' || item.status === selectedStatus.value;
        
        // 关键字搜索：支持姓名、单位、手机号、注册号
        const searchLower = searchText.value.toLowerCase().trim();
        const matchesSearch = !searchLower || 
            (item.name && item.name.toLowerCase().includes(searchLower)) ||
            (item.companyName && item.companyName.toLowerCase().includes(searchLower)) ||
            (item.phone && item.phone.includes(searchLower)) ||
            (item.managerPhone && item.managerPhone.includes(searchLower)) ||
            (item.contactPhone && item.contactPhone.includes(searchLower)) ||
            (item.regNo && item.regNo.toLowerCase().includes(searchLower));

        return isCompetition && matchesRole && matchesStatus && matchesSearch;
    });
});

const competitionStats = computed(() => {
    const stats = { total: 0, athlete: 0, coach: 0, referee: 0, club: 0 };
    // 使用所有赛事数据进行统计，不受当前筛选影响
    // 防御性检查确保 list.value 是数组
    const listData = Array.isArray(list.value) ? list.value : [];
    listData.filter(item => item.serviceId === '13').forEach(item => {
        stats.total++;
        if (item.competitionRole === 'athlete') stats.athlete++;
        else if (item.competitionRole === 'coach') stats.coach++;
        else if (item.competitionRole === 'referee') stats.referee++;
        else if (item.competitionRole === 'club') stats.club++;
    });
    return stats;
});

const statusOptions = [
  { text: '待处理', value: '待处理' },
  { text: '处理中', value: '处理中' },
  { text: '已完成', value: '已完成' },
  { text: '已取消', value: '已取消' },
];

const dateRange = computed(() => {
    if (startDate.value && endDate.value) {
        return `${formatDateShort(startDate.value)} - ${formatDateShort(endDate.value)}`;
    }
    return '全部时间';
});

const formatDateShort = (date) => {
    const d = new Date(date);
    return `${d.getMonth() + 1}/${d.getDate()}`;
};

const formatDate = (dateStr) => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
};

const itemValue = (val) => val || '-';

const getStatusClass = (status) => {
    return status === '已完成' ? 'text-success' : 'text-warning';
};

const onConfirmDate = (values) => {
    const [start, end] = values;
    showCalendar.value = false;
    startDate.value = start;
    endDate.value = end;
    fetchData();
};

const fetchData = async () => {
    try {
        const params = {
            role: userRole.value // 传递当前角色以进行权限过滤
        };
        if (startDate.value) params.startDate = startDate.value;
        if (endDate.value) params.endDate = endDate.value;

        const res = await axios.get('/api/list', { params });
        
        // 确保数据是数组格式
        let data = res.data;
        if (typeof data === 'string') {
            try {
                data = JSON.parse(data);
            } catch (e) {
                console.error('Failed to parse list data:', e);
                data = [];
            }
        }
        if (!Array.isArray(data)) {
            console.error('API /api/list did not return an array:', data);
            data = [];
        }
        
        // 初始化选择状态
        list.value = data.map(item => ({ ...item, selected: false }));
        
        // 如果是DSL管理员且当前在不可见标签，自动切换到赛事管理
        if (userRole.value === 'dsl_admin' && activeTab.value !== 3) {
            activeTab.value = 3;
        }
    } catch (error) {
        showFailToast('获取数据失败');
        console.error(error);
    }
};

const fetchCases = async () => {
    try {
        const res = await axios.get('/api/cases');
        // 兼容：数组 / {data: []} / {cases: []}
        const raw = Array.isArray(res.data) ? res.data : (res.data?.data || res.data?.cases || []);
        cases.value = Array.isArray(raw)
            ? raw.map(c => ({
                ...c,
                cover: normalizeMediaUrl(c.cover),
                media: Array.isArray(c.media) ? c.media.map(m => ({ ...m, url: normalizeMediaUrl(m?.url) })) : c.media,
            }))
            : [];
    } catch (error) {
        showFailToast('获取案例数据失败');
        console.error(error);
    }
};

const fetchUsers = async () => {
    try {
        const res = await axios.get('/api/users');
        // 兼容：数组 / {data: []} / {users: []}
        const raw = Array.isArray(res.data) ? res.data : (res.data?.data || res.data?.users || []);
        users.value = Array.isArray(raw) ? raw : [];
    } catch (error) {
        showFailToast('获取用户数据失败');
        console.error(error);
    }
};

// 研学整合配置逻辑
const innerAddStudyItem = () => {
    studyEditingIndex.value = -1;
    studyEditingItem.value = { title: '', desc: '', image: '' };
    showStudyItemEditPopup.value = true;
};

const innerEditStudyItem = (idx) => {
    const it = editingService.value.studyShowcase[idx];
    studyEditingIndex.value = idx;
    studyEditingItem.value = { ...it };
    showStudyItemEditPopup.value = true;
};

const confirmStudyItemEdit = () => {
    const it = studyEditingItem.value;
    if (!it || !editingService.value) return;
    
    if (!editingService.value.studyShowcase) editingService.value.studyShowcase = [];
    
    if (studyEditingIndex.value >= 0) {
        editingService.value.studyShowcase[studyEditingIndex.value] = { ...it };
    } else {
        editingService.value.studyShowcase.push({ ...it });
    }
    showStudyItemEditPopup.value = false;
};

const onReadStudyImage = async (file) => {
    try {
        const formData = new FormData();
        formData.append('file', file.file);
        const res = await axios.post('/api/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
        if (res.data.success && studyEditingItem.value) {
            studyEditingItem.value.image = res.data.url;
        } else {
            showFailToast('上传失败');
        }
    } catch (error) {
        showFailToast('上传失败');
        console.error(error);
    }
};

// 全局服务配置管理
const DEFAULT_HOME_CONFIG = {
    headerImage: '',
    headerImagePosition: 'center',
    notices: ['交享点无人机外卖配送正式上线', '新开通江心屿无人机外卖配送']
};

const fetchAllServiceConfigs = async () => {
    try {
        const res = await axios.get('/api/services/config');
        allServiceConfigs.value = res.data.data || {};
        homeConfig.value = JSON.parse(JSON.stringify(allServiceConfigs.value._home || DEFAULT_HOME_CONFIG));
    } catch (error) {
        showFailToast('获取服务配置失败');
    }
};

const serviceConfigEntries = computed(() => {
    const entries = Object.entries(allServiceConfigs.value || {}).filter(([id]) => /^\d+$/.test(String(id)));
    entries.sort((a, b) => Number(a[0]) - Number(b[0]));
    return entries;
});

const homeConfig = ref(JSON.parse(JSON.stringify(DEFAULT_HOME_CONFIG)));
const editingHomeConfig = ref(JSON.parse(JSON.stringify(DEFAULT_HOME_CONFIG)));
const showHomeConfigPopup = ref(false);

const editHomeConfig = () => {
    editingHomeConfig.value = JSON.parse(JSON.stringify(homeConfig.value || DEFAULT_HOME_CONFIG));
    if (!Array.isArray(editingHomeConfig.value.notices)) {
        editingHomeConfig.value.notices = [...DEFAULT_HOME_CONFIG.notices];
    }
    showHomeConfigPopup.value = true;
};

const onReadHomeHeaderImage = async (file) => {
    showLoadingToast({ message: '上传中...', forbidClick: true });
    const url = await uploadFile(file);
    closeToast();
    if (url && editingHomeConfig.value) {
        editingHomeConfig.value.headerImage = normalizeMediaUrl(url);
        showSuccessToast('背景图已上传');
    }
};

const saveHomeConfig = async () => {
    try {
        showLoadingToast({ message: '保存中...', forbidClick: true });
        const newConfigs = { ...allServiceConfigs.value };
        newConfigs._home = editingHomeConfig.value;
        await axios.post('/api/services/config', { config: newConfigs });
        allServiceConfigs.value = newConfigs;
        homeConfig.value = JSON.parse(JSON.stringify(editingHomeConfig.value));
        closeToast();
        showSuccessToast('保存成功');
        showHomeConfigPopup.value = false;
    } catch (error) {
        closeToast();
        showFailToast('保存失败');
    }
};

const editServiceConfig = (id) => {
    editingServiceId.value = id;
    editingService.value = JSON.parse(JSON.stringify(allServiceConfigs.value[id]));
    if (!editingService.value.projects) editingService.value.projects = [];
    if (!editingService.value.advantages) editingService.value.advantages = [];
    if (['9', '6'].includes(id)) {
        if (!editingService.value.highlights) editingService.value.highlights = [];
        if (!editingService.value.studyShowcase) editingService.value.studyShowcase = [];
    }
    showServiceEditPopup.value = true;
};

const saveServiceConfig = async () => {
    try {
        showLoadingToast({ message: '保存中...', forbidClick: true });
        const newConfigs = { ...allServiceConfigs.value };
        newConfigs[editingServiceId.value] = editingService.value;
        
        await axios.post('/api/services/config', { config: newConfigs });
        allServiceConfigs.value = newConfigs;
        
        closeToast();
        showSuccessToast('保存成功');
        showServiceEditPopup.value = false;
    } catch (error) {
        closeToast();
        showFailToast('保存失败');
    }
};

const onReadServiceFile = async (file, field) => {
    showLoadingToast({ message: '上传中...', forbidClick: true });
    const url = await uploadFile(file);
    closeToast();
    if (url && editingService.value) {
        editingService.value[field] = normalizeMediaUrl(url);
        showSuccessToast('上传成功');
    }
};

const toggleUserRole = async (user) => {
    if (!isSuperAdmin.value) {
        showFailToast('仅超级管理员可调整权限');
        return;
    }
    if (user.phone === SUPER_ADMIN_PHONE) {
        showFailToast('超级管理员权限不可修改');
        return;
    }
    const newRole = user.role === 'admin' ? 'user' : 'admin';
    try {
        await axios.post('/api/user/role', {
            id: user.id,
            role: newRole
        });
        
        user.role = newRole;
        showSuccessToast('权限更新成功');
    } catch (error) {
        showFailToast('权限更新失败');
        console.error(error);
    }
};

const handleSelectiveExport = () => {
    if (selectedIds.value.length === 0) {
        showToast('请先选择要导出的数据');
        return;
    }
    
    let url = `/api/export?ids=${selectedIds.value.join(',')}&role=${userRole.value}`;
    window.open(url, '_blank');
};

const handleExport = () => {
    let url = `/api/export?role=${userRole.value}&`;
    if (startDate.value) url += `startDate=${startDate.value.toISOString()}&`;
    if (endDate.value) url += `endDate=${endDate.value.toISOString()}&`;
    
    // 如果在赛事管理页导出，添加赛事特定筛选
    if (activeTab.value === 3) {
        url += `serviceId=13&`;
        if (selectedRole.value !== 'all') url += `competitionRole=${selectedRole.value}&`;
        if (selectedStatus.value !== 'all') url += `status=${selectedStatus.value}`;
    }
    
    window.open(url, '_blank');
};

const showDetail = (item) => {
    currentItem.value = { ...item }; // Clone to avoid direct mutation before save
    showDetailPopup.value = true;
};

const editCase = (caseItem) => {
    currentCase.value = JSON.parse(JSON.stringify(caseItem)); // Deep clone
    if (!currentCase.value.media) currentCase.value.media = [];
    if (!currentCase.value.highlights) currentCase.value.highlights = [];
    if (!currentCase.value.coverType) currentCase.value.coverType = 'image'; // Default
    showCaseEditPopup.value = true;
};

const createCase = () => {
    currentCase.value = {
        title: '',
        description: '',
        location: '',
        date: '',
        fullDescription: '',
        coverType: 'image',
        cover: '',
        media: [],
        highlights: [],
        categoryId: 1, // Default category
        service: '无人机物流' // Default service
    };
    showCaseEditPopup.value = true;
};

// Image Upload Helpers (Upload to Server)
const uploadFile = async (file) => {
    const formData = new FormData();
    formData.append('file', file.file); // van-uploader passes object with 'file' property
    
    try {
        const res = await axios.post('/api/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' }
        });
        if (res.data.success) {
            return normalizeMediaUrl(res.data.url);
        } else {
            showFailToast('上传失败');
            return null;
        }
    } catch (err) {
        console.error(err);
        showFailToast('上传出错');
        return null;
    }
};

const onReadCover = async (file) => {
    showLoadingToast({ message: '上传中...', forbidClick: true });
    const url = await uploadFile(file);
    closeToast();
    if (url) {
        currentCase.value.cover = normalizeMediaUrl(url);
        showSuccessToast('                                                         封面已上传');
    }
};

const onReadMedia = async (file, index) => {
    showLoadingToast({ message: '上传中...', forbidClick: true });
    const url = await uploadFile(file);
    closeToast();
    if (url) {
        currentCase.value.media[index].url = normalizeMediaUrl(url);
        showSuccessToast('资源已上传');
    }
};

const addMedia = () => {
    currentCase.value.media.push({ type: 'image', url: '' });
};

const removeMedia = (index) => {
    currentCase.value.media.splice(index, 1);
};

const addHighlight = () => {
    currentCase.value.highlights.push('');
};

const removeHighlight = (index) => {
    currentCase.value.highlights.splice(index, 1);
};

const onUpdateStatus = async ({ selectedOptions }) => {
    const newStatus = selectedOptions[0].value;
    if (!currentItem.value) return;

    try {
        await axios.post('/api/update', {
            id: currentItem.value.id,
            status: newStatus
        });
        
        currentItem.value.status = newStatus;
        // Update list locally
        const index = list.value.findIndex(item => item.id === currentItem.value.id);
        if (index !== -1) {
            list.value[index].status = newStatus;
        }
        
        showSuccessToast('状态更新成功');
        showStatusPicker.value = false;
    } catch (error) {
        showFailToast('更新状态失败');
        console.error(error);
    }
};

const onSaveCase = async () => {
    if (!currentCase.value) return;

    try {
        if (currentCase.value.id) {
            // Update
            await axios.post('/api/cases/update', currentCase.value);
            const index = cases.value.findIndex(c => c.id === currentCase.value.id);
            if (index !== -1) {
                cases.value[index] = currentCase.value;
            }
        } else {
            // Create
            const res = await axios.post('/api/cases/create', currentCase.value);
            currentCase.value.id = res.data.id;
            cases.value.unshift(currentCase.value);
        }
        
        showSuccessToast('保存成功');
        showCaseEditPopup.value = false;
    } catch (error) {
        showFailToast('保存失败');
        console.error(error);
    }
};

const onDeleteCase = () => {
    showConfirmDialog({
        title: '确认删除',
        message: '确定要删除这个案例吗？删除后无法恢复。',
    })
    .then(async () => {
        try {
            await axios.post('/api/cases/delete', { id: currentCase.value.id });
            const index = cases.value.findIndex(c => c.id === currentCase.value.id);
            if (index !== -1) {
                cases.value.splice(index, 1);
            }
            showSuccessToast('删除成功');
            showCaseEditPopup.value = false;
        } catch (error) {
            showFailToast('删除失败');
            console.error(error);
        }
    })
    .catch(() => {
        // cancel
    });
};

onMounted(async () => {
    await refreshCurrentUser();
    fetchData();
    fetchCases();
    fetchUsers();
    fetchAllServiceConfigs();
});
</script>

<style scoped>
.admin-page {
    background-color: #f7f8fa;
    min-height: 100vh;
    padding-bottom: 20px;
}

.filter-section {
    background: #fff;
    padding: 12px;
    margin-bottom: 12px;
}

.action-buttons {
    margin-top: 12px;
    padding: 0 16px;
}

.detail-content {
    padding: 16px 0;
}

.text-success {
    color: #07c160;
}

.text-warning {
    color: #ff976a;
}
.aspect-preview-container {
    width: 100%;
    padding: 10px 0;
}
.aspect-label {
    font-size: 12px;
    color: #86868b;
    margin-bottom: 8px;
}
.preview-box {
    width: 100%;
    background: #f5f5f7;
    border-radius: 12px;
    overflow: hidden;
    position: relative;
    border: 1px solid #eee;
}
.banner-box {
    aspect-ratio: 5 / 3;
}
.preview-box img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}
.preview-box.pos-top img { object-position: top; }
.preview-box.pos-center img { object-position: center; }
.preview-box.pos-bottom img { object-position: bottom; }

.safe-area-marker {
    position: absolute;
    bottom: 10px;
    left: 10px;
    background: rgba(0,113,227,0.6);
    color: #fff;
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 4px;
}
</style>
