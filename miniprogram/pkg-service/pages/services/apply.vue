<template>
  <view class="pub-page" :style="{ paddingTop: topPad + 'px' }">
    <!-- 顶栏（与发布页同款） -->
    <view class="pub-nav">
      <view class="pub-back" hover-class="pub-fade" @tap="goBack">‹</view>
      <view class="pub-nav-title">{{ serviceName }}</view>
    </view>

    <!-- 服务暂未开放提示 -->
    <view v-if="!isServiceAvailable" class="pub-empty">
      <view class="pub-empty-mark">!</view>
      <view class="pub-empty-title">该服务申请功能即将开放</view>
      <view class="pub-empty-desc">{{ serviceName }}功能正在建设中，敬请期待！</view>
      <view class="pub-empty-desc">如有需求，请联系客服：023-55558188</view>
      <view class="pub-btn pub-btn--ghost empty-back-btn" hover-class="pub-btn--active" @tap="goBack">返回服务列表</view>
    </view>

    <!-- 服务信息填写区 -->
    <template v-else>
      <!-- 活动摘要卡（研学专属：从 storage 读研学信息） -->
      <view v-if="serviceId === '9' && studyTour" class="study-summary">
        <view class="summary-bar"></view>
        <view class="summary-body">
          <text class="summary-title">{{ studyTour.title || '研学活动' }}</text>
          <text class="summary-meta">📅 {{ summaryDate }} · 📍 {{ summaryLoc }}</text>
        </view>
        <view class="summary-price">
          <text class="summary-price-num">¥{{ studyPrice }}</text>
          <text class="summary-price-unit">/人</text>
        </view>
      </view>

      <!-- 表单头部 -->
      <view class="pub-form-intro">
        <view class="pub-form-intro-h2">{{ serviceName }}</view>
        <view class="pub-form-intro-p">请填写以下信息，我们将尽快与您联系</view>
      </view>

      <!-- 基本信息 (除培训6和赛事13) -->
      <view v-if="serviceId !== '6' && serviceId !== '13'" class="pub-section">
        <view class="pub-section-title">基本信息</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <view class="pub-field-label">联系人</view>
            <input class="pub-input" v-model="formData.contactName" placeholder="请输入联系人姓名" placeholder-class="pub-placeholder" />
          </view>
          <view class="pub-field">
            <view class="pub-field-label">联系电话</view>
            <input class="pub-input" v-model="formData.contactPhone" type="number" placeholder="请输入联系电话" placeholder-class="pub-placeholder" />
          </view>
        </view>
      </view>

      <!-- 服务详情 -->
      <view class="pub-section">
        <view class="pub-section-title">{{ serviceId === '13' ? '赛事报名' : '服务详情' }}</view>
        <view class="pub-form-card">
          <!-- 无人机物流服务 (ID: 1) -->
          <template v-if="serviceId === '1'">
            <view class="pub-field">
              <view class="pub-field-label">客户类型</view>
              <radio-group class="pub-radio-group" @change="(e) => formData.customerType = e.detail.value">
                <label class="pub-radio-label"><radio value="personal" :checked="formData.customerType === 'personal'" color="#0A66C2" />个人</label>
                <label class="pub-radio-label"><radio value="enterprise" :checked="formData.customerType === 'enterprise'" color="#0A66C2" />企业</label>
              </radio-group>
            </view>
            <view class="pub-field" v-if="formData.customerType === 'enterprise'">
              <view class="pub-field-label">企业名称</view>
              <input class="pub-input" v-model="formData.companyName" placeholder="请输入企业名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">货物类型</view>
              <picker class="pub-select-field" :range="cargoTypeOptions" @change="(e) => formData.cargoType = cargoTypeOptions[e.detail.value]">
                <text :class="formData.cargoType ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.cargoType || '请选择货物类型' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field" v-if="formData.cargoType === '其他'">
              <view class="pub-field-label">具体类型</view>
              <input class="pub-input" v-model="formData.cargoTypeOther" placeholder="请输入具体类型" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">货物重量</view>
              <view class="pub-field-hint">kg</view>
              <input class="pub-input" v-model="formData.cargoWeight" type="digit" placeholder="请输入重量" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">货物体积</view>
              <input class="pub-input" v-model="formData.cargoVolume" placeholder="长×宽×高 (cm³)" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">起运地</view>
              <input class="pub-input address-input" v-model="formData.startAddress" placeholder="搜索或输入起运地" placeholder-class="pub-placeholder" />
              <textarea class="pub-input pub-input--textarea pub-input--textarea-sm" v-model="formData.startAddressDetail" placeholder="详细地址（楼栋、门牌号）" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">目的地</view>
              <input class="pub-input address-input" v-model="formData.endAddress" placeholder="搜索或输入目的地" placeholder-class="pub-placeholder" />
              <textarea class="pub-input pub-input--textarea pub-input--textarea-sm" v-model="formData.endAddressDetail" placeholder="详细地址（楼栋、门牌号）" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">运输时效</view>
              <picker class="pub-select-field" :range="urgencyOptions" @change="(e) => formData.deliveryUrgency = urgencyOptions[e.detail.value]">
                <text :class="formData.deliveryUrgency ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.deliveryUrgency || '请选择时效' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">期望时间</view>
              <picker class="pub-select-field" mode="date" @change="(e) => formData.expectedTime = e.detail.value">
                <text :class="formData.expectedTime ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.expectedTime || '请选择日期' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field upload-field">
              <view class="pub-field-label">货物照片</view>
              <view class="pub-field-hint">选填</view>
              <view class="pub-upload-row">
                <view v-for="(img, idx) in formData.fileList" :key="idx" class="pub-photo" @tap="previewImage(img)">
                  <image :src="img" mode="aspectFill" class="pub-photo-img" />
                  <view class="pub-photo-remove" @tap.stop="delImage(idx)">×</view>
                </view>
                <view class="pub-add-photo" hover-class="pub-fade" v-if="formData.fileList.length < 5" @tap="chooseImage">＋</view>
              </view>
              <view class="pub-upload-tip upload-tip-inline">最多 5 张，点击可预览</view>
            </view>
          </template>

          <!-- 政务服务 (ID: 2) -->
          <template v-if="serviceId === '2'">
            <view class="pub-field">
              <view class="pub-field-label">巡检类型</view>
              <input class="pub-input" v-model="formData.inspectionType" placeholder="如：环保监测、安全巡查" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">巡检区域</view>
              <input class="pub-input" v-model="formData.inspectionArea" placeholder="请输入巡检具体路段或区域" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">巡检时间</view>
              <input class="pub-input" v-model="formData.inspectionDate" placeholder="如：2025-01-25 或 每周一" placeholder-class="pub-placeholder" />
            </view>
          </template>

          <!-- 无人机托管 (ID: 3) -->
          <template v-if="serviceId === '3'">
            <view class="pub-field">
              <view class="pub-field-label">机型</view>
              <input class="pub-input" v-model="formData.droneModel" placeholder="请输入机型" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">托管数量</view>
              <input class="pub-input" v-model="formData.droneCount" type="number" placeholder="请输入数量" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">托管期限</view>
              <picker class="pub-select-field" :range="durationOptions" @change="(e) => formData.trusteeDuration = durationOptions[e.detail.value]">
                <text :class="formData.trusteeDuration ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.trusteeDuration || '请选择期限' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
          </template>

          <!-- 无人机吊运 (ID: 4) -->
          <template v-if="serviceId === '4'">
            <view class="pub-field">
              <view class="pub-field-label">吊运物品</view>
              <input class="pub-input" v-model="formData.liftItemType" placeholder="请输入物品名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">物品重量</view>
              <view class="pub-field-hint">kg</view>
              <input class="pub-input" v-model="formData.liftItemWeight" type="digit" placeholder="请输入重量" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">吊运高度</view>
              <view class="pub-field-hint">m</view>
              <input class="pub-input" v-model="formData.liftHeight" type="number" placeholder="请输入高度" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">作业地点</view>
              <input class="pub-input" v-model="formData.workLocation" placeholder="请输入作业具体地点" placeholder-class="pub-placeholder" />
            </view>
          </template>

          <!-- 飞手培训 (ID: 6) -->
          <template v-if="serviceId === '6'">
            <view class="pub-field">
              <view class="pub-field-label">姓名</view>
              <input class="pub-input" v-model="formData.traineeName" placeholder="请输入学员姓名" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">手机号</view>
              <input class="pub-input" v-model="formData.traineePhone" type="number" placeholder="请输入学员手机号" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">性别</view>
              <radio-group class="pub-radio-group" @change="(e) => formData.traineeGender = e.detail.value">
                <label class="pub-radio-label"><radio value="male" :checked="formData.traineeGender === 'male'" color="#0A66C2" />男</label>
                <label class="pub-radio-label"><radio value="female" :checked="formData.traineeGender === 'female'" color="#0A66C2" />女</label>
              </radio-group>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">证件号码</view>
              <input class="pub-input" v-model="formData.traineeIdCard" placeholder="请输入身份证号" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">考试机型</view>
              <picker class="pub-select-field" :range="examModelOptions" @change="(e) => formData.examModel = examModelOptions[e.detail.value]">
                <text :class="formData.examModel ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.examModel || '请选择机型' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">证照级别</view>
              <picker class="pub-select-field" :range="licenseLevelOptions" @change="(e) => formData.licenseLevel = licenseLevelOptions[e.detail.value]">
                <text :class="formData.licenseLevel ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.licenseLevel || '请选择级别' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
          </template>

          <!-- 低空研学报名 (ID: 9) -->
          <template v-if="serviceId === '9'">
            <view class="pub-field">
              <view class="pub-field-label">学校/机构<text class="pub-required">*</text></view>
              <input class="pub-input" v-model="formData.studyOrg" placeholder="请输入学校/机构名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">年级段<text class="pub-required">*</text></view>
              <picker class="pub-select-field" :range="studyGradeOptions" @change="(e) => formData.studyGrade = studyGradeOptions[e.detail.value]">
                <text :class="formData.studyGrade ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.studyGrade || '请选择年级段' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">参与人数<text class="pub-required">*</text></view>
              <view class="stepper">
                <view class="stepper-btn" :class="{ disabled: participants <= 1 }" @tap="changeParticipants(-1)">−</view>
                <input class="stepper-input" v-model="formData.studyParticipants" type="number" @input="onParticipantsInput" />
                <view class="stepper-btn" @tap="changeParticipants(1)">+</view>
              </view>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">期望日期<text class="pub-required">*</text></view>
              <picker class="pub-select-field" mode="date" @change="(e) => formData.studyDate = e.detail.value">
                <text :class="formData.studyDate ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.studyDate || '请选择日期' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
          </template>

          <!-- 无人机维修服务 (ID: 12) -->
          <template v-if="serviceId === '12'">
            <view class="pub-field">
              <view class="pub-field-label">服务类型</view>
              <radio-group class="pub-radio-group" @change="(e) => formData.maintenanceType = e.detail.value">
                <label class="pub-radio-label"><radio value="repair" :checked="formData.maintenanceType === 'repair'" color="#0A66C2" />故障维修</label>
                <label class="pub-radio-label"><radio value="care" :checked="formData.maintenanceType === 'care'" color="#0A66C2" />定期保养</label>
              </radio-group>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">机型</view>
              <input class="pub-input" v-model="formData.droneModel" placeholder="请输入无人机型号" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">是否在保</view>
              <radio-group class="pub-radio-group" @change="(e) => formData.isWarranty = e.detail.value">
                <label class="pub-radio-label"><radio value="yes" :checked="formData.isWarranty === 'yes'" color="#0A66C2" />在保修期</label>
                <label class="pub-radio-label"><radio value="no" :checked="formData.isWarranty === 'no'" color="#0A66C2" />已过保</label>
              </radio-group>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">购买日期</view>
              <picker class="pub-select-field" mode="date" @change="(e) => formData.purchaseDate = e.detail.value">
                <text :class="formData.purchaseDate ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.purchaseDate || '选填' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">故障/需求描述</view>
              <textarea class="pub-input pub-input--textarea" v-model="formData.remark" placeholder="请详细描述设备故障情况或保养需求" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field upload-field">
              <view class="pub-field-label">设备照片</view>
              <view class="pub-field-hint">选填</view>
              <view class="pub-upload-row">
                <view v-for="(img, idx) in formData.fileList" :key="idx" class="pub-photo" @tap="previewImage(img)">
                  <image :src="img" mode="aspectFill" class="pub-photo-img" />
                  <view class="pub-photo-remove" @tap.stop="delImage(idx)">×</view>
                </view>
                <view class="pub-add-photo" hover-class="pub-fade" v-if="formData.fileList.length < 5" @tap="chooseImage">＋</view>
              </view>
              <view class="pub-upload-tip upload-tip-inline">最多 5 张，点击可预览</view>
            </view>
          </template>

          <!-- 无人机赛事报名 (ID: 13) -->
          <template v-if="serviceId === '13'">
            <view class="pub-field">
              <view class="pub-field-label">注册类型</view>
              <picker class="pub-select-field" :range="competitionRoleLabels" @change="onCompetitionRoleChange">
                <text :class="formData.competitionRoleText ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.competitionRoleText || '请选择注册角色' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
          </template>
        </view>
      </view>

      <!-- 赛事角色表单 -->
      <view v-if="serviceId === '13' && formData.competitionRole" class="pub-section">
        <view class="pub-section-title">表单报名</view>
        <view class="pub-form-card">
          <!-- 裁判员 -->
          <template v-if="formData.competitionRole === 'referee'">
            <view class="pub-field">
              <view class="pub-field-label">单位名称</view>
              <input class="pub-input" v-model="formData.companyName" placeholder="请输入单位名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">姓名</view>
              <input class="pub-input" v-model="formData.name" placeholder="请输入姓名" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">性别</view>
              <radio-group class="pub-radio-group" @change="(e) => formData.gender = e.detail.value">
                <label class="pub-radio-label"><radio value="male" :checked="formData.gender === 'male'" color="#0A66C2" />男</label>
                <label class="pub-radio-label"><radio value="female" :checked="formData.gender === 'female'" color="#0A66C2" />女</label>
              </radio-group>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">证件号</view>
              <input class="pub-input" v-model="formData.idCard" placeholder="请输入证件号" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">组别</view>
              <picker class="pub-select-field" :range="competitionGroupOptions" @change="(e) => formData.competitionGroup = competitionGroupOptions[e.detail.value]">
                <text :class="formData.competitionGroup ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.competitionGroup || '请选择组别' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">联系电话</view>
              <input class="pub-input" v-model="formData.phone" type="number" placeholder="请输入联系电话" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">电子邮箱</view>
              <input class="pub-input" v-model="formData.email" placeholder="请输入邮箱（选填）" placeholder-class="pub-placeholder" />
            </view>
          </template>

          <!-- 教练员 -->
          <template v-if="formData.competitionRole === 'coach'">
            <view class="pub-field">
              <view class="pub-field-label">单位名称</view>
              <input class="pub-input" v-model="formData.companyName" placeholder="请输入单位名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">姓名</view>
              <input class="pub-input" v-model="formData.name" placeholder="请输入姓名" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">性别</view>
              <radio-group class="pub-radio-group" @change="(e) => formData.gender = e.detail.value">
                <label class="pub-radio-label"><radio value="male" :checked="formData.gender === 'male'" color="#0A66C2" />男</label>
                <label class="pub-radio-label"><radio value="female" :checked="formData.gender === 'female'" color="#0A66C2" />女</label>
              </radio-group>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">证件号</view>
              <input class="pub-input" v-model="formData.idCard" placeholder="请输入证件号" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">组别</view>
              <picker class="pub-select-field" :range="competitionGroupOptions" @change="(e) => formData.competitionGroup = competitionGroupOptions[e.detail.value]">
                <text :class="formData.competitionGroup ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.competitionGroup || '请选择组别' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">联系电话</view>
              <input class="pub-input" v-model="formData.phone" type="number" placeholder="请输入联系电话" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">电子邮箱</view>
              <input class="pub-input" v-model="formData.email" placeholder="请输入邮箱（选填）" placeholder-class="pub-placeholder" />
            </view>
          </template>

          <!-- 运动员 -->
          <template v-if="formData.competitionRole === 'athlete'">
            <view class="pub-field">
              <view class="pub-field-label">单位/学校</view>
              <input class="pub-input" v-model="formData.companyName" placeholder="请输入单位/学校名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">姓名</view>
              <input class="pub-input" v-model="formData.name" placeholder="请输入姓名" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">性别</view>
              <radio-group class="pub-radio-group" @change="(e) => formData.gender = e.detail.value">
                <label class="pub-radio-label"><radio value="male" :checked="formData.gender === 'male'" color="#0A66C2" />男</label>
                <label class="pub-radio-label"><radio value="female" :checked="formData.gender === 'female'" color="#0A66C2" />女</label>
              </radio-group>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">证件号</view>
              <input class="pub-input" v-model="formData.idCard" placeholder="请输入证件号" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">组别</view>
              <picker class="pub-select-field" :range="athleteGroupOptions" @change="(e) => formData.athleteGroup = athleteGroupOptions[e.detail.value]">
                <text :class="formData.athleteGroup ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.athleteGroup || '请选择组别' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">参赛项目</view>
              <picker class="pub-select-field" :range="competitionProjectOptions" @change="(e) => formData.competitionProject = competitionProjectOptions[e.detail.value]">
                <text :class="formData.competitionProject ? 'pub-select-value' : 'pub-placeholder'">
                  {{ formData.competitionProject || '请选择项目' }}
                </text>
                <text class="pub-arrow">›</text>
              </picker>
            </view>
            <view class="pub-field">
              <view class="pub-field-label">联系电话</view>
              <input class="pub-input" v-model="formData.phone" type="number" placeholder="请输入联系电话" placeholder-class="pub-placeholder" />
            </view>
          </template>

          <!-- 俱乐部 -->
          <template v-if="formData.competitionRole === 'club'">
            <view class="pub-field">
              <view class="pub-field-label">单位名称</view>
              <input class="pub-input" v-model="formData.companyName" placeholder="请输入单位名称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">单位简称</view>
              <input class="pub-input" v-model="formData.companyShortName" placeholder="请输入单位简称" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">所在地</view>
              <input class="pub-input" v-model="formData.location" placeholder="请输入单位所在地" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">负责人</view>
              <input class="pub-input" v-model="formData.manager" placeholder="请输入负责人姓名" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">负责人电话</view>
              <input class="pub-input" v-model="formData.managerPhone" type="number" placeholder="请输入负责人电话" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">对接人</view>
              <input class="pub-input" v-model="formData.contactPerson" placeholder="请输入对接人姓名" placeholder-class="pub-placeholder" />
            </view>
            <view class="pub-field">
              <view class="pub-field-label">对接人电话</view>
              <input class="pub-input" v-model="formData.contactPhone" type="number" placeholder="请输入对接人电话" placeholder-class="pub-placeholder" />
            </view>
          </template>
        </view>
      </view>

      <!-- 通用备注 (非物流、非政务、非维修、非研学) -->
      <view v-if="!['1', '2', '12', '9'].includes(serviceId)" class="pub-section">
        <view class="pub-section-title">备注/其他需求</view>
        <view class="pub-form-card">
          <view class="pub-field">
            <textarea class="pub-input pub-input--textarea" v-model="formData.remark" placeholder="请详细说明您的需求，以便我们为您提供更精准的服务。" placeholder-class="pub-placeholder" />
          </view>
        </view>
      </view>

      <!-- 研学隐私说明 -->
      <view v-if="serviceId === '9'" class="privacy-note">报名信息仅用于活动组织，受隐私政策保护</view>
    </template>

    <!-- 固定底部操作区：研学带费用预估，其他服务单一提交按钮 -->
    <view v-if="isServiceAvailable" class="pub-sticky">
      <view v-if="serviceId === '9'" class="sticky-price">
        <text class="sticky-price-label">费用预估</text>
        <view class="sticky-price-row">
          <text class="sticky-price-symbol">¥</text>
          <text class="sticky-price-num">{{ displayFee }}</text>
        </view>
        <text class="sticky-price-note">最终费用以机构确认为准</text>
      </view>
      <view
        class="pub-btn pub-btn--primary"
        :class="{ 'cta-glow': serviceId === '9' }"
        hover-class="pub-btn--active"
        @tap="handleSubmit"
      >{{ submitButtonText }}</view>
    </view>

    <HomeFloatButton />
  </view>
