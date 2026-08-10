package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// RunMigrationsFromDir reads and executes all .up.sql files from the given directory in order.
func (s *Store) RunMigrationsFromDir(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}
	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" && filepath.Base(e.Name())[:8] > "00000000" {
			// Match .up.sql files
			if len(e.Name()) > 7 && e.Name()[len(e.Name())-7:] == ".up.sql" {
				upFiles = append(upFiles, e.Name())
			}
		}
	}
	sort.Strings(upFiles)
	for _, f := range upFiles {
		path := filepath.Join(dir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		if _, err := s.pool.Exec(ctx, string(data)); err != nil {
			return fmt.Errorf("run migration %s: %w", f, err)
		}
	}
	return nil
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
