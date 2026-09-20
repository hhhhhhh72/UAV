package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"drone-platform/internal/repository"
)

// TestTranslateSlotConflict 库级排他约束冲突（23P01）必须被翻译成 repository.ErrSlotTaken，
// 否则用户看到的是笼统 500（pgx 原文还会被 fail() 的脱敏换掉，等于什么都没说）。
// 反向也要守住：23505（唯一约束）与普通错误**不能**被误当成"时段被占"。
func TestTranslateSlotConflict(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"23P01 排他约束冲突", &pgconn.PgError{Code: "23P01", ConstraintName: "excl_testsite_occupied_slot"}, true},
		{"包了一层的 23P01", fmt.Errorf("insert booking: %w", &pgconn.PgError{Code: "23P01"}), true},
		{"23505 唯一约束（不该被当成占位冲突）", &pgconn.PgError{Code: "23505"}, false},
		{"普通错误", errors.New("connection reset"), false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		got := errors.Is(translateSlotConflict(c.err), repository.ErrSlotTaken)
		if got != c.want {
			t.Errorf("%s: errors.Is(..., ErrSlotTaken)=%v, want %v", c.name, got, c.want)
		}
	}
}
