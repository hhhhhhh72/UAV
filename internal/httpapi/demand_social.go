package httpapi

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var _socialMu sync.RWMutex

// GET /api/v1/demands/stats — 批量获取点赞+浏览+评论
func (s *Server) demandStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{}
	entries, _ := os.ReadDir(".")
	for _, e := range entries {
		n := e.Name()
		if !strings.HasPrefix(n, "demand_") { continue }
		switch {
		case strings.HasSuffix(n, "_likes.json"):
			id := n[7 : len(n)-11]
			var c int
			readJSON(n, &_socialMu, &c)
			stats["like_"+id] = c
		case strings.HasSuffix(n, "_views.json"):
			id := n[7 : len(n)-11]
			var c int
			readJSON(n, &_socialMu, &c)
			stats["view_"+id] = c
		case strings.HasSuffix(n, "_comments.json"):
			id := n[7 : len(n)-14]
			var list []map[string]any
			readJSON(n, &_socialMu, &list)
			if list != nil { stats["cmt_"+id] = len(list) }
		}
	}
	respond(w, r, http.StatusOK, stats)
}

// POST /api/v1/demands/{id}/like
func (s *Server) likeDemand(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f := "demand_" + id + "_likes.json"
	ensureFile(f, []byte("0"))
	var count int
	readJSON(f, &_socialMu, &count)
	count++
	writeJSON(f, &_socialMu, count)
	respond(w, r, http.StatusOK, map[string]any{"likes": count})
}

// GET /api/v1/demands/{id}/comments
func (s *Server) getDemandComments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f := "demand_" + id + "_comments.json"
	ensureFile(f, []byte("[]"))
	type cmt struct {
		ID        string `json:"id"`
		Content   string `json:"content"`
		UserName  string `json:"userName"`
		CreatedAt string `json:"createdAt"`
	}
	var list []cmt
	readJSON(f, &_socialMu, &list)
	if list == nil { list = []cmt{} }
	respond(w, r, http.StatusOK, list)
}

// POST /api/v1/demands/{id}/comment
func (s *Server) commentDemand(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct{ Content string `json:"content"` }
	if err := decode(r, &body); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	f := "demand_" + id + "_comments.json"
	ensureFile(f, []byte("[]"))
	type cmt struct {
		ID        string `json:"id"`
		Content   string `json:"content"`
		UserName  string `json:"userName"`
		CreatedAt string `json:"createdAt"`
	}
	var list []cmt
	readJSON(f, &_socialMu, &list)
	c := cmt{ID: "c" + time.Now().Format("150405") + randomSuffix(2), Content: body.Content, UserName: "用户", CreatedAt: time.Now().Format(time.RFC3339)}
	list = append(list, c)
	writeJSON(f, &_socialMu, list)
	respond(w, r, http.StatusCreated, c)
}
