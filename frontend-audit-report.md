# frontend/ 代码阅读审计报告（无人机产业综合服务平台 Web 管理后台）

> 范围：D:/w-yao/frontend/ 全部文件（排除 node_modules / dist）。
> 实读：57 个 .vue（13,356 行）+ src 下 15 个 .js（817 行）+ css / index.html / vite.config.js / package.json / nginx.conf / Dockerfile / .env / *.md / 启动服务.bat / public 资源，合计 118 个文件，逐个完整读取。
> 行号以 read 工具为准（ServiceConfigList.vue 含孤立 CR 字符，个别编辑器显示行号可能偏大约 100 行）。
> 所有涉及后端契约的结论，均已在 internal/httpapi 的路由注册与 handler 中交叉核对。

---

## 1. 文件清单

### 1.1 构建 / 配置 / 部署（11）

| 路径 | 行数 | 作用 |
|---|---|---|
| frontend/package.json | 27 | Vue3.4 + Arco2.58 + ECharts6 + pinia + vue-cropper；**无 lint / test / typecheck 脚本** |
| frontend/vite.config.js | 66 | base 为 /、@ 别名、Arco 按需自动导入、manualChunks(echarts/vendor)、/api 与 /uploads 代理、dev 安全响应头 |
| frontend/vite.config.js.timestamp-1785308041682-1eb63e27b24b.mjs | 55 | **Vite 依赖预构建临时产物被误提交**（死文件） |
| frontend/index.html | 43 | SPA 宿主；内联全局 onerror / unhandledrejection 红条 |
| frontend/.env.development | 1 | VITE_API_TARGET=http://localhost:8080 |
| frontend/.env.production | 1 | VITE_API_TARGET=（空，生产同源） |
| frontend/nginx.conf | 127 | 生产反代模板：443+证书、gzip、限流 zone、/api、/api/v1/image 缓存层、/uploads、SPA 回退 |
| frontend/Dockerfile | 21 | 仅 nginx:alpine 阶段（构建阶段被注释，要求本地先 npm run build） |
| frontend/启动服务.bat | 13 | 本地 npm run dev |
| frontend/public/8fb21d8a400679656053bfee9d9a746c.txt、QyhAYrdnQ3.txt | 1 / 1 | 两个来源不明的 32/40 位十六进制串，随构建发布到网站根目录 |
| frontend/public/{icons,images,static,video,logo.png} | — | 15 个 SVG 图标 + 6 张 SVG 插图 + 3 张 JPG + 5 个 mp4（约 38MB 视频入库） |

### 1.2 入口与基础设施（8）

| 路径 | 行数 | 作用 |
|---|---|---|
| src/main.js | 37 | createApp + Pinia + Router；全局 patch window.ResizeObserver（rAF 节流）；引入 arco.css / global.css / utils/http 副作用 |
| src/App.vue | 31 | 仅 router-view + fade 过渡 |
| src/router/index.js | 121 | 43 条子路由（全部懒加载）+ /login；beforeEach 守卫（token + role 白名单 + meta.roles 拦截） |
| src/utils/http.js | 257 | **全局 axios 单例**：内存+sessionStorage 令牌、Idempotency-Key 自动注入、data 信封解包、401 单飞刷新 + pendingQueue、错误信封解包、旧 localStorage 迁移 |
| src/utils/feedback.js | 41 | Arco Message/Modal 封装：showToast / showFailToast / showSuccessToast / showLoadingToast / closeToast / showConfirmDialog |
| src/utils/mask.js | 44 | maskPhone / maskIdCard / maskEmail / maskAddress / maskContact（脱敏统一实现） |
| src/styles/global.css | 113 | 全局 reset + 品牌 CSS 变量 + Arco primary 变量（必须 RGB 分量格式）+ 表头/卡片样式 |
| src/hooks/useListRequest.js | 163 | 通用列表 hook：分页/搜索/排序/选中/批量/loading + 请求序号竞态保护 |

### 1.3 API 模块（7，全部薄封装）

| 路径 | 行数 | 作用 |
|---|---|---|
| src/api/admin/common.js | 33 | useAdminApi(resource) 生成 /api/v1/admin/{resource} 的 list/get/create/update/delete |
| src/api/admin/demand.js | 29 | getDemandList / approveDemand / closeDemand / setOfflineAmount / deleteDemand / rejectDemand |
| src/api/admin/enterprise.js | 42 | getEnterpriseList / reviewEnterprise / batchReviewEnterprise / updateEnterprise（PATCH，非 admin 前缀） |
| src/api/admin/order.js | 15 | getOrderList / updateOrderStatus / reviewAftersale |
| src/api/admin/review.js | 14 | getReviewList / updateReviewStatus(approve 或 reject) / deleteReview |
| src/api/admin/competition.js | 14 | getCompetitionList / updateCompetition / deleteCompetition |
| src/api/admin/user.js | 20 | getUserList / updateUserRole / resetUserPassword / deleteUser |

### 1.4 视图（57 个 .vue）

公共组件与布局：

| 路径 | 行数 | 作用 |
|---|---|---|
| views/admin/AdminLayout.vue | 653 | 顶栏（搜索/通知/全屏/主题/头像）+ 侧栏 11 个菜单 + 内容区；消息轮询；暗色主题变量 |
| views/login/Index.vue | 298 | 手机号+密码登录（POST /api/auth/login）；DEV-only 开发者快捷登录 |
| views/admin/Dashboard.vue | 465 | 数据看板：4 KPI + 单指标趋势 + 模块柱图 + 需求类型雷达 + 状态饼图；range = 7d/30d/90d/12m |
| views/admin/components/CrudList.vue | 457 | **核心配置化列表**：columns/searchFields/batchActions + CSV 导出 + 批量删除 |
| views/admin/components/BizOverview.vue | 179 | 聚合页指标卡 + ECharts 折线/饼图，按 path 点号取值 |
| views/admin/components/DataToolbar.vue | 51 | 工具条插槽（filters / actions） |
| views/admin/components/MetricCard.vue | 71 | 纯展示指标卡 |
| views/admin/composables/useAuth.js | 49 | 从 localStorage.user 读角色；isPlatformAdmin / isAssociationAdmin / canManage + refreshCurrentUser |
| views/admin/composables/useMedia.js | 66 | normalizeMediaUrl（本地域名降级为相对路径）+ uploadFile(POST /api/v1/upload) |
| src/components/RichEditor.vue | 72 | contenteditable + document.execCommand 轻量富文本编辑器 |
| views/admin/config/ImageCropper.vue | 180 | vue-cropper 裁剪弹窗，confirm 回调输出 File |