</template>

<script setup>
import { safeBack } from '../../../utils/nav'
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import HomeFloatButton from '@/components/HomeFloatButton.vue'
import { getStoredUser, request, authStorage } from '../../../utils/request'
import { useSafeTop } from '../../../utils/safeTop'

const { topPad, initSafeTop } = useSafeTop(true)

const serviceId = ref('')
const submitting = ref(false)
const isServiceAvailable = computed(() => ['1', '2', '3', '4', '6', '9', '12', '13'].includes(serviceId.value))

const serviceNames = {
  '1': '无人机物流', '2': '政务服务', '3': '无人机托管', '4': '无人机吊运',
  '5': '无人机表演', '6': '飞手培训', '7': '无人机租赁', '8': '无人机外卖',
  '9': '低空研学', '10': '无人机销售', '11': '无人机金融服务',
  '12': '无人机维修服务', '13': '无人机赛事'
}

const serviceName = computed(() => serviceNames[serviceId.value] || '服务')

const submitButtonText = computed(() => {
  if (['1', '4', '8'].includes(serviceId.value)) return '立即下单'
  if (['6', '9', '13'].includes(serviceId.value)) return '参与报名'
  return '提交申请'
})

const formData = ref({
  contactName: '', contactPhone: '',
  customerType: 'personal', companyName: '', cargoType: '', cargoTypeOther: '',
  cargoWeight: '', cargoVolume: '', startAddress: '', startAddressDetail: '',
  endAddress: '', endAddressDetail: '', deliveryUrgency: '', expectedTime: '', fileList: [],
  inspectionType: '', inspectionArea: '', inspectionDate: '',
  droneModel: '', droneCount: '', trusteeDuration: '',
  liftItemType: '', liftItemWeight: '', workLocation: '', liftHeight: '',
  traineeName: '', traineePhone: '', traineeGender: 'male', traineeIdCard: '',
  examModel: '', licenseLevel: '',
  studyOrg: '', studyGrade: '', studyParticipants: '', studyDate: '',
  maintenanceType: 'repair', isWarranty: 'yes', purchaseDate: '',
  competitionRole: '', competitionRoleText: '', competitionGroup: '',
  competitionProject: '', athleteGroup: '',
  name: '', gender: 'male', idCard: '', phone: '', email: '',
  companyShortName: '', location: '', manager: '', managerPhone: '',
  contactPerson: '',
  remark: ''
})

