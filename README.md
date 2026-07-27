# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 7 大业务系统。

> 🟢 **当前状态**: Go 后端 100% | 代码质量 A- (P0/P1清零) | 小程序 60% | 后台 35%  
> 📋 **团队协作**: [四人并行开发方案 PRD](docs/项目管理/PRD-四人并行开发方案.md)

---

## 🚀 快速启动

### 环境准备
- **Go** >= 1.22
- **PostgreSQL** >= 15（可选，默认使用内存存储）
- **Node.js** >= 16

### 一键启动（前端 + 后端）

```bash
# Windows — 双击运行
start.bat

# Windows — 停止
stop.bat

# Linux/macOS
bash start.sh
```

启动后访问:
- 前端页面: http://localhost:5173
- 管理后台: http://localhost:5173/admin
- Swagger API: http://localhost:8080/swagger/index.html
- **独立 API 文档** (无需服务): 打开 `docs/swagger-standalone.html`

### 分别启动

**后端 API**：
```bash
go build -o drone-api.exe ./cmd/api
set ADMIN_DEV_MODE=true
drone-api.exe
```

**前端 H5**：
```bash
cd frontend && npm run dev
```

**小程序**：
```bash
cd miniprogram && npm install
# 微信开发者工具 → 导入 miniprogram 目录 → 编译
```

**Docker 部署**：
```bash
docker-compose up -d   # PG + API
```

---

## 👥 四人协作 — GitHub 工作流

### 仓库

```
git@github.com:hhhhhhh72/UAV.git
```

### 分支结构

```
master              ← 生产就绪
  └── develop         ← 集成开发
        ├── feat/a-*  ← A: 小程序核心页 + 全部后台(27模块)
        ├── feat/b-pages  ← B: 纯小程序(14页)
        ├── feat/c-pages  ← C: 纯小程序(14页)
        └── feat/d-pages  ← D: 纯小程序(13页)
```

### 日常节奏

```bash
# 上午
git checkout feat/b-pages
git pull --rebase origin develop

# 写代码...

# 下午 5 点
git add -A && git commit -m "feat(模块): xxx"
git push origin feat/b-pages
# → 提 PR 到 develop → A 审查 merge
```

### 铁律

| # | 规则 |
|---|------|
| 1 | **每人只编辑自己分配的文件目录** |
| 2 | **共享层冻结** — `components/`、`App.vue`、`pages.json`、`frontend/` 仅 A 可改 |
| 3 | **先 Pull 再 Push** — 每次 git pull --rebase |
| 4 | **Go 代码红线** — 禁止 `_, _ := json.Marshal`、禁止裸 `return err` |
| 5 | **禁止直推 master/develop** — 必须 PR + 1 人 Approve |
| 6 | **🚨 禁止擅自 push GitHub** — 任何人 push 前须经 A 确认，AI 写代码须人工审核 |

> 📋 完整分工、Sprint 计划、AI Prompt 模板 → [PRD](docs/项目管理/PRD-四人并行开发方案.md)

---

## 📱 7 大业务系统

| 系统 | 能力 | 后端 | 小程序 | 后台 |
|------|------|:--:|:--:|:--:|
| **供需对接** | 需求大厅、竞标报价、双确认成交、搜索匹配 | ✅ | 🟡 | ✅ |
| **会员资源** | 企业入驻、专家库、资源台账、品牌展示 | ✅ | 🟡 | 🟡 |
| **人才教育** | 培训课程、CAAC 证书、赛事、职位、院校研学 | ✅ | 🟡 | 🟡 |
| **产学研** | 科技成果、研发难题、课题攻关、测试场地、成果转化 | ✅ | 🟡 | 🆕 |
| **应急协同** | 应急资源调度、救援案例、部门对接 | ✅ | 🟡 | 🆕 |
| **活动品牌** | 行业活动、展会、行业报告 | ✅ | 🟡 | 🆕 |
| **合规政策** | 政策资讯、合规知识库、团体标准、优秀案例 | ✅ | 🟡 | 🟡 |

---

## 📂 项目结构

