<template>
  <div class="apply-page page-container">
    <van-nav-bar
      :title="pageDisplayTitle"
      left-arrow
      @click-left="$router.back()"
      fixed
      placeholder
    />
    <HomeFloatButton />

    <div class="content-wrapper">
      <!-- 服务暂未开放提示 (5-14号服务) -->
      <template v-if="!isServiceAvailable">
        <van-empty
          image="search"
          description="该服务申请功能即将开放"
        >
          <div class="empty-tips">
            <p style="color: #969799; margin-bottom: 12px;">
              {{ serviceName }}功能正在建设中，敬请期待！
            </p>
            <p style="color: #969799; font-size: 13px; margin-bottom: 20px;">
              如有需求，请联系客服：0577-55558188
            </p>
            <van-button 
              type="primary" 
              round 
              size="small"
              @click="$router.back()"
            >
              返回服务列表
            </van-button>
          </div>
        </van-empty>
      </template>

      <!-- 服务信息提示 (1-4, 6号服务) -->
      <template v-else>
        <van-notice-bar
          color="#667eea"
          background="#f0f2ff"
          left-icon="info-o"
        >
          请填写以下信息，我们将尽快与您联系
        </van-notice-bar>

        <!-- 申请表单 -->
        <van-form ref="formRef" @submit="onSubmit" @failed="onFailed" style="margin-top: 16px;">
        <!-- 基本信息 -->
        <div class="form-section" v-if="serviceId !== '6' && serviceId !== '13'">
          <h3 class="form-title">基本信息</h3>
          <van-field
            v-model="formData.contactName"
            name="contactName"
            label="联系人"
            placeholder="请输入联系人姓名"
            :rules="[{ required: true, message: '请输入联系人姓名' }]"
          />
          <van-field
            v-model="formData.contactPhone"
            name="contactPhone"
            type="tel"
            label="联系电话"
            placeholder="请输入联系电话"
            :rules="[
              { required: true, message: '请输入联系电话' }
            ]"
          />
        </div>

        <!-- 服务详情 - 根据服务类型显示不同字段 -->
        <div class="form-section">
          <h3 class="form-title" v-if="serviceId !== '13'">服务详情</h3>
          <h3 class="form-title" v-else>板块报名</h3>
          
          <!-- 无人机赛事报名 -->
          <template v-if="serviceId === '13'">
            <van-field
              v-model="formData.competitionRoleText"
              is-link
              readonly
              label="注册类型"
              placeholder="请选择注册角色"
              @click="showCompetitionRolePicker = true"
              @click-input="showCompetitionRolePicker = true"
              :rules="[{ required: true, message: '请选择注册角色' }]"
            />
            <van-popup :show="showCompetitionRolePicker" @update:show="val => showCompetitionRolePicker = val" position="bottom">
              <van-picker
                :columns="competitionRoleOptions"
                @confirm="onCompetitionRoleConfirm"
                @cancel="showCompetitionRolePicker = false"
                title="选择注册角色"
              />
            </van-popup>

            <!-- 选择角色后展开表单 -->
            <template v-if="formData.competitionRole">
              <h3 class="form-title" style="margin-top: 20px;">表单报名</h3>
              
              <!-- 裁判员 字段 -->
              <template v-if="formData.competitionRole === 'referee'">
                <van-field v-model="formData.companyName" label="单位名称" placeholder="请输入单位名称" :rules="[{ required: true, message: '请输入单位名称' }]" />
                <van-field v-model="formData.name" label="姓名" placeholder="请输入姓名" :rules="[{ required: true, message: '请输入姓名' }]" />
                <van-field name="gender" label="性别">
                  <template #input>
                    <van-radio-group v-model="formData.gender" direction="horizontal">
                      <van-radio name="male">男</van-radio>
                      <van-radio name="female">女</van-radio>
                    </van-radio-group>
                  </template>
                </van-field>
                <van-field v-model="formData.idCard" label="证件号" placeholder="请输入证件号" :rules="[{ required: true, message: '请输入证件号' }, { pattern: /^\d{17}[\dXx]$/, message: '证件号必须为18位' }]" />
                <van-field
                  v-model="formData.competitionGroup"
                  is-link
                  readonly
                  label="组别"
                  placeholder="请选择组别（足球/FPV/航拍）"
                  @click="showGroupPicker = true"
                  :rules="[{ required: true, message: '请选择组别' }]"
                />
                <van-field v-model="formData.phone" type="tel" label="联系电话" placeholder="请输入联系电话" :rules="[{ required: true, message: '请输入电话' }, { pattern: /^1\d{10}$/, message: '电话号码必须为11位' }]" />
                <van-field v-model="formData.email" label="电子邮箱" placeholder="请输入电子邮箱" />
              </template>

              <!-- 教练员 字段 -->
              <template v-if="formData.competitionRole === 'coach'">
                <van-field v-model="formData.companyName" label="单位名称" placeholder="请输入单位名称" :rules="[{ required: true, message: '请输入单位名称' }]" />
                <van-field v-model="formData.name" label="姓名" placeholder="请输入姓名" :rules="[{ required: true, message: '请输入姓名' }]" />
                <van-field name="gender" label="性别">
                  <template #input>
                    <van-radio-group v-model="formData.gender" direction="horizontal">
                      <van-radio name="male">男</van-radio>
                      <van-radio name="female">女</van-radio>
                    </van-radio-group>
                  </template>
                </van-field>
                <van-field v-model="formData.idCard" label="证件号" placeholder="请输入证件号" :rules="[{ required: true, message: '请输入证件号' }, { pattern: /^\d{17}[\dXx]$/, message: '证件号必须为18位' }]" />
                <van-field
                  v-model="formData.competitionGroup"
                  is-link
                  readonly
                  label="组别"
                  placeholder="请选择组别（足球/FPV/航拍）"
                  @click="showGroupPicker = true"
                  :rules="[{ required: true, message: '请选择组别' }]"
                />
                <van-field v-model="formData.phone" type="tel" label="联系电话" placeholder="请输入联系电话" :rules="[{ required: true, message: '请输入电话' }, { pattern: /^1\d{10}$/, message: '电话号码必须为11位' }]" />
                <van-field v-model="formData.email" label="电子邮箱" placeholder="请输入电子邮箱" />
              </template>

              <!-- 运动员 -->
              <template v-if="formData.competitionRole === 'athlete'">
                <van-field v-model="formData.companyName" label="单位名称" placeholder="请输入单位/学校名称" :rules="[{ required: true, message: '请输入单位/学校名称' }]" />
                <van-field v-model="formData.name" label="姓名" placeholder="请输入姓名" :rules="[{ required: true, message: '请输入姓名' }]" />
                <van-field name="gender" label="性别">
                  <template #input>
                    <van-radio-group v-model="formData.gender" direction="horizontal">
                      <van-radio name="male">男</van-radio>
                      <van-radio name="female">女</van-radio>
                    </van-radio-group>
                  </template>
                </van-field>
                <van-field v-model="formData.idCard" label="证件号" placeholder="请输入证件号" :rules="[{ required: true, message: '请输入证件号' }, { pattern: /^\d{17}[\dXx]$/, message: '证件号必须为18位' }]" />
                <van-field
                  v-model="formData.athleteGroup"
                  is-link
                  readonly
                  label="组别"
                  placeholder="请选择组别（成人/中学/小学）"
                  @click="showAthleteGroupPicker = true"
                  :rules="[{ required: true, message: '请选择组别' }]"
                />
                <van-field
                  v-model="formData.competitionProject"
                  is-link
                  readonly
                  label="参赛项目"
                  placeholder="请选择项目（足球/FPV/航拍）"
                  @click="showProjectPicker = true"
                  :rules="[{ required: true, message: '请选择参赛项目' }]"
                />
                <van-field v-model="formData.phone" type="tel" label="联系电话" placeholder="请输入联系电话" :rules="[{ required: true, message: '请输入电话' }, { pattern: /^1\d{10}$/, message: '电话号码必须为11位' }]" />
              </template>

              <!-- 俱乐部 -->
              <template v-if="formData.competitionRole === 'club'">
                <van-field v-model="formData.companyName" label="单位名称" placeholder="请输入单位名称" :rules="[{ required: true, message: '请输入单位名称' }]" />
                <van-field v-model="formData.companyShortName" label="单位简称" placeholder="请输入单位简称" />
                <van-field v-model="formData.location" label="所在地" placeholder="请输入单位所在地" />
                <van-field name="logoUploader" label="LOGO">
                  <template #input>
                    <van-uploader v-model="formData.logoFileList" :max-count="1" accept="image/*" />
                  </template>
                </van-field>
                <van-field v-model="formData.manager" label="负责人" placeholder="请输入负责人姓名" :rules="[{ required: true, message: '请输入负责人' }]" />
                <van-field v-model="formData.managerPhone" type="tel" label="联系电话" placeholder="请输入负责人电话" :rules="[{ required: true, message: '请输入负责人电话' }, { pattern: /^1\d{10}$/, message: '电话号码必须为11位' }]" />
                <van-field v-model="formData.contactPerson" label="主要对接人" placeholder="请输入对接人姓名" :rules="[{ required: true, message: '请输入对接人' }]" />
                <van-field v-model="formData.contactPhone" type="tel" label="联系电话" placeholder="请输入对接人电话" :rules="[{ pattern: /^1\d{10}$/, message: '电话号码必须为11位' }]" />
              </template>
            </template>
          </template>

          <!-- 无人机物流服务 -->
          <template v-if="serviceId === '1'">
            <!-- 客户类型 -->
            <van-field name="customerType" label="客户类型">
              <template #input>
                <van-radio-group v-model="formData.customerType" direction="horizontal">
                  <van-radio name="personal">个人</van-radio>
                  <van-radio name="enterprise">企业</van-radio>
                </van-radio-group>
              </template>
            </van-field>

            <!-- 企业名称（企业客户时显示） -->
            <van-field
              v-if="formData.customerType === 'enterprise'"
              v-model="formData.companyName"
              label="企业名称"
              placeholder="请输入企业名称"
            />

            <!-- 货物信息 -->
            <van-field
              v-model="formData.cargoType"
              is-link
              readonly
              label="货物类型"
              placeholder="请选择货物类型"
              @click="showCargoTypePicker = true"
              @click-input="showCargoTypePicker = true"
            />
            <van-popup :show="showCargoTypePicker" @update:show="val => showCargoTypePicker = val" position="bottom">
              <van-picker
                :columns="cargoTypeOptions"
                @confirm="onCargoTypeConfirm"
                @cancel="showCargoTypePicker = false"
                title="选择货物类型"
              />
            </van-popup>

            <!-- 其他货物类型（选择"其他"时显示） -->
            <van-field
              v-if="formData.cargoType === '其他'"
              v-model="formData.cargoTypeOther"
              label="具体类型"
              placeholder="请输入具体货物类型"
            />

            <van-field
              v-model="formData.cargoWeight"
              type="number"
              label="货物重量"
              placeholder="请输入货物重量"
            >
              <template #button>
                <span style="color: #969799;">kg</span>
              </template>
            </van-field>

            <van-field
              v-model="formData.cargoVolume"
              label="货物体积"
              placeholder="请输入货物体积（长×宽×高）"
            >
              <template #button>
                <span style="color: #969799;">cm³</span>
              </template>
            </van-field>

            <!-- 起运地 -->
            <van-field
              v-model="formData.startAddress"
              label="起运地"
              placeholder="请输入起运地"
              :rules="[]"
            />

            <van-field
              v-model="formData.startAddressDetail"
              label="详细地址"
              placeholder="请输入起运地详细地址（楼栋门牌号）"
              type="textarea"
              rows="2"
            />

            <!-- 目的地 -->
            <van-field
              v-model="formData.endAddress"
              label="目的地"
              placeholder="请输入目的地"
              :rules="[]"
            />

            <van-field
              v-model="formData.endAddressDetail"
              label="详细地址"
              placeholder="请输入目的地详细地址（楼栋门牌号）"
              type="textarea"
              rows="2"
            />

            <!-- 运输时效 -->
            <van-field
              v-model="formData.deliveryUrgency"
              is-link
              readonly
              label="运输时效"
              placeholder="请选择运输时效"
              @click="showUrgencyPicker = true"
              @click-input="showUrgencyPicker = true"
            />
            <van-popup :show="showUrgencyPicker" @update:show="val => showUrgencyPicker = val" position="bottom">
              <van-picker
                :columns="urgencyOptions"
                @confirm="onUrgencyConfirm"
                @cancel="showUrgencyPicker = false"
                title="选择运输时效"
              />
            </van-popup>

            <!-- 期望运输时间 -->
            <van-field
              v-model="formData.expectedTime"
              is-link
              readonly
              label="期望运输时间"
              placeholder="请选择期望运输时间"
              @click="showTimePicker = true"
              @click-input="showTimePicker = true"
            />
            <van-popup :show="showTimePicker" @update:show="val => showTimePicker = val" position="bottom">
              <van-date-picker
                v-model="expectedDate"
                title="选择日期"
                :min-date="minDate"
                @confirm="onTimeConfirm"
                @cancel="showTimePicker = false"
              />
            </van-popup>

            <!-- 附件上传 -->
            <van-field name="uploader" label="货物照片">
              <template #input>
                <van-uploader 
                  v-model="formData.fileList" 
                  :max-count="5"
                  :after-read="afterRead"
                  accept="image/*"
                >
                  <van-button icon="plus" type="primary" size="small">上传图片</van-button>
                </van-uploader>
              </template>
            </van-field>

            <van-field
              v-model="formData.remark"
              label="备注说明"
              placeholder="请输入其他需要说明的信息"
              type="textarea"
              rows="3"
            />
          </template>

          <!-- 政务服务 -->
          <template v-if="serviceId === '2'">
            <van-field
              v-model="formData.inspectionType"
              label="服务类型"
              placeholder="请输入服务类型，如：环保监测、安全巡查、设施检查等"
              type="text"
            />
            <van-field
              v-model="formData.inspectionArea"
              label="巡检区域"
              placeholder="请输入巡检区域，如：某某区域、某某路段等"
              type="text"
            />
            <van-field
              v-model="formData.inspectionDate"
              label="巡检时间"
              placeholder="请输入巡检时间，如：2025-01-25 或 每周一上午"
              type="text"
            />
            <van-field
              v-model="formData.inspectionRequire"
              label="需求说明"
              placeholder="请详细描述您的巡检需求"
              type="text"
            />
          </template>

          <!-- 无人机托管服务 -->
          <template v-if="serviceId === '3'">
            <van-field
              v-model="formData.droneModel"
              label="无人机型号"
              placeholder="请输入无人机型号"
            />
            <van-field
              v-model="formData.droneCount"
              type="number"
              label="托管数量"
              placeholder="请输入托管数量"
            />
            <van-field
              v-model="formData.trusteeDuration"
              label="托管时长"
              placeholder="请选择托管时长"
              readonly
              is-link
              @click="showDurationPicker = true"
            />
          </template>

          <!-- 无人机吊运服务 -->
          <template v-if="serviceId === '4'">
            <van-field
              v-model="formData.liftItemType"
              label="吊运物品"
              placeholder="请输入吊运物品类型"
            />
            <van-field
              v-model="formData.liftItemWeight"
              type="number"
              label="物品重量(kg)"
              placeholder="请输入物品重量"
            />
            <van-field
              v-model="formData.workLocation"
              label="作业地点"
              placeholder="请输入作业地点"
            />
            <van-field
              v-model="formData.liftHeight"
              type="number"
              label="吊运高度(m)"
              placeholder="请输入吊运高度"
            />
          </template>

          <!-- 飞手培训服务 -->
          <template v-if="serviceId === '6'">
            <van-field
              v-model="formData.traineeName"
              label="姓名"
              placeholder="请输入姓名"
              :rules="[{ required: true, message: '请输入姓名' }]"
            />
            <van-field
              v-model="formData.traineePhone"
              type="tel"
              label="联系电话"
              placeholder="请输入联系电话"
              :rules="[{ required: true, message: '请输入联系电话' }]"
            />
            <van-field name="traineeGender" label="性别">
              <template #input>
                <van-radio-group v-model="formData.traineeGender" direction="horizontal">
                  <van-radio name="male">男</van-radio>
                  <van-radio name="female">女</van-radio>
                </van-radio-group>
              </template>
            </van-field>
            <van-field
              v-model="formData.traineeBirthday"
              label="出生日期"
              placeholder="请输入出生日期（如：1990-01-01）"
              :rules="[{ required: true, message: '请输入出生日期' }]"
            />
            <van-field
              v-model="formData.traineeIdCard"
              label="身份证号"
              placeholder="请输入身份证号"
              :rules="[{ required: true, message: '请输入身份证号' }]"
            />
            <van-field
              v-model="formData.examModel"
              is-link
              readonly
              label="考试机型"
              placeholder="请选择考试机型"
              @click="showExamModelPicker = true"
              @click-input="showExamModelPicker = true"
            />
            <van-popup :show="showExamModelPicker" @update:show="val => showExamModelPicker = val" position="bottom">
              <van-picker
                :columns="examModelOptions"
                @confirm="onExamModelConfirm"
                @cancel="showExamModelPicker = false"
                title="选择考试机型"
              />
            </van-popup>
            <van-field
              v-model="formData.licenseLevel"
              is-link
              readonly
              label="证照级别"
              placeholder="请选择证照级别"
              @click="showLicenseLevelPicker = true"
              @click-input="showLicenseLevelPicker = true"
            />
            <van-popup :show="showLicenseLevelPicker" @update:show="val => showLicenseLevelPicker = val" position="bottom">
              <van-picker
                :columns="licenseLevelOptions"
                @confirm="onLicenseLevelConfirm"
                @cancel="showLicenseLevelPicker = false"
                title="选择证照级别"
              />
            </van-popup>
            <van-field
              v-model="formData.hasExperienceText"
              is-link
              readonly
              label="有无基础"
              placeholder="请选择有无基础"
              @click="showExperiencePicker = true"
              @click-input="showExperiencePicker = true"
            />
            <van-popup :show="showExperiencePicker" @update:show="val => showExperiencePicker = val" position="bottom">
              <van-picker
                :columns="experienceOptions"
                @confirm="onExperienceConfirm"
                @cancel="showExperiencePicker = false"
                title="选择有无基础"
              />
            </van-popup>
          </template>

          <!-- 低空研学报名（服务 9） -->
          <template v-if="serviceId === '9'">
            <van-field
              v-model="formData.studyOrg"
              label="学校/机构"
              placeholder="请输入学校/机构名称"
              :rules="[{ required: true, message: '请输入学校/机构名称' }]"
            />
            <van-field
              v-model="formData.studyGrade"
              label="年级/年龄段"
              placeholder="如：四-六年级 / 10-12岁（可选）"
            />
            <van-field
              v-model="formData.studyParticipants"
              type="number"
              label="参与人数"
              placeholder="请输入参与人数"
              :rules="[{ required: true, message: '请输入参与人数' }]"
            />
            <van-field
              v-model="formData.studyDate"
              is-link
              readonly
              label="期望日期"
              placeholder="请选择日期"
              @click="showStudyDatePicker = true"
              @click-input="showStudyDatePicker = true"
              :rules="[{ required: true, message: '请选择日期' }]"
            />
            <van-popup :show="showStudyDatePicker" @update:show="val => showStudyDatePicker = val" position="bottom">
              <van-date-picker
                v-model="studyDate"
                title="选择日期"
                :min-date="minDate"
                @confirm="onStudyDateConfirm"
                @cancel="showStudyDatePicker = false"
              />
            </van-popup>

            <van-field
              v-model="formData.studySessionText"
              is-link
              readonly
              label="场次"
              placeholder="请选择上午/下午场次"
              @click="showStudySessionPicker = true"
              @click-input="showStudySessionPicker = true"
              :rules="[{ required: true, message: '请选择场次' }]"
            />
            <van-popup :show="showStudySessionPicker" @update:show="val => showStudySessionPicker = val" position="bottom">
              <van-picker
                :columns="studySessionOptions"
                @confirm="onStudySessionConfirm"
                @cancel="showStudySessionPicker = false"
                title="选择场次"
              />
            </van-popup>

            <van-field
              v-model="formData.remark"
              label="备注"
              type="textarea"
              rows="3"
              maxlength="200"
              show-word-limit
              placeholder="可填写：集合方式/是否需要发票/其他需求（可选）"
            />
          </template>

          <!-- 无人机维修服务 -->
          <template v-if="serviceId === '12'">
            <van-field name="maintenanceType" label="服务类型">
              <template #input>
                <van-radio-group v-model="formData.maintenanceType" direction="horizontal">
                  <van-radio name="repair">故障维修</van-radio>
                  <van-radio name="care">定期保养</van-radio>
                </van-radio-group>
              </template>
            </van-field>

            <van-field
              v-model="formData.droneModel"
              label="无人机型号"
              placeholder="请输入无人机型号"
              :rules="[{ required: true, message: '请输入无人机型号' }]"
            />

            <van-field name="isWarranty" label="是否在保">
              <template #input>
                <van-radio-group v-model="formData.isWarranty" direction="horizontal">
                  <van-radio name="yes">在保修期内</van-radio>
                  <van-radio name="no">已过保</van-radio>
                </van-radio-group>
              </template>
            </van-field>

            <van-field
              v-model="formData.purchaseDate"
              label="购买日期"
              placeholder="请选择购买日期（选填）"
              is-link
              readonly
              @click="showDatePicker = true"
            />

            <van-field
              v-model="formData.remark"
              label="故障/需求描述"
              type="textarea"
              placeholder="请详细描述设备故障情况或保养需求"
              rows="3"
              show-word-limit
              maxlength="200"
              :rules="[{ required: true, message: '请描述故障或需求' }]"
            />

            <!-- 故障图片上传 -->
            <van-field name="uploader" label="设备照片">
              <template #input>
                <van-uploader 
                  v-model="formData.fileList" 
                  :max-count="5"
                  :after-read="afterRead"
                  accept="image/*"
                >
                  <div class="upload-slot">
                    <van-icon name="photograph" size="24" color="#dcdee0" />
                    <span>上传照片</span>
                  </div>
                </van-uploader>
              </template>
            </van-field>
          </template>

          <!-- 通用备注（物流、政务服务、维修已有备注，其他服务显示） -->
          <van-field
            v-if="serviceId !== '1' && serviceId !== '2' && serviceId !== '12'"
            v-model="formData.remark"
            type="textarea"
            :label="serviceId === '13' ? '备注' : '需求说明'"
            :placeholder="serviceId === '13' ? '请输入备注信息' : '请描述您的具体需求'"
            rows="3"
            maxlength="200"
            show-word-limit
          />
        </div>

        <!-- 提交按钮 -->
        <div style="margin: 24px 16px;">
          <van-button round block type="primary" native-type="button" @click="manualSubmit">
            {{ submitButtonText }}
          </van-button>
        </div>
      </van-form>
      </template>
    </div>

    <van-popup :show="showDatePicker" @update:show="val => showDatePicker = val" position="bottom">
      <van-date-picker
        v-model="selectedDateArray"
        title="选择日期"
        @confirm="onDateConfirm"
        @cancel="showDatePicker = false"
      />
    </van-popup>

    <van-popup :show="showDurationPicker" @update:show="val => showDurationPicker = val" position="bottom">
      <van-picker
        :columns="durationOptions"
        @confirm="onDurationConfirm"
        @cancel="showDurationPicker = false"
        title="选择托管期限"
      />
    </van-popup>

    <van-popup :show="showGroupPicker" @update:show="val => showGroupPicker = val" position="bottom">
      <van-picker
        :columns="groupOptions"
        @confirm="onGroupConfirm"
        @cancel="showGroupPicker = false"
        title="选择组别"
      />
    </van-popup>

    <van-popup :show="showAthleteGroupPicker" @update:show="val => showAthleteGroupPicker = val" position="bottom">
      <van-picker
        :columns="athleteGroupOptions"
        @confirm="onAthleteGroupConfirm"
        @cancel="showAthleteGroupPicker = false"
        title="选择组别"
      />
    </van-popup>

    <van-popup :show="showProjectPicker" @update:show="val => showProjectPicker = val" position="bottom">
      <van-picker
        :columns="projectOptions"
        @confirm="onProjectConfirm"
        @cancel="showProjectPicker = false"
        title="选择参赛项目"
      />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast, showDialog, showLoadingToast, closeToast, showFailToast } from 'vant'
