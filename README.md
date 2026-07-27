# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 12 大业务模块。

## 🚀 快速启动

### 环境准备
- **Go**: 版本 >= 1.22
- **PostgreSQL**: 版本 >= 15（可选，默认使用内存存储）
- **Node.js**: 版本 >= 16

### 一键启动（前端 + 后端）

```bash
# Windows — 双击运行，或命令行执行
start.bat

# Windows — 停止所有服务
stop.bat

# Linux/macOS/Git Bash
bash start.sh
```

> 首次运行会自动编译 Go 后端、安装 npm 依赖。
> 启动后访问:
> - 前端页面: http://localhost:5173
> - 管理后台: http://localhost:5173/admin
> - Swagger API 文档: http://localhost:8080/swagger/index.html

### 分别启动

**后端 API**：
```bash
# 构建 + 启动
go build -o drone-api.exe ./cmd/api
set ADMIN_DEV_MODE=true
drone-api.exe
```

**前端 H5**：
```bash
cd frontend && npm install && npm run dev
```
访问 `http://localhost:5173`。

**小程序**：
```bash
# 安装依赖
cd miniprogram && npm install && cd ..

# 打开微信开发者工具 → 导入 miniprogram 目录 → 编译
```

**管理后台**：
访问 `http://localhost:5173/admin`（Element Plus 桌面 UI）

**Docker 部署**：
```bash
docker-compose up -d   # 启动 PG + API
```

---

## 📱 功能特性

### 12 大业务模块

| 模块 | 能力 | 状态 |
|------|------|:--:|
| 用户认证 | 微信一键登录、Token 鉴权、角色权限（4级RBAC） | ✅ |
| 企业入驻 | 提交资料 → 协会审核 → 批准/驳回/补件 | ✅ |
| 需求大厅 | 多类型发布（巡检/植保/农药/租赁/清洗/其他）→ 竞标 → 双确认成交 → 争议 | ✅ |
| 招聘求职 | 职位发布、简历管理、投递、面试、录用 | ✅ |
| 社区内容 | 发帖、评论、举报、先审后发 | ✅ |
| 二手交易 | 商品发布、收藏、下架 | ✅ |
| 用工派遣 | 用工需求发布、报价竞标、人员分配 | ✅ |
| 合同签约 | 合同创建、模板管理、作废、签章回调 | ✅ |
| 培训认证 | CAAC 证书、大疆 UTC、人社等级、飞手认证、教练管理、课程报名 | ✅ |
| 无人机交易 | 整机买卖、维修服务、配件商城、交易订单 | ✅ |
| 保险服务 | 保单管理、年审提醒 | ✅ |
| 金融服务 | 分期贷款申请与审批 | ✅ |
| 资金托管 | 学费第三方存管、充值/冻结/释放/退款、交易记录 | ✅ |
| 信用评价 | 多目标评价、星级评分、审核流程 | ✅ |
| 场地预约 | 场地发布、时间预约、冲突检测 | ✅ |
| 行业资讯 | 政策法规、行业动态、合规预警、标准文件 | ✅ |

### 用户端功能（微信小程序）
- **首页**: 城市定位、轮播图、8 快捷入口、热门需求流、分类筛选
- **搜索**: 全局搜索需求与企业、搜索历史
- **需求大厅**: 列表筛选排序、详情、竞标报价、双确认完成、争议处理
- **社区**: 帖子列表、详情、评论、举报
- **消息中心**: 站内信列表、未读红点、标记已读
- **我的**: 个人信息、钱包余额、订单统计、功能菜单

### 后台管理（Web）
- **数据概览**: 指标卡片 + ECharts 趋势图 + 类型/状态分布饼图
- **企业审核**: 列表筛选、通过/驳回/补件操作
- **需求管理**: 需求列表、审核
- **举报管理**: 举报列表
- **行业资讯**: 新建、发布文章
- **评价审核**: 评价列表、通过/驳回审核流程
- **数据导出**: CSV 导出需求与企业数据（UTF-8 BOM，Excel 兼容）
- **Token 工具**: 角色 Token 生成

---

## 📂 项目结构

