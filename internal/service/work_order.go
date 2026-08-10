package service

import (
	"errors"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// WorkOrderService implements the 接单派单闭环 (PRD FR-6.2 ~ FR-6.5):
// 企业确认/拒绝意向 → 生成订单 → 飞手开始/完成 → 企业验收/整改/取消。
type WorkOrderService struct {
	orders  repository.WorkOrderRepository
	demands repository.DemandRepository
	intents repository.IntentRepository
}

func NewWorkOrderService(o repository.WorkOrderRepository, d repository.DemandRepository, i repository.IntentRepository) *WorkOrderService {
	return &WorkOrderService{orders: o, demands: d, intents: i}
}

// AcceptIntent 企业确认接单：意向 → contacted，其他意向 → closed，生成订单（pending）。
// amountFen 为订单金额（企业确认时填写，面议为 0）。
func (s *WorkOrderService) AcceptIntent(a domain.Actor, demandID, intentID string, amountFen int64) (domain.WorkOrder, error) {
	d, err := s.demands.FindByID(demandID)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("demand %s: %w", demandID, err)
	}
	if d.PublisherID != a.ID {
		return domain.WorkOrder{}, errors.New("只有需求发布者可以确认接单")
	}
	if d.Status != domain.DemandPublished {
		return domain.WorkOrder{}, errors.New("只有已发布的需求可以确认接单")
	}
	intents, err := s.intents.ListByDemand(demandID)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	var it *domain.DemandIntent
	for i := range intents {
		if intents[i].ID == intentID {
			it = &intents[i]
			break
		}
	}
	if it == nil {
		return domain.WorkOrder{}, errors.New("意向不存在")
	}
	if it.Status != "pending" {
		return domain.WorkOrder{}, errors.New("该意向已处理")
	}

	// 确认接单：本意向 → contacted，其余意向 → closed
	if _, err := s.intents.UpdateStatus(it.ID, "contacted"); err != nil {
		return domain.WorkOrder{}, err
	}
	others, err := s.intents.ListByDemand(demandID)
	if err == nil {
		for _, o := range others {
			if o.ID != it.ID && o.Status == "pending" {
				_, _ = s.intents.UpdateStatus(o.ID, "closed")
			}
		}
	}

	now := time.Now()
	wo := domain.WorkOrder{
		ID:            fmt.Sprintf("wo-%d-%d", now.UnixNano(), nextSeq()),
		OrderNo:       fmt.Sprintf("WO%d%06d", now.Year()%100, now.Unix()%1000000),
		DemandID:      demandID,
		PublisherID:   d.PublisherID,
		PublisherName: d.PublisherName,
		WorkerID:      it.IntentorID,
		WorkerName:    it.IntentorName,
		AmountFen:     amountFen,
		Status:        domain.WorkOrderPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return s.orders.Create(wo)
}

// RejectIntent 企业拒绝接单：意向 → closed。
func (s *WorkOrderService) RejectIntent(a domain.Actor, demandID, intentID string) error {
	d, err := s.demands.FindByID(demandID)
	if err != nil {
		return fmt.Errorf("demand %s: %w", demandID, err)
	}
	if d.PublisherID != a.ID {
		return errors.New("只有需求发布者可以处理意向")
	}
	intents, err := s.intents.ListByDemand(demandID)
	if err != nil {
		return err
	}
	for i := range intents {
		if intents[i].ID == intentID {
			if intents[i].Status != "pending" {
				return errors.New("该意向已处理")
			}
			_, err := s.intents.UpdateStatus(intents[i].ID, "closed")
			return err
		}
	}
	return errors.New("意向不存在")
}

// StartWork 飞手确认开始作业：pending → ongoing。
func (s *WorkOrderService) StartWork(a domain.Actor, orderID string) (domain.WorkOrder, error) {
	wo, err := s.orders.FindByID(orderID)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	if wo.WorkerID != a.ID {
		return domain.WorkOrder{}, errors.New("只有接单飞手可以开始作业")
	}
	if wo.Status != domain.WorkOrderPending {
		return domain.WorkOrder{}, errors.New("订单状态不允许开始作业")
	}
	return s.orders.UpdateStatus(orderID, domain.WorkOrderOngoing)
}

// CompleteWork 飞手确认作业完成：ongoing → awaiting_accept，可上传成果照片。
func (s *WorkOrderService) CompleteWork(a domain.Actor, orderID string, photos []string) (domain.WorkOrder, error) {
	wo, err := s.orders.FindByID(orderID)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	if wo.WorkerID != a.ID {
		return domain.WorkOrder{}, errors.New("只有接单飞手可以确认完成")
	}
	if wo.Status != domain.WorkOrderOngoing {
		return domain.WorkOrder{}, errors.New("订单状态不允许确认完成")
	}
	if len(photos) > 0 {
		if _, err := s.orders.UpdatePhotos(orderID, photos); err != nil {
			return domain.WorkOrder{}, err
		}
	}
	return s.orders.UpdateStatus(orderID, domain.WorkOrderAwaitingAccept)
}

// AcceptCompletion 企业验收通过：awaiting_accept → completed。
func (s *WorkOrderService) AcceptCompletion(a domain.Actor, orderID string) (domain.WorkOrder, error) {
	wo, err := s.orders.FindByID(orderID)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	if wo.PublisherID != a.ID {
		return domain.WorkOrder{}, errors.New("只有需求方可以验收")
	}
	if wo.Status != domain.WorkOrderAwaitingAccept {
		return domain.WorkOrder{}, errors.New("订单状态不允许验收")
	}
	return s.orders.UpdateStatus(orderID, domain.WorkOrderCompleted)
}

// RequestRework 企业提出整改：awaiting_accept → ongoing，记录整改要求。
func (s *WorkOrderService) RequestRework(a domain.Actor, orderID, note string) (domain.WorkOrder, error) {
	wo, err := s.orders.FindByID(orderID)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	if wo.PublisherID != a.ID {
		return domain.WorkOrder{}, errors.New("只有需求方可以提出整改")
	}
	if wo.Status != domain.WorkOrderAwaitingAccept {
		return domain.WorkOrder{}, errors.New("订单状态不允许整改")
	}
	if _, err := s.orders.UpdateRework(orderID, note); err != nil {
		return domain.WorkOrder{}, err
	}
	return s.orders.UpdateStatus(orderID, domain.WorkOrderOngoing)
}

// RequestCancel 任意一方取消订单（填写原因）：→ cancelled。
func (s *WorkOrderService) RequestCancel(a domain.Actor, orderID, reason string) (domain.WorkOrder, error) {
	wo, err := s.orders.FindByID(orderID)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	if wo.PublisherID != a.ID && wo.WorkerID != a.ID {
		return domain.WorkOrder{}, errors.New("只有订单双方可以取消")
	}
	if wo.Status == domain.WorkOrderCompleted || wo.Status == domain.WorkOrderCancelled {
		return domain.WorkOrder{}, errors.New("订单已结束，不能取消")
	}
	if _, err := s.orders.UpdateCancel(orderID, reason); err != nil {
		return domain.WorkOrder{}, err
	}
	return s.orders.UpdateStatus(orderID, domain.WorkOrderCancelled)
}

// FindByID 订单详情：仅订单双方可查看。
func (s *WorkOrderService) FindByID(a domain.Actor, orderID string) (domain.WorkOrder, error) {
	wo, err := s.orders.FindByID(orderID)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	if wo.PublisherID != a.ID && wo.WorkerID != a.ID {
		return domain.WorkOrder{}, errors.New("只有订单双方可以查看")
	}
	return wo, nil
}

// ListMine 我的订单：企业=我发出的，飞手=我接到的，双方视角合并。
func (s *WorkOrderService) ListMine(a domain.Actor) ([]domain.WorkOrder, error) {
	pub, err := s.orders.ListByPublisher(a.ID)
	if err != nil {
		return nil, err
	}
	wk, err := s.orders.ListByWorker(a.ID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(pub)+len(wk))
	out := make([]domain.WorkOrder, 0, len(pub)+len(wk))
	for _, wo := range append(pub, wk...) {
		if !seen[wo.ID] {
			seen[wo.ID] = true
			out = append(out, wo)
		}
	}
	return out, nil
}
