package service

import (
	"errors"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// IntentService records contact intents on published demands (联系对接模式).
//
// 简版范围（V1）：登记意向 + 发布方查看意向列表 + 意向方查看自己的意向记录。
// 状态流转（contacted / done / closed）与管理端成交标记留待 V2。
type IntentService struct {
	repo  repository.IntentRepository
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
func (s *IntentService) Create(a domain.Actor, demandID string, in CreateIntentInput) (domain.DemandIntent, error) {
	if demandID == "" {
		return domain.DemandIntent{}, errors.New("demand_id is required")
	}
	if in.Contact == "" {
		return domain.DemandIntent{}, errors.New("contact is required")
	}
	d, err := s.demands.FindByID(demandID)
	if err != nil {
		return domain.DemandIntent{}, fmt.Errorf("demand %s: %w", demandID, err)
	}
	if d.Status != domain.DemandPublished {
		return domain.DemandIntent{}, errors.New("只有已发布的需求可以登记对接意向")
	}
	if d.PublisherID == a.ID {
		return domain.DemandIntent{}, errors.New("不能登记自己发布的需求")
	}
	name := in.IntentorName
	if name == "" {
		name = a.ID
	}
	now := time.Now()
	it := domain.DemandIntent{
		ID:           fmt.Sprintf("intent-%d", now.UnixNano()),
		DemandID:     demandID,
		IntentorID:   a.ID,
		IntentorName: name,
		Contact:      in.Contact,
		Remark:       in.Remark,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.Create(it)
}

// ListByDemand returns intents for a demand. Only the publisher or admins.
func (s *IntentService) ListByDemand(a domain.Actor, demandID string) ([]domain.DemandIntent, error) {
	d, err := s.demands.FindByID(demandID)
	if err != nil {
		return nil, fmt.Errorf("demand %s: %w", demandID, err)
	}
	if d.PublisherID != a.ID && a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("只有需求发布者或管理员可以查看对接意向")
	}
	return s.repo.ListByDemand(demandID)
}

// ListMine returns intents registered by the current user.
func (s *IntentService) ListMine(a domain.Actor) ([]domain.DemandIntent, error) {
	return s.repo.ListByIntentor(a.ID)
}
