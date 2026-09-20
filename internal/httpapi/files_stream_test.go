package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// pngSig PNG 文件头（用数字字面量写，避免编辑器/工具对转义的处理差异）：
// sniffAllowedType 依魔数判定类型，测试内容必须以它开头。
var pngSig = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

func streamUploadServer(t *testing.T) (*Server, string) {
	t.Helper()
	tokens, err := NewTokenManager("01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	fs := service.NewFileService(dir)
	srv := NewServer(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		fs, nil, nil, nil, nil, nil, nil, nil, nil,
		memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens,
	)
	return srv, dir
}

// streamUploadBody 构造 multipart 请求体。fieldFirst=false 时把 private 字段排在
// **文件段之后** —— 真实客户端（uni.uploadFile）的字段顺序并不保证，而顺序错了就可能
// 把营业执照写进公开目录，所以两条路径都要有测试。
func streamUploadBody(t *testing.T, content []byte, private bool, fieldFirst bool) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	field := func() {
		if err := mw.WriteField("private", fmt.Sprintf("%v", private)); err != nil {
			t.Fatal(err)
		}
	}
	if fieldFirst {
		field()
	}
	fw, err := mw.CreateFormFile("file", "license.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatal(err)
	}
	if !fieldFirst {
		field()
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, mw.FormDataContentType()
}

func doStreamUpload(t *testing.T, srv *Server, body *bytes.Buffer, ct string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	r.Header.Set("Content-Type", ct)
	r = r.WithContext(contextWithActor(r, domain.Actor{ID: "u-stream", Role: domain.RoleEnterprise}))
	w := httptest.NewRecorder()
	srv.uploadFile(w, r)
	return w
}

func uploadURL(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var out struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse upload response: %v (%s)", err, w.Body.String())
	}
	return out.Data.URL
}

// TestUploadPrivateBothFieldOrders 私密性判定与字段顺序无关。
func TestUploadPrivateBothFieldOrders(t *testing.T) {
	content := append(append([]byte{}, pngSig...), bytes.Repeat([]byte("x"), 1024)...)
	for _, fieldFirst := range []bool{true, false} {
		for _, private := range []bool{true, false} {
			srv, _ := streamUploadServer(t)
			body, ct := streamUploadBody(t, content, private, fieldFirst)
			w := doStreamUpload(t, srv, body, ct)
			if w.Code != http.StatusCreated {
				t.Fatalf("fieldFirst=%v private=%v: want 201, got %d (%s)", fieldFirst, private, w.Code, w.Body.String())
			}
			url := uploadURL(t, w)
			isPrivate := strings.HasPrefix(url, "/uploads/private/")
			if isPrivate != private {
				t.Fatalf("fieldFirst=%v private=%v: url=%q —— 营业执照落错目录等于把敏感件写进公开目录", fieldFirst, private, url)
			}
		}
	}
}

// TestUploadMissingFile400 没有 file 段仍是 400（错误文案不变）。
func TestUploadMissingFile400(t *testing.T) {
	srv, _ := streamUploadServer(t)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("private", "true"); err != nil {
		t.Fatal(err)
	}
	mw.Close()
	w := doStreamUpload(t, srv, &buf, mw.FormDataContentType())
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing file part: want 400, got %d (%s)", w.Code, w.Body.String())
	}
}

// TestUploadOversize400AndNoResidue 超过单文件上限仍是 400，且磁盘不留残file。
// 改流式前由 ParseMultipartForm 提前挡住；改后 MaxBytesReader 要到读满 40MiB 才报错，
// 若不翻译就会变成笼统 500 —— 这条测试守住那个翻译。
func TestUploadOversize400AndNoResidue(t *testing.T) {
	srv, dir := streamUploadServer(t)
	content := make([]byte, maxUploadBytes+1024)
	copy(content, pngSig)
	body, ct := streamUploadBody(t, content, false, true)
	w := doStreamUpload(t, srv, body, ct)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("oversize upload: want 400, got %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "file too large") {
		t.Fatalf("oversize upload should say file too large, got %s", w.Body.String())
	}
	if n := countFiles(t, dir); n != 0 {
		t.Fatalf("oversize upload left %d file(s) on disk, want 0", n)
	}
}

func countFiles(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			n++
		}
		return nil
	})
	return n
}

// TestUploadConcurrentMemoryBounded 200 并发上传的内存必须与文件大小无关。
//
// 这是「200 家企业同时入驻能扛住吗」那个问题的回归测试：营业执照必传
//（pkg-eco/pages/enterprise/register.vue:208），而此前用 ParseMultipartForm
// 把整个文件段读进内存 —— 实测 200 并发 1MB 峰值堆 310MB、2MB 711MB，
// 容器上限只有 512MB。改成流式后 200 并发共 3MB。
// 阈值 120MB 取得很宽：流式实测 3MB、缓存式实测 310MB，差两个数量级，
// 不会因为 -race 的额外开销误报。
func TestUploadConcurrentMemoryBounded(t *testing.T) {
	if testing.Short() {
		t.Skip("并发内存测量在 -short 下跳过")
	}
	const (
		conc = 200
		size = 1 << 20
	)
	srv, _ := streamUploadServer(t)

	// 请求体写到临时文件后 200 个 goroutine 各开各的句柄：这样测试自身不占堆，
	// 量到的增量全部来自 handler。
	content := make([]byte, size)
	copy(content, pngSig)
	body, ct := streamUploadBody(t, content, true, true)
	bodyPath := filepath.Join(t.TempDir(), "body.bin")
	if err := os.WriteFile(bodyPath, body.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	body = nil
	runtime.GC()

	var base runtime.MemStats
	runtime.ReadMemStats(&base)
	var peak uint64
	stop := make(chan struct{})
	var sampler sync.WaitGroup
	sampler.Add(1)
	go func() {
		defer sampler.Done()
		var m runtime.MemStats
		for {
			select {
			case <-stop:
				return
			default:
				runtime.ReadMemStats(&m)
				if m.HeapInuse > atomic.LoadUint64(&peak) {
					atomic.StoreUint64(&peak, m.HeapInuse)
				}
				time.Sleep(time.Millisecond)
			}
		}
	}()

	var ok int32
	var wg sync.WaitGroup
	for i := 0; i < conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f, err := os.Open(bodyPath)
			if err != nil {
				return
			}
			defer f.Close()
			r := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", f)
			r.Header.Set("Content-Type", ct)
			r = r.WithContext(contextWithActor(r, domain.Actor{ID: "u-mem", Role: domain.RoleEnterprise}))
			w := httptest.NewRecorder()
			srv.uploadFile(w, r)
			if w.Code == http.StatusCreated {
				atomic.AddInt32(&ok, 1)
			}
		}()
	}
	wg.Wait()
	close(stop)
	sampler.Wait()

	if ok != conc {
		t.Fatalf("并发上传成功 %d/%d，先查功能再查内存", ok, conc)
	}
	delta := int64(atomic.LoadUint64(&peak)) - int64(base.HeapInuse)
	t.Logf("200 并发 1MB 上传：峰值堆增量 %.1f MB（每请求 %.2f MB）",
		float64(delta)/(1<<20), float64(delta)/(1<<20)/conc)
	if delta > 120<<20 {
		t.Fatalf("200 并发上传峰值堆 %.0f MB 超过 120MB —— 文件段又被读进内存了（容器上限 512MB）",
			float64(delta)/(1<<20))
	}
}
