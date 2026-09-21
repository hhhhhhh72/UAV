# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 7 大业务系统。

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 API | Go 1.25+，标准库 net/http，`mux.HandleFunc` 注册点 **508 处**（只数源码、不数测试；循环体里的按 1 处计），管理端路由探针 204 条（见 `perm_routes_test.go`），311 个 Go 文件（**128 源码 + 183 测试**） |
| 数据库 | PostgreSQL 16（生产） / 内存存储（开发），**92 张存活表**（96 次 CREATE TABLE 减去被 drop 的 shops/demand_bids/association_members；**121 组迁移，242 个 SQL 文件**） |
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
├── migrations/                    # 121 组迁移（242 个 SQL 文件，92 张存活表）
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
| 微信小程序 | `miniprogram/` | uni-app + Vue3 `<script setup>` + 自研 u- 组件库 | **105 页**（主包 31 + 分包 74），5 Tab，**5 分包**（`node scripts/check-miniprogram-routes.cjs` 可复核注册页数并查死链） |
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
| **三个已废弃的发布页还挂在注册表里，另有一条死链藏在白名单**（2026-09-20） | 小程序**没有构建期路由校验**：`pages.json` 注册了什么就编译什么，没人核对「这页还有人跳吗」。实际查出 6 个零引用页面（`pages/webview/index`、`pkg-eco/pages/portfolios/list`、`pkg-service/pages/publish/{product,service,course}`、`pkg-app/pages/applications/index`），其中三个旧发布页**早已退化成跳转壳**（16 行，`redirectTo` 到统一的 `pages/publish/form`），却仍占着注册表；更隐蔽的是首页横幅白名单 `ALLOWED_ROUTES` 里躺着 `/pkg-eco/pages/shops/index` —— `shops` 表随迁移 `000113` 一起删了，目录都不存在，**任何横幅配到这条就是白屏**，而且白名单本意是「只放行安全路径」，等于自己开了一道后门 | **已修复**：① 删掉三个废弃发布页（含 `pages.json` 三条注册）；② 从 `ALLOWED_ROUTES` 摘掉 `shops/index`；另外用户确认后再删两个：`pkg-eco/pages/mall/index`（旧版「服务」tab 页，已被 `pages/services/index` 取代，连 `Layout :current="1"` 的 tab 位都对不上）与 `pkg-app/pages/applications/index`（858 行，功能与「我的需求/我发布的」重叠，且 `pages.json` 标题写「我的申请」而页内导航栏写「我的业务」）——后者一删 `pkg-app` 整个分包就空了，分包条目一并移除（页数 110 → **105**）。删 `applications` 时同步摘掉了 `scripts/verify-filter-alignment.js` 里的硬编码条目，否则那个脚本会 `[MISSING]` 永久失败；③ 新增 `scripts/check-miniprogram-routes.cjs` 并接进 CI（frontend job）—— 校验「页面文件 ↔ pages.json ↔ 代码里写死的路由」三者，两类问题即 fail：注册了没文件、路由没注册；另附零引用页面提示。它认 `ROUTE_MAP` 这类别名表的**键**（旧路径故意不存在，`/pages/demand/list` → `/pages/demands/list`），不当死链。**注意脚本必须放 `scripts/`**：`.tools/` 被 gitignore，CI 拿不到。变异验证：抽掉一个页面文件 → 报「缺失 …vue」并 exit 1；往白名单塞一条不存在的路径 → 报「死链 …」并 exit 1；还原后全绿。**它当场就抓到了我自己删除的后果**：删掉 `mall/index` 后，首页横幅白名单里那条立刻被报成死链 —— 这正是「删页面要连带改所有引用点」的机器化兜底。仍保留 2 个孤儿（`pages/webview/index` 是写好没接的受限外链容器，留着待用；`pkg-eco/pages/portfolios/list` 用户明确说先不接） |
| **给供给大厅卡片加动效，连着两次都被否**（2026-09-20） | 用户要求「给供给大厅的卡片呈现做个动画」。第一次做卡片入场（淡入 + 上移 + 错开延迟）→「算了有点丑」；第二次改做按压反馈（scale 0.985）与图片渐显 →「太奇怪了」。**根因是设计规范本来就说了不要**：`visual-foundations.md:82` 明写「不做卡片批量入场、弹跳、光扫或视差炫技」，而 `:78` 说按压是「0.98–0.985 缩放**或**轻微透明变化」——这页原本的 `hover-class="tap-fade"`（opacity 0.85）**已经符合规范**，我加的两样都是多余的 | **已回退**：两次 `git revert`（`5c64bab`、`81b7d62`），文件 sha 与改动前逐字节一致，残渣扫描 5 个关键词全 0。三条给下次的结论：① 用户明确要求做规范不允许的东西时，**先用规范说服一次再动手**，别直接做 —— 我做了两轮、两轮被否；② 「卡片呈现」这类**页面级**动效在这个 app 的既有语言里是**克制**的（全仓只有 experts/jobs 两个列表页有卡片入场），新加之前先数一数同类页面有几个在做；③ 一次只改一样东西 —— 我第二次一口气加了按压 + 图片渐显，被否后连「是哪个奇怪」都判断不了，只能回头再问一轮 |
| **三个页面在接口挂掉时把演示数据当真实数据渲染**（2026-09-20 契约体检发现） | `miniprogram/utils` 下 6 个 mock 模块（mockProjects/mockChallenges/mockBrands/mockAchievements/mockReports/mockExhibitions）**自身没有任何守卫**，导出的就是完整真实数组，生产是否安全完全取决于调用点。逐点核对 15 个引用点后发现 3 处漏改：`projects/detail.vue` **整个文件零守卫**（全文搜不到任何 NODE_ENV/DEV），且失败分支 `d.value = null` 时**没有置 err** → 生产上是无错误态的空白页；`projects/list.vue` 与 `challenges/list.vue` 在接口首次加载失败时把编造的课题攻关/研发难题列表当真实数据渲染。**而隔壁 `challenges/detail.vue:137` 明明写着「生产环境禁止演示数据回退……（数字诚实铁律）」并做了 `isProduction` 守卫** —— 同一个业务域一个页守住了、另一个没守，说明是漏改而非有意。后两处更隐蔽：「演示数据」横幅写的是 `v-if="mockMode && isDev"`，生产构建里 `mockMode` 是个**死标志**，假数据结构照渲染却没有任何提示 | **已修复**：① 列表页复用文件里**已经声明过**的 `isDev`（原本只用于横幅），改成 `if (isDev && MOCK_X.length)` —— 生产走 `else` 分支置 `err`，如实呈现失败态；② 详情页补 `isProduction` 常量并把两处回退包进 `if (!isProduction)`，同时补上缺失的 `err.value = true`；③ 新增 `scripts/check-mock-guards.cjs` 并接进 CI —— 判定「引用假数据的文件必须带**已应用**的守卫（窗口 ±25 行）」，两条不算数：**声明行不算应用**（`const isDev = ...` 只说明文件里有这个标志，不说明它罩住了这一行）、**注释里的守卫不算**（先剥注释，否则 config.js 里那句提到 NODE_ENV 的说明文字就能让任何文件蒙混过关），import 行与 `export const MOCK_*` 定义行不算使用。自检 13/13。**变异验证**：摘掉 list 页的 `isDev` → RED 并指出 544/545；把 detail 页退回修复前形态 → RED 并指出 332/337；两次还原后 sha256 一致。**已知假阴性**（写进脚本头部）：同函数内两个相邻回退只摘掉其一，会被另一个的守卫落在窗口内掩盖 —— 实测过缩进/括号块级判定修不了它（仓库里列表页主导写法是「守卫在调用点深缩进 + 假数据在顶层 useMock() 助手函数里」，按缩进会把 portfolios/reports/exhibitions 三处**正确**的守卫判成违规）。同时提交 `scripts/check-api-contract.cjs`（501 条后端路由 × 303 个调用点，匹配器自检 10/10）|
| **200 家企业同时入驻会把 API 容器顶爆**（2026-09-20，用户问「我现在假如有200家企业同时入驻我能抗住嘛？」） | 逐层实测后，**CPU / 数据库 / nginx 都不是问题**：生产 2 核上 bcrypt 实测 **99ms/次**、饱和吞吐 **21 req/s**（200 并发注册约 9.5s 完成，远低于 `WriteTimeout 30s`，硬上限约 630 并发）；入驻主流程走微信静默登录，`auth_wechat.go` 全文**无 bcrypt**；pgxpool 50 / PG `max_connections=100` 对毫秒级插入绰绰有余；nginx 限流按客户端 IP 分桶（`limit_req_zone $binary_remote_addr`，100r/s + burst 80），200 家不同 IP 各自独立桶。**真问题是上传**：企业入驻必传营业执照（`pkg-eco/pages/enterprise/register.vue:208`），而 `internal/httpapi/files.go` 的 `maxUploadBytes = 40<<20` 被当作 `r.ParseMultipartForm` 的**内存阈值**，于是每个文件段整份驻留内存。实测 200 并发上传的峰值堆：照片 300KB → 113MB、1MB → **310MB**、2MB → **711MB**，而容器内存上限只有 **512MB**（另有 512MB swap）—— 顶爆后疯狂换页，30s 超时开始吃掉请求，用户重试又加大压力。`service/files.go:35` 的 `uploadQuotaMu` 全局锁**管不住它**：锁只串行化「查配额→写盘→记账」，而缓冲发生在它之前的 handler 里，200 个请求体是同时进内存的 | **已修复**：① `uploadFile` 改用 `r.MultipartReader()` 逐段流式解析，文件段直接交给 service（`service/files.go:188` 本来就是 `io.TeeReader` 边算 SHA-256 边落盘，天然流式，无需二次拷贝）；② 既有语义全部保留 —— 魔数检测（只读前 512 字节）、private/public 分流、每日配额、SHA-256、图片尺寸；③ 新增 `failUploadErr`：超限改流式后要到读满 40MiB 才由 `MaxBytesReader` 报错，必须 `errors.As` 翻译回既有的 **400「file too large」**（否则退化成笼统 500）；④ **字段顺序陷阱**：`private` 字段可能排在文件段之后（客户端不保证顺序），若直接流式交给 service 会把营业执照写进**公开**目录 —— 私密性未确定时先把文件段落临时文件，确定后再交，并加测试覆盖两种顺序。**实测效果**：200 并发 1MB 上传峰值堆 **12.6MB**（0.06MB/请求，改前 310MB），且更快（200 并发 2MB 墙钟 123ms vs 506ms）。新增 `internal/httpapi/files_stream_test.go` 四条：两种字段顺序 × 私密性、缺 file 段 400、超限 400 且磁盘无残留、200 并发内存上限。**变异验证**：把文件段改回 `io.ReadAll` → 内存测试 RED（194MB > 120MB 阈值）并点名 files.go 的读法；还原后 sha256 一致；本地全量 `go build/vet/test ./internal/...` 全绿。**顺带纠正一个我本来会给出的错误建议**：把 `ParseMultipartForm` 阈值从 40MiB 调到 4MiB **毫无作用**（实测每请求 5.26MB vs 5.25MB）——小于阈值的文件本来就整份进内存，唯一解法是流式。容器内存 512M→1G 只把悬崖从 1MB 照片推到 2MB，治标；上传并发闸门可作第二道防线（降级成「部分人重试」而不是「全员卡死」）|
| **研学报名会超卖**（2026-09-20，顺着「其他模块呢」的并发审计查出） | 用户问「需求和培训能不能扛住」。逐条审计后结论是**分层的**：① **需求侧最结实** —— 接单是补偿式事务（`work_order.go:94-118`）：工单唯一索引 `uniq_work_orders_intent`（000065/000066，库级「一意向一单」）+ 意向 CAS + **需求状态 CAS `published→assigned`**（库级「一需求一单」）+ CAS 失败撤单回滚意向，多实例也安全。② **课程报名**：去重有库级唯一索引 `uniq_training_enrollments_user_course`（000064），容量只有**进程内键锁**（`phase3.go:137`）—— 实测 2 座位 / 200 并发正好成功 2 条 ✅（单实例正确）。③ **研学报名**（`study_tour_enrollments.go:50-79`）：容量校验是纯 check-then-act（`ListByTour` 求和 → 比容量 → `Create`），**既没有 `lockByKey` 也没有任何库级唯一索引**（migrations 里 study_tour_enrollments 零唯一索引）。**实测容量 10、200 人并发报名 → 收下 87 人，超卖 77 人**（题外话：修好后再变异摘锁只超卖 1 人 —— 超卖幅度随调度而变，这正是测试要断言不变量而不是断言数字的原因） | **已修复**：给 `Create` 加研学维度键锁 `lockByKey("study-enroll|"+tourID)`，与课程报名、测试场地/场馆/赛事/活动报名同构（全仓共 8 处键锁）。新增 `internal/service/study_tour_concurrency_test.go` 两条：研学容量（容量 10 / 200 并发 → 恰好 10 人）与**课程容量正向对照**（2 座位 / 200 并发 → 恰好 2 条，证明「键锁这个方案没问题，是研学漏了」）。变异验证：摘掉键锁 → RED，还原 sha256 一致。**遗留（未做，需产品决策）**：容量类不变量全仓**没有任何库级兜底** —— `BumpEnrolled` 是**无条件** `enrolled_count + $2`（PG `phase3_repos.go:270`），8 处键锁全是进程内锁，**多实例部署会一起静默超卖**，而没有任何机制保证单实例。durable 修法是条件更新（`UPDATE ... WHERE max_students = 0 OR enrolled_count < max_students`，判 `RowsAffected`），仓库里**已有现成范式**：退款额度占用的 `ReserveRefund` 就是「仅当 remaining >= amount 时成功」的原子占用（`memory_payment.go:57`）。另外测试场地/场馆预约同样有锁无索引，研学报名连「同一用户重复报名」都没拦（是否有意为之待确认）|
| **「单实例」是个没人验证过的前提**（2026-09-20，接上一条） | 上一条查出「8 处进程内键锁只在单实例下成立，多实例会一起静默超发」，而**没有任何机制保证或验证单实例**。这正是备份静默失败/告警假绿那一类病：保护措施依赖一个没人盯着的条件 | **已修复（第一道：让它可验证）**：`ops-status.sh` 新增 `instances` 一节，两个独立信号 —— ① 跑着几个 `uav-api` 容器；② 有几个**不同 client_addr** 连到本库。应用连接走 TCP 会被计数，运维脚本走 `docker exec` 的本地 socket（client_addr 为 NULL）不参与，生产实测单实例 = 1。**取不到数据即判失守**（宁可报，不可假绿）。已接进顶层 `ok` 与 `NOTIFY_STATUS_SECTIONS`（`lib-notify.sh` 唯一一份，12 项 meta-check 自动比对）。`verify-ops-status.sh` 加到 7 项：阈值收紧到 0 必须变红（证明不是摆设）、查不到连接来源必须报、正常环境必须绿并报出真实计数。**变异验证（两次，第一次方法错了）**：先试「宿主机裸 TCP 连 5433」→ **没触发**，因为没发启动包的连接根本不会进 `pg_stat_activity`（信号只统计**已认证**连接，第二台 API 带 `MinConns=5` 的连接池连上来才会被算到）；改用一条**已认证**的额外连接后 → `db_client_addrs: 2` → `instances.ok=false` 且**顶层 ok 一起变红**，释放后恢复绿。全程用临时 `OUT`，未碰生产快照（演练脚本末尾自证 mtime 未变）。**尚未做的**：名额类不变量仍无库级兜底（条件更新），场地/场馆时段重叠仍无排他约束 —— 见下两条 |
| **课程容量补上库级兜底（ReserveSeat）**（2026-09-20，Step 2） | 承接上一条：课程容量此前只靠进程内课程维度键锁，而 `BumpEnrolled` 是**无条件** `enrolled_count + $2`（PG `phase3_repos.go:270`），多一个 API 实例就会一起超卖 | **已修复**：`CourseRepository` 新增 `ReserveSeat(ctx, id) (bool, error)` —— PG 侧把容量判断与 +1 放进**同一条 UPDATE**（`WHERE id=$1 AND (max_students = 0 OR enrolled_count < max_students)`，判 `RowsAffected`），内存侧在同一把锁内做同样判断；`Enroll` 改以它作为**权威门禁**（原先那次 `FindByID` 容量检查降级为「友好文案」，另加一次「已满」兜底），占座失败即 `course is full`，落库失败仍走既有 `BumpEnrolled(-1)` 补偿。**课程不存在返回 `(false, nil)`** —— 调用方按「课程不存在」处理，与 PG 的 `RowsAffected=0` 语义对齐，避免历史无课程路径被误报成满员。**库级验证（不是推断）**：交叉编译测试二进制到生产 PG 的**临时库**（`drone_platform_test`，跑完即删、磁盘回到 26%）**直压仓储层、完全绕过 service 键锁** —— 50 并发抢 3 个座位 → 正好 3 个、`remain=0`；不限量（`max_students=0`）30 并发全通过；不存在的课程 `(false, nil)`。**变异验证**：摘掉 WHERE 里的容量条件（退化成旧的无条件 +1）→ RED「**超卖：座位 3，ReserveSeat 成功 50 次，enrolled_count=50**」；还原后 sha256 一致、PASS。`verify-ops-status` 的 `instances` 一节仍是第一道防线（检测多实例），本条是第二道（即使多实例也不超卖）。**剩余未做**：研学/赛事/活动报名仍只有进程内锁（研学容量是**求和式**，durable 修法要在事务里 `SELECT … FOR UPDATE` 锁住研学行再求和插入，不能照抄条件更新）；测试场地/场馆的**时段重叠**要用 PG 排他约束 `EXCLUDE USING gist (site_id WITH =, tsrange(start_time,end_time) WITH &&)`（需 `btree_gist`），唯一索引表达不了区间重叠 |
| **场地/场馆的时段重叠补上库级兜底**（2026-09-20，Step 3） | 承接前两条：场地/场馆的冲突判定只在 service 层（`lockByKey` 进程内锁 + check-then-insert），而**唯一索引表达不了区间重叠**，库层面一直没有兜底。**动手前先只读扫生产数据**，扫出真问题：`test_site_bookings` 有 **7 组活跃行时段重叠**（6 组 pending×pending、1 组 pending×approved）—— 全是 2026-08-12/08-18 的历史数据，而「pending/approved 均占位」这条规则是 `8991dd2`（**08-25**）才加的，**晚于那些数据**（git log -S 实证）；若直接把 pending 纳入约束，迁移会当场失败 —— 而 `cmd/api/main.go:281-283` 对迁移失败是 `os.Exit(1)`，**API 根本起不来** | **已修复**：migration `000120`（`btree_gist` + **生成列** + `EXCLUDE USING gist`）。三个关键设计：① **PG 的排他约束不支持 `WHERE` 部分约束** —— 直接对 `(site_id, tstzrange(start,end))` 建约束会把**被取消/驳回的历史行**也算作互斥，某时段一旦取消过就再也订不回来；改用生成列（非占用状态取 NULL，而 **NULL 在排他约束里不冲突**，与唯一索引对 NULL 的处理一致）精确表达「只有真正占用的行才互斥」；② 上界用 `'[]'` **闭区间**，与 service 层 `!(end < start_b || start > end_b)` 一致（前一场 12:00 结束、后一场 12:00 开始算冲突）；`tstzrange(...,text)` 三参重载**实测在 pg_proc 里是 immutable**（`provolatile='i'`），可以进生成列；③ **场地预约只约束 `approved`**（approved 之间重叠实测为 0，先钉死最强的那条不变量），pending 期间的互斥仍由键锁 + 审批前复检承担。**上线前的安全顺序**：先在**从最近一次完成备份还原的临时库**上应用同一份迁移文件（真实数据上不冲突 ✓），再在**生产同一事务内应用并登记 `schema_migrations`**，让启动时的自动迁移直接跳过 —— 彻底消除「迁移失败→服务起不来」这个风险。**行为矩阵 13/13**（临时库）：approved 完全重叠被拒、**紧邻（闭区间）被拒**、不重叠放行、pending/cancelled 与 approved 重叠放行、**把 pending 改成 approved 也被拒**（审批路径同样受保护）；场馆 booked/pending 被拒、cancelled 放行；**变异**：摘掉约束后同样的重叠插入不再被拒 → 证明拦住的正是这条约束；清理冲突行后可重新 `ADD CONSTRAINT`（对应以后收紧 pending 的路径）。**应用层**：新增哨兵 `repository.ErrSlotTaken` + `translateSlotConflict`（`errors.As` 判 SQLSTATE **23P01**，与既有 23505 范式同构），service 翻成既有的「time slot conflicted」/「该时段已有预约，审批冲突」，不让用户看到笼统 500；单测覆盖 23P01/包裹层/23505/普通错误/nil 五种输入。**遗留**：那 7 组历史重叠数据**未清理**（属业务数据，需确认后再删；清理后可把条件收紧为 `status IN ('approved','pending')`）|
| **研学容量的库级兜底 + 场地约束收紧**（2026-09-20，Step 4/5） | ① 研学容量是**求和式**的（没有 `enrolled_count` 那样的计数列），所以不能照抄课程的条件更新；此前只有进程内键锁（Step 1 的 `instances` 巡检只能**发现**多实例，不能防超卖）。② Step 3 的场地约束当时只覆盖 `approved`，因为生产有 7 组历史重叠（6 行全部**已过期**：1 组 08-12 的 pending+approved、4 行 08-19 同时段重复 pending），把 pending 一并纳入会让迁移失败 | **① 已修复**：`StudyTourEnrollmentRepository` 新增 `CreateWithCapacity(ctx, e, headcount, capacityLimit)` —— PG 侧在**同一事务**里 `SELECT id FROM study_tours WHERE id=$1 FOR UPDATE` 锁住研学行 → 求和活跃人数（pending/approved 的成人+儿童）→ 未超容量才 INSERT；内存侧在同一把锁内做同样判断；超了返回哨兵 `repository.ErrCapacityFull`，service 翻成既有的「该研学活动名额已满」（预检保留，只为友好文案）。**库级验证**：生产 PG 临时库（跑完即删）**直压仓储层、绕过 service 键锁** —— 50 并发抢 3 个名额 → 正好 3；**变异**：只摘掉 `FOR UPDATE` 这一个词 → RED「**超卖：容量 3，成功 6 次，占用 6 人**」，还原 sha256 一致。内存侧另加一条「直压仓储层」的 200 并发测试（容量 10 → 正好 10）。**② 已修复**：先在**从最近一次完成备份还原的副本**上验证整套动作（7→0 重叠、approved 3 行未被误改、恰好作废 5 行、`ADD CONSTRAINT` 成功、两条重叠 pending 被拒），再对生产执行 —— 清理只针对「**已过期 + 与同场地其它预约重叠 + 仍是 pending**」的行（置 `rejected` + `review_note` 说明），approved 一行未动、总行数仍 15；随后 migration `000121` 把生成列表达式改为 `status IN ('approved','pending')`（生成列不能 ALTER，只能 drop column 重建）并重新登记版本，生产 schema **000121**。**踩坑**：① 脚本通过 stdin 喂给 bash 时，未重定向 stdin 的 `docker exec -i` 会**把脚本剩余部分当输入吃掉**（脚本莫名中断）；② `docker exec … psql -f /tmp/x.sql` 的路径是**容器内**的，宿主机文件要用 `< /tmp/x.sql` 重定向。两条都当场定位并改对了 |
| **容量门禁读的计数列在生产上漂移**（2026-09-20，Step 6） | 上一条收尾时我说「赛事名额仍只有进程内锁」—— **这句是错的**，我只按方法名找了一遍就下了结论。实际读代码：PG 的 `compRepo.CreateReg`（`biz_repos3.go:162-171`）**早就有库级原子占位** —— `UPDATE competitions SET reg_count = reg_count + 1 WHERE id=$1 AND (max_teams <= 0 OR reg_count < max_teams)`，与 INSERT 同事务、`RowsAffected=0` 即回滚报满，与课程 `ReserveSeat` 同型（活动也有，`biz_repos3.go:290`）。**真正的问题在它依赖的列上**：容量门禁读的是**计数列**而不是实时 count，而这一列在生产上漂移了 —— `competitions` **4 行** + `association_events` **6 行**的 `reg_count` 是种子迁移 `000048` 写死的演示数字（全库只有 1 条真实报名），于是 **comp-4 是 max_teams=300 / reg_count=340 → 该赛事永远报名失败**，`evt-2026-006` 是 30 个名额里**只剩 4 个可报**，`evt-2026-005` 还有反向漂移（reg_count=0 而实际 1 → 会多收 1 个）。另外 `training_courses` 有一门课 **3 条真实报名（全为 enrolled）却 enrolled_count=0**（8 月的历史报名，计数机制之后才有），列表显示「已报 0 / 30」，而 `ReserveSeat` 会在此之上再收 30 个 | **已修复**：① **补上守护测试** `internal/repository/postgres/competition_reg_capacity_test.go` —— 直压仓储层、**绕过 service 的赛事维度键锁**，50 并发抢 3 个名额 → 必须正好 3 条，并断言 `reg_count == 报名行数`（代码注释声称的语义）；**变异**：把条件更新退化成 `WHERE id=$1`（无条件 +1）→ RED「**超卖：名额 3，CreateReg 成功 50 次，实际报名行 50**」，还原 sha256 一致。② **对账**（先在备份还原副本上验，再动生产）：`reg_count` 拉回真实行数 → 生产漂移 **4+6 行 → 0**，comp-4 恢复可报名、evt-006 回到 30 个真名额；课程那门课按「**只往少算方向修**」（`enrolled_count` 小于活跃报名数才是无条件错的；多算可能是设计使然 —— 被驳回的报名按设计不释放座位）→ 0/30 → **3/30、remain 27**。③ **机器守卫**：`ops-status.sh` 新增 `counters` 一节做四类一致性检查（赛事 reg_count、活动 reg_count、课程 remain 派生公式、课程 enrolled_count **少算**），取不到数据即判失守，接进顶层 ok 与 `NOTIFY_STATUS_SECTIONS`（现 **9 个分项**，meta-check 自动比对）。演练加到 **9 项**全过；生产快照 `counters` 四项全 0、9 个分项全绿。**教训**：「某个方法名不存在」不等于「那个保证不存在」—— 找代码要读实现，不是搜名字 |
| **09-20 的日报没发出去**（2026-09-21 用户报「昨天的日报生成失败了为什么」） | 用户问得对（我上一轮答错了日期基准：当时服务器已是 09-21，「昨天」是 09-20）。日志：`19:31:01 fetch 第 1 次失败 / 19:33:06 第 2 次 / 19:35:26 第 3 次`，`fatal: Failed to connect to github.com port 443`，`19:35:31 git fetch 三次都失败 —— 不发旧数据`，紧跟一条 `推送成功 HTTP 200` —— **那是"取不到数据"的通知，不是日报**；当天 16 个提交一条没汇报。**根因**：这台机器**只有 `github.com:443` 不通**（实测 `api.github.com`→200、`codeload`→301、`raw`、`ssh.github.com:443`、`github.com:22` **全通**），而裸库的 origin 正是被封的那个 https 地址。**且它是间歇性的**：09-20 17:11 与 17:30 的干跑都成功取到数，19:31 起连续失败；09-21 09:13 实测第 3 次又通了。原来的重试只有 60/120/180 秒（约 5 分钟）一轮，扛不住几十分钟的封锁。**另一个关键事实**：19:30 失败时，本地裸库其实**已经有当天全部提交**（17:11 那次 fetch 成功，末条 `93b5fb6` 16:55），只是脚本"无法确认是否最新"就按设计拒发了 —— 于是有数据却发不出 | **已修复**：① **重试窗口从 5 分钟拉到 2.5 小时**（`RETRY_WINDOW_SECONDS=9000`，每轮内部仍是 60/120/180 秒递进三次，轮间 `RETRY_GAP_SECONDS=300`），失败通知里带上"重试了几轮/等了多久"；② **加 `SKIP_FETCH=1`** —— 人工补发通道（本地裸库已有数据时不必再 fetch），这正是本次补发 09-20 用的路径；③ **加 SSH over 443 备用通道**（`REPORT_REMOTE_FALLBACK`），https 每次失败后立刻试它，成功则记「https 不通 → 备用通道成功」，需要仓库侧加只读 Deploy key（已在 `/root/.ssh/uav_report_ed25519.pub` 生成待加）。**已补发 09-20**：`SKIP_FETCH=1 DAY=2026-09-20 work-report.sh` → 236 字、`HTTP 200`、心跳前移。**告警链路这次是对的**：09-21 01:32 与 07:32 各推一条 `[告警] 运维快照异常：jobs`（`work_report_age_hours` 37.5 > 30），补发后 09:09:46 自动推 `[恢复] jobs 已恢复正常` —— 失败与恢复都在群里，没假绿。演练：用不可达 remote + `RETRY_WINDOW_SECONDS=1` 验证轮次与超窗文案；用真实仓库验证两条通道都在尝试（且 https 间歇可用）。**遗留**：`github.com:443` 的间歇封锁无法从这台机器根治，Deploy key 加上后 SSH 通道即成为稳定路径；**cron 仍是每天 19:30 单次**（靠脚本内 2.5 小时窗口兜底，未加第二个 cron 点 —— 加了就还需要"今天已发过就不重发"的幂等判断）|
| **彻查稳定性与数据安全：备份只备了库、上传的文件根本没备**（2026-09-21，用户要求「给我彻查」） | 服务端本身**没有在出错**（应用日志 09-15 至今 ERROR **0** 条、nginx 5xx **0** 个、400/405 全是 `wp-admin`/`wp-json` 这类互联网扫描器）。彻查扫出 7 个真问题，最要命的是：**`db-backup.sh` 全文 0 处提及 uploads** —— 每天只 `pg_dump`，而 `/uploads` 里是**不可再生的原始凭据**（营业执照、身份证影像、证件照、案例视频，实测 29MB/127 项）。后果不是"少个文件"，是「**库恢复了、文件全没了**」：库里还留着引用它们的 URL → 全站 404，企业资质审核凭据一起丢（合规层面的损失）。此外：备份与库**在同一块盘**（`lsblk` 只有 vda，无异地副本）；PG 慢查询日志是关的（`log_min_duration_statement=-1`）；**17 个部署文件 world-writable 而它们由 root 的 cron 执行**（`deploy/*.sh` 全 777）、`docker-compose.yml` 666 且内联 `DATABASE_URL`/`POSTGRES_PASSWORD`；**培训报名的身份证/手机号是明文入库**（实测 `training_enrollments.id_card=500202100766642255`、`phone=19823864146`，而同仓 `certified_pilots`/`competition_registrations` 的 id_card 都是密文 → 加密只做了两条链路）；**`ENCRYPTION_KEY` 只在本机 `.env` 里而备份里是密文** → 整机损毁后即使备份完好，密文也永久解不开（`restore-drill.sh` 0 处校验解密） | **已修复两项 + 权限组**：① **上传文件纳入每日备份** —— `db-backup.sh` 追加 tar 卷（`docker volume inspect` 动态取挂载点），写**独立日志** `uploads-backup.log`（不混进 `backup.log`，否则会悄悄改变 ops-status「最近一次完成的备份」那条判定的语义），校验非空 + `tar -tzf` 完整性 + 与库备份同款保留策略；`ops-status.sh` 把上传包**折进 `backup` 一节**（分项列表不用动），演练加「只备库没传包必须失守」用例 → **10/10 通过**；实跑产出 `uav-uploads-*.tar.gz`（28MB/134 项，600）。② **PG 慢查询日志打开**（`ALTER SYSTEM SET log_min_duration_statement='1s'` + reload，实测 `SELECT pg_sleep(1.2)` 被记为 `duration: 1202.801 ms`，已持久化）。③ **权限组**：17 个 world-writable 文件 → 脚本 755、yml/env 600，实跑 ops-status 验证仍正常、生产快照未动。**纠正一处我自己的误报**：上一轮说「日志里有 222 个身份证号、287 个手机号」——**是错的**，那是 `course-1787039313772501705-1` 这类**纳秒 ID 的数字碎片**被我的宽松正则截中；日志里真正含 PII 的只有形如 `user-<手机号>` 的**用户 ID**（手机号注册账号的 ID 嵌了手机号），实际只有超级管理员一个，日志里**没有**真实身份证号。**遗留（未做，需用户决策/资源）**：异地副本（用户需给对象存储；**腾讯云自动快照策略**是最便宜的一条，需他在控制台确认是否已开）；培训报名 PII 加密（涉及存量行加密迁移）；备份携带 `ENCRYPTION_KEY` 并让还原演练校验解密；证书 48 天无自动续期 |
| **「备份能恢复」不等于「恢复出来的数据能读」：密钥在机器上、密文在备份里**（2026-09-21，用户批准后做的第 2 项） | 备份里躺着 PII 密文，而 `ENCRYPTION_KEY` 只存在于本机 `/root/UAV/.env`。备份再完好，密钥一丢或被轮换，身份证/手机号就**永久解不开** —— 而此前没有任何检查会发现：gzip 校验过、演练数过 92 张表、`users=11`，**全绿**。更糟的是**三份加密列清单互不一致**：`cmd/reencrypt`（轮换密钥的工具）列了 5 列，`cmd/keycheck`（我这次新建）列了 3 列，而生产实测**加密列实际有 7 列** —— reencrypt 漏掉 `competition_registrations.id_card/.phone`，**用它轮换密钥会跳过那两列，旧密钥一丢就永久损坏**（`decRegPII` 解不开时保留原值 → 界面直接把密文当身份证显示） | **已修复**：① 新增 `cmd/keycheck`（复用 `internal/crypto` 的**同一份实现**，不在脚本里另写解密逻辑），逐列逐行尝解密；② **清单合并成唯一一份** `internal/crypto.EncryptedColumns`（7 列），keycheck 与 reencrypt 都 import 它 —— 从结构上消灭"清单漂移"这个类别；③ `restore-drill.sh` 在还原出的临时库上跑 keycheck，**工具缺失/密钥取不到/源文件比工具新（工具过期）一律判失败**（fail-closed，检查跑不起来时最不该沉默放过）；④ 离线托管材料 `/root/UAV-secrets-escrow.txt`（600，含 ENCRYPTION_KEY / AUTH_SECRET / SIGNING_SECRET + 指纹，**不在备份目录内**），指纹 `0e5c72d641b57001`（可对外核对而不暴露密钥）。**验证**：真跑演练 `密文行 4：可解密 4` ✓；**变异 A** 换一把合法但不对的密钥 → exit=1、`加密字段无法解密`、`可解密 0 / 无法解密 4` + 失败样本；**变异 B** 删掉工具 → exit=1；**变异 C** 把源文件时间改到工具之后 → exit=1 且明说"工具过期，需重建后重传"；恢复后全绿。**顺带扫出的真值表**（对生产只读）：7 列 16 行，**14 行密文全部可解、0 行解不开**，`enterprises.account_name` 有 **2 行明文**（`渝航智能科技`、`山城测绘`，与其它种子同一纳秒前缀 = 加密路径之前建的）—— 不是密钥问题，是**加密覆盖不一致**，App 解不开时保留原值所以界面正常。**顺手纠正一处我自己的过窄判定**：最初只把"身份证/手机号格式"当明文，于是那两行中文企业名被误报成"无法解密"、演练变红；改成「能否 base64 解码」当分界线（GCM 密文必是合法 base64，中文与 18 位身份证号都不是）。**遗留**：那 2 行明文企业名未加密（reencrypt 按设计把解不开的值当历史明文跳过，要覆盖需单独处理）；培训报名 PII 仍未加密 |
| **培训报名的实名信息明文入库；发布链路还依赖一个极慢的 Alpine 源**（2026-09-21，用户「按顺序来 除了1」） | ① 生产实测 `training_enrollments.id_card=500202100766642255`、`phone=19823864146` 是**明文**，而同一把 cipher、同一个仓里 `certified_pilots` / `competition_registrations` 早已加密 —— 只有培训报名这条链路漏了。② 发布时 `apk add git ca-certificates` **失败过一次**（exit 1）：Alpine CDN 从这台机器下载**每个包 60–190 秒**，12 个包十几分钟，中途一抖动整个发布就断 | **已修复**：① `enrollRepo` 加 `cipher`，写入前加密、所有读取路径解密（`encPII`/`decPII`，**解不开就置空、绝不回传密文** —— 与 `compRepo.decRegPII` 同一约定）；`main.go` 接线同一个 `cipher`；`internal/crypto.EncryptedColumns` 补上 `training_enrollments.id_card/phone`（现 **9 列**）。**完整性检查**：把 `training_enrollments` 的 19 处引用全看了一遍 —— 5 处 SELECT 全部过解密，另两处（`ListOrphanFreezes` 只做 `SELECT 1 ... NOT EXISTS`；`user_content_plan.go` 是**擦除计划**只写不读）不需要 ✓。② **顺序是关键**：先发布"读取时解密"的代码、**再**加密存量行（反过来旧代码会把密文当身份证显示 —— 而发布第一次恰好因 apk 失败卡住，这个顺序正好挡住了）。用 `cmd/reencrypt` 新增的 `-encrypt-legacy`（old==new 同一把 key）把历史明文**补加密**：`training_enrollments` 各 4 行 + `enterprises.account_name` 2 行 → 全库 **9 列 24 行、明文 0、解不开 0**（keycheck 复核）。③ **Dockerfile 的 apk 源换阿里云镜像**（实测快两个数量级）。**踩坑**：我给 `restore-drill.sh` 加的"工具新鲜度"检查**按 mtime 判定，部署后必然误报** —— `deploy-api.sh` 每次解包都对 `*.go` 跑 `sed -i` 去 CRLF，源文件 mtime 每次被刷新（实测部署完演练立刻变红，**差点推一条假告警给群里**）；改成**内容哈希**比对，且指纹必须在**服务器上**用服务器实际持有的字节生成（PowerShell 的 Get-FileHash 与 sha256sum 算出的不一致）。变异复核：源码追加一行 → 演练红并报「记录 34622174 / 实测 c2d3a384」→ 还原转绿。**教训**：检查器自己也会假绿/假红 —— 陈旧判定要基于内容，不基于时间 || **证书到期只能靠人去腾讯云控制台重新下载**（2026-09-21，用户「按顺序来 除了1」的第 4 项） | 证书 11-08 到期，**没有任何自动续期**：`certbot certificates` 报「No certificates found」，`/etc/nginx/certs/api.cqnarc.cn.fullchain.crt` 是 Aug 10 手工放进去的（同日还有一张 `uavtest.cloud.crt`，签发者 `TrustAsia DV TLS RSA CA 2024` 正是腾讯云免费证书的 CA —— 两张都是控制台下载后手工装的），crontab 里零条证书任务。**动手前先试标准解 HTTP-01，被 Let's Encrypt 直接挡回**：`Detail: 43.174.225.201: Invalid response from https://dnspod.qcloud.com/static/webblock.html?d=api.cqnarc.cn` —— 根因是**域名未做 ICP 备案**，腾讯云把**境外来源的 80 端口**劫持到备案提示页，而 LE 的多视角校验节点全在境外。同一境外出口只换端口实测：`http://api.cqnarc.cn/` → 302 到 dnspod 拦截页，`https://api.cqnarc.cn/ops-status.json` → 200 真实内容，**443 没被劫持**。certbot 2.9 的 nginx 插件与 standalone 插件源码 `supported_challenges` 都只返回 `[challenges.HTTP01]`（实测配 `--preferred-challenges tls-alpn-01` 一律 "None of the preferred challenges are supported"）；lego 能做 ALPN 但**强制要 `--email`**（用户明确说「不用邮箱，跟报错报告一样发企业微信就行」） | **已修复**：改用 `acme.sh`（v3.1.6 单文件脚本，哈希 `c7d68b…8d7` 钉死在 `deploy/cert-renew.sh` 里，不符即拒绝执行）+ **TLS-ALPN-01**（在 443 上校验，不需要邮箱、不需要 DNS API）。`deploy/cert-renew.sh` 每天 04:30 由 cron 跑：先看证书还剩几天，**没到 30 天阈值就直接退出、一次也不碰 nginx**；到期才 `cert-hooks.sh pre`（停 nginx 让出 443）→ acme.sh → `post`（拉回 nginx）。**实测一次续期停机 24–30 秒**，逐帧看日志发现 LE 3 秒内就校验完了、剩下 25 秒全是 acme.sh 自己的轮询下载 —— 所以 `pre` 在停服务**之前**先挂一个 300 秒**死人开关**（`setsid`+`nohup`，显式 `9>&-`），即使被 SIGKILL、`post` 根本没跑，站点也会自己活过来。**配套**：`ops-status.sh` 证书一节改成读 **nginx 实际在服务的那张证书**（`openssl s_client` 打 127.0.0.1:443）并与文件 serial 比对 —— 只看文件的话「换了证书但没 reload」是全绿假绿，而续期脚本每天都在写那个文件；`jobs` 加续期日志心跳（26h）；`certbot.timer` 已 disable（没有任何 lineage，只会每天两次空转，留着反而让人以为证书归 certbot 管）。**演练**：`verify-cert-renew.sh` 9 项 + `verify-cert-hooks.sh` 9 项 + `verify-ops-status.sh` 15 项 + `verify-alerts.sh` 13 项，全绿。**演练当场揪出两个真 bug**：① 续期后读不到服务端证书时 `after_serial` 是空串，而空串 ≠ 旧 serial 会被判成「换成功了」→ 报假绿，已加空值拦截；② 死人开关**继承了 flock 的 fd 9**，于是续期结束后 5 分钟内再跑都被判成「已有进程在跑」而静默跳过。**顺带查明**：境外访问 `http://` 会看到腾讯云备案提示页（`https://` 正常），这是没备案的直接后果。**遗留**：一旦完成 ICP 备案，把 `--alpn` 换回 webroot HTTP-01 即可（80 端口那段 `/.well-known/acme-challenge/` 一直留着），那时连这 25 秒停机都不需要；25 秒也能靠 nginx `stream`+`ssl_preread` 按 ALPN 分流做到零停机，但那要把 443 整个挪到内网口，结构性改动不值当 |


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