9 个聚合页（consolidated/，指标卡 + a-tabs 嵌子列表）：MembersPage(37)、TradingPage(43)、ContentPage(37)、NewsPage(33)、TalentPage(53)、InnovationPage(43)、PromotionPage(41)、EmergencyPage(33)、SettingsPage(20)。

43 个业务列表页（括号为行数）：

- 会员生态：users/UserList(269)、enterprises/EnterpriseList(435)、experts/ExpertList(215)、resources/ResourceList(255)、messages/NotifyList(178)
- 供需对接：demands/DemandList(377)、orders/OrderList(246)、reviews/ReviewList(128)、products/ProductList(328)、serviceListings/ServiceListingList(274)
- 产学研：achievements/AchievementList(258)、challenges/ChallengeList(239)、projects/ProjectList(297)、testsites/TestSiteList(234)、bookings/BookingList(143)、transformations/TransList(194)
- 合规政策：compliance/ComplianceList(419，双 Tab 文档+标准)、cases/CaseList(356)、articles/ArticleList(358)
- 人才教育：training/CourseList(406)、training/CertList(218)、competition/CompetitionList(401)、jobs/JobList(240)、colleges/CollegeList(671)、study/StudyList(213)、pilots/PilotList(161)、enrollments/EnrollmentList(264)
- 活动品牌：events/EventList(261)、portfolios/PortfolioList(207)、exhibitions/ExhibitionList(252)、exhibitions/BoothApplicationList(137)、reports/ReportList(222)
- 应急协同：emergency/ResourceList(236)、emergency/DispatchList(226)
- 审计/配置：audit/AuditLogList(487)、workbench/ReviewWorkbench(260)、account/AccountSettings(143)、config/ServiceConfigList(1505)

### 1.5 文档（3，均为中文）

