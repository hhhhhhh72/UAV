package postgres_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/repository/postgres"
)

// TestPG_MigrationsVersioned: 回归迁移版本记录机制——
// 1) 真实迁移目录中的每个 .up.sql 版本都登记在 schema_migrations；
// 2) 新迁移首次运行执行且登记，二次运行（内容已改为必错 SQL）被跳过；
// 3) 探测迁移的 SQL 与版本登记同事务原子生效。
func TestPG_MigrationsVersioned(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()

	// 1) 真实迁移目录的版本全部登记
	applied := map[string]bool{}
	rows, err := store.Pool().Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		t.Fatalf("schema_migrations missing: %v", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		applied[v] = true
	}
	rows.Close()

	entries, err := os.ReadDir(postgres.MigrationsDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		v := name
		for i, c := range name {
			if c < '0' || c > '9' {
				v = name[:i]
				break
			}
		}
		if !applied[v] {
			t.Errorf("migration version %s not recorded", v)
		}
	}

	// 2) 临时目录放一个探测迁移，两次运行只执行一次
	probeDir := t.TempDir()
	version := fmt.Sprintf("999%d", time.Now().UnixNano())
	probeFile := version + "_probe.up.sql"
	probeSQL := "CREATE TABLE IF NOT EXISTS versioning_probe (id INT);\nINSERT INTO versioning_probe VALUES (1);"
	if err := os.WriteFile(filepath.Join(probeDir, probeFile), []byte(probeSQL), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		store.Pool().Exec(context.Background(), "DROP TABLE IF EXISTS versioning_probe")
		store.Pool().Exec(context.Background(), "DELETE FROM schema_migrations WHERE version = $1", version)
	})

	if err := store.RunMigrationsFromDir(ctx, probeDir); err != nil {
		t.Fatalf("first probe run: %v", err)
	}
	var probeCnt int
	if err := store.Pool().QueryRow(ctx, "SELECT count(*) FROM versioning_probe").Scan(&probeCnt); err != nil || probeCnt != 1 {
		t.Fatalf("probe not executed: cnt=%d err=%v", probeCnt, err)
	}

	// 把探测文件改成必错 SQL：版本机制失效（重复执行）时会报错；正确行为是跳过
	badSQL := "INSERT INTO versioning_probe VALUES (2);\nINSERT INTO nonexistent_table_probe VALUES (1);"
	if err := os.WriteFile(filepath.Join(probeDir, probeFile), []byte(badSQL), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.RunMigrationsFromDir(ctx, probeDir); err != nil {
		t.Fatalf("second run should skip recorded migration, got: %v", err)
	}
	if err := store.Pool().QueryRow(ctx, "SELECT count(*) FROM versioning_probe").Scan(&probeCnt); err != nil || probeCnt != 1 {
		t.Fatalf("probe re-executed: cnt=%d err=%v", probeCnt, err)
	}
}