import axios from 'axios'
import HomeFloatButton from '@/components/HomeFloatButton.vue'

const route = useRoute()
const router = useRouter()
// 使用 computed 获取路由参数，保持响应式
const serviceId = computed(() => String(route.params.id))

onMounted(() => {
  // 处理从详情页带过来的角色参数
  const queryRole = route.query.role
  if (serviceId.value === '13' && queryRole) {
    const roleMap = {
      'referee': '裁判员',
      'coach': '教练员',
      'athlete': '运动员',
      'club': '俱乐部'
    }
    const roleText = roleMap[queryRole]
    if (roleText) {
      formData.value.competitionRole = queryRole
      formData.value.competitionRoleText = roleText
    }
  }
})

// 所有11项服务名称 (ID 8是外卖，ID flight在列表页直接跳转)
const serviceNames = {
  '1': '无人机物流',
  '2': '政务服务',
  '3': '无人机托管',
  '4': '无人机吊运',
  '5': '无人机表演',
  '6': '飞手培训',
  '7': '无人机租赁',
  '8': '无人机外卖',
  '9': '低空研学',
  '10': '无人机销售',
  '11': '无人机金融服务',
  '12': '无人机维修服务',
  '13': 'DSL'
}

const serviceName = computed(() => serviceNames[serviceId.value] || '服务')