```
.
├── cmd/api/main.go              # 启动入口，组装依赖
├── internal/
│   ├── httpapi/                 # 路由、中间件、请求响应处理
│   │   ├── server.go            # 路由注册（120 条端点）
│   │   ├── auth.go              # Token 鉴权 + 公开路径白名单
│   │   ├── admin.html           # 管理后台 SPA（35KB，含 ECharts）
│   │   └── *.go                 # 各业务 handler
│   ├── service/                 # 业务规则、权限校验、状态流转
│   ├── repository/              # 数据持久化（PostgreSQL / 内存双模式）
│   │   ├── postgres/            # PostgreSQL 实现
│   │   └── memory/              # 内存实现（开发用）
│   ├── domain/models.go         # 业务实体与常量定义
│   ├── config/config.go         # 集中配置 + 验证 + 打印
│   ├── logger/logger.go         # 结构化日志（控制台 + 每日文件轮转）
│   ├── cache/cache.go           # 内存 TTL 缓存（60s 默认，5min 自动清理）
│   ├── middleware/middleware.go  # 输入消毒 + 统一错误格式
│   └── crypto/                  # 敏感字段 AES-256-GCM 加密
├── migrations/                  # 数据库迁移脚本（30 张表）
├── miniprogram/                 # 微信小程序（Vant Weapp 1.11）
│   ├── pages/                   # 27 个页面 + 7 个分包
│   ├── utils/                   # API 封装 + Token 刷新 + 常量
│   └── app.js/json/wxss         # 入口 + 设计系统（30 CSS 变量）
├── docs/                        # 项目文档
├── uploads/                     # 文件上传目录
└── README.md
```

---

## 🛠 技术架构

```
请求 → 幂等去重 → 限流(100/s) → 链路追踪 → 异常捕获 → 跨域控制 → Token鉴权 → 业务处理
```

| 层 | 职责 | 技术 |
|------|------|------|
| 小程序 | 微信原生 + Vant Weapp 1.11 | 27 页面 + 40+ 组件 |
| Web 管理后台 | 嵌入式 SPA | 原生 HTML/CSS/JS + ECharts 5 |
| API 层 | 路由、中间件、请求处理 | Go 标准库 net/http |
| 业务层 | 业务规则、权限校验、状态流转 | Go service |
| 数据层 | 双模式存储 | PostgreSQL 15+ / 内存 |
| 基础设施 | 配置、日志、缓存 | config + logger + cache |

## 🔒 安全能力

| 能力 | 说明 |
|------|------|
| 身份认证 | Bearer Token（HMAC-SHA256），15 分钟过期，支持刷新 |
| 角色控制 | 平台管理员 / 协会管理员 / 企业 / 个人，四级 RBAC |
| 数据隔离 | 企业只能查看自身合同和用工数据 |
| 操作审计 | 所有写操作自动记录操作者、时间、结果 |
| 接口防刷 | Token bucket 限流 100 次/秒，超限返回 429 |
| 幂等保护 | 写接口支持 `Idempotency-Key`，24 小时内去重 |
| 数据加密 | 手机号、身份证支持 AES-256-GCM 加密存储 |
| 输出脱敏 | 公开接口自动屏蔽联系方式、精确坐标 |
| 管理后台保护 | `ADMIN_DEV_MODE` 环境变量控制，生产环境自动 403 |

## 📊 质量指标

| 指标 | 数据 |
|------|:--:|
| API 端点 | 120 条 |
| 数据库表 | 30 张 |
| 小程序页面 | 27 页（7 分包） |
| 前端 API 对接 | 74 条（覆盖率 76%） |
| Go 源文件 | 53 个 |
| 总代码行数 | 5,300+ |
| 静态检查 | `go vet` 零告警 |

## 📄 环境变量

| 变量 | 必填 | 说明 |
|------|:--:|------|
| `AUTH_SECRET` | ✅ | JWT 签名密钥，至少 32 字节 |
| `DATABASE_URL` | — | PostgreSQL 连接串（不设则用内存存储） |
| `WECHAT_APPID` | — | 小程序 AppID |
| `WECHAT_APPSECRET` | — | 小程序 AppSecret |
| `ADMIN_DEV_MODE` | — | 设为 `true` 启用管理后台 |
| `LOG_DIR` | — | 日志目录，默认 `./logs` |
| `LOG_LEVEL` | — | 日志级别：debug/info/warn/error |
| `CORS_ORIGINS` | — | CORS 允许来源，逗号分隔 |
| `HTTP_ADDR` | — | 监听地址，默认 `:8080` |
| `SIGNING_SECRET` | — | 签章 Webhook 签名密钥 |

## 📄 文档

| 文档 | 说明 |
|------|------|
| [docs/api-contract.md](docs/api-contract.md) | API 契约与接口定义 |
| [docs/data-model.md](docs/data-model.md) | 数据模型与表结构设计 |
| [docs/architecture.md](docs/architecture.md) | 架构设计、角色权限矩阵、生产检查清单 |
| [docs/roadmap.md](docs/roadmap.md) | 迭代路线与历次审计记录 |
