// Package main — Swagger 接口注解
//
// 本文件包含所有 API 端点的 Swagger 注解。
// 运行 `swag init -g cmd/api/main.go -d ./cmd/api --parseInternal` 重新生成 docs/swagger.json。
package main

// =====================================================================
// 系统 & 认证
// =====================================================================

// @Summary      健康检查
// @Tags         系统
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /healthz [get]
func _swagger_system_healthz() {}

// @Summary      API 首页信息
// @Tags         系统
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       / [get]
func _swagger_system_index() {}

// @Summary      小程序微信登录
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  object{code=string}  true  "微信 code"
// @Success      200   {object}  map[string]any
// @Router       /api/v1/auth/wechat/login [post]
func _swagger_auth_wechat_login() {}

// @Summary      刷新 Token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body  body  object{refreshToken=string}  true  "刷新令牌"
// @Success      200   {object}  map[string]any
// @Router       /api/auth/refresh [post]
func _swagger_auth_refresh() {}

// @Summary      登出
// @Tags         认证
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/auth/logout [post]
func _swagger_auth_logout() {}

// @Summary      获取当前用户信息
// @Tags         用户
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/me [get]
func _swagger_user_me_get() {}

// @Summary      更新当前用户信息
// @Tags         用户
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/me [patch]
func _swagger_user_me_patch() {}

// =====================================================================
// 首页 & 搜索
// =====================================================================

// @Summary      首页数据
// @Tags         首页
// @Produce      json
// @Param        city  query  string  false  "城市"
// @Param        lat   query  number  false  "纬度"
// @Param        lng   query  number  false  "经度"
// @Success      200   {object}  map[string]any
// @Router       /api/v1/home [get]
func _swagger_home() {}

// @Summary      全局搜索
// @Tags         搜索
// @Produce      json
// @Param        q  query  string  true  "搜索关键词"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/search [get]
func _swagger_search() {}

// =====================================================================
// 需求大厅
// =====================================================================

// @Summary      需求列表
// @Tags         需求
// @Produce      json
// @Param        biz_type   query  string  false  "业务类型"
// @Param        district   query  string  false  "地区"
// @Param        sort       query  string  false  "排序"
// @Param        page       query  int     false  "页码"
// @Param        page_size  query  int     false  "每页条数"
// @Success      200        {object}  map[string]any
// @Router       /api/v1/demands [get]
func _swagger_demand_list() {}

// @Summary      需求详情
// @Tags         需求
// @Produce      json
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/{id} [get]
func _swagger_demand_detail() {}

// @Summary      发布需求
// @Tags         需求
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/demands [post]
func _swagger_demand_create() {}

// @Summary      更新需求
// @Tags         需求
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/{id} [patch]
func _swagger_demand_update() {}

// @Summary      提交需求
// @Tags         需求
// @Security     BearerAuth
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/{id}/submit [post]
func _swagger_demand_submit() {}

// @Summary      竞标报价
// @Tags         需求
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/{id}/applications [post]
func _swagger_demand_bid() {}

// @Summary      选择中标
// @Tags         需求
// @Security     BearerAuth
// @Param        id             path  string  true  "需求ID"
// @Param        applicationId  path  string  true  "投标ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/{id}/applications/{applicationId}/select [post]
func _swagger_demand_select() {}

// @Summary      完成需求
// @Tags         需求
// @Security     BearerAuth
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/{id}/complete [post]
func _swagger_demand_complete() {}

// @Summary      发起争议
// @Tags         需求
// @Security     BearerAuth
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/{id}/dispute [post]
func _swagger_demand_dispute() {}

// @Summary      我的需求
// @Tags         需求
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/demands/bids/mine [get]
func _swagger_demand_mine() {}

// =====================================================================
// 企业
// =====================================================================

// @Summary      企业列表(公开)
// @Tags         企业
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/enterprises [get]
func _swagger_enterprise_list() {}

// @Summary      企业注册
// @Tags         企业
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/enterprises [post]
func _swagger_enterprise_create() {}

// @Summary      更新企业信息
// @Tags         企业
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "企业ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/enterprises/{id} [patch]
func _swagger_enterprise_update() {}

