<!-- Generated: 2026-08-06 | Files scanned: 160 | Token estimate: ~900 -->

# 前端架构地图（管理后台 + 小程序）

## A. Web 管理后台 frontend/（Vue3 + @arco-design/web-vue ^2.58 + ECharts 6）

⚠️ 2026-08 已从 Element Plus 全面切换 Arco（README/CLAUDE.md 已同步）。

### 目录
```
src/
├── main.js            Pinia(无store) + Router + ArcoVue + http 拦截器挂载
├── router/index.js    39 条 /admin 子路由 + /login，全量懒加载
├── utils/http.js      axios：Token注入、{data}解包、401单飞刷新+排队
├── utils/feedback.js  Arco Message/Modal 封装（兼容旧 El 签名）
├── hooks/useListRequest.js  列表 Hook：分页/搜索/排序/批量
├── api/admin/common.js      useAdminApi(resource) CRUD 工厂
└── views/admin/
    ├── AdminLayout.vue      顶栏+侧边9菜单(按角色过滤)+暗色主题
    ├── Dashboard.vue        看板（4 图，/api/v1/admin/dashboard）
    ├── components/CrudList.vue  核心配置化列表组件
    ├── config/ServiceConfigList.vue + ImageCropper.vue
    ├── consolidated/        8 聚合页（a-tabs 嵌套子列表）
    └── 31 个业务模块目录
```

### CrudList 配置驱动模型
页面只需声明 `columns` + `searchFields` + `batchActions`，内部自动：
`resource → useAdminApi → GET 列表 → a-table + 分页 + 搜索`
- 日期范围合并为 start_date/end_date；批量动作逐个调 api(row) 并**传完整行数据**（后端 PUT 全字段覆盖）
- 支持 apiFunction 覆盖（CaseList 走 JSON 文件后端 /api/cases）

### 与后端对接约定
- 响应信封 `{data:...}` 透明解包（分页保留 total）；baseURL 相对路径，Vite 代理 /api → localhost:8080
- 登录：`POST /api/auth/login`（camelCase token）；开发自动登录：`POST /api/v1/admin/token {role}`
- 401 流程：`POST /api/auth/refresh` 单飞刷新 → 排队重放；失败清登录态跳 /login
- 管理端 CRUD：`/api/v1/admin/{resource}`，PUT 全字段更新语义

## B. 微信小程序 miniprogram/（uni-app + Vue3，78 页）

### 结构
```
pages.json  5 Tab 注册（原生 tabBar 被隐藏）
App.vue     onLaunch → 微信静默登录链路
components/
├── ui/     14 个自研 u- 组件（easycom 免注册: ^u-(.*)→components/ui/u-$1.vue）
│           u-button/cell/field/empty/icon/loading/nav-bar/picker/popup/
│           search/sticky/tabs/tag + u-tabbar(在 TabBar.vue)
└── Layout.vue(骨架+TabBar+悬浮按钮) TabBar.vue(自研5Tab,毛玻璃)
    HomeFloatButton.vue StateView.vue(四态)
utils/
├── request.js   BASE_URL+unwrap{data,total}+401刷新队列；authStorage
├── config.js    后端地址唯一配置点（现 http://192.168.5.141:8080）
├── enums.js     业务术语统一标签（对齐 Go 端 biz_standard.go）
└── nav.js       防连点导航封装
```

### 页面分组（78 页）
```
Tab: home/demands/publish/services/mine
供需: demands/{list,detail,publish,mine} intents/mine tasks/{index,detail}
商城: mall/{index,detail} shops/index portfolios/list
人才: jobs/{list,detail,mine,applications,applicants,resume} pilots/{list,apply}
培训: training/{courses,enroll,register,certificates} study/{index,detail}
创新: achievements/{list,detail} challenges projects colleges/{list,detail}
合规: compliance/{news,knowledge,standards}
赛事: competitions/{list,detail,register} exhibitions/{list,booth}
应急: emergency/{resources,cases,depts,dispatches}
资源: resources/{list,detail} experts/{list,detail} reports cases/{index,detail}
其他: search match messages applications/{index,submit} more
      testsites/{list,detail,booking,pay,result} login register mine/{auth,profile}
      webview services/{index,detail,apply}
```

### 设计规范（App.vue 全局 CSS 变量）
- 品牌 #0A66C2 / 深蓝 #074D92 / 青绿 #1DD4A8（辅）/ 强调橙 #F97316
- 输入框 104rpx 高 #fafafa 底 16rpx 圆角；按钮 100rpx 高 50rpx 圆角 + 蓝投影
- 卡片 `1px solid #EEF1F4` + 8rpx 圆角；金额分→元（budget_fen/100），0 显示「面议」
- 列表页标配 u-sticky(u-search+u-tabs) + StateView 四态

### 认证
- 主链路：App.vue → wx.login() → POST /api/v1/auth/wechat/login（snake_case token）
- 备用：login 页账号密码 POST /api/auth/login（camelCase token，⚠两端格式不一致）
- 页内拦截（非全局守卫）：每个跳转前查 user，无则 goLogin()