| 路径 | 行数 | 说明 |
|---|---|---|
| frontend/README.md | 48 | 技术栈 / 路由结构 / 核心抽象 / 对接约定（**路由数写 40，实际 43**） |
| frontend/PINIA_GUIDE.md | 264 | **描述 3 个并不存在的 store**（useUserStore / useApplicationStore / useServiceStore），指向不存在的 @/stores/* 与 /home 路由 |
| frontend/研学管理员权限修复.md | 153 | 描述已删除的 h5 旧前端（frontend/h5/...）与 Express 后端（backend/routes/admin.js），与当前代码无关 |

---

## 2. 核心数据结构 / 接口签名

### 2.1 http.js 的三件套（全站唯一网络出口）

    // 令牌（内存态 + sessionStorage，localStorage 仅作旧版迁移读取）
    let memAccessToken, memRefreshToken
    const ACCESS_TOKEN_KEY = "accessToken", REFRESH_TOKEN_KEY = "refreshToken"
    export const authStorage = { getAccessToken(), getRefreshToken(), setTokens(a, r), clearTokens }
    export function getAuthHeader()   // 调用时刻读取，供上传等非拦截器场景
    export default axios              // 单例，已装请求/响应拦截器

- 请求拦截器：注入 Authorization: Bearer；对 **POST/PATCH 且非 FormData** 自动注入 Idempotency-Key = "idem-" + djb2(url|JSON.stringify(body)) + "-" + 长度（http.js:46-53, 74-82）。
- 响应拦截器（成功）：data 信封透明解包；含 total 的分页响应**保留整包**（http.js:88-100）。
- 响应拦截器（失败）：把 resp.data.error.message 提升到 error.message；非 401 直接 reject；401 走单飞刷新（isRefreshing + pendingQueue），刷新失败仅当明确 401 才清登录态（http.js:101-199）。

### 2.2 useListRequest(options)（src/hooks/useListRequest.js:24）

入参 { apiFunction, idKey = "id", defaultParams = {}, defaultPageSize = 20 }；
返回 { listData, loading, total, selectedIds, filterParams, loadData, onSearchSubmit, onSortChange, onPageChange, onSelectChange, onBatchAction, resetParams }。

- 竞态保护：let seq = 0，响应回来若 currentSeq !== seq 则丢弃（48-64, 77-88）。
- onSortChange({ prop, order }) 写入 sort_field = prop；sort_order = order === "ascending" ? "asc" : order === "descending" ? "desc" : ""（102-106）。
- **onPageChange / onSelectChange / onBatchAction 三个导出无人使用**（CrudList 自己实现），死代码。

### 2.3 useAdminApi(resource)（src/api/admin/common.js:13）

    const base = "/api/v1/admin/" + resource
    list(params) -> GET base?params    get(id) -> GET base/id
    create(data) -> POST base          update(id, data) -> PUT base/id     delete(id) -> DELETE base/id

35 个资源名全部映射到已注册的 Go 路由（见第 5 节）。

### 2.4 CrudList props / emits / expose（components/CrudList.vue:131-155, 395）

props：resource（必填）、columns（必填）、searchFields（input | select | date | range）、rowKey="id"、selectable=true、creatable=false、batchDelete=true、batchActions=[]、defaultParams={}、apiFunction=null、size / scroll / expandable、showExport=true。
emits：add、sorter-change、loaded(rows, total)；expose：loadData、reload、onSortChange。
batchActions 项：{ key, label, status, confirm?, prompt?: { title, placeholder }, api: (row, promptValue) => Promise }。

### 2.5 角色模型（前端侧）

- 路由守卫白名单：[platform_admin, association_admin]（router/index.js:103）。
- meta.roles 仅 4 条路由声明：audit-logs(23)、users(25)、config(28)、settings(65) —— 全部 platform_admin。
- useAuth：isPlatformAdmin / isAssociationAdmin / canManage（两者取或）/ isSuperAdmin（isPlatformAdmin 的别名）。
- 页面内二次门控：ServiceConfigList 的 onServiceRowClick(952-958) 与分组 roles(822-826)；UserList 操作列 v-if="isPlatformAdmin"(41-45)。
- 后端对应门禁：/api/v1/admin/* 由 adminGate 拦（server.go:418）；dashboard 允许两种管理员（phase3.go:646）。

---

## 3. 关键业务逻辑与状态机

### 3.1 登录 / 令牌生命周期（http.js + login/Index.vue）

1. 密码登录 POST /api/auth/login（Index.vue:150）→ 后端返回 **camelCase** { success, user, accessToken, refreshToken }（h5_compat.go:828-833）。
2. saveSession 只把 { id, role } 写 localStorage（Index.vue:127-132）；令牌写 sessionStorage + 内存。
3. 401 → POST /api/auth/refresh { refresh_token }（蛇形，http.js:164）→ 新 access + refresh **都必须落库**（旧 refresh 已被服务端消费）。
4. 登出（AdminLayout.vue:439-449）：先 fire-and-forget POST /api/auth/logout { refresh_token } 吊销，再清本地并跳 /login。
5. 改密（AccountSettings.vue:117-132）：POST /api/v1/auth/password → clearTokens → 800ms 后跳登录。

### 3.2 需求（系统②）状态机 —— DemandList.vue:56-72

- pending：通过（POST /approve）/ 驳回（理由必填，POST /review { action: "reject" }）/ 详情。
- published：关闭（理由必填，POST /close）/ 登记金额（POST /amount）。
- completed：登记金额。
- cancelled / rejected：删除。
- 金额校验（264-270）：Number.isFinite 且 0 < n <= 99999999.99，再 Math.round(n * 100) 转分。
- 批量：批量通过 / 批量驳回（prompt 必填理由）/ 批量删除（注释说明只允许删已取消/已驳回，在架需求逐条失败）。统计条来自 GET /api/v1/admin/demands/stats（独立全量，不随分页）。

### 3.3 企业认证（系统①）—— EnterpriseList.vue

- 仅 submitted 状态出现通过/驳回（28-32）；驳回必填理由（383-387, 397-400），理由写 review_comment 回显给申请人。
- 编辑走 **PATCH /api/v1/enterprises/{id}**（非 admin 前缀，api/admin/enterprise.js:32）；已审核企业编辑后回到待审核（85-91 有警示条）。
- 详情默认脱敏（联系电话/信用代码/地址），勾选"显示完整"才展开（47-54, 230）。
- 批量仅"批量通过"（232-235）—— 驳回必须逐条填理由。

### 3.4 订单与售后（系统②）—— OrderList.vue:72-90

- 无售后：可手动改状态（pending/paid/shipped/aftersale/completed/cancelled）→ PUT /api/v1/admin/orders/{id} { status }。
- aftersale_status === "pending"：同意退款 / 驳回 → PUT /orders/{id}/aftersale { action }（order.js:13-15）。
- 售后已结案：只读，不给改状态（80-82），防重复退款。
- 统计条（155-165）**只基于当前页**：交易额 / 已完成 / 完成率都是本页口径，仅"订单总数"来自接口 total —— 标签写了"(本页)"，但完成率仍会被误读成全站指标。

### 3.5 报名 → 结业（系统⑤）—— EnrollmentList.vue:28-35, 181-199

- 编辑态状态选项**故意不含 approved**（88-95，注释解释 approved 既不能结业也不能回退会卡死）。
- 结业：POST /api/v1/enrollments/{id}/complete（释放托管学费 + 发证，enrolled/paid → completed，completed 幂等可重试补步骤）。

### 3.6 服务配置（系统⑥/⑧）—— ServiceConfigList.vue

- GET/POST /api/services/config（生产无条件注册，h5_compat.go:1466-1468）；整包替换语义：先取全量 allServiceConfigs，改一个 key 后整体 POST（1035-1037）。
- 首页配置保存前**剥离后端注入的运行时字段** baseUrl 与绝对域名图片 URL，并转相对路径（915-925, 938-948）。
- 研学服务（id=9）课程包 CRUD（tab 切换、增删包、裁剪上传）；培训服务（id=6）拆分字段 conditions/prices/features/companyIntro/licenseFunction 在保存时回填（1025-1033）。
- 商业化费率走另一接口 GET/POST /api/v1/admin/config（1276-1305），同样是整包替换；两套配置互不相干。

### 3.7 审核工作台 —— ReviewWorkbench.vue:57-95

7 个待办源，用 useAdminApi(resource).list({ status, page: 1, page_size: 1 }) 只读 total 计数；失败卡片显示"统计失败"并可点击重试。

### 3.8 状态机枚举全集（前端各自硬编码，未集中）

demand(pending/published/completed/cancelled/rejected)、enterprise(draft/submitted/supplement_required/approved/rejected)、order(pending/paid/shipped/aftersale/completed/cancelled)、aftersale(pending/approved/rejected)、product(pending/listed/sold/removed)、course(draft/pending/published/closed)、job(published/closed)、competition(pending/draft/enrolling/closed)、event(published/ongoing/ended/cancelled)、exhibition(draft/recruiting/underway/ended)、study(draft/active/closed)、college(active/inactive)、certificate(pending/approved/rejected)、pilot(pending/approved/rejected)、enrollment(pending/paid/enrolled/rejected/completed)、booking(pending/approved/rejected/completed)、transformation(active/completed/cancelled + stage lab/pilot/industrialized/listed)、achievement(stage lab/pilot/industrialization/launched)、rd_challenge(open/in_progress/closed/resolved/published)、research_project(active|planning|recruiting|ongoing|completed)、dispatch(dispatched/ongoing/completed/cancelled)、emergency_resource(available/in_use/maintenance)、article(draft/published)、case(pending/published/archived)、portfolio(draft|pending/published/rejected)、message(is_read)。

---

## 4. 层间调用关系（前端）

    views/**/*.vue --(props/emits/expose)--> components/CrudList.vue --> hooks/useListRequest.js --> api/admin/common.js(useAdminApi)
          |                                        |                                                   |
          |                                        +--> utils/feedback.js (toast / modal)              |
          +--> api/admin/{demand,enterprise,order,review,competition,user}.js ------------------------+
          +--> utils/http.js (axios 单例 + authStorage + getAuthHeader) <---------------------------+
          +--> composables/useAuth.js (读 localStorage.user) --> GET /api/v1/me
          +--> composables/useMedia.js (uploadFile -> POST /api/v1/upload)
          +--> 直接 axios.get/post("/api/v1/...")（约 60 处"裸调用"，绕过 api/ 目录）

    App.vue --> router-view --> AdminLayout.vue --> router-view --> 43 个子页面
    main.js --> createPinia()（**零 store**）+ router + axios 副作用注册