const cargoTypeOptions = ['生鲜食品', '应急药品', '工业零部件', '电子产品', '文件资料', '日用品', '医疗器械', '其他']
const urgencyOptions = ['加急配送（2小时内）', '标准配送（当日达）', '普通配送（次日达）', '经济配送（3日内）']
const durationOptions = ['1个月', '3个月', '6个月', '12个月', '长期托管']
const examModelOptions = ['小型无人机 (多旋翼)', '中型无人机 (多旋翼)', '垂起固定翼']
const licenseLevelOptions = ['视距内', '超视距']

// ── 研学报名新增 ──
const studyGradeOptions = ['小学低年级（1-3年级）', '小学高年级（4-6年级）', '初中', '高中']
const studyTour = ref(null)
const studyPrice = ref('1280')
const studyDate = ref('')
const studyLoc = ref('')

const summaryDate = computed(() => studyDate.value || '时间待定')
const summaryLoc = computed(() => studyLoc.value || '地点待定')

// 研学价格（按时长兜底，与详情页一致）
const calcStudyPrice = (tour) => {
  const d = (tour && tour.duration) || ''
  if (d.includes('3天')) return '1280'
  if (d.includes('2天')) return '880'
  if (d.includes('1天')) return '480'
  return '980'
}

// 参与人数（步进器）
const participants = computed(() => parseInt(formData.value.studyParticipants) || 1)
const changeParticipants = (delta) => {
  const next = Math.max(1, Math.min(100, participants.value + delta))
  formData.value.studyParticipants = String(next)
  animateFee()
}
const onParticipantsInput = (e) => {
  const v = parseInt(e.detail.value)
  if (!isNaN(v)) {
    if (v < 1) formData.value.studyParticipants = '1'
    else if (v > 100) formData.value.studyParticipants = '100'
    animateFee()
  }
}

