# 「后端服务与仓储层」完整报告

## 一、Repository 接口层（`internal/repository/repositories.go`）

约 **59 个接口**（另含 `AuditWriter`），全部只收/返 domain 类型，禁止业务规则、权限判断与 HTTP 知识。`AuditEntry` 记录每次写操作（actor/action/resource/result/request_id/metadata），由 HTTP 层 audit() 在 create/approve/reject/void 等动作后写入。按业务域分组：

**① 用户与认证**
- `UserRepository`：FindByOpenID / Create / FindByID / All / UpdateRole / UpdateAvatar / UpdateName / UpdateProfile（Phone 明文传入、实现加密落库，空 Phone 表示不修改）/ Delete（事务级联）
- `RefreshTokenRepository`：Store / Find / Revoke（轮转刷新令牌）

**② 产业供需对接**
- `DemandRepository`：Create / Update / FindByID / List（公开语义：仅已发布）/ ListAll（管理端全量，含待审核）/ Search / SetStatus / **CompareAndSetStatus（原子 CAS，并发抢单安全）** / Delete
- `BidRepository`：Create / FindByID / ListByDemand / ListByBidder / UpdateStatus（原子接单）
- `IntentRepository`（联系对接模式）：Create / ListByDemand / ListByIntentor / UpdateStatus
- `WorkOrderRepository`（接单派单闭环）：Create / FindByID / ListByPublisher / ListByWorker / UpdateStatus / UpdatePhotos / UpdateRework / UpdateCancel

**③ 企业 / 招聘 / 合同**
- `EnterpriseRepository`：Create / Update / FindByID / FindByOwner / ListByStatus（分页）/ Pending（待审队列）/ Search / Delete / AddDocument / ListDocuments
- `EmploymentRepository`：Create / ListByEnterprise / ListAll
- `JobRepository`：Create / Update / FindByID / ListByEnterprise / ListAll（含草稿）/ ListPublished / Delete
- `ResumeRepository`：Create / Update / FindByID / ListByUser / ListAll
- `JobApplicationRepository`：Create / FindByID / UpdateStatus / ListByJob / ListByApplicant
- `ContractRepository` + `ContractTemplateRepository`：创建/列表/FindByID/UpdateStatus（签约状态机）

**④ 社区 / 交易**
- `PostRepository` / `CommentRepository` / `ReportRepository`（ListPending 举报队列）
- `ListingRepository`：Create / Update / FindByID / ListByStatus / ListBySeller / AddFavorite / RemoveFavorite
- `LabourOrderRepository`：订单 CRUD + CreateQuote / ListQuotes + CreateAssignment / ListAssignmentsByOrder / ListAssignmentsByWorker
- `TradeOrderRepository`：Create / FindByID / UpdateStatus / **UpdateAftersale（一次性写状态+售后字段，申请/审核结案共用）** / ListByUser / ListAll / Delete

**⑤ 培训 / 证书 / 金融**
- `CertificateRepository` / `CourseRepository` / `InstructorRepository` / `PilotRepository`（Update 支持驳回后覆盖重提）/ `EnrollmentRepository`（FindByUserAndCourse 防重复报名，FindByID 供管理端防回退校验）
- `EscrowRepository`：GetAccount + **Deposit/Freeze/Release/Refund 四个资金方法必须原子**（调整余额与流水同事务），余额不足返回哨兵错误 `ErrInsufficientBalance` / `ErrInsufficientFrozenBalance`；注释明确：旧读改写接口（GetAccount→UpsertAccount）存在并发丢更新，**已移除**

**⑥ 新增业务模块（biz，按协会需求）**
- `ExpertRepository`（智库专家）/ `CaseRepository`（项目案例）/ `ComplianceRepository`（合规文档+团体标准两套 CRUD）/ `AchievementRepository`（科技成果）/ `RDChallengeRepository`（研发难题）/ `ResearchProjectRepository`（课题攻关）/ `ProjectAppRepository`（项目申报）/ `CompetitionRepository`（赛事+报名）/ `EventRepository`（协会活动+报名）/ `PortfolioRepository`（会员品牌展示，List 管理端全量 vs ListPublished）/ `IndustryReportRepository`（行业报告）/ `ResourceRepository`（产业资源+预约 CreateBooking/ListBookingsBy*，C11）/ `EmergencyRepository`（应急资源+调度，C12）/ `ApplicationRepository`（小程序 /api/submit 服务申报）/ `ServiceListingRepository`（服务能力展示 PRD ②-2）/ `ReviewRepository` / `VenueRepository`（场地+预约）/ `MessageRepository`（站内信）/ `ArticleRepository`（资讯）/ `InspectionRepository`（年检）/ `LoanRepository`（贷款）/ `RepairRepository` / `PolicyRepository`（保险）