// @Summary      提交企业审核
// @Tags         企业
// @Security     BearerAuth
// @Param        id  path  string  true  "企业ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/enterprises/{id}/submit [post]
func _swagger_enterprise_submit() {}

// @Summary      上传企业文档
// @Tags         企业
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Param        id    path      string  true  "企业ID"
// @Param        file  formData  file    true  "文件"
// @Success      200   {object}  map[string]any
// @Router       /api/v1/enterprises/{id}/documents [post]
func _swagger_enterprise_docs() {}

// =====================================================================
// 管理后台
// =====================================================================

// @Summary      管理员企业列表
// @Tags         管理后台-企业
// @Security     BearerAuth
// @Produce      json
// @Param        status     query  string  false  "状态"
// @Param        keyword    query  string  false  "搜索"
// @Param        page       query  int     false  "页码"
// @Param        page_size  query  int     false  "每页条数"
// @Success      200        {object}  map[string]any
// @Router       /api/v1/admin/enterprises [get]
func _swagger_admin_enterprise_list() {}

// @Summary      审核企业
// @Tags         管理后台-企业
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "企业ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/enterprises/{id}/review [post]
func _swagger_admin_enterprise_review() {}

// @Summary      批量审核企业
// @Tags         管理后台-企业
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/enterprises/batch-review [post]
func _swagger_admin_enterprise_batch() {}

// @Summary      管理员需求列表
// @Tags         管理后台-需求
// @Security     BearerAuth
// @Produce      json
// @Param        status     query  string  false  "状态"
// @Param        page       query  int     false  "页码"
// @Param        page_size  query  int     false  "每页条数"
// @Success      200        {object}  map[string]any
// @Router       /api/v1/admin/demands [get]
func _swagger_admin_demand_list() {}

// @Summary      审核需求-通过
// @Tags         管理后台-需求
// @Security     BearerAuth
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/demands/{id}/approve [post]
func _swagger_admin_demand_approve() {}

// @Summary      审核需求-驳回
// @Tags         管理后台-��求
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "需求ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/demands/{id}/review [post]
func _swagger_admin_demand_reject() {}

// @Summary      数据看板
// @Tags         管理后台
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/dashboard [get]
func _swagger_admin_dashboard() {}

// @Summary      导出需求 CSV
// @Tags         管理后台
// @Security     BearerAuth
// @Produce      text/csv
// @Success      200  {file}  text/csv
// @Router       /api/v1/admin/export/demands [get]
func _swagger_admin_export_demand() {}

// @Summary      导出企业 CSV
// @Tags         管理后台
// @Security     BearerAuth
// @Produce      text/csv
// @Success      200  {file}  text/csv
// @Router       /api/v1/admin/export/enterprises [get]
func _swagger_admin_export_ent() {}

// @Summary      用户列表(管理)
// @Tags         管理后台
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/users [get]
func _swagger_admin_users() {}

// @Summary      修改用户角色
// @Tags         管理后台
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "用户ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/users/{id}/role [post]
func _swagger_admin_user_role() {}

// =====================================================================
// 培训与认证
// =====================================================================

// @Summary      培训课程列表
// @Tags         培训
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/training-courses [get]
func _swagger_training_list() {}

// @Summary      创建培训课程(管理)
// @Tags         培训
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/training-courses [post]
func _swagger_training_create() {}

// @Summary      课程报名+支付
// @Tags         培训
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "课程ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/training-courses/{id}/pay-and-enroll [post]
func _swagger_training_enroll() {}

// @Summary      课程报名列表
// @Tags         培训
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "课程ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/training-courses/{id}/enrollments [get]
func _swagger_training_enrollments() {}

// @Summary      我的报名
// @Tags         培训
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/enrollments/mine [get]
func _swagger_training_my_enrollments() {}

// @Summary      完成报名
// @Tags         培训
// @Security     BearerAuth
// @Param        id  path  string  true  "报名ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/enrollments/{id}/complete [post]
func _swagger_training_enrollment_complete() {}

// @Summary      我的证书
// @Tags         培训
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/certificates/mine [get]
func _swagger_training_my_certs() {}