- 谁调用 CrudList：35 个列表页；谁被 CrudList 调用：useListRequest + useAdminApi + feedback + 各页自定义 apiFunction。
- 聚合页直接 import 子列表组件（MembersPage → UserList/EnterpriseList/ExpertList 等），因此**同一组件被两条路由复用**。
- **Pinia 被安装但没有任何 store**：main.js:33 调用 createPinia() + 全仓 0 处 defineStore（已 grep 验证），状态全部在组件局部 ref / localStorage。

---

## 5. 路由 / 页面 → 后端接口 全量清单

### 5.1 路由（router/index.js，43 条子路由 + /login，全部懒加载）

| # | 路径 | 组件 | meta.roles | 侧栏可见 |
|---|---|---|---|---|
| — | /login | login/Index.vue | — | — |
| — | / 与 /admin | redirect → /admin/dashboard | — | — |
| 1 | admin/dashboard | Dashboard.vue | — | 数据看板 |
| 2 | admin/workbench | workbench/ReviewWorkbench.vue | — | 审核待办 |
| 3 | admin/audit-logs | audit/AuditLogList.vue | platform_admin | 是 |
| 4 | admin/cases | cases/CaseList.vue | — | 否 |
| 5 | admin/users | users/UserList.vue | platform_admin | 否 |
| 6 | admin/account | account/AccountSettings.vue | — | 否（头像下拉进入） |
| 7 | admin/competition | competition/CompetitionList.vue | — | 否 |
| 8 | admin/config | config/ServiceConfigList.vue | platform_admin | 否 |
| 9 | admin/reviews | reviews/ReviewList.vue | — | 否 |
| 10 | admin/orders | orders/OrderList.vue | — | 否 |
| 11 | admin/products | products/ProductList.vue | — | 否 |
| 12 | admin/service-listings | serviceListings/ServiceListingList.vue | — | 否 |
| 13 | admin/enterprises | enterprises/EnterpriseList.vue | — | 否 |
| 14 | admin/demands | demands/DemandList.vue | — | 否 |
| 15-34 | experts / resources / compliance / training / certs / jobs / colleges / admin-study / achievements / challenges / projects / testsites / transformations / events / portfolios / exhibitions / reports / emergency-resources / emergency-dispatches / messages | 对应 *List.vue | — | 否 |
| 35 | admin/members | consolidated/MembersPage | — | 会员管理 |
| 36 | admin/trading | consolidated/TradingPage | — | 交易管理 |
| 37 | admin/content | consolidated/ContentPage | — | 内容管理 |
| 38 | admin/articles | consolidated/NewsPage | — | 资讯管理 |
| 39 | admin/talent | consolidated/TalentPage（支持 ?tab=pilots 直达） | — | 人才教育 |
| 40 | admin/innovation | consolidated/InnovationPage | — | 产学研 |
| 41 | admin/promotion | consolidated/PromotionPage | — | 运营推广 |
| 42 | admin/emergency | consolidated/EmergencyPage | — | 应急协同 |
| 43 | admin/settings | consolidated/SettingsPage | platform_admin | 系统设置 |

### 5.2 页面 → 后端端点（已逐条核对 Go 路由注册）

| 页面 / 模块 | 端点 | 注册位置 | 备注 |
|---|---|---|---|
| 全部 CrudList 列表 | GET/POST/PUT/DELETE /api/v1/admin/{resource} | admin_list_routes.go 等 | 35 个 resource 全部存在 |
| CrudList 导出 | GET /api/v1/admin/export/{resource} | export_handler.go:188 | **仅支持 7 个资源**（见 7.4） |
| Dashboard / BizOverview | GET /api/v1/admin/dashboard?range= | routes_phase3.go:21 | 返回 trends / trends_detail / category_dist / status_dist / modules / offline_amount_total，键名与前端**完全对齐**（已核对 phase3.go:730-891） |
| DemandList 统计 | GET /api/v1/admin/demands/stats | routes 注册已验证 | 正常 |
| DemandList 动作 | POST /admin/demands/{id}/{approve,close,amount,review} | routes_core.go:90-92、demand_fulfillment.go:95 | 正常 |
| EnterpriseList | GET /admin/enterprises；POST /admin/enterprises/{id}/review；POST /admin/enterprises/batch-review；**PATCH /api/v1/enterprises/{id}** | enterprise.go:130/162/233、routes_core.go:117 | 正常 |
| UserList | GET/POST /admin/users；POST /admin/users/{id}/role；POST /admin/users/{id}/password；DELETE /admin/users/{id} | routes_phase3.go:44-48、admin_users.go:206 | 正常 |
| OrderList | GET/PUT /admin/orders/{id}；PUT /admin/orders/{id}/aftersale | admin_list_routes.go:168-173 | 正常 |
| BookingList | GET /admin/test-sites/bookings；POST /admin/test-sites/bookings/{id}/review | admin_list_routes.go:75、batch1_handlers.go:24 | 正常 |
| PilotList | POST /admin/certified-pilots/{id}/{approve,reject} | routes_core.go:205-206 | 正常 |
| BoothApplicationList | GET /admin/exhibitions/booths；POST /admin/exhibitions/booths/{id}/review | admin_list_routes.go:102-103 | 正常 |
| ProjectList 参与申请 | GET /admin/projects/{id}/joins；POST /admin/projects/{id}/joins/{joinID}/status | biz_handlers.go:83-84 | 正常（注意列表 resource 是 research-projects，join 路径却是 projects） |
| AuditLogList | GET /admin/audit-logs?actor_id&action&resource_type&start&end | audit_handler.go:17-55 | 参数完全一致 |
| Workbench | GET /admin/{enterprises,demands,training-courses,certificates,certified-pilots,competitions,project-applications}?status&page=1&page_size=1 | 各注册点 | 正常 |
| ArticleList | GET /admin/articles（含草稿）；POST/PUT/DELETE /api/v1/articles[/{id}]；POST /api/v1/articles/{id}/publish | admin_list_routes.go:133、routes_phase3.go:32-36 | 正常 |
| AdminLayout 消息 | GET /api/v1/messages；GET /api/v1/messages/unread-count；POST /api/v1/messages/{id}/read | routes_core.go:248-250 | 正常（非分页 respond → 前端 Array.isArray 判断成立） |
| AccountSettings | GET/PATCH /api/v1/me；POST /api/v1/auth/password | routes_core.go:295-298 | 正常 |
| ServiceConfigList | GET/POST /api/services/config；GET/POST /api/v1/admin/config | h5_compat.go:1467-1468、routes_core.go:277-278 | 正常（生产无条件注册） |
| 所有上传 | POST /api/v1/upload | server.go:386 | 已统一到该端点（旧 /api/upload 为 dev-only） |
| 登录 / 刷新 / 登出 | POST /api/auth/{login,refresh,logout} | h5_compat.go:1459-1463 | 正常 |

