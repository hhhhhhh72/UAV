package httpapi

import "net/http"

// P2-1 补齐「七模块管理端列表」：
// 保险/维修/贷款/简历此前只有用户侧「我的」查询，管理端拿不到全量数据，
// 新增 ListAll 分页接口；场地/讲师/校企/应急部门/救援案例复用既有公开
// 列表 handler 挂到 /api/v1/admin/* 下（adminGate 中间件负责鉴权）。

// GET /api/v1/admin/policies
func (s *Server) listAdminPolicies(w http.ResponseWriter, r *http.Request) {
	page, ps := paginationFromQuery(r)
	list, total, err := s.insuranceSvc.ListAllPolicies((page-1)*ps, ps)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, list, total)
}

// GET /api/v1/admin/inspections
func (s *Server) listAdminInspections(w http.ResponseWriter, r *http.Request) {
	list, err := s.insuranceSvc.ListAllInspections()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, list)
}

// GET /api/v1/admin/repairs
func (s *Server) listAdminRepairs(w http.ResponseWriter, r *http.Request) {
	page, ps := paginationFromQuery(r)
	list, total, err := s.tradingSvc.ListAllRepairs((page-1)*ps, ps)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, list, total)
}

// GET /api/v1/admin/loans
func (s *Server) listAdminLoans(w http.ResponseWriter, r *http.Request) {
	page, ps := paginationFromQuery(r)
	list, total, err := s.financeSvc.ListAllLoans((page-1)*ps, ps)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, list, total)
}

// GET /api/v1/admin/resumes
func (s *Server) listAdminResumes(w http.ResponseWriter, r *http.Request) {
	page, ps := paginationFromQuery(r)
	list, total, err := s.jobSvc.ListAllResumes((page-1)*ps, ps)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, list, total)
}
