package httpapi

import "testing"

// 图片地址 → 上传台账 ID 的解析。
//
// 刻意的边界：只有 /uploads/ 路径下的 file-* 才算自家台账。
// 若放宽成「以 file- 开头」，一个外链 https://evil.example/file-abc 也会被拿去查库；
// /static/ 下的种子图则本就查不到，按「尺寸未知」退化为 1:1。
func TestUploadIDFromURL(t *testing.T) {
	cases := map[string]string{
		"/uploads/file-abc123":                        "file-abc123",
		"/uploads/file-abc123?v=2":                    "file-abc123",
		"/uploads/file-abc123#frag":                   "file-abc123",
		"/uploads/private/file-abc123":                "file-abc123",
		"https://api.cqnarc.cn/uploads/file-abc123":   "file-abc123",
		"/static/home/demand-lift.jpg":                "",
		"https://evil.example/file-abc123":            "",
		"/uploads/":                                   "",
		"/uploads/file-":                              "",
		"":                                            "",
	}
	for in, want := range cases {
		if got := uploadIDFromURL(in); got != want {
			t.Errorf("uploadIDFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}
