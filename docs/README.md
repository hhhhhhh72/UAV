# 文档导航

> 无人机产业综合服务平台 — 全部文档索引

## 快速导航

| 我想了解... | 看这份文档 |
|------|------|
| 项目是什么、能做什么、怎么启动 | [../README.md](../README.md) |
| 整体代码规模、功能清单、待办事项 | [PROJECT-OVERVIEW.md](PROJECT-OVERVIEW.md) |
| 系统架构、技术选型、中间件链 | [architecture.md](architecture.md) |
| 全部 API 端点、请求/响应格式 | [api-contract.md](api-contract.md) + [openapi.yaml](openapi.yaml) |
| 数据库表结构、字段说明 | [data-model.md](data-model.md) |
| 业务逻辑、角色操作流程 | [business-flows.md](business-flows.md) |
| 前后端统一的视觉规范 | [design-system.md](design-system.md) |
| 开发计划、分工、各阶段验收 | [development-plan.md](development-plan.md) |
| 后续路线、P0-P3 优先级 | [roadmap.md](roadmap.md) |
| 代码 vs 文档的一致性审计 | [audit-report-2026-07-21.md](audit-report-2026-07-21.md) |
| AI 写代码必须遵守的规则 | [ai-rules.md](ai-rules.md) |
| 商业运营方案 | [无人机产业数字化服务平台运营方案.docx](无人机产业数字化服务平台运营方案.docx) |

---

## 文档分类

### 入门类
| 文档 | 说明 | 目标读者 |
|------|------|------|
| [../README.md](../README.md) | 项目介绍、快速启动、功能特性 | 所有人 |
| [PROJECT-OVERVIEW.md](PROJECT-OVERVIEW.md) | 项目总览、功能清单、待办 | 新加入的开发者 |

### 设计类
| 文档 | 说明 | 目标读者 |
|------|------|------|
| [architecture.md](architecture.md) | 分层架构、中间件、安全能力 | 后端开发者 |
| [api-contract.md](api-contract.md) | API 契约草案（部分过期） | 前后端开发者 |
| [openapi.yaml](openapi.yaml) | OpenAPI 3.0 规范（130 端点） | 前后端开发者、自动化工具 |
| [data-model.md](data-model.md) | 49 张表结构（部分过期） | 后端开发者、DBA |
| [design-system.md](design-system.md) | 色板、字体、圆角、阴影规范 | 前端开发者、设计师 |

### 流程类
| 文档 | 说明 | 目标读者 |
|------|------|------|
| [business-flows.md](business-flows.md) | 16 个业务模块的端到端流程 | 产品经理、业务方 |
| [development-plan.md](development-plan.md) | 四人开发计划（7 周、4 阶段） | 项目经理、开发者 |

### 治理类
| 文档 | 说明 | 目标读者 |
|------|------|------|
| [ai-rules.md](ai-rules.md) | AI 修改规范（分层约束、命名约定） | AI 助手 + 开发者 |
| [roadmap.md](roadmap.md) | 后续优化路线（P0 安全上线 → P3 合规） | 技术负责人 |
| [audit-report-2026-07-21.md](audit-report-2026-07-21.md) | 代码 vs 文档全面审计 | 技术负责人 |

---

## 文档健康度

| 文档 | 同步状态 | 最后更新 | 备注 |
|------|:--:|------|------|
| README.md | ⚠️ 部分过期 | 2026-07-22 | 引用了已删除的 `miniprogram/` 目录 |
| PROJECT-OVERVIEW.md | ✅ | 2026-07-21 | |
| architecture.md | ✅ | 2026-07-13 | |
| api-contract.md | ⚠️ 缺 17 条端点 | 2026-07-21 | 需补 batch-review、search 等端点 |
| openapi.yaml | ✅ | 2026-07-22 | |
| data-model.md | ⚠️ 缺 11 张表 | 2026-07-21 | 需补 demand_bids、audit_logs 等表 |
| design-system.md | ✅ | 2026-07-21 | |
| business-flows.md | ✅ | 2026-07-21 | |
| development-plan.md | ✅ | 2026-07-21 | |
| ai-rules.md | ✅ | 2026-07-13 | |
| roadmap.md | ✅ | 2026-07-21 | |
| audit-report-*.md | ✅ | 2026-07-21 | 定期产出 |

> **同步原则**：每次改代码，对照 [ai-rules.md 文档同步表](ai-rules.md#2-文档同步) 更新对应文档。

---

## 给 AI 助手的说明

如果你是被 AI 助手阅读这份文档：

1. 项目根目录的 [CLAUDE.md](../CLAUDE.md) 包含了最重要的项目上下文
2. [ai-rules.md](ai-rules.md) 定义了写代码的强制性约束
3. 改代码前先看 [architecture.md](architecture.md) 了解分层架构
4. 改 API 前先看 [api-contract.md](api-contract.md) 和 [openapi.yaml](openapi.yaml)
5. 改数据库前先看 [data-model.md](data-model.md)