// 费用预估 = 人数 × 单价（滚动计数）
const feeEstimate = computed(() => {
  const p = participants.value
  const price = parseInt(studyPrice.value) || 0
  return (p * price).toLocaleString()
})
const displayFee = ref('0')
const animateFee = (duration = 400) => {
  const target = parseInt(String(feeEstimate.value).replace(/,/g, '')) || 0
  const from = parseInt(String(displayFee.value).replace(/,/g, '')) || 0
  const start = Date.now()
  // 微信小程序无 requestAnimationFrame，用 setTimeout 模拟帧（16ms）
  const tick = () => {
    const p = Math.min(1, (Date.now() - start) / duration)
    const ease = 1 - Math.pow(1 - p, 3)
    displayFee.value = Math.round(from + (target - from) * ease).toLocaleString()
    if (p < 1) setTimeout(tick, 16)
  }
  tick()
}
const competitionGroupOptions = ['足球', 'FPV', '航拍']
const athleteGroupOptions = ['成人组', '中学组', '小学组']
const competitionProjectOptions = ['足球', 'FPV', '航拍']

const competitionRoleMap = [
  { text: '运动员', value: 'athlete' },
  { text: '教练员', value: 'coach' },
  { text: '裁判员', value: 'referee' },
  { text: '俱乐部', value: 'club' }
]
const competitionRoleLabels = competitionRoleMap.map(r => r.text)

