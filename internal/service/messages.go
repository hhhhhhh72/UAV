package service

import (
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

func (s *MessageService) Send(senderID, receiverID, title, content, resType, resID string) (domain.Message, error) {
	m := domain.Message{
		ID: fmt.Sprintf("msg-%d", time.Now().UnixNano()), SenderID: senderID,
		ReceiverID: receiverID, Title: title, Content: content,
		ResourceType: resType, ResourceID: resID, IsRead: false, CreatedAt: time.Now(),
	}
	return s.repo.Create(m)
}

func (s *MessageService) ListForUser(userID string, unreadOnly bool) ([]domain.Message, error) {
	return s.repo.ListByUser(userID, unreadOnly)
}

func (s *MessageService) Get(msgID string) (domain.Message, error) {
	return s.repo.FindByID(msgID)
}

// MarkRead 归属校验（C10 修复）：仅收件人本人可将消息标为已读，
// 防止任意用户通过消息 ID 修改他人消息状态（IDOR）。
func (s *MessageService) MarkRead(userID, msgID string) (domain.Message, error) {
	m, err := s.repo.FindByID(msgID)
	if err != nil {
		return domain.Message{}, err
	}
	if m.ReceiverID != userID {
		return domain.Message{}, errors.New("message not found")
	}
	return s.repo.MarkRead(msgID)
}

func (s *MessageService) UnreadCount(userID string) (int, error) {
	return s.repo.UnreadCount(userID)
}

func (s *MessageService) ListAll(offset, limit int) ([]domain.Message, int, error) {
	return s.repo.ListAll(offset, limit)
}

func (s *MessageService) Delete(id string) error {
	return s.repo.Delete(id)
}
