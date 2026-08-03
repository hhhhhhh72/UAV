# 设计文档：删除 H5 前台 + 两端脱离 Vant（统一品牌视觉）

- 日期：2026-08-03
- 状态：已确认（brainstorming 完成）
- 范围：frontend（H5 前台删除 + Admin 去 Vant）+ miniprogram（Vant Weapp 移除 + 自研组件库）

## 1. 背景与目标

**背景**：
- 小程序是产品唯一用户入口，frontend 中的 H5 前台（17 路由）与小程序功能重复，维护成本双份。
- 小程序依赖 Vant Weapp（`miniprogram/wxcomponents/` 占 1.2MB，68 页使用 25+ 种组件），前端依赖 Vant 4（H5 17 处 + Admin 内部 10+ 文件，均为 toast 类）。
- 现有页面存在大量非品牌色硬编码（橙/红/黑）与 emoji 图标，视觉不统一。

**目标**：
1. frontend 瘦身为纯 Admin 后台（H5 前台删除，登录入口保留并改造）。
2. 两端彻底移除 Vant 依赖（frontend 中 vant 引用清零；小程序删除 wxcomponents/）。
3. 小程序自研轻量组件库，统一"扁平简约蓝"品牌视觉（参考素材库 `D:\BaiduNetdiskDownload\UI\B-007 figma` 中管理后台风格，建立在既有品牌令牌上）。
4. 重构不改变功能行为，仅替换实现层与统一视觉。

## 2. frontend：H5 前台删除

### 2.1 删除清单

| 类别 | 内容 |
|------|------|
| 页面目录 | `src/views/` 下：home、services、messages、applications、mine、register、study、cases、reviews、demand、layout（H5 布局） |
| 状态 | `src/stores/` 全部：demand、enterprise、home、message、service、user |
| 路由 | router 中 17 条 H5 路由（`/home` `/services` `/messages` `/applications` `/mine` `/study*` `/service-*` `/cases` `/reviews` `/demand/*` `/register`），`/` 重定向到 `/admin` |
| 依赖 | main.js 的 `import Vant from 'vant'` 移除；router 的 `showFailToast from 'vant'` 替换 |

### 2.2 登录页改造（保留 `/login` 路由）

- Admin 路由守卫 6 处 `next('/login')` 与 `utils/http.js` 2 处跳转依赖 `/login`，**路由必须保留**。
- 组件替换：`views/login/Index.vue`（H5 风格）→ 基于既有 `views/auth/Login.vue` 的 Admin 风格登录页。
- 登录流程：密码登录走 `POST /api/auth/login`（生产注册的 bcrypt 校验）；dev 模式保留 `POST /api/v1/admin/token` 自动登录（路由守卫已有逻辑）。

### 2.3 Admin 内部 vant toast 替换

10+ 文件引用 vant toast 函数（CaseList、CompetitionList、Dashboard、DemandList、EnterpriseList、OrderList、ReviewList、useMedia、ImageCropper、ServiceConfigList 等）。

- 新增 `src/utils/feedback.js` 统一封装：
  - `showToast(msg)` / `showFailToast(msg)` / `showSuccessToast(msg)` → `ElMessage`
  - `showLoadingToast()` / `closeToast()` → `ElLoading` 或轻量 loading 状态
  - `showConfirmDialog(opts)` → `ElMessageBox.confirm`
- 各文件 import 改为 `from '@/utils/feedback'`，调用点保持函数名不变（最小 diff）。
- **验收**：`grep -r "from 'vant'" src/` 零命中；`package.json` 移除 vant 依赖。

## 3. miniprogram：自研组件库 + Vant Weapp 移除

### 3.1 设计令牌（扩展现有 App.vue 体系）

| 令牌 | 值 | 用途 |
|------|-----|------|
| `--color-primary` | `#0A66C2`（已有） | 主按钮/激活态/导航 |
| `--color-primary-light` | `#E8F2FC`（已有） | 选中背景/标签浅底 |
| `--color-accent` | `#1DD4A8`（已有） | 点缀：徽标/数据高亮 |
| 圆角/阴影/间距 | 既有 `--radius-*` `--shadow-*` `--space-*` | 大圆角卡片 + 轻阴影 |

**风格规则**：卡片化（白底 + 大圆角 + 轻阴影）；交互色仅品牌蓝/青绿；图标 CSS 绘制或文字标签（不引图标库、不用 emoji）。

