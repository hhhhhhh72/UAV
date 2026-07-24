<template>
  <view class="apply-page">
    <view class="content-wrapper">
      <!-- 服务暂未开放提示 -->
      <view v-if="!isServiceAvailable" class="empty-section">
        <view class="empty-icon">⏳</view>
        <view class="empty-text">该服务申请功能即将开放</view>
        <view class="empty-desc">{{ serviceName }}功能正在建设中，敬请期待！</view>
        <view class="empty-contact">如有需求，请联系客服：0577-55558188</view>
        <button class="back-btn" @tap="goBack">返回服务列表</button>
      </view>

      <!-- 服务信息填写区 -->
      <view v-else>
        <view class="notice-bar">
          <text class="notice-icon">ℹ️</text>
          <text class="notice-text">请填写以下信息，我们将尽快与您联系</text>
        </view>

        <view class="form-body">
          <!-- 基本信息 (除培训6和赛事13) -->
          <view class="form-section" v-if="serviceId !== '6' && serviceId !== '13'">
            <view class="section-title">基本信息</view>
            <view class="form-item">
              <text class="label">联系人</text>
              <input class="input" v-model="formData.contactName" placeholder="请输入联系人姓名" />
            </view>
            <view class="form-item">
              <text class="label">联系电话</text>
              <input class="input" v-model="formData.contactPhone" type="number" placeholder="请输入联系电话" />
            </view>
          </view>

          <!-- 服务详情 -->
          <view class="form-section">
            <view class="section-title">{{ serviceId === '13' ? '赛事报名' : '服务详情' }}</view>
            
            <!-- 无人机物流服务 (ID: 1) -->
            <template v-if="serviceId === '1'">
              <view class="form-item">
                <text class="label">客户类型</text>
                <radio-group class="radio-group" @change="(e) => formData.customerType = e.detail.value">
                  <label class="radio-label"><radio value="personal" :checked="formData.customerType === 'personal'" color="#667eea" />个人</label>
                  <label class="radio-label"><radio value="enterprise" :checked="formData.customerType === 'enterprise'" color="#667eea" />企业</label>
                </radio-group>
              </view>
              <view class="form-item" v-if="formData.customerType === 'enterprise'">
                <text class="label">企业名称</text>
                <input class="input" v-model="formData.companyName" placeholder="请输入企业名称" />
              </view>
              <view class="form-item">
                <text class="label">货物类型</text>
                <picker class="picker" :range="cargoTypeOptions" @change="(e) => formData.cargoType = cargoTypeOptions[e.detail.value]">
                  <view class="picker-value" :class="{ 'placeholder': !formData.cargoType }">
                    {{ formData.cargoType || '请选择货物类型' }}
                  </view>
                </picker>
              </view>
              <view class="form-item" v-if="formData.cargoType === '其他'">
                <text class="label">具体类型</text>
                <input class="input" v-model="formData.cargoTypeOther" placeholder="请输入具体类型" />
              </view>
              <view class="form-item">
                <text class="label">货物重量</text>
                <input class="input" v-model="formData.cargoWeight" type="digit" placeholder="请输入重量" />
                <text class="unit">kg</text>
              </view>
              <view class="form-item">
                <text class="label">货物体积</text>
                <input class="input" v-model="formData.cargoVolume" placeholder="长×宽×高 (cm³)" />
              </view>
              <view class="form-item vertical">
                <text class="label">起运地</text>
                <input class="input" v-model="formData.startAddress" placeholder="搜索或输入起运地" />
                <textarea class="textarea-small" v-model="formData.startAddressDetail" placeholder="详细地址（楼栋、门牌号）" />
              </view>
              <view class="form-item vertical">
                <text class="label">目的地</text>
                <input class="input" v-model="formData.endAddress" placeholder="搜索或输入目的地" />
                <textarea class="textarea-small" v-model="formData.endAddressDetail" placeholder="详细地址（楼栋、门牌号）" />
              </view>
              <view class="form-item">
                <text class="label">运输时效</text>
                <picker class="picker" :range="urgencyOptions" @change="(e) => formData.deliveryUrgency = urgencyOptions[e.detail.value]">
                  <view class="picker-value" :class="{ 'placeholder': !formData.deliveryUrgency }">
                    {{ formData.deliveryUrgency || '请选择时效' }}
                  </view>
                </picker>
              </view>
              <view class="form-item">
                <text class="label">期望时间</text>
                <picker class="picker" mode="date" @change="(e) => formData.expectedTime = e.detail.value">
                  <view class="picker-value" :class="{ 'placeholder': !formData.expectedTime }">
                    {{ formData.expectedTime || '请选择日期' }}
                  </view>
                </picker>
              </view>
              <view class="form-item vertical no-border">
                <text class="label">货物照片 (选填)</text>
                <view class="uploader">
                  <view class="upload-item" v-for="(img, idx) in formData.fileList" :key="idx">
                    <image :src="img" mode="aspectFill" @tap="previewImage(img)" />
                    <text class="del-btn" @tap="delImage(idx)">×</text>
                  </view>
                  <view class="upload-btn" @tap="chooseImage" v-if="formData.fileList.length < 5">
                    <text class="plus">+</text>
                  </view>
                </view>
              </view>
            </template>

            <!-- 政务服务 (ID: 2) -->
            <template v-if="serviceId === '2'">
              <view class="form-item">
                <text class="label">巡检类型</text>
                <input class="input" v-model="formData.inspectionType" placeholder="如：环保监测、安全巡查" />
              </view>
              <view class="form-item">
                <text class="label">巡检区域</text>
                <input class="input" v-model="formData.inspectionArea" placeholder="请输入巡检具体路段或区域" />
              </view>
              <view class="form-item">
                <text class="label">巡检时间</text>
                <input class="input" v-model="formData.inspectionDate" placeholder="如：2025-01-25 或 每周一" />
              </view>
            </template>

            <!-- 无人机托管 (ID: 3) -->
            <template v-if="serviceId === '3'">
              <view class="form-item">
                <text class="label">机型</text>
                <input class="input" v-model="formData.droneModel" placeholder="请输入机型" />
              </view>
              <view class="form-item">
                <text class="label">托管数量</text>
                <input class="input" v-model="formData.droneCount" type="number" placeholder="请输入数量" />
              </view>
              <view class="form-item">
                <text class="label">托管期限</text>
                <picker class="picker" :range="durationOptions" @change="(e) => formData.trusteeDuration = durationOptions[e.detail.value]">
                  <view class="picker-value" :class="{ 'placeholder': !formData.trusteeDuration }">
                    {{ formData.trusteeDuration || '请选择期限' }}
                  </view>
                </picker>
              </view>
            </template>

            <!-- 无人机吊运 (ID: 4) -->
            <template v-if="serviceId === '4'">
              <view class="form-item">
                <text class="label">吊运物品</text>
                <input class="input" v-model="formData.liftItemType" placeholder="请输入物品名称" />
              </view>
              <view class="form-item">
                <text class="label">物品重量</text>
                <input class="input" v-model="formData.liftItemWeight" type="digit" placeholder="请输入重量" />
                <text class="unit">kg</text>
              </view>
              <view class="form-item">
                <text class="label">吊运高度</text>
                <input class="input" v-model="formData.liftHeight" type="number" placeholder="请输入高度" />
                <text class="unit">m</text>
              </view>
              <view class="form-item">
                <text class="label">作业地点</text>
                <input class="input" v-model="formData.workLocation" placeholder="请输入作业具体地点" />
              </view>
            </template>

            <!-- 飞手培训 (ID: 6) -->
            <template v-if="serviceId === '6'">
              <view class="form-item">
                <text class="label">姓名</text>
                <input class="input" v-model="formData.traineeName" placeholder="请输入学员姓名" />
              </view>
              <view class="form-item">
                <text class="label">手机号</text>
                <input class="input" v-model="formData.traineePhone" type="number" placeholder="请输入学员手机号" />
              </view>
              <view class="form-item">
                <text class="label">性别</text>
                <radio-group class="radio-group" @change="(e) => formData.traineeGender = e.detail.value">
                  <label class="radio-label"><radio value="male" :checked="formData.traineeGender === 'male'" color="#667eea" />男</label>
                  <label class="radio-label"><radio value="female" :checked="formData.traineeGender === 'female'" color="#667eea" />女</label>
                </radio-group>
              </view>
              <view class="form-item">
                <text class="label">证件号码</text>
                <input class="input" v-model="formData.traineeIdCard" placeholder="请输入身份证号" />
              </view>
              <view class="form-item">
                <text class="label">考试机型</text>
                <picker class="picker" :range="examModelOptions" @change="(e) => formData.examModel = examModelOptions[e.detail.value]">
                  <view class="picker-value" :class="{ 'placeholder': !formData.examModel }">
                    {{ formData.examModel || '请选择机型' }}
                  </view>
                </picker>
              </view>
              <view class="form-item">
                <text class="label">证照级别</text>
                <picker class="picker" :range="licenseLevelOptions" @change="(e) => formData.licenseLevel = licenseLevelOptions[e.detail.value]">
                  <view class="picker-value" :class="{ 'placeholder': !formData.licenseLevel }">
                    {{ formData.licenseLevel || '请选择级别' }}
                  </view>
                </picker>
              </view>
            </template>

            <!-- 低空研学报名 (ID: 9) -->
            <template v-if="serviceId === '9'">
              <view class="form-item">
                <text class="label">学校/机构</text>
                <input class="input" v-model="formData.studyOrg" placeholder="请输入学校/机构名称" />
              </view>
              <view class="form-item">
                <text class="label">年级段</text>
                <input class="input" v-model="formData.studyGrade" placeholder="如：四-六年级 / 10-12岁" />
              </view>
              <view class="form-item">
                <text class="label">参与人数</text>
                <input class="input" v-model="formData.studyParticipants" type="number" placeholder="请输入参与人数" />
              </view>
              <view class="form-item">
                <text class="label">期望日期</text>
                <picker class="picker" mode="date" @change="(e) => formData.studyDate = e.detail.value">
                  <view class="picker-value" :class="{ 'placeholder': !formData.studyDate }">
                    {{ formData.studyDate || '请选择日期' }}
                  </view>
                </picker>
              </view>
              <view class="form-item">
                <text class="label">场次</text>
                <picker class="picker" :range="studySessionOptions" @change="(e) => formData.studySessionText = studySessionOptions[e.detail.value]">
                  <view class="picker-value" :class="{ 'placeholder': !formData.studySessionText }">
                    {{ formData.studySessionText || '请选择上午/下午场次' }}
                  </view>
                </picker>
              </view>
            </template>

            <!-- 无人机维修服务 (ID: 12) -->
            <template v-if="serviceId === '12'">
              <view class="form-item">
                <text class="label">服务类型</text>
                <radio-group class="radio-group" @change="(e) => formData.maintenanceType = e.detail.value">
                  <label class="radio-label"><radio value="repair" :checked="formData.maintenanceType === 'repair'" color="#667eea" />故障维修</label>
                  <label class="radio-label"><radio value="care" :checked="formData.maintenanceType === 'care'" color="#667eea" />定期保养</label>
                </radio-group>
              </view>
              <view class="form-item">
                <text class="label">机型</text>
                <input class="input" v-model="formData.droneModel" placeholder="请输入无人机型号" />
              </view>
              <view class="form-item">
                <text class="label">是否在保</text>
                <radio-group class="radio-group" @change="(e) => formData.isWarranty = e.detail.value">
                  <label class="radio-label"><radio value="yes" :checked="formData.isWarranty === 'yes'" color="#667eea" />在保修期</label>
                  <label class="radio-label"><radio value="no" :checked="formData.isWarranty === 'no'" color="#667eea" />已过保</label>
                </radio-group>
              </view>
              <view class="form-item">
                <text class="label">购买日期</text>
                <picker class="picker" mode="date" @change="(e) => formData.purchaseDate = e.detail.value">
                  <view class="picker-value" :class="{ 'placeholder': !formData.purchaseDate }">
                    {{ formData.purchaseDate || '选填' }}
                  </view>
                </picker>
              </view>
              <view class="form-item vertical no-border">
                <text class="label">故障/需求描述</text>
                <textarea class="textarea-main" v-model="formData.remark" placeholder="请详细描述设备故障情况或保养需求" />
              </view>
              <view class="form-item vertical no-border">
                <text class="label">设备照片 (选填)</text>
                <view class="uploader">
                  <view class="upload-item" v-for="(img, idx) in formData.fileList" :key="idx">
                    <image :src="img" mode="aspectFill" @tap="previewImage(img)" />
                    <text class="del-btn" @tap="delImage(idx)">×</text>
                  </view>
                  <view class="upload-btn" @tap="chooseImage" v-if="formData.fileList.length < 5">
                    <text class="plus">+</text>
                  </view>
                </view>
              </view>
            </template>

            <!-- 无人机赛事报名 (ID: 13) -->
            <template v-if="serviceId === '13'">
              <view class="form-item">
                <text class="label">注册类型</text>
                <picker class="picker" :range="competitionRoleLabels" @change="onCompetitionRoleChange">
                  <view class="picker-value" :class="{ 'placeholder': !formData.competitionRoleText }">
                    {{ formData.competitionRoleText || '请选择注册角色' }}
                  </view>
                </picker>
              </view>

              <!-- 赛事角色表单 -->
              <template v-if="formData.competitionRole">
                <view class="section-title" style="margin-top: 12px;">表单报名</view>

                <!-- 裁判员 -->
                <template v-if="formData.competitionRole === 'referee'">
                  <view class="form-item">
                    <text class="label">单位名称</text>
                    <input class="input" v-model="formData.companyName" placeholder="请输入单位名称" />
                  </view>
                  <view class="form-item">
                    <text class="label">姓名</text>
                    <input class="input" v-model="formData.name" placeholder="请输入姓名" />
                  </view>
                  <view class="form-item">
                    <text class="label">性别</text>
                    <radio-group class="radio-group" @change="(e) => formData.gender = e.detail.value">
                      <label class="radio-label"><radio value="male" :checked="formData.gender === 'male'" color="#667eea" />男</label>
                      <label class="radio-label"><radio value="female" :checked="formData.gender === 'female'" color="#667eea" />女</label>
                    </radio-group>
                  </view>
                  <view class="form-item">
                    <text class="label">证件号</text>
                    <input class="input" v-model="formData.idCard" placeholder="请输入证件号" />
                  </view>
                  <view class="form-item">
                    <text class="label">组别</text>
                    <picker class="picker" :range="competitionGroupOptions" @change="(e) => formData.competitionGroup = competitionGroupOptions[e.detail.value]">
                      <view class="picker-value" :class="{ 'placeholder': !formData.competitionGroup }">
                        {{ formData.competitionGroup || '请选择组别' }}
                      </view>
                    </picker>
                  </view>
                  <view class="form-item">
                    <text class="label">联系电话</text>
                    <input class="input" v-model="formData.phone" type="number" placeholder="请输入联系电话" />
                  </view>
                  <view class="form-item">
                    <text class="label">电子邮箱</text>
                    <input class="input" v-model="formData.email" placeholder="请输入邮箱（选填）" />
                  </view>
                </template>

                <!-- 教练员 -->
                <template v-if="formData.competitionRole === 'coach'">
                  <view class="form-item">
                    <text class="label">单位名称</text>
                    <input class="input" v-model="formData.companyName" placeholder="请输入单位名称" />
                  </view>
                  <view class="form-item">
                    <text class="label">姓名</text>
                    <input class="input" v-model="formData.name" placeholder="请输入姓名" />
                  </view>
                  <view class="form-item">
                    <text class="label">性别</text>
                    <radio-group class="radio-group" @change="(e) => formData.gender = e.detail.value">
                      <label class="radio-label"><radio value="male" :checked="formData.gender === 'male'" color="#667eea" />男</label>
                      <label class="radio-label"><radio value="female" :checked="formData.gender === 'female'" color="#667eea" />女</label>
                    </radio-group>
                  </view>
                  <view class="form-item">
                    <text class="label">证件号</text>
                    <input class="input" v-model="formData.idCard" placeholder="请输入证件号" />
                  </view>
                  <view class="form-item">
                    <text class="label">组别</text>
                    <picker class="picker" :range="competitionGroupOptions" @change="(e) => formData.competitionGroup = competitionGroupOptions[e.detail.value]">
                      <view class="picker-value" :class="{ 'placeholder': !formData.competitionGroup }">
                        {{ formData.competitionGroup || '请选择组别' }}
                      </view>
                    </picker>
                  </view>
                  <view class="form-item">
                    <text class="label">联系电话</text>
                    <input class="input" v-model="formData.phone" type="number" placeholder="请输入联系电话" />
                  </view>
                  <view class="form-item">
                    <text class="label">电子邮箱</text>
                    <input class="input" v-model="formData.email" placeholder="请输入邮箱（选填）" />
                  </view>
                </template>

                <!-- 运动员 -->
                <template v-if="formData.competitionRole === 'athlete'">
                  <view class="form-item">
                    <text class="label">单位/学校</text>
                    <input class="input" v-model="formData.companyName" placeholder="请输入单位/学校名称" />
                  </view>
                  <view class="form-item">
                    <text class="label">姓名</text>
                    <input class="input" v-model="formData.name" placeholder="请输入姓名" />
                  </view>
                  <view class="form-item">
                    <text class="label">性别</text>
                    <radio-group class="radio-group" @change="(e) => formData.gender = e.detail.value">
                      <label class="radio-label"><radio value="male" :checked="formData.gender === 'male'" color="#667eea" />男</label>
                      <label class="radio-label"><radio value="female" :checked="formData.gender === 'female'" color="#667eea" />女</label>
                    </radio-group>
                  </view>
                  <view class="form-item">
                    <text class="label">证件号</text>
                    <input class="input" v-model="formData.idCard" placeholder="请输入证件号" />
                  </view>
                  <view class="form-item">
                    <text class="label">组别</text>
                    <picker class="picker" :range="athleteGroupOptions" @change="(e) => formData.athleteGroup = athleteGroupOptions[e.detail.value]">
                      <view class="picker-value" :class="{ 'placeholder': !formData.athleteGroup }">
                        {{ formData.athleteGroup || '请选择组别' }}
                      </view>
                    </picker>
                  </view>
                  <view class="form-item">
                    <text class="label">参赛项目</text>
                    <picker class="picker" :range="competitionProjectOptions" @change="(e) => formData.competitionProject = competitionProjectOptions[e.detail.value]">
                      <view class="picker-value" :class="{ 'placeholder': !formData.competitionProject }">
                        {{ formData.competitionProject || '请选择项目' }}
                      </view>
                    </picker>
                  </view>
                  <view class="form-item">
                    <text class="label">联系电话</text>
                    <input class="input" v-model="formData.phone" type="number" placeholder="请输入联系电话" />
                  </view>
                </template>

                <!-- 俱乐部 -->
                <template v-if="formData.competitionRole === 'club'">
                  <view class="form-item">
                    <text class="label">单位名称</text>
                    <input class="input" v-model="formData.companyName" placeholder="请输入单位名称" />
                  </view>
                  <view class="form-item">
                    <text class="label">单位简称</text>
                    <input class="input" v-model="formData.companyShortName" placeholder="请输入单位简称" />
                  </view>
                  <view class="form-item">
                    <text class="label">所在地</text>
                    <input class="input" v-model="formData.location" placeholder="请输入单位所在地" />
                  </view>
                  <view class="form-item">
                    <text class="label">负责人</text>
                    <input class="input" v-model="formData.manager" placeholder="请输入负责人姓名" />
                  </view>
                  <view class="form-item">
                    <text class="label">负责人电话</text>
                    <input class="input" v-model="formData.managerPhone" type="number" placeholder="请输入负责人电话" />
                  </view>
                  <view class="form-item">
                    <text class="label">对接人</text>
                    <input class="input" v-model="formData.contactPerson" placeholder="请输入对接人姓名" />
                  </view>
                  <view class="form-item">
                    <text class="label">对接人电话</text>
                    <input class="input" v-model="formData.contactPhone" type="number" placeholder="请输入对接人电话" />
                  </view>
                </template>
              </template>
            </template>

            <!-- 通用备注 (非物流、非政务、非维修、非研学) -->
            <view class="form-item vertical no-border" v-if="!['1', '2', '12', '9'].includes(serviceId)">
              <text class="label">备注/其他需求</text>
              <textarea class="textarea-main" v-model="formData.remark" placeholder="请详细说明您的需求，以便我们为您提供更精准的服务。" />
            </view>
          </view>
        </view>

        <view class="submit-section">
          <button class="submit-btn" type="primary" @tap="handleSubmit">{{ submitButtonText }}</button>
        </view>
      </view>
    </view>
    <HomeFloatButton />
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import HomeFloatButton from '@/components/HomeFloatButton.vue'
import { getStoredUser, request, authStorage } from '../../utils/request'