**⑦ Batch1-3 补充模块**
- Batch1：`ResourcePoolRepository`（资源池+成员）、`TestSiteRepository`（测试场地+预约审核，含我的预约/管理端全量）、`ExhibitionRepository`（展会+展位）
- Batch2：`TransformationRepository`（成果转化）、`CollegeRepository`（院校）、`StudyTourRepository`（研学）、`CooperationRepository`（校企共建）
- Batch3：`RescueCaseRepository`（救援案例）、`EmergencyDeptRepository`（应急部门+联合演练）、`AssociationMemberRepository`（协会成员 7/8 级权限）

## 二、PostgreSQL 实现（`internal/repository/postgres/`，非测试文件）

**Store 与连接**：`type Store struct { pool *pgxpool.Pool; cipher *crypto.Cipher }`；`NewStore` 用 `pgxpool.ParseConfig` 解析 `DATABASE_URL`、`MaxConns=50`、`pgxpool.NewWithConfig` 后 `Ping` 验证；暴露 `Pool()` / `Close()`。敏感字段（联系方式、营业执照 URL、对公账户名、手机号）通过 `cipher`（AES-256-GCM）加密落库、读时解密。

**迁移机制（migration.go）**：`RunMigrationsFromDir(ctx, dir)` 扫描目录下 `*.up.sql`（按 `versionOf` 提取数字前缀排序），执行未登记版本。要点：
1. 用固定键 `pg_advisory_lock` 咨询锁（挂在独立 Acquire 连接上）保证多实例并发启动（docker compose --scale）只有一个执行迁移；
2. `schema_migrations` 版本表记录 applied；
3. `applyMigration` 把"执行 SQL + 登记版本"放在**同一事务**，任一步失败整体回滚，不会出现"表已建但版本未记"；
4. 旧库兼容注释：此前版本每次启动全量重放且无版本表，首次带版本表启动会全量执行一遍（所有迁移均幂等）并登记；
5. `MigrationsDir()` 优先 `MIGRATIONS_DIR` 环境变量（Docker -trimpath 编译下 runtime.Caller 不可靠）。

**审计**：`audit.go` 定义 `AuditEntry` 与 `Store.WriteAudit`（写 `audit_logs` 表，metadata JSON 序列化）；`audit_adapter.go` 的 `AuditAdapter`（NewAuditAdapter）把 pg Store 桥接成 `repository.AuditWriter` 接口，供 HTTP 层统一调用。

**postgres.go（2295 行）实现的 repo**（约 25 个）：Demand、Enterprise、Employment、Contract、ContractTemplate、Job、Resume、JobApplication、Post、Comment、Report、Listing、LabourOrder、User、RefreshToken、Intent、WorkOrder、Bid、College、StudyTour、Exhibition、TestSite、Transformation——其余分布在 phase3_repos(_2).go、biz_repos(_2,_3).go、service_listing_repos.go、batch3_repos.go。

**已知问题/踩坑注释**：
- `jsonbSlice[T]` 保证 JSONB 数组列写入非 NULL——pgx v5 把 nil slice 编码为 SQL NULL 会违反 `NOT NULL DEFAULT '[]'`（此前 `training_courses.tags` 踩过错误码 23502）；
- `userRepo.Delete` 必须事务内先删 `refresh_tokens`/`user_roles` 再删 users，否则外键阻塞删除；
- `scanEmploymentPaged`/`scanContractsPaged` 注明 where 子句必须是编译期常量，绝不可传用户输入（防注入）；
- 历史遗留 `// scanContracts removed — replaced by scanContractsPaged`；
- `biz_repos3.go` 的 `nullableEndTime` 把零值时间转 NULL（进行中/待响应调度无结束时间，回归测试覆盖）；
- `phase3_repos2.go`：Escrow 账户不存在返回零值账户（与内存实现一致）。

