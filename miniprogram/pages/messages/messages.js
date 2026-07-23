const app = getApp()
Page({
  data: { messages: [] },
  onShow() {
    app.get('/api/v1/messages').then(res => {
      this.setData({ messages: res.data || [] })
    }).catch(() => {})
  }
})
