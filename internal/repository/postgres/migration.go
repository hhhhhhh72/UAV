package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationLockKey 迁移串行化用的固定会话级咨询锁键：
// 多实例并发启动（docker compose --scale）时只允许一个实例执行迁移。
const migrationLockKey = int64(0x64726F6E65) // "drone"

// RunMigrationsFromDir reads .up.sql files from dir in order and executes only
// the ones not yet recorded in schema_migrations.
// 旧库兼容：此前版本每次启动全量重放且无版本表。首次带版本表启动时
// schema_migrations 为空，会全量执行一遍（全部迁移均为幂等写法，与旧行为
// 等价）并登记版本，此后仅执行新增迁移。
func (s *Store) RunMigrationsFromDir(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}
	var upFiles []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".up.sql") || versionOf(name) == "" {
			continue
		}
		upFiles = append(upFiles, name)
	}
	sort.Strings(upFiles)

	// 咨询锁挂在独立连接上（会话级锁），避免锁与池内其他请求的连接纠缠；
	// 迁移期间所有操作都走这条连接，整个流程结束后再归还池。
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", migrationLockKey)
	}()

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := loadMigrationVersions(ctx, conn)
	if err != nil {
		return fmt.Errorf("load schema_migrations: %w", err)
	}

	for _, f := range upFiles {
		if applied[versionOf(f)] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		if err := applyMigration(ctx, conn, versionOf(f), string(data)); err != nil {
			return fmt.Errorf("run migration %s: %w", f, err)
		}
	}
	return nil
}

// versionOf extracts the leading numeric version prefix of a migration file name
// ("000031_enrollment_birthday_date.up.sql" → "000031"); empty if no digit prefix.
func versionOf(fileName string) string {
	i := 0
	for i < len(fileName) && fileName[i] >= '0' && fileName[i] <= '9' {
		i++
	}
	return fileName[:i]
}

func loadMigrationVersions(ctx context.Context, conn *pgxpool.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// applyMigration runs one migration file and records its version in the same
// transaction: 任一步失败整体回滚，不会出现"表已建但版本未记"的半成品状态。
func applyMigration(ctx context.Context, conn *pgxpool.Conn, version, sql string) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // commit 后回滚为 no-op
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT (version) DO NOTHING", version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func MigrationsDir() string {
	// Docker 部署用 MIGRATIONS_DIR 显式指定（-trimpath 编译下 runtime.Caller
	// 返回模块路径 drone-platform/...，运行时推导不可靠）
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	projectRoot := filepath.Join(dir, "..", "..", "..")
	abs, _ := filepath.Abs(projectRoot)
	return filepath.Join(abs, "migrations")
}