## 三、内存实现（`internal/repository/memory/`，非测试文件）

**实现方式**：每个 repo 都是 `struct { mu sync.RWMutex; items []T }`（或 depts/drills 双切片），写操作 `mu.Lock()`、读操作 `mu.RLock()`，语义与 PG 对齐（如 Demand List 仅返回已发布、ListAll 按 status/district/biz_type 过滤）。部分构造器内嵌演示种子数据（如 `NewDemandRepository` 预置 demand-001「重庆巡航科技」）。分文件：`memory.go`（2481 行，核心 25 个 repo）、`memory_new.go`/`memory_new2.go`（新增业务模块）、`memory_batch1/2/3.go`（Batch1-3 模块）。

**仅内存实现、无 PG 的 repo（历史遗留）**：`batch3_repos.go` 头部注释明确指出——资源池（resource pools）、校企合作（cooperation programs）、救援案例（rescue cases）、应急部门（emergency departments）、协会成员（association members）**共 5 个模块此前即使在 PG 模式下也回退到内存存储**（即 coop/rescueCase/emergDept/assocMember 4 个 + 资源池）；该文件现已补齐这 5 个的 PG 实现，遗留问题修复。另测试注释提到"内存实现忽略过滤，分类 tab 永远全量/空列表"（过滤参数差异）。

## 四、Service 层（`internal/service/services.go` + `biz_modules.go` 等）

**通用模式**（严格分层：Handler→Service→Repository，Service 禁触 HTTP/JSON/SQL/环境变量）：
1. **角色校验**：入口先查 `a.Role`（如 `RoleEnterprise`/`RolePlatformAdmin`/`RoleAssociationAdmin`），无权返回 error；
2. **归属校验**：`FindByID` 后核对 `d.PublisherID == a.ID` 等，防止越权操作；
3. **状态机迁移**：用 `validTransitions map[状态][]状态` 显式定义合法迁移（如 ContractService 的 Draft→Sent→Signing→Signed/Voided/Expired），或逐方法硬编码"当前状态必须为 X"；
4. **调 repo**：校验通过后调用 Repository 接口；ID 由 `fmt.Sprintf("xxx-%d", time.Now().UnixNano())` 生成；
5. 记录 `slog.Info` 业务日志。

**典型业务逻辑样例**：
- `DemandService` 完整生命周期：Create（校验角色/必填，兼容小程序"元→分"换算）→ 管理端 Review（approve/reject/supplement，驳回理由写入 `BizFields["reject_reason"]`）→ Submit（仅 rejected 可重提）→ Complete/Cancel（仅发布者）→ CloseByAdmin / Delete（仅已取消/已驳回可删，防误删在审数据）→ SetOfflineAmount（登记线下成交额，撮合价值度量）；
- `WorkOrderService` 接单派单闭环（FR-6.2~6.5）：AcceptIntent 确认意向（本意向→contacted，其余 pending 意向→closed 并生成 pending 订单）→ StartWork → CompleteWork（可传成果照片）→ AcceptCompletion / RequestRework / RequestCancel，全程校验角色与当前状态，`ListMine` 合并发布者+接单者双视角去重；
- `EscrowService`：Deposit/Freeze/Release/Refund 校验金额为正，流水与余额调整同事务原子提交；
- `ComplianceService`：英文分类→中文规范值归一（`complianceDocCategoryAliases`/`complianceStandardCategoryAliases`），兼容历史英文传值。

**全部 Service 清单（50 个）**：
- 核心：DemandService、EnterpriseService、EmploymentService、ContractService、ContractTemplateService、JobService、WorkOrderService、IntentService、MatchingService（含 Bid）、ListingService、LabourService、TradingService、EscrowService、CommunityService、HomeService、FileService、MessageService、NewsService、ReviewService、VenueService、ApplicationService
- 培训/证书：TrainingService、EnrollmentService、CertificateService（phase3）、ExpiryService、ServiceListingService、InsuranceService、FinanceService
- 新增业务：ExpertService、CaseService、ComplianceService、ReportService（行业报告）、PortfolioService、AchievementService、RDChallengeService、ResearchProjectService、ProjectAppService、CompetitionService、EventService、ResourceService、EmergencyService
- Batch：ResourcePoolService、TestSiteService、ExhibitionService、TransformationService、CollegeService、CooperationService、StudyTourService、RescueCaseService、EmergencyDeptService、AssociationMemberService

