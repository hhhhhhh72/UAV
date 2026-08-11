package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/config"
)

// withPlatformSnapshot 记录当前平台配置，测试结束后写回（含文件），
// 避免 syncHomeConfigToPlatform 的落盘污染其他测试
func withPlatformSnapshot(t *testing.T) {
	t.Helper()
	snapshot := config.GetPlatformConfig()
	t.Cleanup(func() {
		if err := config.SavePlatformConfig(snapshot); err != nil {
			t.Logf("restore platform config: %v", err)
		}
	})
}

func TestBuildPlatformBanners(t *testing.T) {
	cases := []struct {
		name  string
		raw   []any
		want  int
		first string // 第一个 banner 的 ImageURL
	}{
		{
			name: "映射 image/link 到 image_url/link_url",
			raw: []any{
				map[string]any{"image": "/uploads/a.jpg", "link": "/pages/demand/list"},
				map[string]any{"image": "/uploads/b.jpg", "link": ""},
			},
			want:  2,
			first: "/uploads/a.jpg",
		},
		{
			name:  "空列表返回空",
			raw:   []any{},
			want:  0,
			first: "",
		},
		{
			name: "无图项跳过",
			raw: []any{
				map[string]any{"image": "", "link": "/x"},
				map[string]any{"image": "/uploads/c.jpg", "link": "/y"},
			},
			want:  1,
			first: "/uploads/c.jpg",
		},
		{
			name: "非对象元素忽略",
			raw: []any{
				"not-a-map",
				map[string]any{"image": "/uploads/d.jpg"},
			},
			want:  1,
			first: "/uploads/d.jpg",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildPlatformBanners(c.raw)
			if len(got) != c.want {
				t.Fatalf("len=%d want %d", len(got), c.want)
			}
			if c.want > 0 {
				if got[0].ImageURL != c.first {
					t.Fatalf("first image=%q want %q", got[0].ImageURL, c.first)
				}
				if got[0].Status != "active" || got[0].SortOrder != 1 {
					t.Fatalf("first banner meta: status=%q sort=%d", got[0].Status, got[0].SortOrder)
				}
			}
		})
	}
}

func TestSyncHomeConfigToPlatform(t *testing.T) {
	withPlatformSnapshot(t)

	cfg := map[string]any{
		"_home": map[string]any{
			"banners": []any{
				map[string]any{"image": "/uploads/ban1.jpg", "link": "/pages/demand/list"},
			},
			"notices": []any{"公告一", "  ", "公告二"},
		},
	}
	if err := syncHomeConfigToPlatform(cfg); err != nil {
		t.Fatalf("sync: %v", err)
	}

	got := config.GetPlatformConfig()
	if len(got.Banners) != 1 {
		t.Fatalf("banners=%d want 1", len(got.Banners))
	}
	if got.Banners[0].ImageURL != "/uploads/ban1.jpg" {
		t.Fatalf("image=%q", got.Banners[0].ImageURL)
	}
	if got.Banners[0].LinkURL != "/pages/demand/list" {
		t.Fatalf("link=%q", got.Banners[0].LinkURL)
	}
	// 空白公告被过滤，其余保留顺序
	if len(got.Notices) != 2 || got.Notices[0] != "公告一" || got.Notices[1] != "公告二" {
		t.Fatalf("notices=%v", got.Notices)
	}
}

func TestSyncHomeConfigNoHome(t *testing.T) {
	withPlatformSnapshot(t)
	// 无 _home 键：不报错、不改配置
	if err := syncHomeConfigToPlatform(map[string]any{"other": "x"}); err != nil {
		t.Fatalf("sync: %v", err)
	}
}

func TestResolveBannerImageURL(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/home", nil)
	r.Host = "uav.example.com"

	cases := []struct {
		name string
		url  string
		env  string
		hdr  string
		want string
	}{
		{name: "非上传路径原样返回", url: "https://img.cdn.com/a.jpg", want: "https://img.cdn.com/a.jpg"},
		{name: "无 BASE_URL 默认 http+Host", url: "/uploads/x.jpg", want: "http://uav.example.com/uploads/x.jpg"},
		{name: "nginx 反代走 X-Forwarded-Proto", url: "/uploads/x.jpg", hdr: "https", want: "https://uav.example.com/uploads/x.jpg"},
		{name: "BASE_URL 优先", url: "/uploads/x.jpg", env: "https://api.example.com", want: "https://api.example.com/uploads/x.jpg"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("BASE_URL", c.env)
			r.Header.Set("X-Forwarded-Proto", c.hdr)
			if got := resolveBannerImageURL(r, c.url); got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}