// @Summary      创建证书(管理)
// @Tags         培训
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/certificates [post]
func _swagger_training_cert_create() {}

// @Summary      审核证书
// @Tags         培训
// @Security     BearerAuth
// @Param        id  path  string  true  "证书ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/certificates/{id}/approve [post]
func _swagger_training_cert_approve() {}

// @Summary      到期证书提醒
// @Tags         培训
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/certificates/expiring [get]
func _swagger_training_cert_expiring() {}

// =====================================================================
// 招聘
// =====================================================================

// @Summary      职位列表
// @Tags         招聘
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/jobs [get]
func _swagger_jobs_list() {}

// @Summary      发布职位
// @Tags         招聘
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/jobs [post]
func _swagger_jobs_create() {}

// @Summary      我的职位
// @Tags         招聘
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/jobs/mine [get]
func _swagger_jobs_mine() {}

// @Summary      发布职位(管理)
// @Tags         招聘
// @Security     BearerAuth
// @Param        id  path  string  true  "职位ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/jobs/{id}/publish [post]
func _swagger_jobs_publish() {}

// @Summary      关闭职位
// @Tags         招聘
// @Security     BearerAuth
// @Param        id  path  string  true  "职位ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/jobs/{id}/close [post]
func _swagger_jobs_close() {}

// @Summary      简历管理
// @Tags         招聘
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/resumes [post]
func _swagger_resume_create() {}

// @Summary      我的简历
// @Tags         招聘
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/resumes/mine [get]
func _swagger_resume_mine() {}

// @Summary      更新简历
// @Tags         招聘
// @Security     BearerAuth
// @Param        id  path  string  true  "简历ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/resumes/{id} [patch]
func _swagger_resume_update() {}

// @Summary      投递申请
// @Tags         招聘
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/applications [post]
func _swagger_apply_create() {}

// @Summary      更新申请状态
// @Tags         招聘
// @Security     BearerAuth
// @Param        id  path  string  true  "申请ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/applications/{id}/status [patch]
func _swagger_apply_status() {}

// =====================================================================
// 社区/二手/用工
// =====================================================================

// @Summary      帖子列表
// @Tags         社区
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/posts [get]
func _swagger_posts_list() {}

// @Summary      发帖
// @Tags         社区
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/posts [post]
func _swagger_posts_create() {}

// @Summary      评论
// @Tags         社区
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "帖子ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/posts/{id}/comments [post]
func _swagger_posts_comment() {}

// @Summary      举报管理
// @Tags         管理后台
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/reports [get]
func _swagger_admin_reports() {}

// @Summary      二手商品列表
// @Tags         交易
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/listings [get]
func _swagger_listings() {}

// @Summary      用工需求列表
// @Tags         用工
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/labour-orders [get]
func _swagger_labour_list() {}

// @Summary      发布用工需求
// @Tags         用工
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/labour-orders [post]
func _swagger_labour_create() {}

// @Summary      用工报价
// @Tags         用工
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "订单ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/labour-orders/{id}/quote [post]
func _swagger_labour_quote() {}

// =====================================================================
// 合同
// =====================================================================

// @Summary      合同列表
// @Tags         合同
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/contracts [get]
func _swagger_contracts_list() {}

// @Summary      创建合同
// @Tags         合同
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/contracts [post]
func _swagger_contracts_create() {}

// @Summary      合同作废
// @Tags         合同
// @Security     BearerAuth
// @Param        id  path  string  true  "合同ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/contracts/{id}/void [post]
func _swagger_contracts_void() {}

// @Summary      合同模板列表
// @Tags         合同
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/contract-templates [get]
func _swagger_contracts_templates() {}

// =====================================================================
// 交易/保险/金融
// =====================================================================

// @Summary      无人机产品列表
// @Tags         交易
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/products [get]
func _swagger_products() {}

// @Summary      维修订单
// @Tags         交易
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/repairs [post]
func _swagger_repairs() {}

// @Summary      保险保单
// @Tags         保险
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/policies [post]
func _swagger_policies() {}

// @Summary      我的保单
// @Tags         保险
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/policies/mine [get]
func _swagger_policies_mine() {}

