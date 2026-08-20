package postgres_test

// 回填迁移 SQL 语法/内容回归测试（000073_backfill_enrollment_paid_amount）。
// 背景：000070 只加列不回填，升级前付费报名 paid_amount_fen=0 → completeEnrollment
// 跳过释放，学费滞留 escrow frozen。000073 按 escrow freeze 流水回填固化金额。
// 本测试在无 PG 环境下做离线校验（文件存在、版本顺序、关键子句、语句括号/引号配平）；
// 有 PG 时 migration_test.go 的版本化测试会真实执行该迁移，做真正语法验证。

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"drone-platform/internal/repository/postgres"
)

func TestBackfillEnrollmentPaidAmountMigration(t *testing.T) {
	dir := postgres.MigrationsDir()
	upPath := filepath.Join(dir, "000073_backfill_enrollment_paid_amount.up.sql")
	downPath := filepath.Join(dir, "000073_backfill_enrollment_paid_amount.down.sql")

	up, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read %s: %v", upPath, err)
	}
	down, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read %s: %v", downPath, err)
	}
	upSQL, downSQL := string(up), string(down)

	// 1) 版本顺序：000073 必须晚于迁移目录中所有既有版本（生产已应用 000070，只能新增迁移）。
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	var versions []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		ver := name
		for i, c := range name {
			if c < '0' || c > '9' {
				ver = name[:i]
				break
			}
		}
		if ver != "" {
			versions = append(versions, ver)
		}
	}
	sort.Strings(versions)
	if len(versions) > 0 && versions[len(versions)-1] != "000073" {
		t.Fatalf("000073 must be the newest migration, got max=%s (all: %v)", versions[len(versions)-1], versions)
	}

	// 2) 关键子句：回填语句必须命中 escrow 冻结流水并按用户+课程取最近一条。
	required := []string{
		"UPDATE training_enrollments",
		"paid_amount_fen",
		"escrow_transactions",
		"tx_type='freeze'",
		"reference_type='training_course'",
		"t.from_user = e.user_id",
		"t.reference_id = e.course_id",
		"t.status='completed'",
		"ORDER BY t.created_at DESC LIMIT 1",
		"WHERE e.paid_amount_fen = 0",
	}
	for _, clause := range required {
		if !strings.Contains(upSQL, clause) {
			t.Errorf("up migration missing clause %q", clause)
		}
	}

	// 3) 幂等保守性：只填 0 值记录、COALESCE 兜底。
	if !strings.Contains(upSQL, "COALESCE") {
		t.Errorf("up migration must COALESCE to 0 when no freeze tx found")
	}

	// 4) 语句级配平检查（无 PG 环境下的最轻量"语法"校验）：
	//    去注释后按分号切分，每条语句非空且括号、单引号配平。
	checkStatementBalance(t, upSQL, "up")
	checkStatementBalance(t, downSQL, "down")

	// 5) down 为空操作（无破坏性回退：不清零 paid_amount_fen）。
	if got := strings.TrimSpace(downSQL); got != "" {
		// 允许纯注释文件；不允许真实 SQL 语句
		for _, line := range strings.Split(downSQL, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}
			t.Errorf("down migration must be a no-op (comments only), found statement: %q", line)
		}
	}
}

// checkStatementBalance 去注释后按分号切分语句，校验每条语句非空且括号/单引号配平。
func checkStatementBalance(t *testing.T, sql, label string) {
	t.Helper()
	var buf strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue // 注释行
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	cleaned := buf.String()
	stmt := &strings.Builder{}
	for _, part := range strings.Split(cleaned, ";") {
		if strings.TrimSpace(part) == "" {
			continue
		}
		stmt.Reset()
		stmt.WriteString(part)
		s := stmt.String()
		if !balancedParens(s) {
			t.Errorf("%s migration: unbalanced parentheses in statement: %q", label, truncateSQL(s))
		}
		if !balancedQuotes(s) {
			t.Errorf("%s migration: unbalanced single quotes in statement: %q", label, truncateSQL(s))
		}
	}
}

func balancedParens(s string) bool {
	depth := 0
	inStr := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\'' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

func balancedQuotes(s string) bool {
	return strings.Count(s, "'")%2 == 0
}

func truncateSQL(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}
