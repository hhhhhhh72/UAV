Page({ data: { keyword: '' }, onInput(e) { this.setData({ keyword: e.detail.value }) }, onSearch() { wx.showToast({ title: '搜索: ' + this.data.keyword, icon: 'none' }) } })
