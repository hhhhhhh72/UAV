# 「后端认证与中间件」报告

## 1. auth.go —— Token 签发/验证与鉴权中间件

**TokenManager**：自研 HMAC-SHA256 令牌体系（非 JWT 标准库），构造时强校验 `AUTH_SECRET` ≥ 32 字节，否则拒绝启动。支持两种签发格式：
- `Issue`（旧式/兼容格式）：`base64url(payload) + "." + base64url(HMAC-SHA256(payload))`，payload 仅含 `sub`（用户 ID）、`role`、`exp`，无 header 与 iat，标注 Deprecated；
- `IssueJWT`（标准 JWT）：三段式 `header.payload.sig`，alg=HS256，含 `iat`。

`Verify` 透明兼容两段/三段：按 `.` 分段数判定格式，用 `hmac.Equal` 常量时间比较防时序攻击；解码 payload 后校验 `sub` 非空且 `exp` 未过期，返回 `domain.Actor{ID, Role}`。Access Token 默认 15 分钟。

**authenticate 中间件**（`Server.authenticate`）按路径分级放行：
- **直接放行白名单**：`/healthz`、`/`、`/admin`、`/favicon.ico`、公开 `/uploads/`（但 `/uploads/private/`（身份证影像等）必须走 token 校验）、`/swagger/`、`/api/v1/admin/token`、`/api/v1/auth/*`、`/api/auth/*`、`/api/v1/webhooks/*`；
- **`/api/services/*` 配置层**：GET 公开，POST 会覆盖全局配置（platform_config/services_config），中间件只"尽力解析" token 注入 context，由 handler 自行校验角色；
- **公开 GET 端点**（`isPublicPath` 前缀匹配）：命中且携带有效 token 时解析 actor 入 context（供 `jobs/mine`、`certificates/mine` 等区分登录态）；`isPublicPath` 只放行精确路径和单层详情（`/api/v1/jobs/{id}`），嵌套子资源（如 `demands/{id}/applications`）不放行——注释明确记载此前前缀匹配 bug 曾导致竞标者信息未鉴权泄露；`certified-pilots/mine` 被特判为需要认证；
- **其余路径**：强制 `Authorization: Bearer`，验证失败返回 401。

actor 通过 `contextWithActor`/`authenticatedActor` 在 context 中传递。

**adminGate**：包在 authenticate 之内，只拦截 `/api/v1/admin/*`（`/api/v1/admin/token` 豁免），要求 `platform_admin` 或 `association_admin`，否则 403。

## 2. auth_wechat.go —— 微信登录 + refresh 轮转

**`wechatLogin`（POST /api/v1/auth/wechat/login）**：读取 `WECHAT_APPID/APPSECRET` → `service.WeChatLogin` 调 `jscode2session`（10s 超时、URL 编码参数）→ 按返回 openid 查用户，不存在则自动创建（ID 为 `user-时间戳`，role=individual）。**关键兜底**：code2Session 失败且 `adminDevMode()` 为真时，所有 code 一律映射为 `openid="dev-fixed"` 的共享开发账号（避免每次登录新建行）——这是踩坑记录中"微信一键登录变成同一个号"的根因链。角色为空时按 `SUPER_ADMIN_PHONE` 匹配则升为 platform_admin。签发 access（IssueJWT，15min）+ refresh（`GenerateRefreshToken` 随机 32B，库中只存 `HashToken` 的 SHA-256 哈希，7 天过期）。

**`refreshToken`（POST /api/v1/auth/refresh）**：按 refresh 哈希 `Find` → 校验未撤销、未过期 → **先 `Revoke` 旧令牌，再签发新 access + 新 refresh**（轮转，旧令牌即作废）。注意三点：① 此处 access 用旧式 `Issue` 而非 IssueJWT，与登录路径不一致；② 若用户查不到则 role 默认 individual；③ 若新 refresh Store 失败，旧令牌已被 Revoke → 用户被锁死。

**`logout`**：撤销 refresh token 并写审计。**`updateMe`**：校验手机号格式（`^1[3-9]\d{9}$`），phone 由 repository 加密落库。**`me`**：返回用户资料 + demand_count/cert_count——其中 `demand_count` 实为 `s.demands.List(空 filter)` 的全量计数，不是当前用户的需求数（**疑似 bug**）；手机号从 `u.PhoneCipher`（FindByID 后已解密的明文）读出。

## 3. auth_sms.go —— 短信验证码登录

验证码存**进程内 `sync.Map`**（phone → code+过期时间）：5 分钟 TTL、60 秒防重发（按剩余有效期判断）。`genSMSCode` 用 crypto/rand 生成 6 位数字。**TODO 注释明确**：未接入腾讯云/阿里云短信商，发送处为空实现，仅 `adminDevMode()` 时响应回显 `dev_code` 便于调试。`loginWithSMS`（POST /api/auth/login-code）：校验 code 后即登录，未注册手机号自动注册（`WechatOpenID="phone:"+手机号` 保证 UNIQUE 不冲突，与踩坑记录一致），签发与微信登录相同的 token 对。**风险**：验证码无尝试次数限制、无 IP 维度限流、比较非常量时间、内存态无法水平扩展；一旦真实短信未接入而上线，任何人可任设手机号+任意 6 位码（配合 dev_code 更甚）。

## 4. admin_handler.go —— ADMIN_DEV_MODE 开发令牌

`adminDevMode()` 即 `os.Getenv("ADMIN_DEV_MODE")=="true"`。`adminDevLogin`（POST /api/v1/admin/token）：非 dev 模式直接 403（"dev token disabled in production"）；dev 模式接受请求体 `role`（非法值回退 platform_admin），签发 **2 小时** JWT，用户 ID 固定为 `"admin-dev"`（**不落 users 表**，注释说明固定 ID 是为保证站内消息按 receiver 隔离时历史消息可见）。生产环境必须关闭此开关。

