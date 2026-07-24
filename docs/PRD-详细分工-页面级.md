# PRD 详细分工 — 页面级（更新：基于实际进度）

> 最后更新: 2026-07-24 | 已有18页 | 缺24页 | 4人 × 3周

---

## 实际进度盘点

### 已完成页面（18 个，灰色 = 不需要新建）

| # | 页面 | 路径 | 对应系统 | 状态 |
|:--:|------|------|:--:|:--:|
| — | 首页 | `pages/home/index` | 通用 | ✅ |
| — | 登录 | `pages/login/index` | 通用 | ✅ |
| — | 注册 | `pages/register/index` | 通用 | ✅ |
| — | 消息中心 | `pages/messages/index` | 通用 | ✅ |
| — | 个人中心 | `pages/mine/index` | 通用 | ✅ |
| — | 实名认证 | `pages/mine/auth` | 通用 | ✅ |
| — | 个人信息 | `pages/mine/profile` | 通用 | ✅ |
| — | 业务大厅 | `pages/services/index` | 通用 | ✅ |
| — | 服务详情 | `pages/services/detail` | 通用 | ✅ |
| — | 服务申请 | `pages/services/apply` | 通用 | ✅ |
| — | 我的业务 | `pages/applications/index` | 通用 | ✅ |
| — | 案例库列表 | `pages/cases/index` | ④ 合规 | ✅ |
| — | 案例详情 | `pages/cases/detail` | ④ 合规 | ✅ |
| — | 研学列表 | `pages/study/index` | ⑤ 人才 | ✅ |
| — | 研学详情 | `pages/study/detail` | ⑤ 人才 | ✅ |
| — | 管理后台 | `pages/admin/index` | 后台 | ✅ 框架 |
| — | H5 需求详情 | `frontend/views/demand/Detail` | ② 供需 | ✅ |
| — | H5 需求发布 | `frontend/views/demand/Publish` | ② 供需 | ✅ |
| — | WebView | `pages/webview/index` | 通用 | ✅ |

### 缺失页面（24 个）

| 系统 | 缺页数 | 页面 |
|------|:--:|------|
| ① 会员资源 | 6 | 企业注册/审核状态/专家列表/专家详情/资源台账/资源详情 |
| ② 供需对接 | 4 | 需求大厅列表/竞标报价/我的需求/智能推荐 |
| ③ 产学研 | 6 | 成果库/成果详情/研发难题/课题攻关/测试预约/成果转化 |
| ⑤ 人才教育 | 6 | 培训课程/课程报名/我的证书/赛事列表/赛事报名/职位/简历/院校 |
| ⑥ 活动品牌 | 6 | 活动列表/活动报名/品牌展示/展会列表/展位申请/行业报告 |
| ⑦ 应急协同 | 4 | 应急资源/调度记录/救援案例/部门对接 |

**④合规：案例库已覆盖（cases页面），还缺政策资讯/合规知识库/团体标准/项目申报 4 页**

---

## 4 人分工（更新）

### 成员 A：供需对接 + 管理后台

| Sprint | 页面 | 路径 |
|:--:|------|------|
| W1 | 需求大厅列表 | `pages/demands/list` |
| | 需求详情 | `pages/demands/detail` |
| | 竞标报价 | `pages/demands/bid` |
| | 我的需求 | `pages/demands/mine` |
| W2 | 智能推荐 | `pages/match/recommend` |
| | 管理后台-看板 | `pages/admin/dashboard` |
| | 管理后台-审核 | `pages/admin/review` |
| W3 | H5对齐 + 全局搜索 + 联调 |

### 成员 B：会员资源 + 合规补充

| Sprint | 页面 | 路径 |
|:--:|------|------|
| W1 | 企业注册 | `pages/enterprise/register` |
| | 企业审核状态 | `pages/enterprise/status` |
| | 专家智库列表 | `pages/experts/list` |
| | 专家详情 | `pages/experts/detail` |
| W2 | 产业资源台账 | `pages/resources/list` |
| | 资源详情+预约 | `pages/resources/detail` |
| | 政策资讯列表 | `pages/compliance/news` |
| | 合规知识库 | `pages/compliance/knowledge` |
| | 团体标准库 | `pages/compliance/standards` |
| | 项目申报 | `pages/applications/submit` |
| W3 | H5对齐 + 联调 |

### 成员 C：人才教育 + 应急协同

| Sprint | 页面 | 路径 |
|:--:|------|------|
| W1 | 培训课程列表 | `pages/training/courses` |
| | 课程详情+报名 | `pages/training/enroll` |
| | 我的证书 | `pages/training/certificates` |
| | 赛事列表 | `pages/competitions/list` |
| W2 | 赛事报名 | `pages/competitions/register` |
| | 职位列表 | `pages/jobs/list` |
| | 简历管理 | `pages/jobs/resume` |
| | 院校展示 | `pages/colleges/list` |
| | 应急资源列表 | `pages/emergency/resources` |
| W3 | 救援案例库 | `pages/emergency/cases` |
| | 调度记录 | `pages/emergency/dispatches` |
| | 应急部门对接 | `pages/emergency/depts` |
| | H5对齐 + 联调 |

### 成员 D：产学研 + 活动品牌

| Sprint | 页面 | 路径 |
|:--:|------|------|
| W1 | 成果库列表 | `pages/achievements/list` |
| | 成果详情 | `pages/achievements/detail` |
| | 研发难题广场 | `pages/challenges/list` |
| | 活动列表 | `pages/events/list` |
| W2 | 课题攻关列表 | `pages/projects/list` |
| | 测试场地预约 | `pages/testsites/book` |
| | 成果转化追踪 | `pages/transformations/track` |
| | 活动详情+报名 | `pages/events/detail` |
| | 品牌展示 | `pages/portfolios/list` |
| | 展会列表 | `pages/exhibitions/list` |
| | 展位申请 | `pages/exhibitions/booth` |
| | 行业报告 | `pages/reports/list` |
| W3 | H5对齐 + 联调 |

---

## Sprint 总览

| Sprint | A | B | C | D | 合计 |
|:--:|:--:|:--:|:--:|:--:|:--:|
| W1 | 4 | 4 | 4 | 4 | **16** |
| W2 | 2 | 6 | 5 | 8 | **21** |
| W3 | 联调 | 联调 | 联调 | 联调 | — |

---

## 不做的页面

以下已被现有页面覆盖，**不需要重新开发**：

| 已有页面 | 覆盖了哪些 PRD 需求 |
|------|------|
| `pages/cases/*` | ④合规-企业案例库（列表+详情） |
| `pages/study/*` | ⑤人才-研学展示（课程详情） |
| `pages/services/*` | 通用服务入口 |
| `pages/applications/*` | 我的业务（需求/竞标/合同） |
| `frontend/views/demand/*` | ②供需-需求详情+发布（H5已做，小程序需补） |