const onCompetitionRoleChange = (e) => {
  const idx = e.detail.value
  formData.value.competitionRole = competitionRoleMap[idx].value
  formData.value.competitionRoleText = competitionRoleMap[idx].text
}

onLoad((options) => {
  initSafeTop()
  serviceId.value = options.id || '1'
  const user = getStoredUser()
  if (user) {
    formData.value.contactName = user.name || ''
    formData.value.contactPhone = user.phone || ''
    formData.value.traineeName = user.name || ''
    formData.value.traineePhone = user.phone || ''
  }
  // 研学报名：优先取列表/详情页经 storage 传入的活动（摘要卡）；读取后立即清理
  // 防陈旧残留。storage 为空时支持 tourId 参数（详情页/分享直达报名页）走接口自取。
  if (serviceId.value === '9') {
    const tour = uni.getStorageSync('study_tour_detail')
    uni.removeStorageSync('study_tour_detail')
    if (tour && tour.id) {
      fillStudyTour(tour)
    } else if (options && options.tourId) {
      request({ url: '/api/v1/study/tours/' + encodeURIComponent(options.tourId) })
        .then((res) => {
          const d = (res && res.data) || res
          if (d && d.id) fillStudyTour(d)
        })
        .catch(() => { /* 活动不可用时不填摘要，提交由后端校验 */ })
    }
  }
})