const serviceId = ref('')
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
  studyOrg: '', studyGrade: '', studyParticipants: '', studyDate: '', studySessionText: '',
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
const studySessionOptions = ['上午场 (9:00-11:30)', '下午场 (13:30-16:00)', '全天 (9:00-16:00)']
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
  serviceId.value = options.id || '1'
  const user = getStoredUser()
  if (user) {
    formData.value.contactName = user.name || ''
    formData.value.contactPhone = user.phone || ''
    formData.value.traineeName = user.name || ''
    formData.value.traineePhone = user.phone || ''
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

const previewImage = (current) => {
  uni.previewImage({ current, urls: formData.value.fileList })
}

const delImage = (index) => {
  formData.value.fileList.splice(index, 1)
}

const goBack = () => {
  uni.navigateBack()
}

const handleSubmit = async () => {
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
        uni.switchTab({ url: '/pages/applications/index' })
      }
    })
  } catch (error) {
    const mock = uni.getStorageSync('mock_applications') || []
    mock.unshift(submitData)
    uni.setStorageSync('mock_applications', mock)

    uni.hideLoading()
    uni.showModal({
      title: '✅ 提交成功',
      content: `申请单号：${orderNo}\n我们将尽快与您联系并确认服务方案！`,
      showCancel: false,
      confirmText: '查看我的申请',
      success: () => {
        uni.switchTab({ url: '/pages/applications/index' })
      }
    })
  }
}
</script>