**结论：未发现指向不存在路由的前端调用**（这是本次审计中少有的"没问题"的部分）。

---

## 6. 值得注意的实现细节 / 命名习惯 / 重复模式

### 6.1 命名与结构约定

- 组件文件一律 XxxList.vue 放在 views/admin/<模块复数>/；聚合页放 consolidated/XxxPage.vue；公共组件放 components/。
- 页面脚本区高度模板化：statusLabel / statusTag(或 statusColor) / formatDate / batchActions / searchFields / columns / detailVisible / currentItem / formVisible / formEdit / formLoading / form / resetForm / openForm / submitForm / guardClose / handleCancel / handleDelete —— 每页 30~60 行是同构样板。
- "未保存守卫"是全站统一模式：let formSnapshot = ""；guardClose() 比对 JSON.stringify(form) 有差异就 Modal.confirm，配合 :on-before-cancel（Arco 2.58 无 before-close prop，代码里反复注释强调：CaseList.vue:281-283、ChallengeList.vue:186-187、ProjectList.vue:209-210 等）。CollegeList 还额外加了 onBeforeRouteLeave（563-579）。
- 批量动作**一律传完整行**（api.update(row.id, { ...row, status: "x" })），因为后端 PUT 是全字段覆盖；几乎所有页面都写了这条注释。
- "白名单 payload"模式：Events / Exhibitions / EmergencyResources / Dispatch 等较新页面改成显式列举可写字段（EventList.vue:214-224、DispatchList.vue:180-189），而 Achievements / Study / Transforms / Certs / Colleges 仍是 { ...form } / { ...row } 整行回传。

### 6.2 重复代码（可量化）

- formatDate：**约 25 个页面各自实现**（DemandList:168、EnterpriseList:212、OrderList:122、JobList:102 …），格式还不统一（有的到分钟、有的到日、有的带 isNaN 兜底）。
- maskPhone / maskIdCard：utils/mask.js 已提供，但 EnterpriseList.vue:220-229、BookingList.vue:53-57、EnrollmentList.vue:124-133、PilotList.vue:70-74 又各自实现；**EnterpriseList 的 maskCode 是 4+8星+2，utils 的 maskIdCard 是 6+8星+4**，同一类数据两种脱敏形态。
- statusLabel / statusTag 映射表每页重复一遍，且同一枚举在不同页取值/颜色不一致（例如 event 状态在 EventList 是 published/ongoing/ended/cancelled，在 ExhibitionList 是 draft/recruiting/underway/ended）。
- 枚举转中文：Dashboard.vue:226-229 的 BIZ_LABEL 与 DemandList.vue:149-156 的 bizTypeLabel **同一枚举两套文案**（cable_inspection = 巡检 vs 工业巡检；plant_transport = 植保 vs 植保运输；spray_pesticide = 农药 vs 农药喷洒）。
- fullUrl：EnrollmentList.vue:121 与 CertList.vue:112 各写一遍 (u) => u && u.startsWith("http") ? u : u || "" —— **恒等于 u || "" 的空操作函数**，函数名与实现不符。
- 图片上传逻辑：beforeUpload(5MB/图片) + uploadRequest(custom-request) 在 Experts / Courses / Colleges / Competitions 里几乎逐行复制（Experts:122-146、Courses:193-234、Colleges:268-333、Competition:182-206）。
- onImageChange 提取真实 URL：ProductList:236-242、ServiceListingList:190-197、CollegeList:338-344 三份近似实现（且行为不一致，见 7.3）。
- 每页手写 import Message from "@arco-design/web-vue/es/message" + "@arco-design/web-vue/es/message/style/css" 两行 —— 而 utils/feedback.js 已封装同样的东西，两套提示 API 并存（部分页用 Message.xxx，部分页用 showFailToast）。

### 6.3 值得保留的好实践（客观记录）

- CrudList 的列插槽透传（slotColumns 只透传声明了 slotName 的列，91-93），避免把 search-extra / batch 误传给 a-table。
- 导出 CSV 加 BOM + csvEscape 引号转义（175-188）；审计导出同样处理（AuditLogList:338-364）。
- Document 级 ResizeObserver 警告抑制（main.js:14-23）与 index.html 白名单过滤。
- 请求序号竞态保护、刷新单飞与 pendingQueue、刷新失败只在明确 401 才登出（http.js:147-197）—— 并发细节比一般后台项目扎实。
- 上传场景统一提供 getAuthHeader() 动态读 token（避免组件创建时快照旧 token），Competition / Experts / Courses / Colleges 都用了。

---

## 7. 发现的 BUG、隐患、坏味道、不一致（按严重度）

### 🔴 P0-1 幂等键把 /api/auth/login、/api/auth/refresh 也覆盖了 → 登出后重登会拿到已被吊销的令牌

- 前端：http.js:77-82 对所有非 multipart 的 POST/PATCH 注入 Idempotency-Key，键是 idem-<djb2(url|JSON.stringify(body))>-<len>（46-53）——**确定性**，同一手机号+密码永远同一个键。
- 后端：internal/httpapi/server.go:1074-1104 对 POST/PATCH 做 24h 响应回放；未认证请求的命名空间是 anon:<ip>:<path>:<key>（1093）。
- 后果（可复现）：
  1. 同一 IP 用同一账号密码登录 → 登出（refresh token 被 Revoke）→ 24h 内再次登录 → 命中回放，返回**上一次的 access_token 与已吊销的 refresh_token**，页面提示"登录成功"；
  2. access 15 分钟内可用，之后 401 → 刷新失败 → 被踢出登录：用户看到"登录成功却马上掉线"，且服务端日志完全正常（请求根本没走到登录 handler）；
  3. 同理 POST /api/auth/refresh 的键由 refresh_token 决定：旧令牌被消费后，同一 IP 再拿旧令牌刷新会命中回放并拿到一组**有效的新令牌**，等于绕过 refresh 轮转/吊销语义 24 小时。
