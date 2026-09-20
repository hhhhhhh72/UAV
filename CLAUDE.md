# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 7 大业务系统。

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 API | Go 1.25+，标准库 net/http，`mux.HandleFunc` 注册点 **508 处**（只数源码、不数测试；循环体里的按 1 处计），管理端路由探针 204 条（见 `perm_routes_test.go`），311 个 Go 文件（**128 源码 + 183 测试**） |
| 数据库 | PostgreSQL 16（生产） / 内存存储（开发），**92 张存活表**（96 次 CREATE TABLE 减去被 drop 的 shops/demand_bids/association_members；**119 组迁移，238 个 SQL 文件**） |
| 部署 | Docker 多阶段构建 + docker-compose（PG + API 双容器） |
| CI/CD | GitHub Actions（build + vet + test + integration） |

> 上表数字是快照，会随开发漂移。复核方式：
> `(Get-ChildItem migrations -Filter *.up.sql).Count`（迁移组数）；
> `node .tools/gen-perm-routes.cjs`（管理端路由探针数，过期时 `perm_routes_test.go` 会直接报错）；
> `(Get-ChildItem internal\httpapi -Recurse -Filter *.go | Where-Object { $_.Name -notlike '*_test.go' } | Select-String -SimpleMatch 'mux.HandleFunc(').Count`（路由注册点）；
> `(Get-ChildItem internal,cmd -Recurse -Filter *.go)`（Go 文件数）。

## 项目结构

```
.
├── cmd/api/main.go               # 启动入口，组装依赖
├── internal/
│   ├── httpapi/                   # 路由、中间件、Handler（解析请求→调Service→respond）
│   │   ├── server.go              # Server struct + 路由注册 + 中间件链
│   │   ├── auth.go                # Token 签发/验证 + 鉴权中间件
│   │   ├── auth_wechat.go         # 微信 code2Session + Token 刷新
│   │   ├── admin_handler.go       # 管理后台 Token（ADMIN_DEV_MODE）
│   │   ├── compat_routes.go       # 旧版 API 兼容 (/api/auth/*)
│   │   ├── h5_compat.go           # H5/Vue 前端兼容层（auth 路由生产注册，JSON 文件路由仅 dev）
│   │   └── *.go                   # 各业务 Handler（batch1-3 + biz + phase3）
│   ├── service/                   # 业务规则、权限校验、状态机流转
│   ├── repository/                # 数据持久化
│   │   ├── repositories.go        # Repository 接口定义（61 interface）
│   │   ├── postgres/              # PostgreSQL 实现（pgxpool）
│   │   └── memory/                # 内存实现（开发用，sync.RWMutex）
│   ├── domain/                    # 业务实体与常量（96 个 struct，含 models_batch*/models_new）
│   ├── config/config.go           # 集中配置 + 验证 + 脱敏打印
│   ├── logger/logger.go           # 结构化日志（slog + 每日文件轮转）
│   ├── cache/cache.go             # 内存 TTL 缓存（60s 默认，5min 自动清理）
│   ├── middleware/middleware.go    # 输入消毒 + 统一错误格式
│   └── crypto/                    # AES-256-GCM 加密 + 脱敏函数
├── migrations/                    # 119 组迁移（238 个 SQL 文件，92 张存活表）
├── docs/                          # 项目文档（33 份 Markdown，中文）
├── icons/                         # 15 个 SVG 图标
└── docker-compose.yml
```

## 分层约束（铁律）

这是最重要的编码规范，每次修改必须遵守：

```
Handler     → 只做：解析请求、调 Service、调 respond/fail
              禁做：SQL 查询、业务规则判断、直接读写库

Service     → 只做：角色校验、归属校验、状态机迁移、调 Repository 接口
              禁做：http.Request、http.ResponseWriter、JSON 编解码、环境变量

Repository  → 只做：SQL 执行、数据映射、返回 Domain 对象
              禁做：业务规则、权限判断、HTTP 相关
```

## 命名约定

| 层级 | 约定 | 示例 |
|------|------|------|
| 路由 | `/api/v1/资源复数` | `/api/v1/demands` |
| Handler | `func (s *Server) verbNoun(...)` | `createDemand` |
| Service | `type NounService struct` | `DemandService` |
| Repository 接口 | `type NounRepository interface` | `DemandRepository` |
| PG 实现 | `type nounRepo struct` (小写开头) | `demandRepo` |
| 迁移文件 | `00000X_description.up.sql` | `000001_init.up.sql` |

## 中间件链

```
请求 → 限流(100/s, 按RemoteAddr) → 链路追踪(requestID) → Panic恢复 → 安全头
  → CORS → Token鉴权(白名单+可选解析) → 幂等去重(24h, actor命名空间)
  → 管理端门禁(/api/v1/admin/* 仅admin) → 输入消毒(JSON去HTML标签) → 业务处理
```

## 响应规范