## 5. H5 兼容层 —— 生产注册 vs dev-only

**生产无条件注册（`registerH5AuthRoutes`）**：`POST /api/auth/login`（h5AuthLogin，bcrypt 校验，DB 优先、users.json 遗留兜底）、`register`（bcrypt 落库 + 同步写 users.json）、`GET /api/auth/me`（手动解析 token，DB→JSON 回退）、`refresh`、`logout`、`send-code`、`login-code`、`GET|POST /api/services/config`（POST 校验管理员角色，另同步 _home.banners/notices 到 platform_config.json）、`POST /api/submit`（h5SubmitApplication，落 `service_applications` 表，FormData JSONB 全量入库）。

**仅 dev 注册（`registerH5Compat` + `registerCompatRoutes`，`adminDevMode()` 门控，server.go 中明确警告 "NOT FOR PRODUCTION"）**：全部 JSON 文件路由——applications/cases/case-categories/reviews/users.json 的增删改查、管理后台 stats/用户/角色修改、WeChat OAuth URL（mock）、SSO（恒 400）、image proxy、client-ip、upload。`init()` 会在工作目录自动创建这些 JSON 文件。**C9 修复注释**：曾存在未注册的 `passwordLogin/passwordRegister/getMeLegacy` 死代码——passwordLogin 不校验密码即签发 token 且 refresh 明文入库，一旦重新挂载即安全事故，已整体删除。另外 `wxPhone`（dev 下直接收 phone、生产路径返回伪造脱敏号）仅 dev 注册。

**h5ImageProxy 风险**：`/uploads/` 路径做了 filepath.Clean + 前缀校验防穿越，但 http/https 任意 URL 直接 302 跳转 → 开放重定向。

## 6. middleware.go —— 输入消毒 + 统一错误格式

- `SanitizeBody` **实为空操作**：只检查 Method/Content-Type 后直接放行，从不读 body；注释自述"消毒在 decode/respond 层处理，此处仅防御纵深"——实际无纵深；
- `SanitizeString`：正则去 HTML 标签、TrimSpace、超 10000 字符截断；`SanitizeMap` 递归处理字符串值；
- `WriteError`：统一 `{"error":{"code","message"}}`；`WriteJSON`：统一 `{"data":...}`。

## 7. config.go —— 配置加载与验证

`Load()` 集中读环境变量：HTTP_ADDR（默认 :8080）、ENV、CORS_ORIGINS、BASE_URL、AUTH_SECRET、ACCESS_TOKEN_TTL（默认 900s）、REFRESH_TOKEN_TTL（默认 604800s）、SUPER_ADMIN_PHONE、WECHAT_APPID/APPSECRET、DATABASE_URL（非空即用 PG，否则内存存储）。`Validate()`：AUTH_SECRET 缺失/短于 32B 为 Error；生产环境微信密钥缺失为 Error；PG 未用 `sslmode=require` 与未设 DATABASE_URL 为 Warning。`Print()` 对密钥类字段脱敏（保留首尾 4 位）。

## 8. crypto/ —— AES-256-GCM 加密与脱敏

`Cipher`：`NewCipher` 要求 base64 解码后恰好 32 字节密钥；`Encrypt` 用 crypto/rand 生成随机 nonce，`gcm.Seal(nonce, nonce, ...)` 将 nonce 前置后整体 base64（同明文每次密文不同，测试验证）；`Decrypt` 拆 nonce 校验长度后 Open。空串不加密（返回空串）。脱敏函数：`MaskPhone`（138\*\*\*\*5678）、`MaskCreditCode`（首 4 尾 4）、`MaskIDCard`（首 3 尾 4），短串全星号。

## 9. cache.go 与 logger.go

**cache**：`sync.RWMutex + map` 的内存 TTL 缓存，默认 60s，`New` 启动 5 分钟周期的后台清理协程（无 Close/停止机制）；`Get` 惰性删除过期项；`GetOrSet` 无 single-flight（并发下缓存击穿会重复执行 fn）；`Global` 为全局实例，定义 `Key*` 常量（仿 JS CacheKeys）供存储层使用。

**logger**：双通道——slog（生产 JSON/开发 Text 到 stdout）+ 自写 `writeFile` 追加到 `LOG_DIR` 下按日命名 `YYYY-MM-DD.log`；级别由 LOG_LEVEL 与 ENV 双重决定（内部 int 映射 0-3 与 slog.Level 两套状态，存在双源不一致隐患）；`Fatal` 记录后直接 `os.Exit(1)`（跳过 defer 清理）；`RequestLog`/`ErrorWithContext` 附带 method/path/ip/status/duration 结构化字段，body 截断 500 字符。

## 10. 已知问题汇总

1. `me` 接口 demand_count 是全表计数而非个人计数；
2. refresh 轮转先撤销后落库，Store 失败即锁号；且 refresh 路径签发的 access 格式与登录不一致（旧式 vs JWT）；
3. 微信 dev 兜底 `dev-fixed` 共享账号——打包误含 .env 导致生产 code2Session 失败时会全员复用同一账号（踩坑记录，须 `--exclude='.env'`）；
4. SMS 验证码无尝试限制/无真实短信商/内存态，仅适合开发；
5. `SanitizeBody` 形同虚设；
6. h5ImageProxy 开放重定向；
7. JSON 文件层（applications/cases/reviews/users.json）若 ADMIN_DEV_MODE 误开即暴露无鉴权写路径；
8. cache 无 single-flight、清理协程不可停；logger 级别双源。
