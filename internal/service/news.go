package service

import (
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type NewsService struct {
	repo repository.ArticleRepository
}

func NewNewsService(repo repository.ArticleRepository) *NewsService {
	return &NewsService{repo: repo}
}

func (s *NewsService) Create(title, content, category, source string) (domain.Article, error) {
	now := time.Now()
	a := domain.Article{ID: fmt.Sprintf("article-%d", now.UnixNano()), Title: title, Content: content,
		Summary: truncate(content, 100), Category: category, Source: source, Status: "draft",
		Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repo.Create(a)
}

// Update 编辑资讯内容（保留 ID/状态/创建时间，摘要重新截取）。
func (s *NewsService) Update(id, title, content, category, source string) (domain.Article, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Article{}, err
	}
	a.Title = title
	a.Content = content
	a.Summary = truncate(content, 100)
	a.Category = category
	a.Source = source
	a.UpdatedAt = time.Now()
	return s.repo.Update(a)
}

func (s *NewsService) Publish(id string) (domain.Article, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Article{}, err
	}
	a.Status = "published"
	a.UpdatedAt = time.Now()
	return s.repo.Update(a)
}

func (s *NewsService) ListByCategory(category string, page, pageSize int) ([]domain.Article, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByCategory(category, offset, pageSize)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
