<template>
  <view class="doc-page">
    <u-nav-bar :title="title" show-back @back="goBack" />

    <view class="doc-card">
      <view v-for="(sec, i) in sections" :key="i" class="doc-section">
        <text class="doc-title">{{ sec.title }}</text>
        <text v-for="(p, j) in sec.paras" :key="j" class="doc-para">{{ p }}</text>
      </view>
    </view>

    <view class="doc-foot">生效日期：2026年8月10日</view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const type = ref('terms')

const title = computed(() => (type.value === 'privacy' ? '隐私政策' : '用户协议'))

const TERMS = [
  {
    title: '一、总则',
    paras: [
      '本协议是您与无人机产业综合服务平台（以下简称"本平台"）之间就使用本平台服务所订立的协议。您注册、登录或使用本平台，即视为已阅读并同意本协议全部条款。',
    ],
  },
  {
    title: '二、账号注册与安全',
    paras: [
      '您应保证注册信息真实、准确、有效，并妥善保管账号与密码。因您保管不善造成的损失，由您自行承担。',
      '本平台账号不可转让、出借或出租。如发现账号异常使用，本平台有权采取限制措施。',
    ],
  },
  {
    title: '三、用户行为规范',
    paras: [
      '您不得利用本平台从事任何违法违规活动，包括但不限于发布虚假需求、侵犯他人知识产权、泄露国家秘密等。',
      '本平台上的供需对接、交易行为应符合相关法律法规，双方应诚信履约。',
    ],
  },
  {
    title: '四、服务变更与终止',
    paras: [
      '本平台可能因业务调整、技术升级等原因变更或暂停服务，将尽合理努力提前通知。',
      '如您违反本协议或相关法律法规，本平台有权暂停或终止向您提供服务。',
    ],
  },
  {
    title: '五、免责声明',
    paras: [
      '本平台为信息展示与对接的第三方平台，不对用户之间达成的交易结果承担担保责任。',
      '因不可抗力、网络故障、第三方原因导致的服务中断，本平台不承担相应责任。',
    ],
  },
  {
    title: '六、争议解决',
    paras: [
      '本协议的订立、执行与解释均适用中华人民共和国法律。因本协议产生的争议，双方应友好协商解决；协商不成的，可向平台运营方所在地人民法院提起诉讼。',
    ],
  },
]

const PRIVACY = [
  {
    title: '一、我们收集的信息',
    paras: [
      '注册与登录信息：您注册时提供的手机号、密码，以及使用微信登录时的微信授权信息（昵称、头像、OpenID）。',
      '您主动提交的信息：需求、招聘、认证等业务表单中填写的内容，以及上传的图片、文件等资料。',
      '使用信息：包括浏览记录、操作日志、设备信息（操作系统、网络环境）等，用于保障服务安全和优化体验。',
    ],
  },
  {
    title: '二、信息的使用',
    paras: [
      '用于向您提供平台服务、完成认证与审核、匹配供需资源，以及与您沟通服务相关事项。',
      '在法律法规允许的范围内，可能对脱敏后的数据进行统计分析，用于改进产品与服务。',
    ],
  },
  {
    title: '三、信息的存储与保护',
    paras: [
      '您的信息存储于境内服务器，我们将采取加密存储、访问控制等技术措施保护您的信息安全。',
      '未经您同意，我们不会向第三方出售您的个人信息。',
    ],
  },
  {
    title: '四、信息共享与披露',
    paras: [
      '仅在以下情形下共享您的信息：经您明确同意；法律法规要求；为完成业务对接，在必要范围内向对接方展示（如您的企业名称、联系方式）。',
    ],
  },
  {
    title: '五、您的权利',
    paras: [
      '您有权查询、更正、删除您的个人信息，有权注销账号。可通过"我的-个人信息"或联系平台客服行使上述权利。',
      '注销账号后，我们将依法删除或匿名化处理您的个人信息（法律法规要求保留的除外）。',
    ],
  },
  {
    title: '六、未成年人保护',
    paras: [
      '本平台主要面向成年人提供服务。若您为未成年人，请在监护人陪同下使用，并征得监护人同意。',
    ],
  },
  {
    title: '七、政策更新',
    paras: [
      '我们可能适时修订本政策，修订后将在平台内公布。重大变更将另行通知。持续使用本平台即视为您接受修订后的政策。',
    ],
  },
]

const sections = computed(() => (type.value === 'privacy' ? PRIVACY : TERMS))

const goBack = () => uni.navigateBack()

onLoad((opts) => {
  if (opts?.type === 'privacy') type.value = 'privacy'
})
</script>

<style scoped>
.doc-page {
  min-height: 100vh;
  background: #F4F6F8;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

.doc-card {
  background: #fff;
  border: 1rpx solid #EEF1F4;
  border-radius: 16rpx;
  box-shadow: 0 8rpx 32rpx rgba(16, 24, 40, 0.06);
  margin: 24rpx;
  padding: 32rpx 28rpx;
}

.doc-section { margin-bottom: 32rpx; }
.doc-section:last-child { margin-bottom: 0; }

.doc-title {
  display: block;
  font-size: 30rpx;
  font-weight: 700;
  color: #17212B;
  margin-bottom: 16rpx;
}

.doc-para {
  display: block;
  font-size: 26rpx;
  line-height: 1.8;
  color: #475467;
  margin-bottom: 12rpx;
}
.doc-para:last-child { margin-bottom: 0; }

.doc-foot {
  text-align: center;
  font-size: 22rpx;
  color: #98A2B3;
}
</style>
