package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// GET /api/v1/applications（求职者视角）的响应契约：
//  1. 顶层保留 job_id/status 等原字段——招聘大厅与职位详情按 job_id 判断"是否已投递"；
//  2. 新增 job 快照（标题/地点/类型/薪资/状态），前端不再显示裸 ID；
//  3. 无投递时 data 必须是 []（不是 null）。
func TestMyApplicationsIncludeJobSnapshot(t *testing.T) {
	app := newServer(t)

	jw := requestAs(t, app, http.MethodPost, "/api/v1/jobs",
		[]byte(`{"title":"无人机运维工程师","description":"机队维护","location":"重庆·南岸","job_type":"实习","salary_fen":1000000}`),
		"enterprise-1", domain.RoleEnterprise)
	if jw.Code != http.StatusCreated {
		t.Fatalf("create job: %d %s", jw.Code, jw.Body.String())
	}
	jobID := dataID(t, jw)
	if pw := requestAs(t, app, http.MethodPost, "/api/v1/jobs/"+jobID+"/publish", nil, "enterprise-1", domain.RoleEnterprise); pw.Code != http.StatusOK {
		t.Fatalf("publish job: %d %s", pw.Code, pw.Body.String())
	}

	rw := requestAs(t, app, http.MethodPost, "/api/v1/resumes",
		[]byte(`{"title":"我的简历","name":"张三","phone":"13800000000","visibility":"public"}`),
		"user-1", domain.RoleIndividual)
	if rw.Code != http.StatusCreated {
		t.Fatalf("create resume: %d %s", rw.Code, rw.Body.String())
	}
	resumeID := dataID(t, rw)

	aw := requestAs(t, app, http.MethodPost, "/api/v1/applications",
		[]byte(`{"job_id":"`+jobID+`","resume_id":"`+resumeID+`"}`), "user-1", domain.RoleIndividual)
	if aw.Code != http.StatusCreated {
		t.Fatalf("apply: %d %s", aw.Code, aw.Body.String())
	}

	w := requestAs(t, app, http.MethodGet, "/api/v1/applications", nil, "user-1", domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("list applications: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data []struct {
			ID             string `json:"id"`
			JobID          string `json:"job_id"`
			Status         string `json:"status"`
			EnterpriseName string `json:"enterprise_name"`
			Job            *struct {
				ID        string `json:"id"`
				Title     string `json:"title"`
				Location  string `json:"location"`
				JobType   string `json:"job_type"`
				SalaryFen int64  `json:"salary_fen"`
				Status    string `json:"status"`
			} `json:"job"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v (body=%.200s)", err, w.Body.String())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expect 1 application, got %d (body=%.300s)", len(resp.Data), w.Body.String())
	}
	got := resp.Data[0]
	if got.ID == "" || got.JobID != jobID || got.Status != string(domain.AppSubmitted) {
		t.Fatalf("top-level fields changed: %+v", got)
	}
	if got.Job == nil {
		t.Fatalf("job snapshot missing: %.300s", w.Body.String())
	}
	if got.Job.ID != jobID || got.Job.Title != "无人机运维工程师" || got.Job.Location != "重庆·南岸" ||
		got.Job.JobType != "实习" || got.Job.SalaryFen != 1000000 || got.Job.Status != string(domain.JobPublished) {
		t.Fatalf("job snapshot mismatch: %+v", *got.Job)
	}
	// 本装配未注入企业仓库 → 单位名留空，而不是编一个出来
	if got.EnterpriseName != "" {
		t.Fatalf("enterprise_name should be empty without enterprise repo, got %q", got.EnterpriseName)
	}

	// 无投递者：data 必须是 []，不能是 null（避免前端到处兜底）
	ew := requestAs(t, app, http.MethodGet, "/api/v1/applications", nil, "user-2", domain.RoleIndividual)
	if ew.Code != http.StatusOK {
		t.Fatalf("empty list: %d %s", ew.Code, ew.Body.String())
	}
	if !strings.Contains(ew.Body.String(), `"data":[]`) {
		t.Fatalf("empty applications should serialize as [], got %.200s", ew.Body.String())
	}
}