// 页面显示的标题
const pageDisplayTitle = computed(() => {
  if (serviceId.value === '13') return 'DSL 预注册'
  return serviceName.value + '申请'
})

const submitButtonText = computed(() => {
  // 物流、吊运 -> 下单
  if (['1', '4', '8'].includes(serviceId.value)) {
    return '立即下单'
  }
  // 赛事报名 -> 提交
  if (serviceId.value === '13') {
    return '提交'
  }
  // 培训、研学 -> 报名
  if (['6', '9'].includes(serviceId.value)) {
    return '参与报名'
  }
  return '提交申请'
})

// 判断服务是否可申请 (一期开放：1-4, 6, 9 以及 13 号)
const isServiceAvailable = computed(() => ['1', '2', '3', '4', '6', '9', '13'].includes(serviceId.value))

// 表单数据
const formData = ref({
  contactName: '',
  contactPhone: '',
  // 培训服务
  traineeName: '',
  traineePhone: '',
  traineeGender: 'male',
  traineeBirthday: '',
  traineeIdCard: '',
  examModel: '',
  licenseLevel: '',
  hasExperience: 'no',
  // 物流服务 - 完整字段
  customerType: 'personal', // 客户类型：personal/enterprise
  companyName: '', // 企业名称
  cargoType: '', // 货物类型
  cargoTypeOther: '', // 其他货物类型
  cargoWeight: '', // 货物重量
  cargoVolume: '', // 货物体积
  startAddress: '', // 起运地
  startAddressDetail: '', // 起运地详细地址
  endAddress: '', // 目的地
  endAddressDetail: '', // 目的地详细地址
  deliveryUrgency: '', // 运输时效
  expectedTime: '', // 期望运输时间
  fileList: [], // 上传的文件列表
  remark: '', // 备注说明
  // 巡检
  inspectionType: '',
  inspectionArea: '',
  inspectionDate: '',
  // 托管
  droneModel: '',
  droneCount: '',
  trusteeDuration: '',
  // 吊运
  liftItemType: '',
  liftItemWeight: '',
  workLocation: '',
  liftHeight: '',
  // 研学报名（服务 9）
  studyOrg: '', // 学校/机构
  studyGrade: '', // 年级/年龄段
  studyParticipants: '', // 人数
  studyDate: '', // 期望日期
  studySession: '', // am/pm
  studySessionText: '', // 上午/下午
  // 维修
  maintenanceType: 'repair',
  isWarranty: 'yes',
  purchaseDate: '',
  // 赛事报名
  competitionRole: '',
  competitionRoleText: '',
  competitionContent: '',
  regNo: '',
  companyName: '',
  companyShortName: '',
  location: '',
  name: '',
  gender: 'male',
  idCard: '',
  competitionGroup: '',
  phone: '',
  email: '',
  level: '',
  validDate: '',
  athleteGroup: '',
  competitionProject: '',
  logoFileList: [],
  manager: '',
  managerPhone: '',
  contactPerson: '',
  contactPhone: '',
  // 通用
  remark: ''
})