- 修复方向：请求拦截器排除 /api/auth/ 与 /api/v1/auth/ 前缀；或让 key 带时间戳（但那会削弱幂等本意）。

### 🔴 P0-2 index.html 全局错误处理器把错误内容直接 insertAdjacentHTML → DOM XSS

    // frontend/index.html:18-25
    var errorHtml = "<div ...>" + "<p>" + msg + "</p>" + "<p>URL: " + url + "</p>" + ...;
    document.body.insertAdjacentHTML("beforeend", errorHtml);
    // frontend/index.html:29-34
    "<p>" + event.reason + "</p>"

- 拼接的是**未转义字符串**并作为 HTML 注入。而 http.js:107-110 恰好把**后端返回的 error.message 写进 error.message**，任何被 reject 的 Promise（含 axios 错误）都会进入 unhandledrejection → event.reason 可以是含 HTML 的字符串。
- 攻击面：后端某处把用户可控内容拼进 4xx 错误文案（例如 fail(w, r, 400, errBadRequest("invalid X: " + userInput))）即可在管理后台执行脚本；即便概率低，这仍是明确的 HTML 注入汇聚点。
- 修复：改用 textContent / createElement，或对 msg / url / reason 做 HTML 转义。

### 🟠 P1-1 BookingList 的审核确认根本没等用户点确定 —— await Modal.confirm 不是 Promise

    // frontend/src/views/admin/bookings/BookingList.vue:98-120
    const ok = await Modal.confirm({ title: "驳回预约", content: h("div", ...),
      onOk: () => { if (!inputValue.trim()) { ...; return Promise.reject() } note = inputValue.trim(); return Promise.resolve() } })
    if (!ok) return

- 已核对 Arco 源码 node_modules/@arco-design/web-vue/es/modal/index.js:74-88：Modal.confirm() 返回的是 { close, update } **普通对象，没有 then**。await 立即返回该真值对象 → if (!ok) 为假 → **代码继续往下走，直接 POST 审核接口**。
- 后果：点"驳回"的瞬间就以 note = "" 提交驳回；弹窗还开着，用户填完理由点确定时请求早已发出。124-128 行的 try { await Modal.confirm(...) } catch (e) { return } 同理，永远不 catch。
- 对照正确写法：exhibitions/BoothApplicationList.vue:106-115 用 await new Promise(resolve => Modal.confirm({ onOk: () => resolve(true), onCancel: () => resolve(false) })) —— 全站仅此一处正确。

### 🟠 P1-2 排序参数前后端不一致 → 评价时间升序永远无效

- 前端 hooks/useListRequest.js:104：sort_order = order === "ascending" ? "asc" : order === "descending" ? "desc" : ""。
- 后端 internal/httpapi/reviews_resources.go:142：if sort_field == "created_at" && sort_order == "ascending" { ... } —— 只认 ascending。
- 结果：ReviewList.vue:86 的 sortable: true 升序点击后端不翻转（永远 DESC）；且全仓只有 reviews 一个接口读 sort_field / sort_order（grep 仅 4 处命中，其中 3 处是 domain/models.go 的无关字段），CaseList.vue:183、ComplianceList.vue:254/282 的 sortable 列把参数发给后端后被静默忽略，只剩 Arco 的当前页客户端排序 —— 用户看到的"排序"语义混乱。

### 🟠 P1-3 图片上传失败时把 blob: 预览地址写进表单 → 保存后入库死链

    // frontend/src/views/admin/products/ProductList.vue:236-242（ServiceListingList.vue:190-197 同）
    form.images = fileList
      .map((f) => f.response?.data?.url || f.response?.url || (f.status === "uploading" ? "" : f.url))
      .filter(Boolean)

- 只排除了 uploading，**没有排除 error**。上传失败时 Arco 的 fileItem.status === "error" 且 f.url 仍是 blob:http://... ，会被写进 form.images → PUT 落库。
- 对照：CollegeList.vue:342 已经写对 (f.status === "uploading" || f.status === "error" ? "" : f.url) —— 同一仓库两套写法，说明前两个是漏改。

### 🟠 P1-4 /api/v1/admin/export/{resource} 只支持 7 个资源，其余页面按钮静默降级

- 后端 internal/httpapi/export_handler.go:204-327 只处理 enterprises / demands / training-courses / certificates / certified-pilots / enrollments / competitions，其余走 default: 404（324-326）。
- 前端 components/CrudList.vue:204-221：任何非 403 的失败都**静默**回退到"导出当前页"且不提示，只有 403 才 Message.warning。于是 orders / products / experts / jobs / colleges / events / exhibitions …（约 28 个页面）的"导出 CSV"实际只导出当前 20 行，用户以为拿到了全量。
- 佐证作者意图：CrudList.vue:164-170 的 EXPORT_NAME 里列了 users:"用户管理"、audit-logs:"操作审计"，但后端根本没有这两个 case。

### 🟠 P1-5 侧栏只暴露 11 项，32 条路由成为"无高亮孤岛"

- 侧栏菜单 AdminLayout.vue:212-232 共 11 项；selectedKeys = computed(() => [route.path])（322）。
- Workbench 的卡片把用户送到 /admin/enterprises、/admin/demands、/admin/training、/admin/certs、/admin/competition、/admin/projects（ReviewWorkbench.vue:58-64），这些路径**不在 allMenus 里** → 进入后侧栏无任何项高亮，用户只能改 URL 或返回。
- 说明"43 条旧路由"与"9 个聚合页"两代信息架构同时在线，旧路由既未下线也未从侧栏可达。

### 🟡 P2-1 菜单可见性与路由守卫不一致（点了才被弹回）

- AdminLayout.vue:231 把 /admin/settings 的 roles 写成 [platform_admin, association_admin]，而 router/index.js:65 的 meta.roles 是 [platform_admin]。
- 协会管理员能看到"系统设置"入口，点击后守卫弹"该页面仅平台管理员可用"并跳回 dashboard —— 菜单与守卫两套权限判断，应由同一份数据驱动。

