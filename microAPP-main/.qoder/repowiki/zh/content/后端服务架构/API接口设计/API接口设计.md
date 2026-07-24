# API接口设计

<cite>
**本文引用的文件**
- [backend/index.js](file://backend/index.js)
- [backend/config.js](file://backend/config.js)
- [backend/middleware/auth.js](file://backend/middleware/auth.js)
- [backend/middleware/validation.js](file://backend/middleware/validation.js)
- [backend/middleware/error.js](file://backend/middleware/error.js)
- [backend/storage.js](file://backend/storage.js)
- [backend/platformAuth.js](file://backend/platformAuth.js)
- [backend/routes/auth.js](file://backend/routes/auth.js)
- [backend/routes/admin.js](file://backend/routes/admin.js)
- [backend/routes/medical.js](file://backend/routes/medical.js)
- [frontend/h5/src/utils/http.js](file://frontend/h5/src/utils/http.js)
- [frontend/miniprogram/utils/request.js](file://frontend/miniprogram/utils/request.js)
- [frontend/h5/src/stores/medical.js](file://frontend/h5/src/stores/medical.js)
- [backend/package.json](file://backend/package.json)
- [backend/OPTIMIZATION_SUMMARY.md](file://backend/OPTIMIZATION_SUMMARY.md)
- [backend/QUICK_REFERENCE.md](file://backend/QUICK_REFERENCE.md)
- [docs/接入文档/Apifox使用说明.md](file://docs/接入文档/Apifox使用说明.md)
- [docs/医疗配送模块功能设计.md](file://docs/医疗配送模块功能设计.md)
</cite>

## 目录
1. [简介](#简介)
2. [项目结构](#项目结构)
3. [核心组件](#核心组件)
4. [架构总览](#架构总览)
5. [详细组件分析](#详细组件分析)
6. [依赖分析](#依赖分析)
7. [性能考虑](#性能考虑)
8. [故障排查指南](#故障排查指南)
9. [结论](#结论)
10. [附录](#附录)

## 简介
本文件面向低空飞行服务平台的API接口设计，基于现有后端代码库进行系统化梳理与规范化输出，涵盖RESTful设计原则、接口命名约定、HTTP状态码使用、请求/响应格式、参数校验、错误处理、版本管理、速率限制与安全防护，并提供接口文档模板、常见业务场景实现示例、错误码定义与处理流程，以及设计新接口与兼容演进的最佳实践。

**更新** 本次更新重点关注医疗配送模块的API接口重构，包括认证系统、订单处理逻辑、估价系统从费用计算改为时间估算，以及新增的超时检查和短信通知功能。

## 项目结构
后端采用Express框架，按职责拆分为配置、中间件、路由与存储层；前端提供H5与小程序两套客户端，统一通过拦截器处理鉴权与刷新流程。新增的医疗配送模块包含独立的路由和存储层。

```mermaid
graph TB
subgraph "后端"
A["入口 index.js"]
B["配置 config.js"]
C["中间件<br/>auth.js / validation.js / error.js"]
D["路由<br/>routes/auth.js / routes/admin.js / routes/medical.js"]
E["存储 storage.js"]
F["平台对接 platformAuth.js"]
end
subgraph "前端"
G["H5 http.js"]
H["小程序 request.js"]
I["H5 医疗配送 store"]
end
G --> A
H --> A
I --> A
A --> B
A --> C
A --> D
D --> E
D --> F
```

**图表来源**
- [backend/index.js:1-1673](file://backend/index.js#L1-L1673)
- [backend/config.js:1-123](file://backend/config.js#L1-L123)
- [backend/middleware/auth.js:1-168](file://backend/middleware/auth.js#L1-L168)
- [backend/middleware/validation.js:1-192](file://backend/middleware/validation.js#L1-L192)
- [backend/middleware/error.js:1-90](file://backend/middleware/error.js#L1-L90)
- [backend/storage.js:1-361](file://backend/storage.js#L1-L361)
- [backend/platformAuth.js:1-177](file://backend/platformAuth.js#L1-L177)
- [backend/routes/auth.js:1-591](file://backend/routes/auth.js#L1-L591)
- [backend/routes/admin.js:1-514](file://backend/routes/admin.js#L1-L514)
- [backend/routes/medical.js:1-1094](file://backend/routes/medical.js#L1-L1094)
- [frontend/h5/src/utils/http.js:1-99](file://frontend/h5/src/utils/http.js#L1-L99)
- [frontend/miniprogram/utils/request.js:1-144](file://frontend/miniprogram/utils/request.js#L1-L144)
- [frontend/h5/src/stores/medical.js:1-301](file://frontend/h5/src/stores/medical.js#L1-L301)

**章节来源**
- [backend/index.js:1-1673](file://backend/index.js#L1-L1673)
- [backend/package.json:1-34](file://backend/package.json#L1-L34)

## 核心组件
- 配置中心：集中管理JWT、微信、SSO、数据库、上传、服务器等配置项，并提供配置校验与打印。
- 中间件体系：认证与权限、输入消毒与校验、速率限制、全局错误处理。
- 路由模块：认证路由（登录/注册/刷新/登出/SSO/微信登录）、管理路由（统计/申请/用户/服务配置/评价等）、**医疗配送路由**（起降场管理、寄件人认证、常用联系人、配送订单、评价）。
- 存储层：统一读写封装，支持JSON文件与PostgreSQL双模式，内置缓存与一致性保障，**新增医疗配送专用存储**。
- 平台对接：与"畅行温州"平台的SM2/SM4加解密与签名交互。
- 前端拦截器：统一注入Authorization头，401自动刷新令牌并重试，**新增医疗配送状态管理store**。

**更新** 新增医疗配送模块的完整路由和存储实现，包括认证、订单管理、起降场管理等功能。

**章节来源**
- [backend/config.js:1-123](file://backend/config.js#L1-L123)
- [backend/middleware/auth.js:1-168](file://backend/middleware/auth.js#L1-L168)
- [backend/middleware/validation.js:1-192](file://backend/middleware/validation.js#L1-L192)
- [backend/middleware/error.js:1-90](file://backend/middleware/error.js#L1-L90)
- [backend/storage.js:1-361](file://backend/storage.js#L1-L361)
- [backend/platformAuth.js:1-177](file://backend/platformAuth.js#L1-L177)
- [frontend/h5/src/utils/http.js:1-99](file://frontend/h5/src/utils/http.js#L1-L99)
- [frontend/miniprogram/utils/request.js:1-144](file://frontend/miniprogram/utils/request.js#L1-L144)
- [frontend/h5/src/stores/medical.js:1-301](file://frontend/h5/src/stores/medical.js#L1-L301)

## 架构总览
下图展示从客户端到后端路由、中间件、存储与外部平台的整体调用链路，包括新增的医疗配送模块。

```mermaid
sequenceDiagram
participant FE as "前端(H5/小程序)"
participant INT as "拦截器/中间件"
participant RT as "路由(auth/admin/medical)"
participant ST as "存储(storage)"
participant MF as "医疗配送存储"
participant PF as "平台对接(platformAuth)"
FE->>INT : 发起请求(带Authorization)
INT->>RT : 校验/限流/参数清洗
RT->>ST : 读取/写入基础数据
RT->>MF : 读取/写入医疗配送数据
RT->>PF : 调用外部平台(SSO)
PF-->>RT : 返回平台数据
ST-->>RT : 返回业务数据
MF-->>RT : 返回医疗配送数据
RT-->>FE : 统一响应(success/data/错误)
```

**图表来源**
- [backend/routes/auth.js:1-591](file://backend/routes/auth.js#L1-L591)
- [backend/routes/admin.js:1-514](file://backend/routes/admin.js#L1-L514)
- [backend/routes/medical.js:1-1094](file://backend/routes/medical.js#L1-L1094)
- [backend/middleware/auth.js:1-168](file://backend/middleware/auth.js#L1-L168)
- [backend/middleware/validation.js:1-192](file://backend/middleware/validation.js#L1-L192)
- [backend/storage.js:1-361](file://backend/storage.js#L1-L361)
- [backend/platformAuth.js:1-177](file://backend/platformAuth.js#L1-L177)
- [frontend/h5/src/utils/http.js:1-99](file://frontend/h5/src/utils/http.js#L1-L99)
- [frontend/miniprogram/utils/request.js:1-144](file://frontend/miniprogram/utils/request.js#L1-L144)

## 详细组件分析

### 认证与权限中间件
- authRequired：校验Bearer Token，支持过期与无效错误细分。
- roleRequired：基于用户角色的权限控制。
- optionalAuth：可选认证，不影响请求流程。
- rateLimit：基于内存的滑动窗口限流，默认15分钟窗口与最大请求数可配。

```mermaid
flowchart TD
Start(["进入路由"]) --> CheckAuth["校验Authorization头"]
CheckAuth --> |缺失| Resp401["返回401 未提供认证令牌"]
CheckAuth --> |存在| Verify["jwt.verify校验"]
Verify --> |失败| Resp401b["返回401 无效的认证令牌/TOKEN_EXPIRED"]
Verify --> |成功| AttachUser["附加用户信息到req.user"]
AttachUser --> RoleCheck{"角色允许?"}
RoleCheck --> |否| Resp403["返回403 权限不足"]
RoleCheck --> |是| Next["进入业务处理"]
```

**图表来源**
- [backend/middleware/auth.js:11-75](file://backend/middleware/auth.js#L11-L75)

**章节来源**
- [backend/middleware/auth.js:1-168](file://backend/middleware/auth.js#L1-L168)

### 输入验证与消毒中间件
- sanitizeBody/sanitizeString/sanitizeObject：移除脚本标签、事件属性、javascript协议，清洗字符串。
- requireFields：必填字段校验。
- validateQuery：查询参数类型、范围、枚举与长度校验。
- validateFileUpload：文件类型与大小校验。

```mermaid
flowchart TD
Enter(["请求进入"]) --> Sanitize["sanitizeBody清洗请求体"]
Sanitize --> Fields["requireFields校验必填"]
Fields --> Query["validateQuery校验查询参数"]
Query --> File["validateFileUpload校验文件"]
File --> Ok{"全部通过?"}
Ok --> |否| Err400["返回400 参数验证失败"]
Ok --> |是| Proceed["进入业务处理"]
```

**图表来源**
- [backend/middleware/validation.js:52-149](file://backend/middleware/validation.js#L52-L149)

**章节来源**
- [backend/middleware/validation.js:1-192](file://backend/middleware/validation.js#L1-L192)

### 全局错误处理
- errorHandler：统一捕获错误，区分Multer、JWT、数据库约束等错误类型，生产环境不暴露堆栈。
- notFoundHandler：404接口不存在。
- asyncHandler：异步函数错误包装，避免try-catch重复。

```mermaid
flowchart TD
Try(["业务处理Promise"]) --> Catch["asyncHandler捕获异常"]
Catch --> Log["记录错误日志"]
Log --> Type{"错误类型?"}
Type --> |JWT过期| R401Exp["返回401 TOKEN_EXPIRED"]
Type --> |JWT无效| R401Inv["返回401 无效的认证令牌"]
Type --> |文件过大| R400Size["返回400 文件过大"]
Type --> |约束冲突| R400Constr["返回400 数据违反约束"]
Type --> |其他| R500["返回500 服务器内部错误"]
```

**图表来源**
- [backend/middleware/error.js:9-83](file://backend/middleware/error.js#L9-L83)

**章节来源**
- [backend/middleware/error.js:1-90](file://backend/middleware/error.js#L1-L90)

### 存储与缓存
- 支持JSON文件与PostgreSQL两种存储后端，自动切换。
- 读取使用缓存，写入主动清缓存，缓存键与过期策略明确。
- 提供初始化与双序列化兼容处理。
- **新增医疗配送专用存储**：订单、认证、起降场、联系人、评价、短信日志等数据的独立存储。

```mermaid
flowchart TD
ReadReq["读取请求"] --> Mode{"USE_POSTGRES?"}
Mode --> |否| ReadFile["读取JSON文件"]
Mode --> |是| ReadPG["查询PostgreSQL"]
ReadFile --> CacheGet["cache.getOrSet"]
ReadPG --> CacheGet
CacheGet --> Return["返回数据"]
WriteReq["写入请求"] --> ClearCache["cache.delete"]
ClearCache --> Persist{"持久化"}
Persist --> |JSON| WriteFile["写入JSON文件"]
Persist --> |PG| WritePG["写入PostgreSQL"]
```

**图表来源**
- [backend/storage.js:42-132](file://backend/storage.js#L42-L132)

**章节来源**
- [backend/storage.js:1-361](file://backend/storage.js#L1-L361)

### 平台对接（SSO）
- 与"畅行温州"平台交互，使用SM2签名与SM4加密，构建请求体并校验响应。
- 提供queryMemberByAuthCode封装调用。

```mermaid
sequenceDiagram
participant API as "后端API"
participant Plat as "平台接口"
API->>API : SM4加密data与hdata
API->>API : SHA1摘要+SM2签名
API->>Plat : POST /member/.../query/V1
Plat-->>API : 返回dataenc与签名
API->>API : SM4解密dataenc
API-->>API : 校验结果/解析数据
```

**图表来源**
- [backend/platformAuth.js:165-172](file://backend/platformAuth.js#L165-L172)

**章节来源**
- [backend/platformAuth.js:1-177](file://backend/platformAuth.js#L1-L177)

### 前端拦截与令牌刷新
- H5与小程序均在请求前注入Authorization头。
- 401时自动调用刷新接口，队列化并发刷新，成功后重试原请求。
- **新增医疗配送store**：统一管理认证状态、起降场、联系人、订单等状态。

```mermaid
sequenceDiagram
participant Client as "前端客户端"
participant Inter as "请求拦截器"
participant Auth as "刷新接口"
Client->>Inter : 发起请求(带Token)
Inter->>Auth : 401触发刷新(refreshToken)
Auth-->>Inter : 返回新AccessToken
Inter->>Client : 重试原请求(新Token)
```

**图表来源**
- [frontend/h5/src/utils/http.js:19-78](file://frontend/h5/src/utils/http.js#L19-L78)
- [frontend/miniprogram/utils/request.js:36-134](file://frontend/miniprogram/utils/request.js#L36-L134)

**章节来源**
- [frontend/h5/src/utils/http.js:1-99](file://frontend/h5/src/utils/http.js#L1-L99)
- [frontend/miniprogram/utils/request.js:1-144](file://frontend/miniprogram/utils/request.js#L1-L144)
- [frontend/h5/src/stores/medical.js:1-301](file://frontend/h5/src/stores/medical.js#L1-L301)

### 医疗配送模块API设计

**更新** 新增完整的医疗配送模块API接口设计，包括认证、订单管理、起降场管理等功能。

#### 认证相关接口
- GET `/api/medical/certification/status` - 查询当前用户认证状态
- POST `/api/medical/certification/apply` - 提交认证申请
- POST `/api/medical/certification/resubmit` - 重新提交认证（驳回后）

#### 起降场相关接口
- GET `/api/medical/pads` - 获取启用的起降场列表（C端）
- GET `/api/medical/pads/all` - 获取全部起降场（管理端）
- POST `/api/medical/pads` - 新增起降场
- PUT `/api/medical/pads/:id` - 编辑起降场
- DELETE `/api/medical/pads/:id` - 删除起降场

#### 订单相关接口
- POST `/api/medical/orders` - 创建配送订单
- GET `/api/medical/orders/my` - 获取我的订单列表（C端）
- GET `/api/medical/orders/:id` - 获取订单详情
- POST `/api/medical/orders/:id/cancel` - 取消订单（C端/管理端）
- POST `/api/medical/orders/:id/reorder` - 再次下单（获取复制数据）
- GET `/api/medical/orders` - 获取全部订单（管理端）
- POST `/api/medical/orders/:id/accept` - 管理端接单
- POST `/api/medical/orders/:id/pickup` - 标记待取件
- POST `/api/medical/orders/:id/deliver` - 标记配送中
- POST `/api/medical/orders/:id/delivered` - 标记已送达
- POST `/api/medical/orders/:id/complete` - 标记已完成
- POST `/api/medical/orders/:id/exception` - 标记异常
- GET `/api/medical/orders/estimate` - 获取预估费用与时间

#### 常用联系人接口
- GET `/api/medical/contacts` - 获取常用联系人列表
- POST `/api/medical/contacts` - 新增常用联系人
- PUT `/api/medical/contacts/:id` - 编辑常用联系人
- DELETE `/api/medical/contacts/:id` - 删除常用联系人

#### 评价相关接口
- POST `/api/medical/orders/:id/rate` - 提交订单评价
- GET `/api/medical/orders/:id/rating` - 获取订单评价
- GET `/api/medical/ratings/stats` - 获取评价统计（管理端）

#### 管理端认证审核接口
- GET `/api/medical/certifications` - 获取认证申请列表
- POST `/api/medical/certifications/:id/approve` - 通过认证
- POST `/api/medical/certifications/:id/reject` - 驳回认证

**章节来源**
- [backend/routes/medical.js:1-1094](file://backend/routes/medical.js#L1-L1094)
- [docs/医疗配送模块功能设计.md:683-746](file://docs/医疗配送模块功能设计.md#L683-L746)

## 依赖分析
- Express：Web框架与路由。
- JWT：令牌签发与校验。
- Bcrypt：密码哈希。
- Multer：文件上传。
- Sharp：图片压缩。
- Axios：HTTP客户端。
- XLSX：Excel导出。
- PG：PostgreSQL驱动。
- Winston：日志（可选）。

```mermaid
graph LR
Express["express"] --> App["应用"]
JWT["jsonwebtoken"] --> AuthMW["认证中间件"]
Bcrypt["bcryptjs"] --> AuthRoute["认证路由"]
Multer["multer"] --> Upload["上传接口"]
Sharp["sharp"] --> Img["图片压缩接口"]
Axios["axios"] --> Plat["平台对接"]
XLSX["xlsx"] --> Admin["导出Excel"]
PG["pg"] --> Store["PostgreSQL存储"]
Winston["winston"] --> Logger["日志模块"]
```

**图表来源**
- [backend/package.json:12-28](file://backend/package.json#L12-L28)

**章节来源**
- [backend/package.json:1-34](file://backend/package.json#L1-L34)

## 性能考虑
- 缓存策略：用户/申请/案例/服务配置分别设置60秒/5分钟缓存，显著降低I/O与响应时间。
- **新增医疗配送缓存策略**：订单30秒、认证60秒、起降场5分钟、联系人60秒、评价60秒、短信日志60秒缓存。
- 存储层：读取走缓存、写入清缓存，避免脏读。
- 图片压缩：失败回退原图，设置合理缓存头。
- 速率限制：登录/注册接口限流，防暴力破解。
- **新增超时检查机制**：每5分钟扫描一次订单超时状态，自动提醒和处理。

**更新** 新增医疗配送模块的缓存策略和超时检查机制。

**章节来源**
- [backend/storage.js:134-181](file://backend/storage.js#L134-L181)
- [backend/index.js:86-114](file://backend/index.js#L86-L114)
- [backend/middleware/auth.js:108-160](file://backend/middleware/auth.js#L108-L160)
- [backend/routes/auth.js:18-21](file://backend/routes/auth.js#L18-L21)
- [backend/routes/medical.js:1022-1091](file://backend/routes/medical.js#L1022-L1091)

## 故障排查指南
- 401未提供/无效令牌：检查Authorization头格式与有效期；确认刷新接口可用。
- 403权限不足：确认用户角色与路由所需角色匹配。
- 400参数错误：检查必填字段、类型、范围与长度；关注文件类型与大小限制。
- 429请求频繁：检查限流配置与客户端重试策略。
- 500服务器错误：查看日志文件定位具体异常；生产环境不暴露堆栈。
- SSO失败：检查平台对接参数、SM2/SM4密钥与签名流程。
- **医疗配送订单异常**：检查认证状态、起降场有效性、物品重量范围、紧急程度选择。
- **短信通知失败**：检查短信模板配置，确认开发环境使用日志模拟。

**更新** 新增医疗配送模块相关的故障排查指导。

**章节来源**
- [backend/middleware/error.js:9-83](file://backend/middleware/error.js#L9-L83)
- [backend/middleware/auth.js:11-75](file://backend/middleware/auth.js#L11-L75)
- [backend/middleware/validation.js:79-149](file://backend/middleware/validation.js#L79-L149)
- [backend/middleware/auth.js:108-160](file://backend/middleware/auth.js#L108-L160)
- [backend/platformAuth.js:135-163](file://backend/platformAuth.js#L135-L163)
- [backend/routes/medical.js:492-629](file://backend/routes/medical.js#L492-L629)

## 结论
本项目在安全、性能与可维护性方面已形成较为完善的API设计与实现体系：统一的中间件链路确保输入安全与权限控制；模块化的路由与存储便于扩展；前端拦截器提供一致的鉴权体验；平台对接采用国密算法保证合规。**新增的医疗配送模块实现了完整的认证、订单管理、起降场管理等功能，估价系统从费用计算改为时间估算，新增超时检查和短信通知机制，为后续的支付和智能调度奠定了基础。** 建议后续在数据库、监控与API文档自动化方面持续演进。

**更新** 强调医疗配送模块的完整实现和功能特性。

## 附录

### RESTful设计规范与命名约定
- 资源命名：使用名词复数形式，如/users、/applications、/reviews。
- 动作映射：GET/POST/PUT/DELETE分别对应读取/创建/更新/删除。
- 版本管理：建议在路径中加入版本号，如/api/v1/，便于未来平滑演进。
- 命名风格：小写短横线或驼峰均可，保持前后端一致。
- **医疗配送命名规范**：使用/api/medical/前缀，保持与现有API的一致性。

**更新** 新增医疗配送模块的命名规范。

### HTTP状态码使用标准
- 200：成功，返回success与data。
- 201：资源创建成功。
- 400：参数错误/数据校验失败。
- 401：未认证/令牌无效/过期。
- 403：权限不足。
- 404：资源不存在。
- 409：资源冲突（如重复注册）。
- 429：请求过于频繁。
- 500：服务器内部错误。

### 请求/响应格式规范
- 成功响应：包含success与data字段；部分接口返回用户信息时包含令牌。
- 失败响应：包含success与message；生产环境不返回堆栈与detail。
- 分页响应：包含page、limit、total、totalPages。
- **医疗配送响应格式**：统一使用success/data结构，错误响应包含message字段。

**更新** 新增医疗配送模块的响应格式规范。

### 参数验证机制
- 必填字段：requireFields。
- 查询参数：validateQuery（类型、范围、枚举、长度）。
- 文件上传：validateFileUpload（类型与大小）。
- 输入消毒：sanitizeBody/sanitizeString。

### 错误处理策略
- 统一错误处理器：errorHandler。
- 404处理：notFoundHandler。
- 异步包装：asyncHandler。
- 日志记录：结构化日志，包含请求上下文与错误详情。

### API版本管理
- 建议在路由前缀加入版本号，如/api/v1/。
- 保持向后兼容，新增字段时默认可选，避免破坏既有客户端。
- 通过配置开关或路由分组管理多版本并行。

### 速率限制
- 登录/注册接口默认限流，防止暴力破解。
- 可根据业务场景调整窗口与阈值，结合Redis实现分布式限流。

### 安全防护措施
- JWT：强密钥、合理TTL、刷新令牌轮换。
- 输入消毒：移除脚本与事件属性，防止XSS。
- 参数校验：严格的数据类型与范围校验。
- 权限隔离：按角色过滤数据与操作范围。
- 平台对接：SM2/SM4加解密与签名，确保数据安全。

### 接口文档模板
- 路径：/api/...（建议加入版本号）
- 方法：GET/POST/PUT/DELETE
- 认证：Bearer Token（部分接口可选）
- 请求头：Content-Type: application/json
- 成功示例：返回success与data；失败示例：返回success与message
- 分页：page、limit、total、totalPages
- 错误码：参考错误处理章节
- **医疗配送接口模板**：包含认证状态查询、订单创建、状态流转等完整流程

**更新** 新增医疗配送模块的接口文档模板。

### 常见业务场景实现示例
- 用户登录：/api/auth/login（用户名/手机号+密码），返回令牌与用户信息。
- 注册：/api/auth/register（手机号+密码+姓名），返回成功信息。
- 获取当前用户：/api/auth/me（需认证）。
- 刷新令牌：/api/auth/refresh（提交refreshToken）。
- 管理员统计：/api/admin/stats（需管理员角色）。
- 申请列表：/api/admin/applications（支持分页与筛选）。
- 服务配置：/api/admin/services/config（按角色可见/可改）。
- 评价管理：/api/admin/reviews（审核/删除）。
- **医疗配送认证**：/api/medical/certification/status（查询认证状态）。
- **医疗配送下单**：/api/medical/orders（创建订单，需认证）。
- **医疗配送订单管理**：/api/medical/orders/my（获取个人订单列表）。
- **医疗配送起降场管理**：/api/medical/pads（获取启用起降场）。

**更新** 新增医疗配送模块的业务场景示例。

### 设计新API接口步骤
- 明确资源与动作，选择合适HTTP方法与路径。
- 定义请求/响应模型，统一字段命名与类型。
- 在路由层编写处理逻辑，引入必要的中间件（认证、校验、限流）。
- 使用存储层读写数据，注意缓存一致性。
- 编写错误处理与日志记录。
- 前端拦截器自动注入Authorization，必要时增加重试与刷新逻辑。
- **医疗配送接口设计**：遵循现有命名规范，使用/api/medical/前缀，确保与前端store一致。

**更新** 新增医疗配送模块的接口设计步骤。

### 实现接口版本兼容
- 在路径中加入版本号，如/api/v1/。
- 保留旧接口一段时间，返回迁移提示或重定向。
- 对新增字段默认可选，避免破坏既有客户端。
- 通过配置或路由分组管理多版本并行。

### 优化API性能与安全性
- 使用缓存减少I/O，写入时清缓存。
- 图片压缩与CDN，失败回退原图并设置缓存头。
- 速率限制与白名单策略。
- 强JWT密钥与短TTL，及时刷新。
- 输入消毒与参数校验，最小权限原则。
- **医疗配送缓存优化**：针对高频访问的起降场、订单等数据设置合理的缓存策略。

**更新** 新增医疗配送模块的性能优化建议。

### 错误码定义与处理流程
- 400：参数/数据校验失败。
- 401：未认证/令牌无效/过期。
- 403：权限不足。
- 404：资源不存在。
- 409：资源冲突。
- 429：请求过于频繁。
- 500：服务器内部错误。

### SSO对接与调试
- 使用Apifox或脚本生成请求体，包含签名与加密字段。
- 注意每次运行生成新的会话号与签名，避免复用。
- 在后端通过queryMemberByAuthCode调用平台接口。

### 医疗配送模块特殊说明
- **估价系统重构**：从费用计算改为时间估算，提供estimated_minutes字段。
- **认证要求**：下单前必须通过寄件人认证，收货人也需通过认证。
- **超时检查**：自动检测订单超时状态，发送提醒和异常处理。
- **短信通知**：一期使用日志模拟，记录通知内容和发送状态。
- **状态流转**：完整的订单状态管理，支持复杂的业务流程。

**更新** 新增医疗配送模块的特殊功能说明。

**章节来源**
- [backend/routes/admin.js:101-110](file://backend/routes/admin.js#L101-L110)
- [backend/middleware/error.js:56-61](file://backend/middleware/error.js#L56-L61)
- [backend/routes/medical.js:453-489](file://backend/routes/medical.js#L453-L489)
- [backend/routes/medical.js:1022-1091](file://backend/routes/medical.js#L1022-L1091)
- [docs/医疗配送模块功能设计.md:186-238](file://docs/医疗配送模块功能设计.md#L186-L238)