```
.
├── cmd/api/main.go              # 启动入口
├── internal/
│   ├── httpapi/                 # 路由 + 中间件 + Handler (212 API)
│   ├── service/                 # 业务规则 + 权限校验 + 状态流转
│   ├── repository/
│   │   ├── postgres/            # PostgreSQL 持久化 (66 表)
│   │   └── memory/              # 内存存储 (开发用)
│   ├── domain/                  # 业务实体 + 常量
│   ├── config/                  # 集中配置
│   ├── logger/                  # 结构化日志 (slog)
│   ├── cache/                   # TTL 内存缓存
│   ├── middleware/              # 输入消毒 + 统一错误
│   └── crypto/                  # AES-256-GCM 加密
├── frontend/                    # Vue 3 + Element Plus 管理后台
│   └── src/views/admin/         # 27 管理模块 (8 已完成)
├── miniprogram/                 # 微信小程序 (uni-app + Vant Weapp)
│   ├── pages/                   # 45 页面 (7大系统)
│   ├── components/              # 共享组件
│   └── utils/                   # API 封装 + Token 刷新
├── migrations/                  # 66 张表迁移脚本
├── docs/                        # 项目文档 + Swagger + PRD
└── prototypes/                  # HTML 原型 (首页+商家页)
```

---

## 🛠 技术架构

```
请求 → 幂等去重 → 限流(100/s) → 链路追踪 → 异常捕获 → 跨域 → Token鉴权 → 业务处理
```

| 层 | 技术 |
|------|------|
| 微信小程序 | uni-app + Vant Weapp 1.11 |
| 管理后台 | Vue 3 + Element Plus 2.14 + ECharts 6 + Pinia |
| API 层 | Go 标准库 net/http + gorilla/mux |
| 业务层 | Go service (接口依赖注入) |
| 数据层 | PostgreSQL 15+ / 内存 (双模式) |
| 基础设施 | config + slog + cache + crypto |

---

## 🔒 安全能力

| 能力 | 实现 |
|------|------|
| 身份认证 | Bearer Token (HMAC-SHA256), 15min 过期, 支持刷新 |
| 四级 RBAC | platform_admin / association_admin / enterprise / individual |
| 数据隔离 | 企业仅见自身合同+用工数据 |
| 操作审计 | 所有写操作记录操作者+时间+结果 |
| 接口防刷 | Token bucket 100次/秒 → 429 |
| 幂等保护 | Idempotency-Key, 24h 去重 |
| 数据加密 | AES-256-GCM 加密手机号+身份证 |
| 输出脱敏 | 公开接口自动屏蔽联系方式+坐标 |
| 管理后台保护 | ADMIN_DEV_MODE 控制, 生产环境 403 |

---

## 📊 质量指标

| 指标 | 数值 |
|------|:--:|
| API 端点 | **212 条** |
| 数据库表 | **66 张** |
| 小程序页面 | **45 页** (vue, 从59精简, 功能不减) |
| 管理后台模块 | **27 模块** (8 已完成) |
| Go 源文件 | **80+** |
| 测试通过率 | **100%** (92 HTTP case) |
| 测试覆盖率 | **45.8%** |
| 静态分析 | `go vet` ✅ 零告警 |
| Lint 工具 | `.golangci.yml` (11 linter) |
| 前端构建 | Vue 3 + Element Plus ✅ |
| 代码质量 | P0/P1 已清零 (0 JSON忽略 + 0裸error) |

---

## 📄 文档

| 文档 | 说明 |
|------|------|
| [四人并行开发方案](docs/项目管理/PRD-四人并行开发方案.md) | 零冲突分工 + Git 策略 + AI Prompt 模板 |
| [API 契约](docs/接口文档/API契约.md) | 212 条端点完整定义 |
| [Code Review Checklist](docs/开发规范/Code-Review-Checklist.md) | 10 大维度 30+ 检查项 |
| [系统架构](docs/系统架构/架构总览.md) | 架构设计 + 中间件链 + 角色权限矩阵 |
| [Swagger 独立页](docs/swagger-standalone.html) | 离线 API 文档 (单文件, 无需服务) |
| [编码规范](docs/开发规范/编码规范.md) | Go + Vue 代码风格指南 |
| [首页原型](prototypes/首页-贴吧式-原型.html) | 贴吧式任务大厅设计 |
| [商家页原型](prototypes/商家页-原型.html) | 同城好店设计 |

---

## 📄 环境变量

| 变量 | 必填 | 说明 |
|------|:--:|------|
| `AUTH_SECRET` | ✅ | JWT 签名密钥，至少 32 字节 |
| `DATABASE_URL` | — | PG 连接串（不设则内存存储） |
| `WECHAT_APPID` | — | 小程序 AppID |
| `WECHAT_APPSECRET` | — | 小程序 AppSecret |
| `ADMIN_DEV_MODE` | — | `true` 启用管理后台 |
| `LOG_DIR` | — | 日志目录，默认 `./logs` |
| `LOG_LEVEL` | — | debug/info/warn/error |
| `CORS_ORIGINS` | — | CORS 允许来源，逗号分隔 |
| `HTTP_ADDR` | — | 监听地址，默认 `:8080` |