## 五、Domain 补充模型（`internal/domain/`）

**models_new.go**：Expert（智库专家）、Application（服务申报，原 JSON 文件后改 service_applications 表）、CaseEntry（项目案例）、ComplianceDoc + StandardDoc（合规文档/团体标准，Category 中文：国家标准/行业标准/团体标准/企业标准）、ProjectApplication（项目申报）、Achievement + Attachment（科技成果）、RDChallenge（研发难题）、ResearchProject（课题攻关）、Competition + CompetitionEvent/Requirement/Prize + CompetitionReg（赛事，字段与小程序 pages/competitions 对齐；C13 补姓名/手机/身份证/证件照列）、AssociationEvent + EventRegistration（协会活动）、MemberPortfolio（品牌展示）、IndustryReport（行业报告）、IndustryResource（产业资源，含四级可见性 public<member<partner<admin）、IndustryResourceBooking（C11 预约）、EmergencyResource + EmergencyDispatch（应急资源/调度，EndTime 可空）

**models_batch1.go**：ResourcePool + ResourcePoolMember（资源池）、TestSite + TestSiteBooking（测试环境预约，含预约联系人/审核）、Exhibition + ExhibitionBooth（展会/展位）、Shop、StudyTour + StudySchedule（研学行程）、ParseTime 工具

**models_batch2.go**：TransformationStage 常量（lab→pilot→industrialized→listed）+ Transformation + TransMilestone（成果转化追踪）、CollegeMajor/CollegePartner/College（院校展示大字段，与小程序 list/detail 页对齐，CoopType 分 research/talent/both）、CooperationProgram（校企共建：定向培养/实习基地/联合实验室/课程）

**models_batch3.go**：RescueCase（救援案例）、EmergencyDept + EmergencyDrill（部门对接+联合演练）、AssociationRole 8 级常量（president 会长 / vice_president 副会长 / secretary 秘书长 / dept_head 部门负责人 / member 普通会员 / partner 副会长单位 / college 合作院校 / guest 访客）+ AssociationMember（含加入/到期日期、active/expired/suspended）

**biz_standard.go（业务术语标准，2026-08-04 功能方案修订版 v2 评审落地）**：定义统一 `BizDomain` 枚举 14 个——inspection 巡检 / survey 测绘 / spray 植保 / logistics 物流 / lifting 吊运 / aerial 航拍 / purchase 采购 / maintain 维修 / calibrate 检测标定 / test_fly 试飞测试 / airspace 空域协调 / training 培训 / trade 买卖租赁 / other 其他，配套中文标签 `BizDomainLabel`；核心约定：把历史上三套互不兼容的枚举（需求 BizType、供给 ProductType、企业注册自由文本分类）统一映射到业务域（`BizTypeDomain`/`ProductTypeDomain`/`DomainProductTypes`），线上枚举值保持不变防破坏既有数据与 API 契约，**新增任何类型必须先在本文件登记补映射、禁止各模块另立标签表**，前端标签统一取自本表；另立第二标准轴 `EnterpriseCategory` 8 类——drone 整机 / part 零部件 / flight_ctrl 飞控 / payload 载荷 / operator 运营服务 / college 实训院校 / airport 通航机场 / inspector 检测机构，`NormalizeCategory` 将历史自由文本归一为合法分类（未知→"other"）；注释标注"清洗粉刷暂归其他，未来如需独立域在本文件登记"。

## 六、已知问题/遗留汇总

- jsonbSlice 防 pgx nil slice→NULL 违反 23502；
- Escrow 旧读改写并发丢更新已移除、改用原子方法+哨兵错误；
- userRepo.Delete 需事务级联删外键行；
- scan*Paged where 必须编译期常量；
- 旧库无版本表首次全量幂等重放；
- batch2/3 五个模块（资源池/校企/救援/应急部门/协会成员）曾 PG 模式回退内存、现已补 PG；
- 内存实现部分过滤参数被忽略（测试注释）；
- `auth_sms.go` 留有 TODO(短信服务商)：接入腾讯云/阿里云 SMS 前 dev_code 回显。
