# PRD 详细分工 — 页面级

> 4 人 × 4 周 = 57 页面 + 组件库

---

## 成员 A（组长）：供需对接 + 管理后台 + 基础架构

### Sprint 1（W1）：列表+详情 — 6 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| A1 | 需求大厅列表 | `pages/demands/list` | ⭐⭐ |
| A2 | 需求详情 | `pages/demands/detail` | ⭐⭐ |
| A3 | 全局搜索 | `pages/search/index` | ⭐⭐⭐ |
| — | 通用组件：需求卡片 | `components/demand-card` | ⭐ |
| — | 通用组件：分类筛选栏 | `components/filter-tabs` | ⭐ |
| — | 通用组件：骨架屏 | `components/skeleton` | ⭐ |

### Sprint 2（W2）：交互 — 4 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| A4 | 需求发布 | `pages/demands/publish` | ⭐⭐⭐ |
| A5 | 竞标报价 | `pages/demands/bid` | ⭐⭐⭐ |
| A6 | 我的需求 | `pages/demands/mine` | ⭐⭐ |
| — | 智能推荐 | `pages/match/recommend` | ⭐⭐⭐ |

### Sprint 3（W3）：管理后台 — 15 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| A7 | 后台框架 | `pages/admin/layout` | ⭐⭐⭐ |
| A8 | 数据看板 | `pages/admin/dashboard` | ⭐⭐⭐ |
| A9 | 企业审核列表 | `pages/admin/enterprises/list` | ⭐⭐ |
| A10 | 企业审核详情 | `pages/admin/enterprises/detail` | ⭐⭐ |
| A11 | 需求审核列表 | `pages/admin/demands/list` | ⭐⭐ |
| A12 | 需求审核详情 | `pages/admin/demands/detail` | ⭐⭐ |
| A13 | 评价审核 | `pages/admin/reviews/list` | ⭐⭐ |
| A14 | 举报处理 | `pages/admin/reports/list` | ⭐ |
| A15 | 用户管理 | `pages/admin/users/list` | ⭐⭐ |
| A16 | 平台配置 | `pages/admin/config/index` | ⭐⭐⭐ |
| A17 | CSV 导出 | `pages/admin/export` | ⭐ |
| A18 | 内容审核(帖子) | `pages/admin/posts/list` | ⭐ |
| A19 | 合同管理 | `pages/admin/contracts/list` | ⭐ |
| A20 | 消息推送 | `pages/admin/messages/send` | ⭐ |
| A21 | 操作日志 | `pages/admin/audit/logs` | ⭐ |

### Sprint 4（W4）：收尾
- H5 ↔ 小程序对齐全部 A 负责页面
- E2E 测试
- 性能优化 + Bug 修复

**A 合计：6 + 4 + 15 = 25 页**

---

## 成员 B：会员资源 + 合规政策

### Sprint 1（W1）：展示型 — 5 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| B1 | 企业注册 | `pages/enterprise/register` | ⭐⭐⭐ |
| B2 | 企业审核状态 | `pages/enterprise/status` | ⭐⭐ |
| B3 | 专家智库列表 | `pages/experts/list` | ⭐⭐ |
| B4 | 专家详情 | `pages/experts/detail` | ⭐ |
| B5 | 政策资讯列表 | `pages/compliance/news` | ⭐⭐ |

### Sprint 2（W2）：交互+资源 — 6 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| B6 | 产业资源台账 | `pages/resources/list` | ⭐⭐ |
| B7 | 资源详情+预约 | `pages/resources/detail` | ⭐⭐⭐ |
| B8 | 合规知识库 | `pages/compliance/knowledge` | ⭐⭐ |
| B9 | 团体标准库 | `pages/compliance/standards` | ⭐⭐ |
| B10 | 项目申报 | `pages/applications/submit` | ⭐⭐⭐ |
| B11 | 案例库列表 | `pages/cases/list` | ⭐⭐ |

### Sprint 3（W3）：收尾 — 0 新页
- B 全部 11 页已完成，协助 A 做管理后台的审核流页面
- H5 ↔ 小程序对齐

**B 合计：5 + 6 = 11 页**

---

## 成员 C：人才教育 + 应急协同