### 3.2 组件库清单（`miniprogram/components/ui/`，u- 前缀，easycom 自动注册）

基于 68 页使用统计（van- 出现次数）自研 15 个：

| 组件 | 对应 van | 使用量 | 关键属性（兼容 van） |
|------|---------|:--:|------|
| u-button | van-button | 39 | type/size/block/disabled/loading |
| u-cell | van-cell | 66 | title/value/label/is-link/is-clickable |
| u-cell-group | van-cell-group | 76 | inset/border |
| u-field | van-field | 39 | label/value/placeholder/type(输入框类型)/textarea |
| u-empty | van-empty | 58 | description/image |
| u-loading | van-loading | 94 | size/color |
| u-tag | van-tag | 52 | type/plain/size |
| u-nav-bar | van-nav-bar | 32 | title/left-text/right-text/fixed |
| u-search | van-search | 14 | placeholder/disabled |
| u-popup | van-popup | 20 | show/position/round/close-on-click-overlay |
| u-tabs | van-tabs | 18 | active/type/swipe-threshold |
| u-tab | van-tab | 23 | title（配合 u-tabs 使用） |
| u-sticky | van-sticky | 18 | offset-top |
| u-icon | van-icon | 25 | name（内置：arrow/search/close/plus/check/success 等，CSS 绘制） |
| u-picker | van-picker | 5 | columns/model-value（单选简化版） |

**低频组件**（rate、notice-bar、datetime-picker、action-sheet、card、image、grid、collapse 等合计 <10 处）：不建组件，页面内 CSS 实现或直接替换为原生结构。

**easycom 配置**（pages.json）：
```json
"easycom": {
  "autoscan": true,
  "custom": { "^u-(.*)": "@/components/ui/u-$1.vue" }
}
```

### 3.3 迁移原则

- 页面迁移 = `van-x` → `u-x` 标签替换 + 高频属性名保持一致 + 样式微调（品牌化）。
- 页面内样式统一走设计令牌（`var(--color-*)`），顺带清理非品牌色硬编码与 emoji 图标。
- 迁移完成后删除 `miniprogram/wxcomponents/` 目录与 pages.json 中 `usingComponents` 的 van- 声明。

## 4. 迁移分批

| 批次 | 内容 | 验证 |
|------|------|------|
| 批 1 | frontend：H5 删除 + login 改造 + Admin toast 替换（feedback.js） | `vite build` 通过；Admin 登录→各模块页浏览无报错 |
| 批 2 | 小程序：设计令牌扩展 + 15 个 u- 组件 + easycom 配置 | 小程序编译通过；组件渲染正常 |
| 批 3 | 高频页面迁移（首页/我的/列表类约 25 页） | 小程序编译通过；逐页浏览 |
| 批 4 | 中低频页面迁移（约 43 页）+ 删除 wxcomponents/ + 清理 van 声明 | 小程序编译通过；68 页全量浏览；无 van- 残留 |

## 5. 验收标准

1. **视觉统一品牌化**：小程序与 Admin 交互色仅品牌蓝/青绿；无 emoji 图标；无非品牌色硬编码残留。
2. **vant 依赖清零**：frontend `grep -r "from 'vant'"` 零命中且 package.json 移除；小程序页面与 pages.json 中 `van-` 引用零残留（`grep -r "van-" miniprogram/pages miniprogram/pages.json` 零命中），wxcomponents/ 目录删除。
3. **功能行为不变**：68 页 + Admin 35 路由重构前后功能等价。
4. **体积**：小程序主包明显减小（wxcomponents 1.2MB 移除）；frontend bundle 减小（vant 移除）。

## 6. 风险与回滚

| 风险 | 缓解 |
|------|------|
| 组件视觉与 van 差异导致页面布局错乱 | 分批迁移 + 逐批编译浏览验证；组件属性兼容降低 diff |
| 登录改造影响 Admin 访问 | 批 1 单独验证登录全流程后再动小程序 |
| 低频组件页面内实现重复代码 | 控制在 <10 处；若某组件使用量迁移中增长，升格为 u- 组件 |
| 删除 H5 后旧链接 404 | 产品决策接受；`/` 重定向到 `/admin` |

回滚：每批独立 commit；若某批破坏功能，单独 revert 该批 commit。
