package httpapi

import "regexp"

// userPhoneIDRe 匹配手机号注册用户 ID：形如 user-<11位手机号>。
// 这类 ID 原样出现在公开响应中会泄露用户手机号，可被批量抓取。
var userPhoneIDRe = regexp.MustCompile(`^user-\d{11}$`)

// maskUserID 公开响应层脱敏：匹配 user-<11位手机号> 的 ID 替换为 "user-***"，
// 其余 ID 原样返回。仅用于公开只读展示接口的响应构造处，不影响业务逻辑
// （订单归属等内部流程仍使用真实 ID）；登录用户自己可见的 mine 类接口与管理端
// 列表不脱敏（管理端需要真实 ID 做归属/审核）。
func maskUserID(id string) string {
	if userPhoneIDRe.MatchString(id) {
		return "user-***"
	}
	return id
}

// phoneMaskRe 手机号脱敏：保留前 3 后 4，中间以 **** 代替（138****8888）。
var phoneMaskRe = regexp.MustCompile(`^(\d{3})\d{4}(\d{4})$`)

// maskPhone 公开接口手机号脱敏：11 位手机号 → 138****8888；非手机号原样返回。
func maskPhone(p string) string {
	if m := phoneMaskRe.FindStringSubmatch(p); m != nil {
		return m[1] + "****" + m[2]
	}
	return p
}