// 选择器状态
const showDatePicker = ref(false)
const showDurationPicker = ref(false)
const showCargoTypePicker = ref(false) // 货物类型选择器
const showUrgencyPicker = ref(false) // 运输时效选择器
const showTimePicker = ref(false) // 期望时间选择器
const showCompetitionRolePicker = ref(false)
const showGroupPicker = ref(false)
const showAthleteGroupPicker = ref(false)
const showProjectPicker = ref(false)
const selectedDateArray = ref([
  String(new Date().getFullYear()),
  String(new Date().getMonth() + 1).padStart(2, '0'),
  String(new Date().getDate()).padStart(2, '0')
])
const expectedDate = ref([
  String(new Date().getFullYear()),
  String(new Date().getMonth() + 1).padStart(2, '0'),
  String(new Date().getDate()).padStart(2, '0')
])
const showExamModelPicker = ref(false)
const showLicenseLevelPicker = ref(false)
const showExperiencePicker = ref(false)
const showBirthdayPicker = ref(false)
// 研学选择器
const showStudySessionPicker = ref(false)
const showStudyDatePicker = ref(false)
const studyDate = ref([
  String(new Date().getFullYear()),
  String(new Date().getMonth() + 1).padStart(2, '0'),
  String(new Date().getDate()).padStart(2, '0')
])
const studySessionOptions = [
  { text: '上午（08:50-11:40）', value: 'am' },
  { text: '下午（13:50-16:40）', value: 'pm' }
]
const currentBirthday = ref(new Date(1990, 0, 1))
const minDate = new Date(1950, 0, 1)
const maxDate = new Date()

