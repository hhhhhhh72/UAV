---
name: miniprogram-structure
description: 微信小程序结构 — 页面、分包、API封装、认证流程、drone-platform-ui改版
metadata:
  type: reference
---

# 微信小程序

## 主小程序 (miniprogram/)

基于原生 + Vant Weapp 1.11，自定义 tabbar。

**TabBar 5栏**: 首页(pages/index) → 分类(pages/training/courses) → 卖机(pages/market) → 任务(pages/tasks/list) → 我的(pages/mine)

**主包页面 (14页)**:
- pages/index, search, demand/list, demand/detail, demand/publish
- community/list, community/detail
- tasks/list
- training/courses, training/certificates, training/pilots
- messages/list, mine/index, market/market

**分包 (7个)**:
- jobs(list/detail/resume), trading(products/listings/orders)
- finance(wallet/insurance/loans), enterprise(list/apply/detail)
- venues(list), reviews(list), news(list)

**全局组件**: 40+ Vant Weapp 组件注册在 app.json usingComponents

**API 封装** (`utils/api.js`, 170行):
- `api.get(url, params)` / `api.post(url, data)` / `api.patch(url, data)` / `api.upload(url, filePath, formData)`
- BASE_URL: http://localhost:8080
- 自动 Bearer Token 注入 + X-Request-ID 追踪
- Token 过期自动刷新(刷新锁防并发) + 队列重试
- 统一错误处理

**认证流程** (`utils/auth.js`):
- wxLogin() → wx.login获取code → POST /auth/wechat/login → 存token到globalData+Storage
- ensureLogin() 确保登录态有效
- logout() 撤销refresh token

**全局常量** (`utils/constants.js`):
- bizTypeMap(6类), certTypeMap(3类), productTypeMap, policyTypeMap
- fenToYuan, formatPrice, formatDate, timeAgo 工具函数

## UI 改版 (drone-platform-ui/)

另一个小程序版本，自定义 tabbar 风格，5个主页面：
- pages/index — 首页(600行JS, 9000行WXML)
- pages/category — 分类
- pages/market — 卖机(5600行WXML)
- pages/tasks — 任务
- pages/mine — 我的(5100行WXML)

## 全局样式 (app.wxss)

CSS 变量设计系统，主色调 #4a90d9，背景 #F5F6FA。
