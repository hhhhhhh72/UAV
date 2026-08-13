package postgres

import "testing"

// TestVersionOf: 迁移文件名 → 版本号解析，schema_migrations 的版本键来源。
// 回归：非数字前缀文件不参与迁移调度，避免脏文件被误执行。
func TestVersionOf(t *testing.T) {
	cases := []struct{ name, want string }{
		{"000001_init.up.sql", "000001"},
		{"000031_enrollment_birthday_date.up.sql", "000031"},
		{"000053_drop_shops.up.sql", "000053"},
		{"000009_escrow.up.sql", "000009"},
		{"000001_init.down.sql", "000001"},
		{"README.md", ""},
		{"up.sql", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := versionOf(c.name); got != c.want {
			t.Errorf("versionOf(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}