// 赛事报名 - 选项数据
const competitionRoleOptions = [
  { text: '裁判员', value: 'referee' },
  { text: '教练员', value: 'coach' },
  { text: '运动员', value: 'athlete' },
  { text: '俱乐部', value: 'club' }
]

const groupOptions = [
  { text: '足球', value: '足球' },
  { text: 'FPV', value: 'FPV' },
  { text: '航拍', value: '航拍' }
]

const athleteGroupOptions = [
  { text: '成人', value: '成人' },
  { text: '中学', value: '中学' },
  { text: '小学', value: '小学' }
]

const projectOptions = [
  { text: '足球', value: '足球' },
  { text: 'FPV', value: 'FPV' },
  { text: '航拍', value: '航拍' }
]

const onCompetitionRoleConfirm = ({ selectedOptions }) => {
  formData.value.competitionRole = selectedOptions[0].value
  formData.value.competitionRoleText = selectedOptions[0].text
  showCompetitionRolePicker.value = false
}

const onGroupConfirm = ({ selectedOptions }) => {
  formData.value.competitionGroup = selectedOptions[0].text
  showGroupPicker.value = false
}

const onAthleteGroupConfirm = ({ selectedOptions }) => {
  formData.value.athleteGroup = selectedOptions[0].text
  showAthleteGroupPicker.value = false
}

