<script>
	export default {
		onLaunch: function() {
			this.tryWxSilentLogin()
		},
		methods: {
			tryWxSilentLogin() {
				// #ifdef MP-WEIXIN
				const token = uni.getStorageSync('accessToken')
				if (token) {
					// 有 token 但 user 缺失（历史遗留/被清）：用 me 恢复用户信息
					const { request } = require('./utils/request')
					request({ url: '/api/auth/me' }).then(meRes => {
						if (meRes?.user) {
							uni.setStorageSync('user', JSON.stringify(meRes.user))
						}
					}).catch(() => { /* token 失效时由请求层处理 */ })
					return
				}
				this.silentLogin()
				// #endif
			},
			silentLogin() {
				const { request, authStorage } = require('./utils/request')
				uni.login({
					provider: 'weixin',
					success: (loginRes) => {
						request({
							url: '/api/v1/auth/wechat/login',
							method: 'POST',
							data: { code: loginRes.code }
						}).then(res => {
							// Backend returns snake_case: { access_token, refresh_token, expires_in, user }
							if (res?.access_token && res?.user) {
								uni.setStorageSync('user', JSON.stringify(res.user))
								authStorage.setTokens(res.access_token, res.refresh_token)
							}
						}).catch(() => {})
					}
				})
			}
		}
	}
</script>

<style>
	/* ===== 全局 CSS 变量 & 工具类 ===== */

	/* --- 语义化颜色 --- */
	page {
		--color-primary: #0A66C2;
		--color-primary-deep: #074D92;   /* 深空蓝：顶部导航等重色场景 */
		--color-primary-light: #E8F2FC;
		--color-accent: #F97316;         /* 强调橙：类型徽章/价格 */
		--color-accent-deep: #E96012;
		--color-accent-light: #FFF0E6;
		--color-success: #34c759;
		--color-warning: #ff9f0a;
		--color-danger: #ff3b30;
		--color-info: #909399;

		--color-bg: #f5f6f8;
		--color-bg-card: #ffffff;
		--color-text: #1a1a1a;
		--color-text-secondary: #969799;
		--color-text-placeholder: #c8c9cc;
		--color-border: #f0f1f3;
		--color-divider: #ebedf0;

		/* --- 圆角 --- */
		--radius-sm: 8rpx;
		--radius-md: 16rpx;
		--radius-lg: 24rpx;
		--radius-round: 999rpx;

		/* --- 阴影 --- */
		--shadow-sm: 0 2rpx 8rpx rgba(0,0,0,0.03);
		--shadow-md: 0 4rpx 16rpx rgba(0,0,0,0.06);
		--shadow-lg: 0 8rpx 32rpx rgba(0,0,0,0.08);

		/* --- 间距 --- */
		--space-xs: 8rpx;
		--space-sm: 16rpx;
		--space-md: 24rpx;
		--space-lg: 32rpx;
		--space-xl: 48rpx;

		/* --- 字体 --- */
		--font-xs: 20rpx;
		--font-sm: 24rpx;
		--font-md: 28rpx;
		--font-lg: 32rpx;
		--font-xl: 36rpx;
		--font-xxl: 40rpx;

		/* --- 布局 --- */
		--tabbar-height: 50px;
		--safe-bottom: env(safe-area-inset-bottom);
		--safe-top: env(safe-area-inset-top);

		/* 组件库扩展令牌 */
		--ui-radius-card: 24rpx;      /* 卡片大圆角 */
		--ui-radius-btn: 50rpx;       /* 按钮圆角 */
		--ui-shadow-card: 0 4rpx 16rpx rgba(0,0,0,0.06);  /* 卡片轻阴影 */
		--ui-color-accent-light: #E6FAF5;  /* 青绿浅底 */
		--ui-color-disabled: #c8c9cc;
		--ui-color-text-secondary: #969799;
		--ui-font-size-lg: 34rpx;
		--ui-font-size-md: 30rpx;
		--ui-font-size-sm: 26rpx;
		--ui-space-card: 24rpx;      /* 卡片内边距 */

		font-size: 28rpx;
		color: var(--color-text);
		background-color: var(--color-bg);
	}

	/* ===== 工具类 ===== */

	/* --- Flex 布局 --- */
	.flex { display: flex; }
	.flex-col { display: flex; flex-direction: column; }
	.flex-center { display: flex; align-items: center; justify-content: center; }
	.flex-between { display: flex; align-items: center; justify-content: space-between; }
	.flex-start { display: flex; align-items: center; justify-content: flex-start; }
	.flex-end { display: flex; align-items: center; justify-content: flex-end; }
	.flex-wrap { flex-wrap: wrap; }
	.flex-1 { flex: 1; min-width: 0; }
	.flex-shrink { flex-shrink: 0; }

	/* --- 文本 --- */
	.text-ellipsis { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.text-ellipsis-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.text-center { text-align: center; }
	.text-left { text-align: left; }
	.text-right { text-align: right; }

	/* --- 颜色 --- */
	.text-primary { color: var(--color-primary); }
	.text-secondary { color: var(--color-text-secondary); }
	.text-success { color: var(--color-success); }
	.text-warning { color: var(--color-warning); }
	.text-danger { color: var(--color-danger); }

	/* --- 背景 --- */
	.bg-white { background-color: #fff; }
	.bg-card { background-color: var(--color-bg-card); }
	.bg-primary { background-color: var(--color-primary); }
	.bg-primary-light { background-color: var(--color-primary-light); }

	/* --- 圆角 --- */
	.radius-sm { border-radius: var(--radius-sm); }
	.radius-md { border-radius: var(--radius-md); }
	.radius-lg { border-radius: var(--radius-lg); }
	.radius-round { border-radius: var(--radius-round); }
	.radius-card { border-radius: var(--radius-md); }

	/* --- 阴影 --- */
	.shadow-sm { box-shadow: var(--shadow-sm); }
	.shadow-md { box-shadow: var(--shadow-md); }

	/* --- 间距 --- */
	.p-xs { padding: var(--space-xs); }
	.p-sm { padding: var(--space-sm); }
	.p-md { padding: var(--space-md); }
	.p-lg { padding: var(--space-lg); }
	.px-sm { padding-left: var(--space-sm); padding-right: var(--space-sm); }
	.px-md { padding-left: var(--space-md); padding-right: var(--space-md); }
	.py-sm { padding-top: var(--space-sm); padding-bottom: var(--space-sm); }
	.py-md { padding-top: var(--space-md); padding-bottom: var(--space-md); }
	.mt-sm { margin-top: var(--space-sm); }
	.mt-md { margin-top: var(--space-md); }
	.mb-sm { margin-bottom: var(--space-sm); }
	.mb-md { margin-bottom: var(--space-md); }
	.gap-xs { gap: var(--space-xs); }
	.gap-sm { gap: var(--space-sm); }
	.gap-md { gap: var(--space-md); }

	/* --- 卡片容器 --- */
	.card {
		background-color: var(--color-bg-card);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-sm);
		overflow: hidden;
	}
</style>