### 🟡 P2-2 /api/v1/me 的完整响应（含明文手机号）被写进 localStorage

- views/admin/composables/useAuth.js:19-29：localStorage.setItem("user", JSON.stringify(current))，而 current 是 GET /api/v1/me 的完整对象；后端返回 id/role/status/name/avatar_url/phone/has_password/gender/birthday/region/bio（auth_wechat.go:371-385），其中 phone 是解密后的**明文**（auth_wechat.go:361-362）。
- 这与登录页的自我约束直接矛盾：login/Index.vue:126-131 注释写"用户信息仅存 { id, role }"，实际每次进 AdminLayout 都被 refreshCurrentUser()（AdminLayout.vue:452）覆盖成完整对象。
- 另外令牌虽已从 localStorage 降级到 sessionStorage（http.js:3-6），但 readToken 仍保留 localStorage 回退（18-26），迁移逻辑之后依然可以被回填。

### 🟡 P2-3 XSS 汇聚点：3 处 v-html

- articles/ArticleList.vue:66：<div class="detail-content" v-html="currentItem.content"></div>
- competition/CompetitionList.vue:51：<span v-html="currentItem.description"></span>
- projects/ProjectList.vue:54：<span v-html="currentItem.description"></span>
- 缓解：后端 middleware.SanitizeJSONBody 有白名单清洗（middleware.go:129-200：允许 p/br/strong/a/img 等，危险标签连内容删除，a/img 仅允许 http(s)/相对路径），所以**当前不可直接利用**；但它是**正则实现**，且只覆盖 "Content-Type: application/json" 的写入路径。任何绕过写入路径的历史数据或正则边界绕过，都会在管理后台直接执行。富文本展示应改用 v-text 或前端同步做一次 DOM 级清洗。

### 🟡 P2-4 失败与空数据不可区分（全站 35 个列表页）

- hooks/useListRequest.js:77-88：catch 里 listData.value = []、total.value = 0 并 toast，页面表格随即显示"暂无数据"。
- 另有页面在 apiFunction 内自己 catch 并 return { data: [], total: 0 }（ArticleList.vue:163-174、CaseList.vue:144-155），把错误彻底吞掉，只留一条 toast。
- 结果：接口 500 与"真的没有数据"在 UI 上完全同形。

### 🟡 P2-5 缺 loading / 持久错误态

- config/ServiceConfigList.vue:858-867：无 loading 态，失败仅 toast + console.error，页面留一片空白卡片，无空态也无重试按钮。
- account/AccountSettings.vue:90-101：load() 无 loading 标记，加载中展示 "-" 占位（与"未绑定"同形），失败只 toast。
- exhibitions/BoothApplicationList.vue:92-102：loadExhibitions 失败静默 → "展会"列退化为裸 ID。
- AdminLayout.vue:278-287：消息拉取失败完全静默（注释说明是有意为之，可接受，但也导致角标长期为 0 无人察觉）。

### 🟡 P2-6 重复点击/相同内容会被幂等中间件吞掉（非 auth 场景）

- 同 P0-1 的机制，但作用在业务接口：NotifyList.vue:137-149 用相同标题+内容连发两次通知 → 第二次命中 24h 回放，**不会新增消息**，页面照样弹"已发送/已广播给全部用户"。
- 同理重复创建同名同参数的商品 / 职位 / 课程。用户无任何提示，属于"看起来成功实际没生效"。

### 🟡 P2-7 normalizeMediaUrl 的 "port === 8090" 判断过宽

    // views/admin/composables/useMedia.js:19-27
    const isLocalish = host === "localhost" || host === "127.0.0.1" || host === "0.0.0.0" || host === "172.17.0.1" || port === "8090"
    if (isLocalish) return u.pathname + u.search + u.hash

- 端口判断没有限定 host：https://cdn.example.com:8090/a.png 会被改写成 /a.png → 图片 404。
- 同时把 Docker 网关 IP 172.17.0.1 硬编码进了前端源码（nginx.conf:63/83/96/107 也硬编码了同一 IP，部署时应使用 compose 服务名）。

### 🟡 P2-8 AdminLayout 里两个"僵尸 Tab"

    // frontend/src/views/admin/AdminLayout.vue:261-262
    const notices = ref([]); const todos = ref([])

- 全文件再无任何赋值，通知面板的"通知(N)""待办(N)"两个页签永远显示"暂无通知/暂无待办"（43-58），角标也永远等于 messages 未读数。属未完成的半成品 UI。
- 另：setInterval(loadMessages, 60000)（455）无 document.hidden 判断，后台标签页持续轮询。

### 🟡 P2-9 登录成功但无权限造成的"半登录"残留

- 非管理员账号可以成功登录（后端 h5AuthLogin 不校验角色，h5_compat.go:660-833），前端 afterLogin 先 router.push("/admin") 再被守卫弹回 /login 并 toast（router/index.js:103-107）。
- 此时令牌与 localStorage.user 都已写入，而登录页**没有登出入口**，用户必须再登录一次覆盖。建议登录页先按 role 判断并给出明确提示。

### 🟢 P3 级坏味道 / 不一致（逐条）

