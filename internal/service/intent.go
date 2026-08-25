package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// nextSeq 提供同纳秒内创建的 ID 唯一性（Windows 时钟精度约 100ns，
// 连续创建两个对象可能拿到相同 UnixNano，导致内存/PG 按 ID 更新时错配）。
var idSeq atomic.Uint64

func nextSeq() uint64 {
	return idSeq.Add(1)
}

// nextID 生成进程内唯一 ID（前缀-UnixNano-原子序号）：
// 纯 UnixNano 在快速连续创建（如 HTTP 顺序请求间隔小于时钟粒度）时
// 可能相同，导致内存 repo 记录互相覆盖、PG 主键冲突 500。
// 所有新建资源的 ID 生成应统一走本函数。
func nextID(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), nextSeq())
}

// lockByKey 进程内互斥锁池：check-then-insert 类操作的并发竞态防护
// （如报名/投递的"先查重再创建"——无锁时双请求同时通过查重导致重复记录，
// 内存 repo 无唯一约束；PG 唯一索引下后到者报 500）。
// key 量级 = 用户×资源（报名/投递量小）；条目带引用计数，refs 归零即从池中删除，
// 避免每个不重复的 (用户,资源) 组合永久新增一个 *sync.Mutex 造成内存单调增长。
var onceLocks sync.Map // key -> *refLockEntry

// refLockEntry 是锁池条目：lockMu 是实际互斥锁；refs 为引用计数（受 refMu 保护），
// 语义为"当前持有者 + 已递增但仍在等待 lockMu 的获取者"数量。
// 计数保证：递增发生在 lockMu.Lock 之前、递减发生在 lockMu.Unlock 之后，
// 因此 refs==0 ⟹ 不存在任何持有者/等待者，此时才可从池中删除条目。
type refLockEntry struct {
	lockMu sync.Mutex // 实际业务互斥锁
	refMu  sync.Mutex // 保护 refs 与池删除
	refs   int
}

func lockByKey(key string) func() {
	for {
		created := false
		v, loaded := onceLocks.LoadOrStore(key, &refLockEntry{refs: 1})
		if !loaded {
			created = true // 本次创建者自带一个引用，下面不再递增
		}
		e := v.(*refLockEntry)
		e.refMu.Lock()
		if e.refs == 0 {
			// 条目刚被并发删除（refs 已归零），本 goroutine 拿到的指针已失效：
			// 释放后重新 Load，下一次要么取到新条目，要么创建新条目。
			e.refMu.Unlock()
			continue
		}
		if !created {
			e.refs++
		}
		e.refMu.Unlock()

		e.lockMu.Lock()
		return func() {
			e.lockMu.Unlock()
			e.refMu.Lock()
			e.refs--
			if e.refs == 0 {
				onceLocks.Delete(key)
			}
			e.refMu.Unlock()
		}
	}
}

// IntentService records contact intents on published demands (联系对接模式).
//
// 简版范围（V1）：登记意向 + 发布方查看意向列表 + 意向方查看自己的意向记录。
// 状态流转（contacted / done / closed）与管理端成交标记留待 V2。
type IntentService struct {
	repo    repository.IntentRepository
	demands repository.DemandRepository
}

func NewIntentService(r repository.IntentRepository, d repository.DemandRepository) *IntentService {
	return &IntentService{repo: r, demands: d}
}

type CreateIntentInput struct {
	IntentorName string `json:"intentor_name"`
	Contact      string `json:"contact"`
	Remark       string `json:"remark"`
}

// Create registers an intent to contact the publisher of a published demand.
func (s *IntentService) Create(ctx context.Context, a domain.Actor, demandID string, in CreateIntentInput) (domain.DemandIntent, error) {
	if demandID == "" {
		return domain.DemandIntent{}, errors.New("demand_id is required")
	}
	if in.Contact == "" {
		return domain.DemandIntent{}, errors.New("contact is required")
	}
	d, err := s.demands.FindByID(ctx, demandID)
	if err != nil {
		return domain.DemandIntent{}, fmt.Errorf("demand %s: %w", demandID, err)
	}
	// assigned 表示已接受意向并生成工单：禁止再登记新意向（此前无此状态，
	// 已派单需求可无限登记）。
	if d.Status != domain.DemandPublished {
		if d.Status == domain.DemandAssigned {
			return domain.DemandIntent{}, errors.New("该需求已确认接单，暂不开放新意向")
		}
		return domain.DemandIntent{}, errors.New("只有已发布的需求可以登记对接意向")
	}
	if d.PublisherID == a.ID {
		return domain.DemandIntent{}, errors.New("不能登记自己发布的需求")
	}
	// P1 修复：同一用户对同一需求只允许一条"待处理"意向（防重复提交）。
	// 已被确认/关闭（contacted/closed 等）的旧意向不阻塞再次登记；
	// 数据库层部分唯一索引 (demand_id, intentor_id) WHERE status='pending' 兜底并发。
	// 预检查询出错必须上抛：静默跳过会把 DB 故障伪装成"无重复"导致重复意向。
	if existing, err := s.repo.ListByDemand(ctx, demandID); err != nil {
		return domain.DemandIntent{}, fmt.Errorf("check existing intents: %w", err)
	} else {
		for _, e := range existing {
			if e.IntentorID == a.ID && e.Status == "pending" {
				return domain.DemandIntent{}, errors.New("已登记过该需求的对接意向，请勿重复提交")
			}
		}
	}
	name := in.IntentorName
	if name == "" {
		name = a.ID
	}
	now := time.Now()
	it := domain.DemandIntent{
		ID:           fmt.Sprintf("intent-%d-%d", now.UnixNano(), nextSeq()),
		DemandID:     demandID,
		IntentorID:   a.ID,
		IntentorName: name,
		Contact:      in.Contact,
		Remark:       in.Remark,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.Create(ctx, it)
}

// ListByDemand returns intents for a demand. Only the publisher or admins.
func (s *IntentService) ListByDemand(ctx context.Context, a domain.Actor, demandID string) ([]domain.DemandIntent, error) {
	d, err := s.demands.FindByID(ctx, demandID)
	if err != nil {
		return nil, fmt.Errorf("demand %s: %w", demandID, err)
	}
	if d.PublisherID != a.ID && a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("只有需求发布者或管理员可以查看对接意向")
	}
	return s.repo.ListByDemand(ctx, demandID)
}

// ListMine returns intents registered by the current user.
func (s *IntentService) ListMine(ctx context.Context, a domain.Actor) ([]domain.DemandIntent, error) {
	return s.repo.ListByIntentor(ctx, a.ID)
}

// Cancel 意向方取消自己的登记（撤回意向）：
// 仅待处理（pending）可取消；已确认洽谈/已关闭的意向不可撤回。取消后状态 closed。
func (s *IntentService) Cancel(ctx context.Context, a domain.Actor, intentID string) error {
	mine, err := s.repo.ListByIntentor(ctx, a.ID)
	if err != nil {
		return err
	}
	for i := range mine {
		if mine[i].ID == intentID {
			if mine[i].Status != "pending" {
				return errors.New("该意向已处理，无法取消")
			}
			_, err := s.repo.UpdateStatus(ctx, intentID, "closed")
			return err
		}
	}
	return errors.New("意向不存在")
}