const onProjectConfirm = ({ selectedOptions }) => {
  formData.value.competitionProject = selectedOptions[0].text
  showProjectPicker.value = false
}

// 培训服务 - 选项数据
const examModelOptions = [
  { text: '小型无人机', value: '小型' },
  { text: '中型无人机', value: '中型' }
]

const licenseLevelOptions = [
  { text: '视距内', value: '视距内' },
  { text: '超视距', value: '超视距' }
]

const experienceOptions = [
  { text: '无', value: 'no' },
  { text: '有', value: 'yes' }
]

const onBirthdayConfirm = (value) => {
  const date = new Date(value)
  formData.value.traineeBirthday = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  showBirthdayPicker.value = false
}

const onExamModelConfirm = ({ selectedOptions }) => {
  formData.value.examModel = selectedOptions[0].text
  showExamModelPicker.value = false
}

const onLicenseLevelConfirm = ({ selectedOptions }) => {
  formData.value.licenseLevel = selectedOptions[0].text
  showLicenseLevelPicker.value = false
}

const onExperienceConfirm = ({ selectedOptions }) => {
  formData.value.hasExperience = selectedOptions[0].value
  formData.value.hasExperienceText = selectedOptions[0].text
  showExperiencePicker.value = false
}

// 物流服务 - 货物类型选项
const cargoTypeOptions = [
  { text: '生鲜食品', value: '生鲜食品' },
  { text: '应急药品', value: '应急药品' },
  { text: '工业零部件', value: '工业零部件' },
  { text: '电子产品', value: '电子产品' },
  { text: '文件资料', value: '文件资料' },
  { text: '日用品', value: '日用品' },
  { text: '医疗器械', value: '医疗器械' },
  { text: '其他', value: '其他' }
]

