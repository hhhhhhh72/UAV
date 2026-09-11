package httpapi

import (
	"testing"

	"drone-platform/internal/domain"
)

// 列表手机号回显：库里有手机号用库里的；老账号（旧表单按用户ID建号、无手机号密文）
// 按 ID 里的手机号回显；两者都没有才留空（前端显示"未绑定"）。
func TestListPhoneMasked(t *testing.T) {
	cases := []struct {
		name string
		user domain.User
		want string
	}{
		{"库内有手机号", domain.User{ID: "user-13800000001", PhoneCipher: "13800000001"}, "138****0001"},
		{"老账号按 ID 回显", domain.User{ID: "user-18623249541"}, "186****9541"},
		{"微信账号无手机号", domain.User{ID: "user-1786412375767536531", PhoneCipher: ""}, ""},
		{"非手机号形态的 ID", domain.User{ID: "admin"}, ""},
		{"ID 里的号段非法（12 开头）", domain.User{ID: "user-12823249541"}, ""},
		{"库内手机号优先于 ID", domain.User{ID: "user-18600000000", PhoneCipher: "13911112222"}, "139****2222"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := listPhoneMasked(tc.user); got != tc.want {
				t.Fatalf("listPhoneMasked(%q, cipher=%q) = %q, want %q", tc.user.ID, tc.user.PhoneCipher, got, tc.want)
			}
		})
	}
}
