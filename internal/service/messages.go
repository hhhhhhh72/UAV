package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type MessageService struct {
	repo repository.MessageRepository
}

func NewMessageService(repo repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) Send(ctx context.Context, senderID, receiverID, title, content, resType, resID string) (domain.Message, error) {
	m := domain.Message{
		ID: fmt.Sprintf("msg-%d", time.Now().UnixNano()), SenderID: senderID,
		ReceiverID: receiverID, Title: title, Content: content,
		ResourceType: resType, ResourceID: resID, IsRead: false, CreatedAt: time.Now(),
	}
	return s.repo.Create(ctx, m)
}

func (s *MessageService) ListForUser(ctx context.Context, userID string, unreadOnly bool) ([]domain.Message, error) {
	return s.repo.ListByUser(ctx, userID, unreadOnly)
}

func (s *MessageService) Get(ctx context.Context, msgID string) (domain.Message, error) {
	return s.repo.FindByID(ctx, msgID)
}

// MarkRead 归属校验（C10 修复）：仅收件人本人可将消息标为已读，
// 防止任意用户通过消息 ID 修改他人消息状态（IDOR）。
func (s *MessageService) MarkRead(ctx context.Context, userID, msgID string) (domain.Message, error) {
	m, err := s.repo.FindByID(ctx, msgID)
	if err != nil {
		return domain.Message{}, err
	}
	if m.ReceiverID != userID {
		return domain.Message{}, errors.New("message not found")
	}
	return s.repo.MarkRead(ctx, msgID)
}

func (s *MessageService) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *MessageService) ListAll(ctx context.Context, offset, limit int) ([]domain.Message, int, error) {
	return s.repo.ListAll(ctx, offset, limit)
}

func (s *MessageService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