// 物流服务 - 运输时效选项
const urgencyOptions = [
  { text: '加急配送（2小时内）', value: '加急' },
  { text: '标准配送（当日达）', value: '标准' },
  { text: '普通配送（次日达）', value: '普通' },
  { text: '经济配送（3日内）', value: '经济' }
]

const durationOptions = [
  { text: '1个月', value: '1个月' },
  { text: '3个月', value: '3个月' },
  { text: '6个月', value: '6个月' },
  { text: '1年', value: '1年' },
  { text: '长期托管', value: '长期托管' }
]

const onDateConfirm = ({ selectedValues }) => {
  const [year, month, day] = selectedValues
  const formatted = `${year}-${month}-${day}`
  
  if (serviceId.value === '12') {
    formData.value.purchaseDate = formatted
  } else {
    formData.value.inspectionDate = formatted
  }
  showDatePicker.value = false
}

const onDurationConfirm = ({ selectedOptions }) => {
  formData.value.trusteeDuration = selectedOptions[0].text
  showDurationPicker.value = false
}

// 物流服务 - 货物类型确认
const onCargoTypeConfirm = ({ selectedOptions }) => {
  formData.value.cargoType = selectedOptions[0].text
  showCargoTypePicker.value = false
}

// 物流服务 - 运输时效确认
const onUrgencyConfirm = ({ selectedOptions }) => {
  formData.value.deliveryUrgency = selectedOptions[0].text
  showUrgencyPicker.value = false
}