const chooseImage = () => {
  uni.chooseImage({
    count: 5 - formData.value.fileList.length,
    success: (res) => {
      formData.value.fileList = [...formData.value.fileList, ...res.tempFilePaths]
    }
  })
}

// fillStudyTour 填充研学活动摘要卡（费用/日期/地点 + 首帧滚动动画）
const fillStudyTour = (tour) => {
  studyTour.value = tour
  studyPrice.value = calcStudyPrice(tour)
  // 日期范围（零值降级）
  const fmt = (v) => {
    if (!v) return ''
    const d = new Date(v)
    if (isNaN(d.getTime()) || d.getFullYear() <= 1) return ''
    const p = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}年${p(d.getMonth() + 1)}月${p(d.getDate())}日`
  }
  const s = fmt(tour.start_date)
  const e = fmt(tour.end_date)
  studyDate.value = s && e ? `${s}-${e.replace(/^\d{4}年/, '')}` : (s || '')
  studyLoc.value = tour.location || tour.destination || ''
  // 费用预估初始滚动（延迟到视图渲染后）
  setTimeout(() => animateFee(600), 200)
}

const previewImage = (current) => {
  uni.previewImage({ current, urls: formData.value.fileList })
}

const delImage = (index) => {
  formData.value.fileList.splice(index, 1)
}

const goBack = () => {
  safeBack()
}

const handleSubmit = async () => {
  if (submitting.value) return
  const user = getStoredUser()
  if (!user) {
    uni.showModal({
      title: '提示',
      content: '您当前未登录，请先登录后再提交申请',
      confirmText: '去登录',
      success: (res) => {
        if (res.confirm) uni.navigateTo({ url: '/pages/login/index' })
      }
    })
    return
  }

  if (serviceId.value !== '6' && serviceId.value !== '13') {
    if (!formData.value.contactName || !formData.value.contactPhone) {
      return uni.showToast({ title: '请填写联系人和电话', icon: 'none' })
    }
  } else if (serviceId.value === '6') {
    if (!formData.value.traineeName || !formData.value.traineePhone) {
      return uni.showToast({ title: '请填写学员姓名和电话', icon: 'none' })
    }
  } else if (serviceId.value === '13') {
    if (!formData.value.competitionRole) {
      return uni.showToast({ title: '请选择注册角色', icon: 'none' })
    }
  }

  submitting.value = true
  uni.showLoading({ title: '提交中...', mask: true })

  const now = new Date()
  const orderNo = 'DK' + now.getTime()
  const applyTime = now.toLocaleString()

  const submitData = {
    ...formData.value,
    serviceId: serviceId.value,
    serviceName: serviceName.value,
    orderNo,
    applyTime,
    status: '处理中',
    userId: user.id
  }

  try {
    await request({
      url: '/api/submit',
      method: 'POST',
      data: submitData
    })

    uni.hideLoading()
    uni.showModal({
      title: '✅ 提交成功',
      content: `申请单号：${orderNo}\n我们将尽快与您联系并确认服务方案！`,
      showCancel: false,
      confirmText: '查看我的申请',
      success: () => {
        uni.navigateTo({ url: '/pkg-app/pages/applications/index' })
      }
    })
  } catch (error) {
    // 展示后端真实错误原因（500/400/401 各有其因），仅 404（接口未上线）才归因"暂未开放"
    uni.hideLoading()
    const status = (error && error.statusCode) || 0
    const backendMsg = (error && error.data && error.data.error && error.data.error.message) || ''
    let content = backendMsg || '提交失败，请稍后重试。'
    if (status === 404) {
      content = '线上申请暂未开放，请直接电话联系客服完成服务申请。'
    }
    uni.showModal({
      title: '提交失败',
      content: content,
      showCancel: false,
      confirmText: '知道了'
    })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
@import '../../../pages/publish/pub-style.css';

.pub-fade { opacity: 0.6; }
.pub-form-intro-h2 {
  font-size: 20px;
  margin: 0 0 4px;
  color: #17212B;
}
.pub-form-intro-p {
  font-size: 12px;
  color: #667085;
  margin: 0;
  line-height: 1.5;
}
.pub-photo-img {
  width: 100%;
  height: 100%;
  display: block;
}
.pub-photo-remove {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 17px;
  height: 17px;
  border-radius: 50%;
  background: rgba(23, 33, 43, 0.55);
  color: #fff;
  font-size: 12px;
  line-height: 17px;
  text-align: center;
}

/* 原生 radio 组（对齐 pub 色板） */
.pub-radio-group { display: flex; gap: 18px; padding-top: 2px; }
.pub-radio-label { display: flex; align-items: center; gap: 4px; font-size: 14px; color: #17212B; }

/* 小号多行输入（起运地/目的地详细地址） */
.pub-input--textarea-sm { min-height: 58px; }
.address-input { margin-bottom: 8px; }

/* 上传区贴齐字段内边距 */
.upload-field .pub-upload-row { padding: 0; }
.upload-tip-inline { margin: 8px 0 0; }

/* 空状态返回按钮 */
.empty-back-btn { flex: none; margin: 16px auto 0; padding: 0 22px; }

/* ── 研学报名：活动摘要卡 ── */
.study-summary {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 13px;
  padding: 14px 13px;
  background: #fff;
  border: 1px solid #EEF1F4;
  border-radius: 10px;
  overflow: hidden;
  animation: summaryIn 0.45s cubic-bezier(0.16, 1, 0.3, 1) both;
}
.summary-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #0A66C2;
}
.summary-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 5px; }
.summary-title { font-size: 15px; font-weight: 750; color: #17212B; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.summary-meta { font-size: 12px; color: #667085; }
.summary-price { flex-shrink: 0; text-align: right; }
.summary-price-num { font-size: 20px; font-weight: 800; color: #F97316; }
.summary-price-unit { font-size: 11px; color: #98A2B3; margin-left: 2px; }

/* ── 研学报名：参与人数步进器 ── */
.stepper { display: flex; align-items: center; gap: 10px; padding-top: 2px; }
.stepper-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #F5F6F8;
  border: 1px solid #EEF1F4;
  color: #0A66C2;
  font-size: 20px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stepper-btn.disabled { color: #98A2B3; background: #F5F6F8; }
.stepper-input {
  width: 64px;
  height: 40px;
  text-align: center;
  font-size: 15px;
  font-weight: 700;
  color: #17212B;
  background: #F5F6F8;
  border: 1px solid #EEF1F4;
  border-radius: 8px;
}

/* ── 研学报名：底栏费用预估 ── */
.sticky-price { flex: 1; min-width: 0; }
.sticky-price-label { font-size: 11px; color: #667085; display: block; }
.sticky-price-row { display: flex; align-items: baseline; gap: 2px; }
.sticky-price-symbol { font-size: 15px; font-weight: 750; color: #F97316; }
.sticky-price-num { font-size: 20px; font-weight: 800; color: #F97316; }
.sticky-price-note { font-size: 10px; color: #98A2B3; display: block; }

/* 研学：隐私说明（底栏上方） */
.privacy-note { padding: 4px 4px 0; text-align: center; font-size: 11px; color: #98A2B3; }

/* 研学：提交按钮呼吸光（对齐 pub 主色） */
.cta-glow { animation: applyGlow 2.5s ease-in-out infinite; }
@keyframes applyGlow {
  0%, 100% { box-shadow: 0 7px 14px rgba(10, 102, 194, 0.20); }
  50% { box-shadow: 0 8px 24px rgba(10, 102, 194, 0.42); }
}

/* 活动摘要卡入场 */
@keyframes summaryIn {
  from {
    transform: translateX(20rpx);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>
