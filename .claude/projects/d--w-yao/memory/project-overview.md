---
name: project-overview
description: 无人机产业综合服务平台 — 全栈项目全景概览
metadata:
  type: project
---

# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 12+ 大业务模块。

## 技术栈

| 端 | 技术 |
|----|------|
| 微信小程序 | 原生 + Vant Weapp 1.11，27 页面 + 7 分包 + 40+ 组件 |
| Web 管理后台 | 嵌入式 SPA（单文件 admin.html 35KB），ECharts 5 |
| 后端 API | Go 1.22+，标准库 net/http，120 条端点 |
| 数据库 | PostgreSQL 15+（生产）/ 内存存储（开发），30 张表，11 个迁移文件 |
| 部署 | Docker + docker-compose（PostgreSQL + API 双容器） |

## 12+ 业务模块

用户认证(微信登录+Token+4级RBAC) → 企业入驻(审核流程) → 需求大厅(发布/竞标/双确认/争议) → 招聘求职 → 社区内容 → 二手交易 → 用工派遣 → 合同签约 → 培训认证(CAAC/UTC/人社) → 无人机交易 → 保险服务 → 金融服务 → 资金托管 → 信用评价 → 场地预约 → 行业资讯

## 项目结构

```
cmd/api/main.go              # 启动入口，组装所有依赖
internal/httpapi/            # 路由+中间件+请求处理(仅解析/调Service/respond)
internal/service/            # 业务规则、权限校验、状态流转
internal/repository/         # 数据持久化接口(repositories.go)+PG实现(postgres/)+内存实现(memory/)
internal/domain/models.go    # 所有业务实体+常量+状态定义(612行)
internal/config/config.go    # 集中配置(环境变量读取+校验+打印)
internal/crypto/             # AES-256-GCM 字段加密
internal/logger/             # 结构化日志(控制台+每日文件轮转)
internal/cache/              # 内存TTL缓存
internal/middleware/         # 输入消毒+统一错误格式
migrations/                  # 11个SQL迁移文件
miniprogram/                 # 微信小程序(Vant Weapp)
drone-platform-ui/           # 小程序UI改版(自定义tabbar)
docs/                        # 架构/API契约/数据模型/路线图/AI规范
```

## 数量指标

| 指标 | 数据 |
|------|:--:|
| API 端点 | 120 |
| 数据库表 | 30 |
| 小程序页面 | 27（7 分包） |
| Go 源文件 | 53 |
| 总代码行数 | 5,300+ |