<style scoped>
.apply-page { min-height: 100vh; background: #f7f8fa; padding-bottom: 40px; }

.empty-section { padding: 60px 40px; text-align: center; }
.empty-icon { font-size: 48px; margin-bottom: 20px; }
.empty-text { font-size: 18px; font-weight: bold; color: #323233; margin-bottom: 12px; }
.empty-desc { font-size: 14px; color: #969799; line-height: 1.6; }
.empty-contact { font-size: 13px; color: #667eea; margin-top: 12px; }
.back-btn { margin-top: 30px; border-radius: 99px; font-size: 15px; color: #fff; background: #667eea; }

.notice-bar { display: flex; align-items: center; padding: 10px 16px; background: #f0f2ff; color: #667eea; }
.notice-icon { margin-right: 8px; font-size: 14px; }
.notice-text { font-size: 13px; }

.form-body { padding: 16px; }
.form-section { background: #fff; border-radius: 12px; padding: 0 16px; margin-bottom: 16px; overflow: hidden; }
.section-title { font-size: 15px; font-weight: 600; padding: 16px 0 8px; color: #323233; border-bottom: 1px solid #f2f3f5; margin-bottom: 8px; }

.form-item { display: flex; align-items: center; padding: 14px 0; border-bottom: 1px solid #f2f3f5; }
.form-item.vertical { flex-direction: column; align-items: flex-start; gap: 10px; }
.form-item.no-border { border-bottom: none; }

.label { width: 90px; font-size: 14px; color: #646566; flex-shrink: 0; }
.input { flex: 1; font-size: 14px; color: #323233; }
.unit { font-size: 14px; color: #969799; margin-left: 8px; }
.picker { flex: 1; }
.picker-value { font-size: 14px; color: #323233; }
.picker-value.placeholder { color: #ccc; }

.radio-group { display: flex; gap: 20px; }
.radio-label { display: flex; align-items: center; font-size: 14px; color: #323233; gap: 4px; }

.textarea-small { width: 100%; height: 60px; font-size: 13px; background: #f7f8fa; border-radius: 6px; padding: 8px; }
.textarea-main { width: 100%; height: 100px; font-size: 14px; background: #f7f8fa; border-radius: 8px; padding: 12px; }

.uploader { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 8px; }
.upload-item { width: 80px; height: 80px; position: relative; border-radius: 8px; overflow: hidden; }
.upload-item image { width: 100%; height: 100%; }
.del-btn { position: absolute; top: 0; right: 0; width: 20px; height: 20px; background: rgba(0,0,0,0.5); color: #fff; text-align: center; line-height: 18px; border-bottom-left-radius: 8px; }
.upload-btn { width: 80px; height: 80px; background: #f7f8fa; border: 1px dashed #ddd; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.plus { font-size: 30px; color: #ccc; }

.submit-section { padding: 20px 16px 40px; }
.submit-btn { border-radius: 99px; height: 48px; line-height: 48px; font-weight: bold; background: #667eea !important; }
</style>