// @Summary      年审管理
// @Tags         保险
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/inspections [post]
func _swagger_inspections() {}

// @Summary      贷款申请
// @Tags         金融
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/loans [post]
func _swagger_loans() {}

// @Summary      我的贷款
// @Tags         金融
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/loans/mine [get]
func _swagger_loans_mine() {}

// =====================================================================
// 资金托管
// =====================================================================

// @Summary      充值
// @Tags         资金
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/escrow/deposit [post]
func _swagger_escrow_deposit() {}

// @Summary      冻结资金
// @Tags         资金
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/escrow/freeze [post]
func _swagger_escrow_freeze() {}

// @Summary      释放资金
// @Tags         资金
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/escrow/release [post]
func _swagger_escrow_release() {}

// @Summary      退款
// @Tags         资金
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/escrow/refund [post]
func _swagger_escrow_refund() {}

// @Summary      余额查询
// @Tags         资金
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/escrow/balance [get]
func _swagger_escrow_balance() {}

// @Summary      交易记录
// @Tags         资金
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/escrow/transactions [get]
func _swagger_escrow_transactions() {}

// =====================================================================
// 消息
// =====================================================================

// @Summary      消息列表
// @Tags         消息
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/messages [get]
func _swagger_messages() {}

// @Summary      未读消息数
// @Tags         消息
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/messages/unread-count [get]
func _swagger_messages_unread() {}

// @Summary      标记已读
// @Tags         消息
// @Security     BearerAuth
// @Param        id  path  string  true  "消息ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/messages/{id}/read [post]
func _swagger_messages_read() {}

// =====================================================================
// 评价/场地/资讯/应急/产学研/活动
// =====================================================================

// @Summary      评价列表(公开)
// @Tags         评价
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/reviews [get]
func _swagger_reviews() {}

// @Summary      创建评价
// @Tags         评价
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/reviews [post]
func _swagger_reviews_create() {}

// @Summary      评价列表(管理)
// @Tags         管理后台-评价
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/reviews [get]
func _swagger_admin_reviews() {}

// @Summary      审核评价
// @Tags         管理后台-评价
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "评价ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/reviews/{id}/approve [post]
func _swagger_admin_review_approve() {}

// @Summary      驳回评价
// @Tags         管理后台-评价
// @Security     BearerAuth
// @Param        id  path  string  true  "评价ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/admin/reviews/{id}/reject [post]
func _swagger_admin_review_reject() {}

// @Summary      场地列表
// @Tags         场地
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/venues [get]
func _swagger_venues() {}

// @Summary      场地预约
// @Tags         场地
// @Security     BearerAuth
// @Accept       json
// @Param        id  path  string  true  "场地ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/venues/{id}/book [post]
func _swagger_venues_book() {}

// @Summary      资讯列表
// @Tags         资讯
// @Produce      json
// @Param        page       query  int  false  "页码"
// @Param        page_size  query  int  false  "每页条数"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/articles [get]
func _swagger_articles() {}

// @Summary      发布资讯(管理)
// @Tags         资讯
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  map[string]any
// @Router       /api/v1/articles [post]
func _swagger_articles_create() {}

// @Summary      应急资源列表
// @Tags         应急
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/emergency-resources [get]
func _swagger_emergency() {}

// @Summary      救援案例库
// @Tags         应急
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/rescue-cases [get]
func _swagger_rescue() {}

// @Summary      科技成果列表
// @Tags         产学研
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/achievements [get]
func _swagger_achievements() {}

// @Summary      科技成果详情
// @Tags         产学研
// @Produce      json
// @Param        id  path  string  true  "成果ID"
// @Success      200  {object}  map[string]any
// @Router       /api/v1/achievements/{id} [get]
func _swagger_achievement_detail() {}

// @Summary      活动列表
// @Tags         活动
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/events [get]
func _swagger_events() {}

// @Summary      赛事列表
// @Tags         活动
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /api/v1/competitions [get]
func _swagger_competitions() {}

// @Summary      文件上传
// @Tags         系统
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "文件"
// @Success      200   {object}  map[string]any
// @Router       /api/v1/files/upload [post]
func _swagger_files_upload() {}
