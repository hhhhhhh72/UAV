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
// key 量级 = 用户×资源（报名/投递量小），无需清理。
var onceLocks sync.Map // key -> *sync.Mutex

func lockByKey(key string) func() {
	l, _ := onceLocks.LoadOrStore(key, &sync.Mutex{})
	l.(*sync.Mutex).Lock()
	return l.(*sync.Mutex).Unlock
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
	if d.Status != domain.DemandPublished {
		return domain.DemandIntent{}, errors.New("只有已发布的需求可以登记对接意向")
	}
	if d.PublisherID == a.ID {
		return domain.DemandIntent{}, errors.New("不能登记自己发布的需求")
	}
	// P1 修复：同一用户对同一需求只允许一条"待处理"意向（防重复提交）。
	// 已被确认/关闭（contacted/closed 等）的旧意向不阻塞再次登记；
	// 数据库层部分唯一索引 (demand_id, intentor_id) WHERE status='pending' 兜底并发。
	if existing, err := s.repo.ListByDemand(ctx, demandID); err == nil {
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