```go
// 成功 — 必须用这些函数，禁止手动拼 JSON
respond(w, r, statusCode, data)
paginatedRespond(w, r, items, totalCount)

// 失败
fail(w, r, statusCode, err)
```

## 双存储模式

| 环境 | 存储 | 触发条件 |
|------|------|------|
| 开发 | 内存 | `DATABASE_URL` 未设置 |
| 生产 | PostgreSQL | `DATABASE_URL` 已设置 |

## 角色权限（4 级 RBAC）

| 角色 | 权限 |
|------|------|
| `platform_admin` | 全部管理权限 |
| `association_admin` | 企业审核/内容管理 |
| `enterprise` | 发布需求/招聘/合同 |
| `individual` | 接单/求职/交易 |

> **协会内部 8 级角色已下线**（原 `association_members` 表：会长/副会长/秘书长/部门负责人/
> 普通会员/副会长单位/合作院校/访客）。它源自一份 .doc 需求但从未启用：生产 0 行数据、
> 前端零调用、8 个角色里只有 `partner` 参与过一次判定（会长与访客权限完全等价）。
> 实体/仓储/服务/handler 已删除，表由迁移 `000113` 删除。

## 7 大业务系统

| 系统 | 核心子模块 |
|------|------|
| ①会员生态资源管控 | 会员注册/专家智库/产业资源台账/人才资源库 |
| ②产业供需智能对接 | 需求大厅/供应展示/意向对接/工单闭环/智能匹配/资源池 |
| ③产学研协同创新 | 科技成果库/研发难题广场/课题攻关/测试预约/成果转化追踪 |
| ④合规政策服务 | 政策资讯/合规知识库/团体标准库/项目申报/企业案例库 |
| ⑤人才教育与产教融合 | 培训认证/赛事管理/招聘求职/院校展示/校企共建 |
| ⑥活动与品牌服务 | 活动管理/会员品牌展示/展会排期/行业报告发布 |
| ⑦低空应急资源协同 | 应急资源/一键调度/救援案例库/部门对接/联合演练 |

## Token 认证

- **格式**: 标准 JWT (HS256) 为主（`IssueJWT`），兼容旧式两段 `payload.sig` 格式
- **过期**: Access Token 15分钟，Refresh Token 7天
- **刷新**: 轮转刷新令牌——先落库新令牌、成功后再撤销旧令牌（防 Store 失败锁号），库中只存 SHA-256 哈希

## 验收标准

每个功能模块完成后必须：

```bash
go build ./...          # 编译通过
go vet ./...            # 零告警
go test ./internal/...  # 全部 PASS
```

## 环境变量

| 变量 | 必填 | 说明 |
|------|:--:|------|
| `AUTH_SECRET` | ✅ | JWT 签名密钥，至少 32 字节 |
| `DATABASE_URL` | — | PostgreSQL 连接串（不设则用内存存储） |
| `WECHAT_APPID` | — | 小程序 AppID |
| `WECHAT_APPSECRET` | — | 小程序 AppSecret |
| `WECHAT_PAY_MCHID` | — | 微信支付商户号（与下面四项**同生共死**：缺一项＝真实支付未开通） |
| `WECHAT_PAY_API_V3_KEY` | — | APIv3 密钥，**必须正好 32 字节**，用于回调解密 |
| `WECHAT_PAY_CERT_SERIAL` | — | 商户 API 证书序列号（请求签名用） |
| `WECHAT_PAY_PRIVATE_KEY_PATH` | — | 商户私钥 apiclient_key.pem 路径 |
| `WECHAT_PAY_NOTIFY_URL` | — | 支付结果回调地址，生产必须是 https 公网地址 |
| `WECHAT_PAY_REFUND_NOTIFY_URL` | — | 退款结果回调地址；**可选**，留空由 `WECHAT_PAY_NOTIFY_URL` 推导（`.../notify` → `.../refund-notify`） |
| `ADMIN_DEV_MODE` | — | 设为 `true` 启用开发令牌 |
| `ENCRYPTION_KEY` | — | AES-256-GCM 加密密钥 |
| `CORS_ORIGINS` | — | CORS 允许来源，逗号分隔 |
| `HTTP_ADDR` | — | 监听地址，默认 `:8080` |

## 前端项目

| 项目 | 位置 | 技术栈 | 规模 |
|------|------|--------|------|
| 微信小程序 | `miniprogram/` | uni-app + Vue3 `<script setup>` + 自研 u- 组件库 | **107 页**（主包 31 + 分包 76），5 Tab，6 分包（`node scripts/check-miniprogram-routes.cjs` 可复核注册页数并查死链） |
| Web 管理后台 | `frontend/` | Vue 3 + Arco Design Vue + ECharts | Admin SPA（**46 条路由** + 聚合页） |

