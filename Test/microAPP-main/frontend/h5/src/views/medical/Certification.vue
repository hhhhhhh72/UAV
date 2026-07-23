<template>
  <div class="certification-page">
    <van-nav-bar title="寄件人认证" left-arrow @click-left="$router.back()" fixed placeholder />

    <div class="form-container">
      <van-notice-bar
        left-icon="info-o"
        text="医疗配送需身份登记后方可下单，请如实填写信息"
      />

      <van-form @submit="onSubmit" ref="formRef">
        <van-cell-group inset title="基本信息">
          <van-field
            v-model="form.real_name"
            label="真实姓名"
            placeholder="请输入真实姓名"
            :rules="[{ required: true, message: '请输入真实姓名' }]"
          />
          <van-field
            v-model="form.phone"
            label="手机号"
            placeholder="请输入手机号"
            type="tel"
            :rules="[{ required: true, message: '请输入手机号' }, { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' }]"
          />
        </van-cell-group>

        <van-cell-group inset title="机构信息">
          <van-field
            v-model="orgTypeText"
            is-link
            readonly
            label="机构类型"
            placeholder="请选择机构类型"
            @click="showOrgTypePicker = true"
            :rules="[{ required: true, message: '请选择机构类型' }]"
          />
          <van-field
            v-model="form.org_name"
            label="机构名称"
            placeholder="请输入机构全称"
            :rules="[{ required: true, message: '请输入机构名称' }]"
          />
          <van-field
            v-model="form.org_address"
            label="机构地址"
            placeholder="请输入机构地址（选填）"
          />
          <van-field
            v-model="form.position"
            label="职务"
            placeholder="如：医生、护士、检验师（选填）"
          />
        </van-cell-group>

        <van-cell-group inset title="授权协议">
          <div class="agreement-section">
            <van-checkbox v-model="agreed" shape="square">
              我已阅读并同意
              <span class="link" @click.stop="showAgreement = true">《医疗配送授权协议》</span>
            </van-checkbox>
          </div>
        </van-cell-group>

        <div class="submit-section">
          <van-button
            type="primary"
            block
            round
            native-type="submit"
            :loading="submitting"
            :disabled="!agreed"
          >
            提交认证
          </van-button>
        </div>
      </van-form>
    </div>

    <!-- 机构类型选择器 -->
    <van-popup v-model:show="showOrgTypePicker" position="bottom" round>
      <van-picker
        :columns="orgTypeOptions"
        @confirm="onOrgTypeConfirm"
        @cancel="showOrgTypePicker = false"
      />
    </van-popup>

    <!-- 协议弹窗 -->
    <van-dialog v-model:show="showAgreement" title="医疗配送授权协议" confirm-button-text="我已了解">
      <div class="agreement-content">
        <p>1. 寄件人保证所寄送物品符合航空运输安全规范</p>
        <p>2. 禁止寄送违禁品、危险品、非法物品</p>
        <p>3. 同意平台对物品进行必要安全检查</p>
        <p>4. 理解无人机配送受天气、空域等客观条件限制</p>
        <p>5. 同意平台在不可抗力情况下的免责条款</p>
        <p>6. 授权平台在配送过程中使用寄件人提供的联系信息</p>
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showFailToast } from 'vant'
import { useMedicalStore } from '@/stores/medical'

const router = useRouter()
const store = useMedicalStore()

const form = reactive({
  real_name: '',
  phone: '',
  org_type: '',
  org_name: '',
  org_address: '',
  position: ''
})

const agreed = ref(false)
const submitting = ref(false)
const showOrgTypePicker = ref(false)
const showAgreement = ref(false)
const orgTypeText = ref('')

const orgTypeOptions = [
  { text: '医院', value: 'hospital' },
  { text: '检验中心', value: 'lab' },
  { text: '药房', value: 'pharmacy' },
  { text: '其他', value: 'other' }
]

function onOrgTypeConfirm({ selectedOptions }) {
  form.org_type = selectedOptions[0].value
  orgTypeText.value = selectedOptions[0].text
  showOrgTypePicker.value = false
}

async function onSubmit() {
  if (!agreed.value) {
    showFailToast('请同意授权协议')
    return
  }

  submitting.value = true
  try {
    const data = { ...form }
    const res = await store.submitCertification(data)
    if (res?.success) {
      showSuccessToast('提交成功')
      router.replace('/medical/certification/status')
    } else {
      showFailToast(res?.message || '提交失败')
    }
  } catch (e) {
    showFailToast(e.response?.data?.message || '提交失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.certification-page {
  min-height: 100vh;
  background: #f5f5f5;
  max-width: var(--page-max-width);
  margin: 0 auto;
}

/* 固定导航栏居中约束 */
.certification-page :deep(.van-nav-bar--fixed) {
  left: 50% !important;
  transform: translateX(-50%) !important;
  width: 100% !important;
  max-width: var(--page-max-width);
}

.form-container {
  padding-bottom: 80px;
}
.agreement-section {
  padding: 16px;
}
.agreement-content {
  padding: 16px;
  font-size: 13px;
  line-height: 2;
  color: #666;
}
.link {
  color: #1989fa;
}
.submit-section {
  padding: 24px 16px;
}
</style>
