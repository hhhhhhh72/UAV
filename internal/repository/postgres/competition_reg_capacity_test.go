package postgres_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"drone-platform/internal/domain"
)

// TestPG_CompetitionRegCapacityAtomic 赛事名额由**数据库**兜底：CreateReg 在同一事务里
// INSERT 报名 + 条件更新 reg_count（仅当 reg_count < max_teams 才 +1，0 行受影响即回滚）。
//
// 这里直接调仓储层、绕过 service 的赛事维度键锁，验证的是库那一层 —— 进程内键锁
// 多实例失效，只有这条 UPDATE 跨进程有效。
//
// 顺带守住 reg_count 的语义（biz_repos3.go:141 声称"reg_count 与 registrations 行数
// 保持一致"）：生产上曾因种子数据写死假数字而漂移（comp-4 max_teams=300 / reg_count=340
// → 该赛事**永远报名失败**），所以这里一并断言两者相等。
func TestPG_CompetitionRegCapacityAtomic(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	repo := store.NewCompetitionRepository(nil) // nil cipher：与 all_repos_test.go:217 同款，PII 加解密对 nil 有守卫

	const (
		maxTeams   = 3
		concurrent = 50
	)
	comp, err := repo.Create(ctx, domain.Competition{
		ID: ug("comp"), Title: "名额并发", MaxTeams: maxTeams, Status: "open",
	})
	if err != nil {
		t.Fatalf("create competition: %v", err)
	}

	var wg sync.WaitGroup
	var ok int32
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, cerr := repo.CreateReg(ctx, domain.CompetitionReg{
				ID: ug("creg"), CompetitionID: comp.ID,
				UserID: fmt.Sprintf("u-%d", i), TeamName: fmt.Sprintf("队%d", i),
				Status: "registered",
			})
			if cerr == nil {
				atomic.AddInt32(&ok, 1)
			}
		}(i)
	}
	wg.Wait()

	regs, err := repo.ListRegs(ctx, comp.ID)
	if err != nil {
		t.Fatalf("list regs: %v", err)
	}
	after, err := repo.FindByID(ctx, comp.ID)
	if err != nil {
		t.Fatalf("find competition: %v", err)
	}
	if int(ok) != maxTeams || len(regs) != maxTeams {
		t.Fatalf("超卖：名额 %d，CreateReg 成功 %d 次，实际报名行 %d", maxTeams, ok, len(regs))
	}
	if after.RegCount != len(regs) {
		t.Fatalf("reg_count 与实际报名行数不一致：reg_count=%d，实际=%d（容量门禁读的就是这一列）",
			after.RegCount, len(regs))
	}
}
