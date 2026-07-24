<script>
	export default {
		onLaunch: function() {
			console.log('App Launch')
			this.tryWxSilentLogin()
		},
		onShow: function() {
			console.log('App Show')
		},
		onHide: function() {
			console.log('App Hide')
		},
		methods: {
			tryWxSilentLogin() {
				// #ifdef MP-WEIXIN
				const token = uni.getStorageSync('accessToken')
				if (token) return

				uni.login({
					provider: 'weixin',
					success: (loginRes) => {
						const { request, authStorage } = require('./utils/request')
						request({
							url: '/api/auth/wx-login',
							method: 'POST',
							data: { code: loginRes.code }
						}).then(res => {
							if (res?.success && !res.isNewUser) {
								uni.setStorageSync('user', JSON.stringify(res.user))
								authStorage.setTokens(res.accessToken, res.refreshToken)
								console.log('微信静默登录成功')
							}
						}).catch(() => {})
					}
				})
				// #endif
			}
		}
	}
</script>

<style>
	/*每个页面公共css */
</style>
