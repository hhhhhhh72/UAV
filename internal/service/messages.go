package service

import (
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

func (s *MessageService) MarkRead(msgID string) (domain.Message, error) {
	return s.repo.MarkRead(msgID)
}

func (s *MessageService) UnreadCount(userID string) (int, error) {
	return s.repo.UnreadCount(userID)
}