// 物流服务 - 期望时间确认
const onTimeConfirm = ({ selectedValues }) => {
  const [year, month, day] = selectedValues
  formData.value.expectedTime = `${year}-${month}-${day}`
  showTimePicker.value = false
}

// 研学 - 日期确认
const onStudyDateConfirm = ({ selectedValues }) => {
  const [year, month, day] = selectedValues
  formData.value.studyDate = `${year}-${month}-${day}`
  showStudyDatePicker.value = false
}

// 研学 - 场次确认
const onStudySessionConfirm = ({ selectedOptions }) => {
  formData.value.studySession = selectedOptions[0].value
  formData.value.studySessionText = selectedOptions[0].text
  showStudySessionPicker.value = false
}
// 文件上传后处理
const afterRead = (file) => {
  console.log('上传文件：', file)
  showToast('上传成功')
}

const manualSubmit = () => {
    formRef.value.submit()
}

const onFailed = (errorInfo) => {
  console.log('failed', errorInfo);
  showFailToast('请填写必填项');
};

const formRef = ref(null)

// Phone validation: allow simple check or relax it
// pattern: /^1\d{10}$/

const onSubmit = async () => {
  // Check if user is logged in
  const userStr = localStorage.getItem('user');
  let user = null;
  
  if (userStr) {
    user = JSON.parse(userStr);
  }

  // 无人机赛事报名 (ID 13) 不需要登录即可提交
  if (!user && serviceId.value !== '13') {
      showDialog({
          title: '提示',
          message: '您当前未登录，请先登录或注册账号后再提交申请。',
          confirmButtonText: '去登录',
          showCancelButton: true
      }).then((action) => {
          if (action === 'confirm') {
              router.push('/login');
          }
      });
      return;
  }

  showLoadingToast({
    message: '提交中...',
    forbidClick: true,
    duration: 0 // 持续展示
  })
  console.log('onSubmit triggered');
  // 获取当前时间作为申请时间
  const now = new Date()
  const applyTime = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
  
  // 生成申请单号
  const orderNo = 'DK' + now.getTime()
  
  try {
    // 构造提交数据
    const submitData = {
      ...formData.value,
      serviceId: serviceId.value,
      serviceName: serviceName.value,
      orderNo,
      applyTime,
      status: '待处理',
      userId: user ? user.id : 'ANONYMOUS' // Add User ID or mark as Anonymous
    };

    // 无人机赛事报名自动生成注册号
    if (serviceId.value === '13') {
      // 假设 openid 存在于 user 对象中，如果不存在则使用 id 作为 fallback
      const openid = (user && (user.openid || user.id)) || 'GUEST';
      submitData.regNo = `REG-${openid.substring(0, 8).toUpperCase()}-${now.getTime().toString().slice(-6)}`;
    }

    const res = await axios.post('/api/submit', submitData);

    closeToast();
    
    let successMessage = `
申请单号：${orderNo}
申请时间：${applyTime}
处理状态：待处理
`;

    if (serviceId.value === '13') {
      successMessage += `注册号：${submitData.regNo}\n`;
    }

    successMessage += `━━━━━━━━━━━━━━━━
📞 客服联系方式：
电话：0577-55558188

我们将在1个工作日内与您联系！`;

    showDialog({
      title: '✅ 提交成功',
      className: 'submit-success-dialog',
      message: successMessage,
      confirmButtonText: serviceId.value === '13' ? '确认' : '查看我的申请'
    }).then(() => {
      if (serviceId.value === '13') {
        router.push('/')
      } else {
        router.push('/applications')
      }
    })
  } catch (error) {
    closeToast();
    showFailToast('提交失败，请重试');
    console.error(error);
  }
}
</script>

<style scoped>
.form-section {
  background: #fff;
  border-radius: 12px;
  padding: 4px 0;
  margin-bottom: 12px;
}

.form-title {
  font-size: 15px;
  font-weight: 500;
  color: #323233;
  padding: 12px 16px;
  border-left: 3px solid #667eea;
  margin-left: 16px;
}

.empty-tips {
  padding: 20px;
  text-align: center;
}

.content-wrapper {
  padding: 16px;
  min-height: calc(100vh - 46px);
}
</style>