**小程序设计规范**:
- 品牌色 `#0A66C2`（深空蓝），辅色 `#1DD4A8`（青绿）
- 全局 CSS 变量定义在 `App.vue` 的 `page` 选择器中
- 输入框: `bg=#fafafa` `radius=24rpx`，按钮: `radius=50rpx` + `box-shadow`
- 不使用 emoji 图标，用 CSS 绘制或文字标签

**小程序 API 调用注意**:
- `request.js` 自动 unwrap Go 后端的 `{ data: {...} }` 响应包（分页响应保留 `total`）
- 微信静默登录用 `POST /api/v1/auth/wechat/login`（生产路由，返回蛇形 `{ access_token, refresh_token, user }`）
- 密码登录用 `POST /api/auth/login`（H5 兼容层，生产注册，bcrypt 校验）
- Token 存储: `authStorage.setTokens(accessToken, refreshToken)` + `uni.setStorageSync('user', ...)`；刷新轮转，须保存新 refreshToken
- 主流程是微信静默登录（`App.vue` → `wx.login()` → `/api/v1/auth/wechat/login`），密码登录只是备用

## 关键踩坑记录

| 问题 | 根因 | 修复 |
|------|------|------|
| 注册 500: `duplicate key violates unique constraint "users_wechat_openid_key"` | 手机注册时 `wechat_openid` 为空字符串，第二次 `""` 违反 UNIQUE | 设置 `WechatOpenID: "phone:"+body.Phone` 确保唯一 |
| H5 兼容路由 404 | `/api/auth/*` 曾只在 `ADMIN_DEV_MODE=true` 时注册，生产 404 | **已修复**：auth 路由（login/register/me/refresh/logout + GET services/config）已无条件生产注册，JSON 文件路由仍 dev-only |
| 登录 401 | 小程序调 `/api/v1/login` 不存在 | 改为 `/api/auth/login` |
| 登录后 Token 不匹配 | 旧代码只存 `token`，新 API 返回 `accessToken` | 改用 `authStorage.setTokens()` |
| 小程序全部请求 `ERR_CONNECTION_TIMED_OUT` | `miniprogram/utils/config.js` 的 `BASE_URL` 写死旧局域网 IP（192.168.5.141），DHCP 换网后 IP 变了 | 用 `ipconfig` 查当前 WLAN IP（如 192.168.5.19），改 `BASE_URL` 后重新编译；真机调试需手机与电脑同一 WiFi |
| 微信一键登录变成"同一个号/多出 dev-fixed 用户" | 打包 tar 未排除 `.env`，本地 `.env` 覆盖服务器 `.env` → `WECHAT_APPID/APPSECRET` 丢失 → code2Session 失败 → `adminDevMode()` 兜底 `openid=dev-fixed` → 每次失败都建共享账号 | 打包命令必须 `--exclude='.env'`；部署后验证 `docker exec uav-api-1 env \| grep WECHAT` 非空；恢复 `~/UAV/.env` 中 WECHAT_APPID/WECHAT_APPSECRET 后 `docker compose up -d api` |
| `GET /api/v1/me` 的 demand_count 是全平台计数 | `me()` 用 `List(空 filter)` 统计全部已发布需求 | **已修复**：新增 `DemandRepository.ListByPublisher` 只统计本人需求（含回归测试） |
| refresh 轮转可能锁号 | 旧实现先 `Revoke` 旧令牌再 `Store` 新令牌，Store 失败即账号锁死；且签发的是旧式两段 Token | **已修复**：先落库新令牌、成功后再撤旧；签发统一 `IssueJWT` |
| 短信验证码可在线爆破 | 错误码无尝试次数限制、比较非常量时间 | **已修复**：5 次错误作废验证码 + `subtle.ConstantTimeCompare` |
| h5ImageProxy 开放重定向 | 任意 http/https URL 直接 302 跳转 | **已修复**：白名单（localhost/127.0.0.1/BASE_URL 域名）外一律 403 |
| `middleware.SanitizeBody` 是空壳 | 只查 Method/Content-Type 就放行，且未挂载 | **已修复**：实现真实 JSON 消毒（去 HTML 标签、password 保真、1MiB 上限）并挂载进中间件链 |
| **定时备份静默失败两天**（2026-09-17 发现） | 从 Windows 打包 tar 部署时 `deploy/*.sh` 丢了可执行位；且 `core.autocrlf=true` 让工作区里的 `deploy/db-backup.sh` 是 CRLF，覆盖服务器上正常的 LF 版本后 `set -euo pipefail` 被读成 `pipefail\r`，cron 只留两行 Permission denied | **已修复**：① 仓库 `.gitattributes` 统一 `* text=auto eol=lf`、`git config core.autocrlf false` 并把工作区重新规范化；② 每次部署后必须 `chmod +x deploy/*.sh *.sh` —— Windows 打包会丢可执行位，**这一步不能省**；③ 新增 `deploy/ops-status.sh`（每 10 分钟快照）+ 探活覆盖备份新鲜度/磁盘/证书/容器；④ 新增 `deploy/restore-drill.sh`（每周把最新备份还原到临时库验证），因为「文件存在」不等于「能恢复」 |
| **平台只能收钱不能退真钱**（2026-09-18 补齐，未上线） | `internal/wechatpay` 此前**完全没有退款能力**：真实支付一旦开通，用户微信付的钱在退款时只会把平台内余额加回去，真金白银不会回到微信钱包。这在消保法口径下站不住（「退款」通常指原路退回），而且当时连「能不能退」都没人知道 | **已修复**：补齐整条退款链路。**关键设计**：退款的对象是**某一笔充值单**而不是某个订单——平台的钱是以「充值进托管余额」的形式进来的，微信退款 API 本就要求 `out_trade_no` + 累计不超过原单金额，正好对上。资金三件套全部复用已有词汇：发起 `Freeze`（余额→冻结，余额不足则发起就失败，不会出现「退了但扣不到」）→ 成功 `Withdraw`（新增动作，冻结扣掉＝钱离开平台）→ 失败 `Refund`（解冻退回余额）。三道幂等：额度由 `refunded_fen` 条件更新（库级，000118）、状态推进是 CAS、出账流水由 `idx_escrow_once_per_ref` 兜底（000119 起覆盖 withdraw）。**「微信明确拒绝」与「结果未知」严格区分**：前者安全回滚，后者**绝不回滚**（微信可能已受理，回滚＝钱退出去还把余额还给用户）。回调沿用「只做触发、一律主动查单」。发起仅平台管理员；测试 10 例覆盖超额/余额不足/微信拒绝/结果未知/重复确认/关闭解冻 |
| **演练时把假警报推给了真实群**（2026-09-18，我造成的） | 我在做告警链路演练时，`ops-status.sh` 默认写**生产快照** `/var/www/ops-status.json`，而 cron 每 10 分钟就会读它。于是「故意写坏的快照」被正在运行的告警脚本读到，**一条假警报推到了真实群里**（`运维快照异常：top_level`），当时系统一切正常。两个错误叠加：变异测试写了生产文件 + 那一刻 `alert.sh` 的分项列表还没同步 `jobs`（列表不全时只会兜成没指向性的 `top_level`） | **已修复**：① 新增 `deploy/verify-alerts.sh` —— 在 **mktemp 临时目录**里跑完整演练矩阵（健康/备份过期/冷却/恢复/磁盘/资金/任务心跳/证书/容器/快照过期/快照缺失，共 11 项），全程 `OUT`/`STATE`/`LOG`/`ENVFILE` 都指向临时文件，**碰不到生产快照与生产配置**，脚本末尾还会用 mtime 自证「生产快照未被改动」；默认干跑不推送，`--send` 才真发一条。② 把「演练要用临时文件」固化成脚本 —— 比写一句提醒可靠，因为我当时以为自己在做的正是这件事。③ 演练脚本自身第一次跑出 3 个 FAIL，全是它的 bug（`check()` 把冷却与恢复依赖的 STATE 一起删了；快照缺失的断言写的是签名而非人看的文本），已修到 11/11 通过 |
| **「设置好了但没人确认它真的在跑」**（2026-09-18 补，第三次同类问题） | 备份静默失败两天、告警填错 URL 假绿、保洁周任务从没跑过 —— 三次都是同一个病根：**设置了但没人确认它在运行**。原有的检查全是**结果类**（备份文件在不在、新不新鲜、能不能还原），它们看不出「任务从没跑过」，因为根本没有结果可看 | **已修复**：`ops-status.sh` 新增 `jobs` 一节做**过程类**检查（心跳），并折叠进顶层 `ok`。只给**没有结果产物**的任务做心跳：`disk-hygiene`（跑完只写日志）与 `alert`（健康时完全静默，另盖 `.alert-heartbeat` 文件，覆盖写不增长）。有结果产物的不加：备份看 `backup.file` 年龄、还原演练看 drill 年龄、快照看 `generated_epoch`——那些更强（证明**结果产出了**，不只是脚本跑了）。**`alert.sh` 的分项列表必须与 ops-status 保持同步**，否则新分项失守时只会兜成没指向性的 `top_level`。已做变异验证：保洁日志改成 3 天前 → `ok=false` 且告警正确报出 `jobs`；恢复 → `ok=true` |
| **磁盘又涨回去**（2026-09-18，用户发现） | 会话开始时 26%，几轮部署后到 34%。三个原因叠在一起：① `deploy/disk-hygiene.sh` 是**周任务**（`0 5 * * 0`）且 9/17 才装，**到周五都没遇到过周日 → 一次都没跑过**；② 它完全没管 `/tmp`，而发布把 tar 包与解包目录丢在那里；③ 更根本的是**部署流程自己不留手**——每次 `docker compose build api` 都攒构建缓存，每次发布留一个包，全靠事后保洁兜底 | **已修复**：① 保洁改为**每天** 05:00；② 新增 `/tmp` 一节，且通配写全（历史上前端包叫 `admin-dist-notify.tgz`/`admin-dist-0911.tgz`…，只匹配 `admin-dist.tar.gz` 一个名字漏了 15 份共 508MB）；③ **新增 `deploy/deploy-api.sh` 把发布固化成脚本**，内含健康检查、schema 变化、重启次数、异常条数，并在检查通过后 `rm -f "$PKG"` 自清理（失败则保留包供排查）；④ `deploy-web.sh` 同样自清理解包目录。实测：磁盘 **34% → 24%**（释放 4.3GB），`/tmp` **693M → 1.4M**，构建缓存 5.4G → 2.0G。`--keep-storage` 在新 Docker 上已改名 `--reserved-space`，脚本两个都试以免升级后静默失效 |
| **快照有了没人看**（2026-09-18 补） | `ops-status.sh` 把磁盘/备份/容器/证书/资金不变量全算进了 `/var/www/ops-status.json`，但**没有任何东西会主动通知人** —— 它是「你想起去看才看得到」。9/16 那次备份静默失败两天，正是这个形态：指标在，没人看。GitHub Actions 探活失败虽会发通知，但那不是每天会打开的地方 | **已修复**：新增 `deploy/alert.sh`（cron 每 10 分钟的第 2/12/22/32/42/52 分钟跑一次，比 ops-status 晚 2 分钟）。读快照 → 判定 → 推送群机器人。**去重与冷却**：同一故障签名最多每 6 小时提醒一次，故障消失立即发「已恢复」；没有这一层，持续故障会每 10 分钟刷屏最后被无视。**快照本身就是告警项**：超过 60 分钟没更新即报（生成脚本自己挂了，正是备份那次的形态）。**未配置通道时明确记日志**「告警通道未配置」，绝不静默。启用只需在 `/root/UAV/alert.env` 填一行 `ALERT_WEBHOOK=<企业微信群机器人 webhook>`（钉钉同款；飞书加 `ALERT_FORMAT=feishu`）。7 种情形已用临时路径逐条实测（正常/异常/冷却/换故障/恢复/快照缺失/快照过期），并用本地 HTTP 监听验证了真实 POST 的报文与 `推送 HTTP 200`。**2026-09-18 已配置生产群机器人并实测连通**（`{"errcode":0,"errmsg":"ok"}`）——注意「群机器人」在企业微信里已改名为「**消息推送**」，且**只能在群聊里创建、不在工作台/管理后台**；工作台里那个「智能机器人」是另一类东西，**没有 webhook**（社区里同类提问很多）。另外补了一道防「假绿」的校验：`send()` 会先校验 webhook 的形状（必须是对应厂商的群机器人地址），再**解析返回体里的 errcode**而不是只看 HTTP 码——因为把管理后台的机器人资料页链接误当成 webhook 时，那个页面同样返回 200，脚本会记一条「推送 HTTP 200」就以为发出去了，群里却什么都没有（与备份静默失败同一类问题）。日志里的 key 一律打码 |
| **日报做成了「平台经营数据」，用户要的是「我今天干了什么」**（2026-09-18） | 第一版把日报做成了经营数据看板（新增用户/订单/托管余额）：链路上线了、也真的推到了群里，但用户一句「日报内容错了，我要的是我每天的工作内容」把它否掉 —— 负责人要的是**工作汇报**（今天推进了什么、走到哪个节点），不是平台指标。这是我理解需求时默认成了「平台日报」，而他要的是「我的工作日报」 | **已修复**：改成按汇报要求生成工作日报（20:00 前、≤300 字、进度推进类写进度节点、群里直接发纯文字）。**生成在服务器、发送在服务器、与开发机是否开机无关**：服务器每天 19:30 由 cron 跑 `deploy/work-report.sh` —— 先 `git fetch` 一份**只含提交与目录、不含文件内容**的裸库（`--filter=blob:none`，**1.4MB**；本机 `.git` 是 524MB，每天 fetch 约 1 秒），再由 `deploy/work_report.py` 生成正文。正文按模块（资金/支付/后端/管理后台/小程序/运维/文档）归类当日提交，标题里的 `;` 拆成独立工作项，超 300 字预算时**整条丢弃而不是把句子切一半**；进度节点不是猜的（今日是否部署过 = `docker inspect uav-api-1` 的 `StartedAt`，补看历史日期时不写这行，否则「今日已部署」会堂而皇之挂在 09-17 的日报上）。**取不到最新提交就直出不发** —— 用旧数据发一份看起来正常的日报正是「假绿」。**送达检查**：心跳文件只在推送成功后覆盖写，`ops-status.sh` 的 jobs 一节看它的年龄（30h ＝ 24h 间隔 + 6h 抖动余量）。开发机上的 `scripts/daily-report.ps1` 退化成**薄壳**（默认拉服务器 `--dry-run` 的正文做预览/写文件/剪贴板，`-Send` 才真发），**不重复实现生成逻辑**；服务器版经营日报 `daily-report.sh` 已整份退役删除。为什么不直接在本机 POST：本机 schannel 拿不到 TLS 凭据，`curl`/`Invoke-WebRequest` 一律报 `(35) SEC_E_NO_CREDENTIALS`，任何 https 都发不出去 |
| **售后「确认收到退货」可重复退款**（2026-09-18，000116 之后剩下的最后一条缝） | `ConfirmReturnReceived` 是**先退款再改状态**的（`phase3.go` 里 `refundForAftersale` 在 `UpdateAftersale` 之前，有意为之：钱动不了就保持 `returned` 供重试，否则会出现「状态已结案、钱还冻着」的死局）。代价是并发确认（两个管理员/卖家同时点「确认收到退货」）时两边都读到 `AftersaleStatus='returned'`，并双双通过 `refundForAftersale` 的 `HasFrozen`+`HasRefunded` 两道 check-then-act 查询。该函数两个分支里「钱已放给卖家」那一支走 `Transfer`（已被 000116 兜住），「钱还在冻结里」这一支走 `Refund`，**此前没有任何库级兜底**。后果不是凭空生钱，而是把同一笔冻结款重复退回去 → 付款方 `frozen` 被多扣，其余仍处于「已付款冻结」的订单再也释放不出来 | **已修复**：migration `000117` 加部分唯一索引 `idx_escrow_refund_once_per_order (from_user, tx_type, reference_type, reference_id) WHERE tx_type='refund' AND reference_type='trade_order' AND reference_id <> ''`。**范围只限 `trade_order`**：一个订单只有一个售后（`AftersaleStatus` 是订单上的单字段），「一单一退款」正是正确的不变量；而培训课程那条键是 `(user, course)`，同一人可合法地反复冻结/退款，加全局唯一会挡住正常业务（`phase3.go:464` 有明确注释）。`commitFundMove` 把 23505 翻译成幂等成功。已做库级变异验证：有索引 PASS，`DROP INDEX` 后 RED 并报「refund 流水应为 1 条，实际 8 条」 |
| **放款「只放一次」在库层面无保障**（2026-09-18） | `escrow_transactions` 上唯一的唯一索引是 `idx_escrow_external UNIQUE (channel, external_txn_id) WHERE external_txn_id <> ''` —— **只覆盖真实渠道入金**。release/transfer 走内部渠道、`external_txn_id` 恒为空串，这条部分索引根本不生效。于是「同一付款方对同一业务单只放款一次」只剩下应用层一道 check-then-act，而 `EscrowService.Release` 的注释自己写明了：「PG 的 Release 只校验 frozen_fen 足够、**没有 (from,ref) 去重，本次查询是唯一防线**」。`completeEnrollment` 的「completed 幂等补齐」分支又没有状态 CAS，两个并发重试会双双通过 `HasReleased=false`；付款方若还有其它冻结资金，两次释放都会成功 → 机构双倍入账 | **已修复**：migration `000116` 加部分唯一索引 `idx_escrow_once_per_ref (from_user, tx_type, reference_type, reference_id) WHERE tx_type IN ('release','transfer') AND reference_id <> ''`。Release 的余额调整与流水写入在同一事务里，所以第二次插入会带着整个事务回滚——付款方一分钱不会被多扣；`postgres.commitFundMove` 把 23505 翻译成幂等成功并回查真实流水。**只覆盖 release/transfer**：freeze/refund 的重复是设计允许的（报名被拒→回滚退款→重试，生产上确有同一 (user,course) 多次 freeze/refund）。已在生产 PG 上用临时库做**库级变异验证**：有索引 PASS，`DROP INDEX` 后 RED 并报「release 流水应为 1 条，实际 8 条」 |
| **42 个端点把 pgx 原文漏给客户端**（2026-09-18 全量详情扫描） | `fail()` 的脱敏只在 `status >= 500` 时生效，**4xx 分支原样透传** `err.Error()`。仓储层没把 `pgx.ErrNoRows` 翻译成 `repository.ErrNotFound` 的路径，于是客户端收到 `404 {"message":"no rows in result set"}`、`"product xxx not found: no rows in result set"`——违反本函数自己的注释「不回显内部错误细节（可能含 SQL/实现信息）」。全库 310 个 4xx 透传点，逐个改不现实也防不住新增 | **已修复**：① `fail()` 在响应边界加 `sanitizeErrorMessage()`，**覆盖全部状态码**，命中存储/驱动特征词（`no rows in result set`、`SQLSTATE`、`violates unique constraint`、`relation "`、`pq: `、连接类错误等）即换中性文案，业务文案原样保留；② 词表用例直接取自线上响应体原文，并做变异验证（摘掉脱敏 → RED）；③ 顺带修正 `workOrderDetail`：它把**所有**错误硬编码成 403，数据库故障会显示成「无权限」——改为基础设施故障回 500，而「不存在/不是双方」仍统一 403（有意的反枚举，见 `TestR4WorkOrderDetailReworkCancel` 的显式断言） |
| **自助充值＝免费印钞**（2026-09-18 发现，生产尚未被用过） | `POST /api/v1/escrow/deposit` 对**任何登录用户**开放自助充值（`target = a.ID`、渠道 `internal_self`），单笔上限 ¥20 万、**不限次数**，且 `DepositFromChannel` 在 `externalTxnID` 为空时**完全不做查重**。它本意是真实支付接入前的联调通道（handler 注释：「模拟通道限额定闸，真实支付接入后由支付校验替代」），但没有任何机制保证它真的会随支付上线而关闭——`escrow.go` 顶部注释还写着「充值仅管理员可操作…任意登录用户可无限充值属 P0 印钞漏洞」，文档与实现自相矛盾 | **已修复**：新增 `simulatedDepositAllowed()` —— 微信支付一旦开通，普通用户自助充值**自动**关闭（管理员代充不受影响），不再依赖人记得去关；确需保留演示通道时显式设 `ALLOW_SIMULATED_DEPOSIT=true`。已加 `TestSimulatedDepositClosedOncePaymentEnabled` 并做变异验证：删掉门禁 → 测试 RED，且失败信息直接打印出一笔 `channel=internal_self, external_txn_id=""` 的凭空入账 |
| **微信支付默认域名带反引号**（2026-09-18 发现，未上线） | `internal/wechatpay/wechatpay.go` 的 `defaultAPIBase` 写成了 `"`https://api.mch.weixin.qq.com`"`——字符串里真的多了两个反引号。`Config.APIBase` 只有测试会覆盖，所以单测全绿，生产每一笔下单都会去解析一个带反引号的域名 → DNS 直接失败 | **已修复**：改为干净的 https 字面量，并加 `TestDefaultAPIBaseIsCleanURL` + `TestNewFallsBackToDefaultBase` 守住「默认基址必须是合法 https URL 且不含引号/空格」。已做变异验证：改回带反引号 → 测试 RED；还原 → SHA-256 一致且 PASS |
| **前端发布差点清空站点根**（2026-09-18 排查） | `/var/www/admin` 下除了构建产物，还躺着 `frontend/public/` 之外的历史媒体（`static/home/*.jpg` 被 10 个商品封面引用）；此前的前端发布习惯是 `rm -rf /var/www/admin/*` 再解包，一旦沿用就会把这些文件连根删掉 → 商品图全部 404 | **已修复**：① 新增 `deploy/deploy-web.sh` —— 先 `cp -a` 备份到 `admin.bak.<ts>`（保留最近 3 份），再**覆盖写入 + 只清理新构建里已不存在的 `assets/*`**，绝不整目录清空；发布后强制校验 `index.html`/`assets`/`static/home/home-bg.jpg`/`images`/`video` 五处；② 把仅存在于服务器的 `home-bg.jpg` 补进 `frontend/public/static/home/`，让站点根内容重新全部由构建产物决定 |
| **编辑 `.ps1` 后 PowerShell 报满屏语法错误**（2026-09-18） | 重写 `scripts/daily-report.ps1` 后，PS 5.1 报出十几条「表达式或语句中包含意外的标记」，报错位置的中文全是乱码（`涓€','浜?`）。根因：写文件的工具产出的是 **UTF-8 无 BOM**，而 PS 5.1 读无 BOM 的 `.ps1` 时按**系统 ANSI 代码页**（简中 936/GBK）解码 —— 中文字符串被拆坏，引号配对全乱，于是语法检查全线报错（脚本本身没问题） | **修法**：`.ps1` 里有中文就必须带 UTF-8 BOM：`[IO.File]::WriteAllText($p,$s,(New-Object Text.UTF8Encoding($true)))`。**每次 edit / 重写都会把 BOM 抹掉**，改完要补一次再跑语法检查（`[System.Management.Automation.Language.Parser]::ParseFile`）。仓库里 `scripts/clean-c.ps1` 是带 BOM 的，可作参照 |
| **每天凌晨 03:02 一条假的备份告警**（2026-09-20，用户报） | 用户收到 `[告警] 运维快照异常：backup（09-20 03:02）`，10 分钟后又是一条「已恢复」。查下来**不是偶发，是每天必发**：crontab 里 `0 3 * * *` 的备份条目排在 `*/10 * * * *` 的快照前面，同一分钟两者同秒启动；备份要边写边落盘（实测 0–1 秒），而快照这一秒里 `ls -1t \| head -1` 正好拿到**还没写完**的那个文件，`gzip -t` 必然失败 → 判 `corrupt` → 03:02 告警；10 分钟后再跑文件已完整 → 03:12 恢复。09-19、09-20 连着两天各发一对，真正的故障反而会被这种噪音淹掉 | **已修复**：备份检查改成取「最近一次**完成**的备份」——解析 `backup.log` 里最后一条 `OK <文件名>`（该行只在备份完整落盘且非空之后才写），而不是按 mtime 取目录里最新的文件；老环境没有 backup.log 时兜底「按时间取最新、但跳过 180 秒内修改过的文件」。新增 `deploy/verify-ops-status.sh` 四项演练（写入中的新文件被跳过 / 真过期 40h 仍报 / 损坏仍被 `gzip -t` 抓住 / 无 backup.log 时的兜底），全程 `mktemp` 临时目录 + 末尾自证生产快照未被改动。**变异验证**：把逻辑改回 `ls -1t \| head -1` → 2 项 RED 且明确指出「又读了正在写的半个文件」，还原后 sha256 一致 |
| **三个已废弃的发布页还挂在注册表里，另有一条死链藏在白名单**（2026-09-20） | 小程序**没有构建期路由校验**：`pages.json` 注册了什么就编译什么，没人核对「这页还有人跳吗」。实际查出 6 个零引用页面（`pages/webview/index`、`pkg-eco/pages/portfolios/list`、`pkg-service/pages/publish/{product,service,course}`、`pkg-app/pages/applications/index`），其中三个旧发布页**早已退化成跳转壳**（16 行，`redirectTo` 到统一的 `pages/publish/form`），却仍占着注册表；更隐蔽的是首页横幅白名单 `ALLOWED_ROUTES` 里躺着 `/pkg-eco/pages/shops/index` —— `shops` 表随迁移 `000113` 一起删了，目录都不存在，**任何横幅配到这条就是白屏**，而且白名单本意是「只放行安全路径」，等于自己开了一道后门 | **已修复**：① 删掉三个废弃发布页（含 `pages.json` 三条注册）；② 从 `ALLOWED_ROUTES` 摘掉 `shops/index`；③ 新增 `scripts/check-miniprogram-routes.cjs` 并接进 CI（frontend job）—— 校验「页面文件 ↔ pages.json ↔ 代码里写死的路由」三者，两类问题即 fail：注册了没文件、路由没注册；另附零引用页面提示。它认 `ROUTE_MAP` 这类别名表的**键**（旧路径故意不存在，`/pages/demand/list` → `/pages/demands/list`），不当死链。**注意脚本必须放 `scripts/`**：`.tools/` 被 gitignore，CI 拿不到。变异验证：抽掉一个页面文件 → 报「缺失 …vue」并 exit 1；往白名单塞一条不存在的路径 → 报「死链 …」并 exit 1；还原后全绿 |

