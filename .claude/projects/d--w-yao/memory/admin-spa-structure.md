---
name: admin-spa-structure
description: 管理后台 SPA 结构 — admin.html 侧边栏导航、页面切换、样式系统
metadata:
  type: reference
---

# 管理后台 SPA (admin.html)

单文件 35KB，位于 `internal/httpapi/admin.html`，Go embed 到 `/admin` 路由（需 ADMIN_DEV_MODE=true）。

## CSS 设计系统

Apple 风格变量：`--bg: #f5f5f7`, `--accent: #0071e3`, `--green: #34c759`, `--orange: #ff9f0a`, `--red: #ff3b30`, `--purple: #5856d6`, `--teal: #5ac8fa`
圆角 12px，侧边栏宽 220px，头部高 56px，字体 SF Pro / PingFang SC。

## 侧边栏导航

分组标签 + nav-item，通过 `switchPage(page, el)` JS 函数切换 page-section 的 display。
CSS class `.page-section.active { display: block }` 控制可见性。

当前页面：
- 数据概览 (dashboard) — 4个指标卡片 + ECharts 趋势图 + 饼图
- 企业审核 (enterprises) — 表格+筛选+通过/驳回/补件
- 需求管理 (demands) — 表格+审核
- 举报管理 (reports) — 表格
- 行业资讯 (articles) — 表格+新建Modal
- 评价审核 (reviews-admin) — 表格+审核
- 用户管理 (users-admin) — 表格+新建/改角色
- 服务配置 (config-admin) — Banner/公告编辑弹窗
- 数据导出 (exports) — CSV导出卡片
- Token工具 (tokens) — 生成/查看Token

## JS 核心函数

- `fetchAPI(url, opts)` — 封装 fetch + auth headers + 错误处理
- `respond(w, r, status, data)` — Go侧统一响应
- `fail(w, r, status, err)` — Go侧统一错误
- `toast(msg, type)` — 浮动提示
- `esc(s)` — HTML 转义
- `fmtDate(d)` — 日期格式化

## 添加新页面步骤

1. 在 sidebar 添加 `<a class="nav-item" data-page="xxx" onclick="switchPage('xxx',this)">`
2. 在 `.content` 添加 `<div class="page-section" id="page-xxx">`
3. 在 `switchPage()` 函数添加页面加载调用
4. 使用现有 CSS 类：`.panel`, `.panel-header`, `.panel-body`, `.btn`, `.tag`, `.modal-overlay`, `.toast`