1. frontend/vite.config.js.timestamp-*.mjs：Vite 临时产物入库，应删除。
2. frontend/PINIA_GUIDE.md：264 行文档描述 3 个不存在的 store（@/stores/user 等）与不存在的 /home 路由；实测全仓 0 处 defineStore。要么补 store，要么删文档。
3. frontend/研学管理员权限修复.md：描述已删除的 frontend/h5/ 与 Express 后端 backend/routes/admin.js，与当前 Go + Arco 架构完全脱节。
4. frontend/README.md:29-32 称"40 条子路由（基础 11 + 业务 20 + 聚合 9）"，实际 **43 条**（基础已是 14：多了 workbench / audit-logs / account）。
5. hooks/useListRequest.js 返回的 onPageChange / onSelectChange / onBatchAction 无人调用（CrudList 全部自实现），死代码；onSortChange 走的是 hook 内部未防抖的 loadData，绕过 CrudList 的 150ms 防抖且不触发 emit("loaded")（对比 CrudList.vue:279-285）。
6. components/BizOverview.vue:7 的 :span="Math.floor(24 / metrics.length)" 与 :md="24 / metrics.length"：TalentPage 传 5 个指标时得到 4 与 4.8（非整数 span），布局塌陷。
7. RichEditor.vue:11 工具栏用 emoji 链接图标，index.html:5 favicon 用 emoji 直升机 —— 与项目"不使用 emoji 图标"的规范不符（该规范写在小程序章节，同一设计体系下建议对齐）。
8. RichEditor.vue:20 的 @keydown.enter.prevent="onEnter" 把**所有**回车都拦成 formatBlock("p")：无序/有序列表里无法回车新建下一项，Shift+Enter 软换行也失效；document.execCommand（36-47）为废弃 API。
9. EnterpriseList.vue:85 同时绑定 :on-before-ok="beforeOkEdit" 与 @ok="confirmEdit"，两处重复同样的必填校验（312-317 与 326-334）。
10. ComplianceList.vue:117 删除确认的 title 写成 "提示"（其他页面都是"删除 X"），文案不一致。
11. compliance/ComplianceList.vue:314-318 的 formSnapshot 是**单变量**，两个 Tab 共用；在 docs 表单改动后切到 standards Tab，守卫会拿 stdForm 与 docs 的快照比对 → 误报"有未保存修改"（或漏报）。
12. CrudList.vue:166 的 EXPORT_NAME 是只为"导出文件名"存在的硬编码表；exportFilename()（172）用 toISOString().slice(0,10) 取 UTC 日期，北京时间晚上导出的文件名会差一天。
13. CrudList.vue:293-302 与 useListRequest.js:120-122 两套 selectedIds 逻辑并存；翻页后 selectedIds 不清空（loadData 未重置），批量条会显示"已选择 N 项"而 selectedRows 为空。
14. ReviewWorkbench.vue:84 在循环里调用 useAdminApi(item.resource) —— 每次 loadAll 重建 7 个 api 对象（无副作用但属误用，composable 应在 setup 期调用一次）。
15. OrderList.vue:155-165 的"交易额/已完成/完成率"是**当前页**口径（只有总数是全量），完成率尤其容易被当成全站指标；DemandList 则专门做了 /stats 全量接口 —— 同类指标两种口径。
16. ProductList.vue:181/199 的 priceYuan = ((row.price_fen || 0) / 100).toString() 会把 null 价格回显成 "0"，提交时写成 0 分而不是 null（与自家 price_fen: null 的"空值不写 0"注释自相矛盾）；ServiceListingList.vue:160 同病（该项目注释明写"0 表示面议"，尚可接受，但 products 页面渲染成金额 ¥0 是错的）。
17. CompetitionList.vue:323 把 fee / min_fee / original_fee / max_teams 用 ?? 0 强制转 0：清空输入框存的是 0 而非"未设置"（同项目其他页面已统一改为 null，属漏改）。
18. CaseList.vue:159-161 及多数批量动作把 { ...row } 原样 PUT 回后端，行内含 status / created_at 等只读字段；虽然当前后端忽略，但破坏了"白名单 payload"约定，只读字段一旦进入 Update 结构体就会被写回。
19. ArticleList.vue:242/266、CaseList.vue:262 用 Message.loading("保存中...", 0) + 手动 Message.clear()，与 utils/feedback 的 showLoadingToast / closeToast 两套风格混用，且 Message.clear() 会顺手清掉页面上其他提示。
20. AdminLayout.vue:435 的 goHome → router.push("/")（靠 / 重定向到 /admin/dashboard）；下拉里"返回首页"实际回到看板，而平台并没有"首页"，属于语义空转。
21. 品牌名不一致：AdminLayout.vue:8 "低空运营后台" vs login/Index.vue:16 "无人机产业综合服务平台" vs index.html:7 标题。
22. index.html:6 的 maximum-scale=1.0, user-scalable=no 禁用缩放，管理后台（大屏为主）无必要且不利于可访问性。
23. 无任何前端测试与静态检查（package.json 只有 dev/build/preview；全仓无 eslint / prettier / tsconfig / *.spec.js），与后端 go vet 的验收标准形成反差。
24. public/ 下 5 个 mp4 共约 38MB 随构建产物发布，其中"泰顺.mp4""杨梅吊运.mp4"命名与项目无关联；两个神秘 txt 文件（32/40 位十六进制串）也在公网可访问。
25. nginx.conf:25-26 证书路径、nginx.conf:63/83/96/107 网关 IP 172.17.0.1:8080 全部硬编码；nginx.conf:94 的 location 正则 ^/api/(auth/|v1/files/) 与更严的 auth_limit 只覆盖 auth 与 files，上传接口走较宽的 api_limit。
26. frontend/Dockerfile 的构建阶段被整体注释，必须"本地先 npm run build 再 COPY dist"，与 CLAUDE.md 描述的 Docker 多阶段构建不符（前端侧）。
27. utils/mask.js:6-10 的 maskPhone 对 length < 7 直接返回原文（不脱敏）；座机号 023-88886666（12 位）会被切成 023****6666，把区号当手机号前缀；maskContact（38-44）的正则 1[3-9]\d{9}|0\d{2,3}-?\d{7,8} 对"023-8888 6666"这类含空格写法匹配不到，会原样输出（潜在 PII 泄漏点，仅影响展示）。

---

## 8. 建议的修复优先级

1. **立刻**：http.js 的 Idempotency-Key 排除 /api/auth/*（P0-1）；index.html 改 textContent（P0-2）。
2. **本迭代**：BookingList 改 Promise 包裹（P1-1）；统一 sort_order 取值常量（P1-2）；ProductList / ServiceListingList 补 status === "error" 排除（P1-3）；CrudList 导出失败改为显式提示，或后端补 export 资源 / 前端隐藏按钮（P1-4）。
3. **随后**：侧栏与路由权限同源（P2-1）；localStorage 只存 { id, role }（P2-2）；v-html 改 v-text 或加前端清洗（P2-3）；列表页增加 error 态（P2-4）。
4. **技术债**：抽 useCrudForm（守卫+快照+提交）、useStatusMap（枚举→文案/颜色），把 formatDate 与脱敏函数收口到 utils；清理死代码（useListRequest 冗余导出、AdminLayout 僵尸 Tab、PINIA_GUIDE / 研学修复文档、vite timestamp 产物）；补 eslint，并至少为 http.js / useListRequest 加单测。