### Sprint 1（W1）：展示型 — 5 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| C1 | 培训课程列表 | `pages/training/courses` | ⭐⭐ |
| C2 | 课程详情+报名 | `pages/training/enroll` | ⭐⭐⭐ |
| C3 | 赛事列表 | `pages/competitions/list` | ⭐⭐ |
| C4 | 院校展示 | `pages/colleges/list` | ⭐⭐ |
| C5 | 应急资源列表 | `pages/emergency/resources` | ⭐⭐ |

### Sprint 2（W2）：交互型 — 7 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| C6 | 我的证书 | `pages/training/certificates` | ⭐⭐ |
| C7 | 赛事报名 | `pages/competitions/register` | ⭐⭐⭐ |
| C8 | 职位列表 | `pages/jobs/list` | ⭐⭐ |
| C9 | 简历管理 | `pages/jobs/resume` | ⭐⭐⭐ |
| C10 | 救援案例库 | `pages/emergency/cases` | ⭐⭐ |
| C11 | 调度记录 | `pages/emergency/dispatches` | ⭐⭐ |
| C12 | 应急部门对接 | `pages/emergency/depts` | ⭐⭐ |

### Sprint 3（W3）：对齐+协助 — 0 新页
- C 全部 12 页完成
- 协助 D 做剩余页面
- H5 ↔ 小程序对齐

**C 合计：5 + 7 = 12 页**

---

## 成员 D：产学研 + 活动品牌

### Sprint 1（W1）：展示型 — 5 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| D1 | 成果库列表 | `pages/achievements/list` | ⭐⭐ |
| D2 | 成果详情 | `pages/achievements/detail` | ⭐ |
| D3 | 研发难题广场 | `pages/challenges/list` | ⭐⭐ |
| D4 | 活动列表 | `pages/events/list` | ⭐⭐ |
| D5 | 品牌展示 | `pages/portfolios/list` | ⭐⭐ |

### Sprint 2（W2）：交互型 — 7 页
| # | 页面 | 路径 | 复杂度 |
|:--:|------|------|:--:|
| D6 | 课题攻关列表 | `pages/projects/list` | ⭐⭐ |
| D7 | 测试场地预约 | `pages/testsites/book` | ⭐⭐⭐ |
| D8 | 成果转化追踪 | `pages/transformations/track` | ⭐⭐⭐ |
| D9 | 活动详情+报名 | `pages/events/detail` | ⭐⭐ |
| D10 | 展会列表 | `pages/exhibitions/list` | ⭐⭐ |
| D11 | 展位申请 | `pages/exhibitions/booth` | ⭐⭐⭐ |
| D12 | 行业报告 | `pages/reports/list` | ⭐ |

### Sprint 3（W3）：收尾 — 0 新页
- D 全部 12 页完成
- H5 ↔ 小程序对齐

**D 合计：5 + 7 = 12 页**

---

## Sprint 4：全员联调（W4）

| 成员 | 任务 |
|------|------|
| **A** | 管理后台 E2E + 性能优化 |
| **B** | 小程序端全部页面 E2E + 微信审核提审 |
| **C** | H5 端全部页面对齐 + Bug 修复 |
| **D** | API 响应格式统一校验 + 文档更新 |

---

## 总览

| 成员 | Sprint1 | Sprint2 | Sprint3 | Sprint4 | 合计 |
|------|:--:|:--:|:--:|:--:|:--:|
| **A** | 6 页 | 4 页 | 15 页 | 联调 | **25** |
| **B** | 5 页 | 6 页 | 协助A | 联调 | **11** |
| **C** | 5 页 | 7 页 | 协助D | 联调 | **12** |
| **D** | 5 页 | 7 页 | 对齐 | 联调 | **12** |
| **合计** | **21** | **24** | **15+** | — | **60** |

---

## 复杂度说明

| 级别 | 含义 | 工时 |
|:--:|------|:--:|
| ⭐ | 纯列表读取，1 个 API，van-cell 渲染 | 0.5 天 |
| ⭐⭐ | 列表 + 详情 + 筛选，2-3 个 API | 1 天 |
| ⭐⭐⭐ | 表单提交 + 支付/预约 + 状态流转 | 1.5-2 天 |