## 本地开发

```bash
# 后端 API（推荐用 run_api.bat，自动设环境变量）
go build -o api.exe ./cmd/api && run_api.bat

# 或手动
go run ./cmd/api     # 后端 → :8080

# 前端 Admin
cd frontend && npm run dev   # → :5173

# 小程序（用 HBuilderX 打开 miniprogram/ 目录编译）
```

## 详细文档

| 想看... | 文档 |
|------|------|
| 项目简介 + 快速开始 | [README.md](README.md) |
| 架构 + 分层 + 中间件 | [docs/系统架构/架构总览.md](docs/系统架构/架构总览.md) |
| 7大业务系统详情 | [docs/业务系统/](docs/业务系统/) |
| 全部 API 契约 | [docs/接口文档/API契约.md](docs/接口文档/API契约.md) |
| 91 张表结构 | [docs/数据设计/数据模型.md](docs/数据设计/数据模型.md) |
| 编码规范 | [docs/开发规范/编码规范.md](docs/开发规范/编码规范.md) |
| Docker + CI | [docs/运维部署/Docker部署.md](docs/运维部署/Docker部署.md) |
| 开发计划 | [PRD-四人并行开发方案.md](docs/项目管理/PRD-四人并行开发方案.md) |

## 前端任务技能规则（必守）

**凡是涉及前端设计/界面/布局/样式的工作（小程序页面、管理后台 UI、原型、组件），动手前必须先加载对应 skill，不得跳过：**

| 任务类型 | 必须加载的 skill | 触发词 |
|---|---|---|
| 界面设计/改版/打磨/审查（小程序、Admin、原型） | `impeccable` | 设计、页面、布局、样式、UI、好看、乱、丑、排版、配色 |
| 做可视化/图表/模拟器/交互原型 | `visualize` | 画图、架构图、图表、模拟器、地图、原型 |
| 生成图片素材（海报/配图/图标背景） | `imagegen` | 海报、配图、生成图、封面图 |
| 前端改动完成后的独立复查 | `review-agent`（或 `code-review`） | 复查、审查、检查改动 |
| 浏览/点击/截图验证本地页面 | `control-in-app-browser` | 打开 localhost、验证页面、网页操作 |

加载方式：在动手前调用 `skill` 工具加载对应技能（如 `skill(name: "impeccable")`），按其规范执行；多类任务重叠时按主要任务类型加载 1 个即可，其余按需补充。
