package httpapi

import "testing"

// 图片地址 → uploads 台账主键。台账主键就是磁盘文件名，所以取路径最后一段。
//
// 关键的边界是「只认站点相对路径」：外链一律不查台账。
// 查询本身只用于取宽高、不涉及授权，但也不该让外部地址进到自家索引里。
func TestImageLedgerKey(t *testing.T) {
	cases := map[string]string{
		// 自家上传（文件名即 file-<32hex>）
		"/uploads/file-abc123":                      "file-abc123",
		"/uploads/file-abc123?v=2":                  "file-abc123",
		"/uploads/file-abc123#frag":                 "file-abc123",
		"/uploads/private/file-abc123":              "file-abc123",
		"https://api.cqnarc.cn/uploads/file-abc123": "file-abc123",
		// 种子图沿用自己的文件名，同样在台账里
		"/uploads/sl-hero.jpg":                      "sl-hero.jpg",
		"/static/home/demand-lift.jpg":              "demand-lift.jpg",
		// 外链与非法输入一律不查
		"https://evil.example/file-abc123":          "",
		"//evil.example/file-abc123":                "",
		"/uploads/":                                 "",
		"":                                          "",
	}
	for in, want := range cases {
		if got := imageLedgerKey(in); got != want {
			t.Errorf("imageLedgerKey(%q) = %q, want %q", in, got, want)
		}
	}
}
